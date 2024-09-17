package commands

import (
	"fmt"
	global "server/global"
)

func ParseList(tokens []string) (string, string, error) {

	partitions := []string{}
	for id:= range global.MountedPartitions {
		partitions = append(partitions, id)		
	}

	// Verificar si hay particiones montadas
	if len(partitions) == 0 {
		return "", "\n No hay particiones montadas\n", nil
	}

	// Crear mensaje de salida
	msg := "\n Particiones montadas:\n"
	for _, partition := range partitions {
		msg += fmt.Sprintf("- %s\n", partition)
	}

	return "", msg, nil

}