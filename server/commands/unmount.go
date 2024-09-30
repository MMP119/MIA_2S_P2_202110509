package commands

import (
	"errors"
	"fmt"
	"regexp"
	global "server/global"
	"server/structures"
	"strings"
)


type UNMOUNT struct {
	id string
}


func ParseUnmount(tokens []string)(*UNMOUNT, string, error){

	cmd := &UNMOUNT{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`(?i)-id=[^\s]+`)

	matches := re.FindAllString(args, -1)

	for _, math := range matches{
		
		kv := strings.SplitN(math, "=", 2)
		if len(kv) != 2 {
			return nil, "ERROR: formato de parámetro inválido", fmt.Errorf("formato de parámetro inválido: %s", math)
		}

		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key{

			case "-id":
				if value == "" {
					return nil, "ERROR: el id es obligatorio", errors.New("el id es obligatorio")
				}
				cmd.id = value

			default:
				return nil, "ERROR: parámetro no reconocido", fmt.Errorf("parámetro no reconocido: %s", key)
		}
	}

	if cmd.id == "" {
		return nil, "ERROR: el id es obligatorio", errors.New("el id es obligatorio")
	}

	msg, err := CommandUnmount(cmd)
	if err != nil {
		return nil, msg, err
	}

	return cmd, "Comando UNMOUNT: Realizado conrrectamente", nil
}


func CommandUnmount(cmd *UNMOUNT)(string, error){

	fmt.Println(cmd.id)

	// Verificar si la partición está montada
	_, path, err := global.GetMountedPartition(cmd.id)
	if err != nil {
		return "La partición no está montada", err
	}


	var mbr structures.MBR

	// Deserializar la estructura MBR desde un archivo binario
	msg,err := mbr.DeserializeMBR(path)
	if err != nil {
		return msg, err
	}

	// Desmontar la partición
	global.UnmountPartition(cmd.id)

	for i := range mbr.Mbr_partitions {
		particion := &mbr.Mbr_partitions[i]
		if strings.Trim(string(particion.Part_id[:]), "\x00") == cmd.id {
			particion.Part_status[0] = '0'
			particion.Part_correlative = int32(0)
			break
		}
	}


	// Serializar la estructura MBR en un archivo binario
	msg, err = mbr.SerializeMBR(path)
	if err != nil {
		return msg, err
	}

	return "Partición desmontada correctamente", nil
}