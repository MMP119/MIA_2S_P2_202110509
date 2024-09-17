package commands

import (
	"errors" 
	"fmt"    
	"regexp" 
	util "server/util" 
	"strings" 
)

// RMDISK estructura que representa el comando rmdisk con su parámetro
type RMDISK struct {
	path string // Ruta del archivo del disco
}


// CommandRmdisk parsea el comando rmdisk y devuelve una instancia de RMDISK
func ParserRmdisk(tokens []string) (*RMDISK, string, error) {
	cmd := &RMDISK{} // Crea una nueva instancia de RMDISK


	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`(?i)-path="[^"]+"|(?i)-path=[^\s]+`)

	matches := re.FindAllString(args, -1)

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return nil, "ERROR: formato de parámetro inválido", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar el parámetro -path
		switch key {

			case "-path":
				if value == "" {
					return nil, "ERROR: el path no puede estar vacío", errors.New("el path no puede estar vacío")
				}
				cmd.path = value
			default:
				// Si el parámetro no es reconocido, devuelve un error
				return nil, "ERROR: parámetro desconocido", fmt.Errorf("parámetro desconocido: %s", key)

		}

	}

	// Verifica que el parámetro -path haya sido proporcionado
	if cmd.path == "" {
		return nil, "ERROR: faltan parámetros requeridos: -path", errors.New("faltan parámetros requeridos: -path")
	}

	successMsg, err := util.DeleteBinaryFile(cmd.path) // Elimina el archivo binario del disco
	if err != nil {
		return nil, "Error al borrar Disco", err // Devuelve un error si no se pudo eliminar el disco
	}

	return cmd, successMsg ,nil // Devuelve el comando RMDISK creado
}
