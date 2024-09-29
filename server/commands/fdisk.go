package commands

import (
	"errors"  
	"fmt"     
	"regexp"  
	"strconv" 
	"strings" 
)



// CommandFdisk parsea el comando fdisk y devuelve una instancia de FDISK
func ParserFdisk(tokens []string) (*FDISK, string, error) {
	cmd := &FDISK{} 

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`(?i)-size=\d+|(?i)-unit=[bBkKmM]|(?i)-fit=[bBfF]{2}|(?i)-path="[^"]+"|(?i)-path=[^\s]+|(?i)-type=[pPeElL]|(?i)-name="[^"]+"|(?i)-name=[^\s]+|(?i)-delete=[^\s]+|(?i)-delete="[^"]+"|(?i)-add=[^\s]+|(?i)-add="[^"]+"`)

	matches := re.FindAllString(args, -1)

	var banderaDelete bool = false
	var banderaAdd bool = false

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
			if value!= "B" && value != "K" && value != "M" {
				return nil, "ERROR: La unidad debe ser B, K o M", errors.New("la unidad debe ser B, K o M")
			}
			cmd.Unit = strings.ToUpper(value)

		case "-path":

			if value == "" {
				return nil, "ERROR: el path no puede estar vacío", errors.New("el path no puede estar vacío")
			}
			cmd.Path = value

		case "-type":

			value = strings.ToUpper(value)
			if value != "P" && value != "E" && value != "L" {
				return nil, "ERROR: el tipo debe ser P, E o L", errors.New("el tipo debe ser P, E o L")
			}
			cmd.TypE = value
		
		case "-fit":

			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return nil, "ERROR: el ajuste debe ser BF, FF o WF", errors.New("el ajuste debe ser BF, FF o WF")
			}
			cmd.Fit = value
		
		case "-name":

			if value == "" {
				return nil, "ERROR: el nombre no puede estar vacío", errors.New("el nombre no puede estar vacío")
			}
			cmd.Name = value

		case "-delete":
			banderaDelete = true
			value = strings.ToLower(value)
			if value != "fast" && value != "full" {
				return nil, "ERROR: el tipo de eliminación debe ser fast o full", errors.New("el tipo de eliminación debe ser fast o full")
			}
			cmd.Delete = value

		case "-add":
			banderaAdd = true
			//verificar que sean números positivos o negativos
			num, err := strconv.Atoi(value)
			if err != nil {
				return nil, "ERROR: el valor de add debe ser un número entero", errors.New("el valor de add debe ser un número entero")
			}
			cmd.Add = num

		default:

			return nil, "ERROR: parámetro desconocido", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que los parámetros -size, -path y -name hayan sido proporcionados
	if cmd.Size == 0 && !banderaDelete && !banderaAdd {
		return nil, "ERROR: faltan parámetros requeridos: -size", errors.New("faltan parámetros requeridos: -size")
	}
	if cmd.Path == "" {
		return nil, "ERROR: faltan parámetros requeridos: -path", errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.Name == "" {
		return nil, "ERROR: Faltan parámetros requeridos: -name", errors.New("faltan parámetros requeridos: -name")
	}

	if cmd.Unit == "" {
		cmd.Unit = "K"
	}

	if cmd.Fit == "" {
		cmd.Fit = "WF"
	}

	if cmd.TypE == "" {
		cmd.TypE = "P"
	}

	if cmd.Delete == "" {
		banderaDelete = false
		cmd.Delete = ""
	}

	if cmd.Add == 0 {
		banderaAdd = false
		cmd.Add = 0
	}

	banderaDelete = false
	banderaAdd = false

	msg, err := CommandFdisk(cmd)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, msg, err
	}

	return cmd, "Comando FDISK: realizado correctamente", nil // Devuelve el comando FDISK creado
}