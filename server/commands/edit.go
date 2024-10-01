package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	global "server/global"
	//structures "server/structures"
	utils "server/util"
	"strings"
)



type EDIT struct{
	Path string
	Contenido string
}



func ParseEdit(tokens []string) (*EDIT, string, error) {

	cmd := &EDIT{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`(?i)-path="[^"]+"|(?i)-path=[^\s]+|(?i)-contenido="[^"]+"|(?i)-contenido=[^\s]+`)

	matches := re.FindAllString(args, -1)

	for _, math := range matches {
		kv := strings.SplitN(math, "=", 2)
		if len(kv) != 2 {
			return nil, "ERROR: formato de parámetro inválido", fmt.Errorf("formato de parámetro inválido: %s", math)
		}

		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key{

			case "-path":
				if value == "" {
					return nil, "ERROR: el path es obligatorio", errors.New("el path es obligatorio")
				}
				cmd.Path = value
			
			case "-contenido":
				if value == "" {
					return nil, "ERROR: el contenido es obligatorio", errors.New("el contenido es obligatorio")
				}
				cmd.Contenido = value

			default: 
				return nil, "ERROR: parámetro no reconocido", fmt.Errorf("parámetro no reconocido: %s", key)
		}
	}

	if cmd.Path == "" {
		return nil, "ERROR: el path es obligatorio", errors.New("el path es obligatorio")
	}	

	if cmd.Contenido == "" {
		return nil, "ERROR: el contenido es obligatorio", errors.New("el contenido es obligatorio")
	}

	msg, err := CommandEdit(cmd)
	if err != nil {
		return nil, msg, err
	}

	return cmd, "Comando EDIT: realizado correctamente", nil
}


func CommandEdit(cmd *EDIT) (string, error) {

	// leer un archivo de mi pc, obtener el contenido y guardarlo en una variable
	contenido, err := ioutil.ReadFile(cmd.Contenido)
	if err != nil {
		return "ERROR: no se pudo leer el archivo", err
	}

	size := len(contenido)

	//obtener el contenido por chunks de 64 bytes
	chunks := utils.SplitStringIntoChunks(string(contenido))

	//obtener los directorios padre y el destino
	parentDirs, destino := utils.GetParentDirectories(cmd.Path)

	idPartition := global.GetIDSession()

	partitionSuperblock, mountedPartition, partitionPath, err := global.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return "Error al obtener la partición montada en el comando login", fmt.Errorf("error al obtener la partición montada: %v", err)
	}

	err = partitionSuperblock.EditeFile(partitionPath, parentDirs, destino, size, chunks)
	if err != nil {
		return "Error al editar el archivo", err
	}


	//serializar el superbloque
	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return "Error al serializar el superbloque", err
	}


	return "Comando EDIT: realizado correctamente", nil
}