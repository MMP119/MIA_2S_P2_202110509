package globales

import (
	"path/filepath" // Para extraer el nombre del archivo
)

//guardar nombreDisco,nombreParticion, idParticion, pathParticion
var(
	partitionsOnDisk  = make(map[string]map[string]map[string]string)
)

func extractDiskName(path string) string {
	return filepath.Base(path) // Extrae la parte final del path (nombre del disco)
}

// Retorna las particiones de un disco en específico
func GetPartitionOnDisk(nombreDisco string) map[string]map[string]string {
	return partitionsOnDisk[nombreDisco]
}

func SetPartitionOnDisk(partitionName string, idPartition string, pathPartition string) {
	diskName := extractDiskName(pathPartition)
	if partitionsOnDisk[diskName] == nil {
		partitionsOnDisk[diskName] = make(map[string]map[string]string)
	}
	if partitionsOnDisk[diskName][partitionName] == nil {
		partitionsOnDisk[diskName][partitionName] = make(map[string]string)
	}
	partitionsOnDisk[diskName][partitionName][idPartition] = pathPartition
}