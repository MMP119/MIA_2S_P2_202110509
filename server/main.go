package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	analyzer "server/analyzer"
	"server/global"
	"server/globales"
	util "server/util"
	"strings"
)

type InputData struct {
	Code string `json:"code"`
}


func handleAnalyze(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    var inputs []string
    err := json.NewDecoder(r.Body).Decode(&inputs)
    if err != nil {
        http.Error(w, "Error al decodificar el JSON", http.StatusBadRequest)
        return
    }

    results, errors := analyzer.Analyzer(inputs)

    response := map[string]interface{}{
        "results": results,
        "errors":  errors,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func handleGetDisk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var inputs struct {
		DiskID string `json:"diskID"`
	}
	err := json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		http.Error(w, "Error al decodificar el JSON", http.StatusBadRequest)
		return
	}

	pathDisks := global.GetPathDisk(inputs.DiskID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pathDisks)
}


func handleGetPartition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var inputs struct {
		DiskName string `json:"diskName"`
	}
	err := json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		http.Error(w, "Error al decodificar el JSON", http.StatusBadRequest)
		return
	}

	particionesDelDisco := globales.GetPartitionOnDisk(inputs.DiskName)

	// Crear un array de particiones
	particiones := []map[string]string{}
	for partitionName, partitionInfo := range particionesDelDisco {
		particiones = append(particiones, map[string]string{
			"partitionName": partitionName,
			"partitionId":   getFirstKey(partitionInfo), // Supongo que el ID es la primera clave
			"path":          partitionInfo[getFirstKey(partitionInfo)],
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(particiones)
}

func getFirstKey(m map[string]string) string {
	for k := range m {
		return k
	}
	return ""
}


func handleGetArchivoCarpetas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var inputs struct {
		IdParticion string `json:"idParticion"`
		Path string `json:"path"`
	}
	err := json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		http.Error(w, "Error al decodificar el JSON", http.StatusBadRequest)
		return
	}

	var archivosCarpetas map[string]string
	archivosCarpetas, err = global.ObtenerArchivosCarpetasRaiz(inputs.IdParticion, inputs.Path)
	if err != nil {
		http.Error(w, "Error al obtener archivos y carpetas", http.StatusBadRequest)
		return
	}

	//verificar si tiene caracteres nulos 
	for key, value := range archivosCarpetas {
		if strings.Contains(key, "\x00") {
			delete(archivosCarpetas, key)
		}
		if strings.Contains(value, "\x00") {
			delete(archivosCarpetas, key)
		}
	}

	//crear un array de archivos y carpetas
	archivosCarpetasArray := []map[string]string{}
	for archivoCarpeta, tipo := range global.ArchivosCarpetas {
		archivosCarpetasArray = append(archivosCarpetasArray, map[string]string{
			"nombre": archivoCarpeta,
			"tipo": tipo,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(archivosCarpetasArray)

	// Limpiar la variable global
	global.ArchivosCarpetas = make(map[string]string)
}


func handleInicioSesion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var inputs struct {
		PartitionId string `json:"partitionId"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		http.Error(w, "Error al decodificar el JSON", http.StatusBadRequest)
		return
	}
	mensaje, err := global.VerificarSesion(inputs.PartitionId, inputs.Username, inputs.Password)
	if err != nil {
		http.Error(w, mensaje, http.StatusBadRequest)
		return
	}

	inicio := global.ComprobarCredenciales(inputs.PartitionId, inputs.Username, inputs.Password)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inicio)
}


func main() {
	// Crea un nuevo multiplexor de servidores
	mux := http.NewServeMux()
	mux.HandleFunc("/analyze", handleAnalyze)
	mux.HandleFunc("/disks", handleGetDisk)
	mux.HandleFunc("/partitions", handleGetPartition)
	mux.HandleFunc("/archivosCarpetas", handleGetArchivoCarpetas)
	mux.HandleFunc("/inicioSesion", handleInicioSesion)

	//cors 
	corsHandler := util.EnableCors(mux)

	// Inicia el servidor en el puerto 8080
	fmt.Println("Servidor escuchando en el puerto 8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}