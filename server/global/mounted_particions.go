package global

import (
	"errors"
	"fmt"
	structures "server/structures"
)


const Carnet string = "09" //202110509

var (
	MountedPartitions map[string]string = make(map[string]string)
)


var(
	ParticionesMontadas = make(map[string]string)
)

// GetMountedPartition obtiene la partición montada con el id especificado
func GetMountedPartition(id string) (*structures.PARTITION, string, error) {
	// Obtener el path de la partición montada
	path := MountedPartitions[id]
	if path == "" {
		return nil, "", errors.New("la partición no está montada")
	}

	// Crear una instancia de MBR
	var mbr structures.MBR

	// Deserializar la estructura MBR desde un archivo binario
	msg,err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil,msg, err
	}

	// Buscar la partición con el id especificado
	partition, err:= mbr.GetPartitionByID(id)
	if partition == nil {
		return nil, "", err
	}

	return partition, path, nil
}


//funcion para desmontar una particion
func UnmountPartition(id string) (string, error) {
	// Eliminar la partición montada de la lista de particiones montadas
	fmt.Println("Desmontando particion con id: " + id)
	fmt.Println(MountedPartitions)
	delete(MountedPartitions, id)
	fmt.Println("---------------------")
	fmt.Println(MountedPartitions)

	return "", nil
}

func UnmountPartition1(nombre string)string{
	for key:= range ParticionesMontadas {
		if key == nombre {
			delete(ParticionesMontadas, key)
			break
		}
	}
	return ""
}


func ObtenerParticion(nombre string)string{
	for key, value := range ParticionesMontadas {
		if key == nombre {
			return value
		}
	}
	return "P"
}


func GetMountedPartitionRep(id string) (*structures.MBR, *structures.SuperBlock, string, error) {
	// Obtener el path de la partición montada
	path := MountedPartitions[id]
	if path == "" {
		return nil, nil, "", errors.New("la partición no está montada")
	}

	// Crear una instancia de MBR
	var mbr structures.MBR

	// Deserializar la estructura MBR desde un archivo binario
	_, err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, nil, "", err
	}

	// Buscar la partición con el id especificado
	partition, err := mbr.GetPartitionByID(id)
	if partition == nil {
		return nil, nil, "", err
	}

	// Crear una instancia de SuperBlock
	var sb structures.SuperBlock

	// Deserializar la estructura SuperBlock desde un archivo binario
	err = sb.Deserialize(path, int64(partition.Part_start))
	if err != nil {
		return nil, nil, "", err
	}

	return &mbr, &sb, path, nil
}


// GetMountedPartitionSuperblock obtiene el SuperBlock de la partición montada con el id especificado
func GetMountedPartitionSuperblock(id string) (*structures.SuperBlock, *structures.PARTITION, string, error) {
	// Obtener el path de la partición montada
	path := MountedPartitions[id]
	if path == "" {
		return nil, nil, "", errors.New("la partición no está montada")
	}

	// Crear una instancia de MBR
	var mbr structures.MBR

	// Deserializar la estructura MBR desde un archivo binario
	_,err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, nil, "", err
	}

	// Buscar la partición con el id especificado
	partition, err := mbr.GetPartitionByID(id)
	if partition == nil {
		return nil, nil, "", err
	}

	// Crear una instancia de SuperBlock
	var sb structures.SuperBlock

	// Deserializar la estructura SuperBlock desde un archivo binario
	err = sb.Deserialize(path, int64(partition.Part_start))
	if err != nil {
		return nil, nil, "", err
	}

	return &sb, partition, path, nil
}