package commands

import (
	"errors"
	"fmt"
	"regexp"
	structures "server/structures"
	"strconv"
	"strings"
)



func ParserMkdisk(tokens []string) (*structures.MKDISK, string,error) {

	cmd := &structures.MKDISK{} 

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`(?i)-size=\d+|(?i)-unit=[kKmM]|(?i)-fit=[bBfFwW]{2}|(?i)-path="[^"]+"|(?i)-path=[^\s]+`)

	matches := re.FindAllString(args, -1)

	for _, match := range matches {

		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return nil, "ERROR: formato de parámetro inválido", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {

			case "-size":
				size, err := strconv.Atoi(value)
				if err != nil || size <= 0 {
					return nil, "ERROR: el tamaño debe ser un número entero positivo", errors.New("el tamaño debe ser un número entero positivo")
				}
				cmd.Size = size

			case "-unit":
				value = strings.ToUpper(value)
				if value != "K" && value != "M" {
					return nil, "ERROR: la unidad debe ser K o M", errors.New("la unidad debe ser K o M")
				}
				cmd.Unit = value

			case "-fit":
				value = strings.ToUpper(value)
				if value != "BF" && value != "FF" && value != "WF" {
					return nil, "ERROR: el ajuste debe ser BF, FF o WF", errors.New("el ajuste debe ser BF, FF o WF")
				}
				cmd.Fit = value

			case "-path":
				if value == "" {
					return nil, "ERROR: El path no puede estar vacío", errors.New("el path no puede estar vacío")
				}
				cmd.Path = value

			default:
				return nil, "ERROR: parámetro desconocido", fmt.Errorf("parámetro desconocido: %s", key)
			}
	}

	if cmd.Size == 0 {
		return nil, "ERROR: faltan parámetros requeridos: -size", errors.New("faltan parámetros requeridos: -size")
	}

	if cmd.Path == "" {
		return nil, "ERROR: faltan parámetros requeridos: -path", errors.New("faltan parámetros requeridos: -path")
	}

	if cmd.Unit == "" {
		cmd.Unit = "M"
	}
	if cmd.Fit == "" {
		cmd.Fit = "FF"
	}

	msg,err := structures.CommandMkdisk(cmd)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, msg, err
	}

	return cmd, "COMANDO MKDISK: Disco Creado Exitosamente", nil // Devuelve el comando MKDISK creado
}

