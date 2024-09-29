package structures

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type MBR struct {
	Mbr_size           int32      
	Mbr_creation_date  float32     
	Mbr_disk_signature int32       
	Mbr_disk_fit       [1]byte      
	Mbr_partitions     [4]PARTITION 
}


func CreateMBR(mkdisk *MKDISK, sizeBytes int) (string, error) {

	var fitByte byte

	switch mkdisk.Fit {
		case "FF":
			fitByte = 'F'
		case "BF":
			fitByte = 'B'
		case "WF":
			fitByte = 'W'
		default:
			fmt.Println("Invalid fit type")
			return "Invalid fit type en el MBR",nil
	}

	mbr := &MBR{
		Mbr_size:           int32(sizeBytes),
		Mbr_creation_date:  float32(time.Now().Unix()),
		Mbr_disk_signature: rand.Int31(),
		Mbr_disk_fit:       [1]byte{fitByte},
		Mbr_partitions: 	[4]PARTITION{
			{Part_status: [1]byte{'2'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: 0, Part_id: [4]byte{'0'}},
			{Part_status: [1]byte{'2'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: 0, Part_id: [4]byte{'0'}},
			{Part_status: [1]byte{'2'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: 0, Part_id: [4]byte{'0'}},
			{Part_status: [1]byte{'2'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: 0, Part_id: [4]byte{'0'}},
		},
	}

	msg, err := mbr.SerializeMBR(mkdisk.Path)
	if err != nil {
		fmt.Println("Error:", err)
		return msg,err
	}

	return "",nil
}


func (mbr *MBR) SerializeMBR(path string) (string, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return "Error al abrir el archivo al serializar MBR",err
	}
	defer file.Close()

	err = binary.Write(file, binary.LittleEndian, mbr)
	if err != nil {
		return "Error al escribir en el archivo al serializar MBR",err
	}

	err = file.Sync()
	if err != nil {
		return "Error al sincronizar los datos en el archivo", err
	}

	return "",nil
}

func (mbr *MBR) DeserializeMBR(path string) (string, error) { // Deserializar un MBR desde un archivo esto es para leer el archivo binario y obtener la información del MBR
	file, err := os.Open(path) // Abrir el archivo
	if err != nil {
		return "Error al deserializar el MBR",err
	}
	defer file.Close()

	mbrSize := binary.Size(mbr) // Tamaño de la estructura MBR
	if mbrSize <= 0 {
		return "Tamaño inválido para el MBR",fmt.Errorf("invalid MBR size: %d", mbrSize)
	}

	buffer := make([]byte, mbrSize) // Crear un buffer para leer el archivo
	_, err = file.Read(buffer) // Leer el archivo
	if err != nil {
		return "Error al leer el archivo al deserializar",err
	}

	reader := bytes.NewReader(buffer) // Crear un lector para el buffer
	err = binary.Read(reader, binary.LittleEndian, mbr) // Leer la estructura desde el buffer
	if err != nil {
		return "Error al crear un nuevo buffer al deserializar",err
	}

	return "",nil
}

func (mbr *MBR) Print() {

	creationTime := time.Unix(int64(mbr.Mbr_creation_date), 0)

	diskFit := rune(mbr.Mbr_disk_fit[0])

	fmt.Printf("MBR Size: %d\n", mbr.Mbr_size)
	fmt.Printf("Creation Date: %s\n", creationTime.Format(time.RFC3339))
	fmt.Printf("Disk Signature: %d\n", mbr.Mbr_disk_signature)
	fmt.Printf("Disk Fit: %c\n", diskFit)
}

func (mbr *MBR) PrintPartitions() {
	for i, partition := range mbr.Mbr_partitions {
		// Convertir Part_status, Part_type y Part_fit a char
		partStatus := rune(partition.Part_status[0])
		partType := rune(partition.Part_type[0])
		partFit := rune(partition.Part_fit[0])

		// Convertir Part_name a string
		partName := string(partition.Part_name[:])
		// Pasar part id a string
		partId := string(partition.Part_id[:])

		fmt.Printf("Partition %d:\n", i+1)
		fmt.Printf("  Status: %c\n", partStatus)
		fmt.Printf("  Type: %c\n", partType)
		fmt.Printf("  Fit: %c\n", partFit)
		fmt.Printf("  Start: %d\n", partition.Part_start)
		fmt.Printf("  Size: %d\n", partition.Part_size)
		fmt.Printf("  Name: %s\n", partName)
		fmt.Printf("  Correlative: %d\n", partition.Part_correlative)
		fmt.Printf("  ID: %s\n", partId)

	}
}

//Para obtener la primera particion disponible
func (mbr *MBR) GetFirstPartitionAvailable()(*PARTITION, int, int, string) {
	
	//cálculo de offset para el inicio (start) de la partición
	offset := binary.Size(mbr)	//tamaño del MBR

	//recorrer las particiones
	for i := 0; i<len(mbr.Mbr_partitions); i++ {
		// si el star de la particion es -1, entonces la particion esta vacia, se puede usar
		if mbr.Mbr_partitions[i].Part_start == -1 {
			// se retorn la particion, el offset y el indice
			return &mbr.Mbr_partitions[i], offset, i, ""  //el & es para retornar la dirección de memoria de la particion
			//EL OFFSET ES EL INICIO DE LA PARTICION
		}else{
			// calcula el nuevo offset para la siguiente particion, suma el tamaño de la particion
			offset += int(mbr.Mbr_partitions[i].Part_size)
		}
	}	
	return nil, -1, -1, ""
}

//Para obtener la particion por nombre
func (mbr *MBR) GetPartitionByName(name string, path string) (*PARTITION, int, string) {

	//recorrer las particiones
	for i, particion := range mbr.Mbr_partitions {
		// convertir el part_name a string y quitar los caracteres nulos
		particionName := strings.Trim(string(particion.Part_name[:]), "\x00")
		// pasar el nombre de la particion a string y quitar los caracteres nulos
		inputName := strings.Trim(name, "\x00")

		// si el nombre de la particion es igual al nombre de la particion que se busca
		if (strings.EqualFold(particionName, inputName)) {
			return &particion, i, "" //retornar la particion y el indice
		}

	}
	return nil, -1, "No se encontró la partición"
}

// Función para obtener una partición por ID
func (mbr *MBR) GetPartitionByID(id string) (*PARTITION, error) {
	for i := 0; i < len(mbr.Mbr_partitions); i++ {
		// Convertir Part_name a string y eliminar los caracteres nulos
		partitionID := strings.Trim(string(mbr.Mbr_partitions[i].Part_id[:]), "\x00 ")
		// Convertir el id a string y eliminar los caracteres nulos
		inputID := strings.Trim(id, "\x00 ")
		// Si el nombre de la partición coincide, devolver la partición
		if strings.EqualFold(partitionID, inputID) {
			return &mbr.Mbr_partitions[i], nil
		}
	}
	return nil, errors.New("partición no encontrada")
}

func (mbr *MBR) UpdatePartitionCorrelatives() {
    correlativo := 1

    // Iterar sobre todas las particiones en el MBR
    for i := 0; i < len(mbr.Mbr_partitions); i++ {
        part := &mbr.Mbr_partitions[i]

        // Si la partición está activa y es primaria, actualizar el correlativo
        if part.Part_status[0] != 0 && part.Part_type[0] == 'P' {
			part.Part_correlative = int32(correlativo)
            correlativo++ // Incrementar solo para particiones primarias
        } else if part.Part_type[0] == 'E' {
            // Si es una partición extendida, el correlativo es 0
            part.Part_correlative = 0
        }
    }
}