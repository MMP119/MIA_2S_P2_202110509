package main

import (
	analyzer "server/analyzer" 
	//"bufio"                     
	"fmt"                       
	//"os"           
	"encoding/json"
	"log"
	"net/http"
	util "server/util"
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



func main() {
	/*
	// Crea un nuevo escáner que lee desde la entrada estándar (teclado)
	scanner := bufio.NewScanner(os.Stdin)

	// Bucle infinito para leer comandos del usuario
	for {
		fmt.Print(">>> ") // Imprime el prompt para el usuario

		// Lee la siguiente línea de entrada del usuario
		if !scanner.Scan() {
			break // Si no hay más líneas para leer, rompe el bucle
		}

		// Obtiene el texto ingresado por el usuario
		input := scanner.Text()

		// Llama a la función Analyzer del paquete analyzer para analizar el comando ingresado
		_, err := analyzer.Analyzer(input)
		if err != nil {
			// Si hay un error al analizar el comando, imprime el error y continúa con el siguiente comando
			fmt.Println("Error:", err)
			continue
		}
	}

	// Verifica si hubo algún error al leer la entrada
	if err := scanner.Err(); err != nil {
		// Si hubo un error al leer la entrada, lo imprime
		fmt.Println("Error al leer:", err)
	}*/

	mux := http.NewServeMux()
	mux.HandleFunc("/analyze", handleAnalyze)

	//cors 
	corsHandler := util.EnableCors(mux)

	// Inicia el servidor en el puerto 8080
	fmt.Println("Servidor escuchando en el puerto 8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}