package commands

import (
	"errors"
	"fmt"
	"regexp"
	//"server/global"
	//utils "server/util"
	"strings"
)


type COPY struct{
	Path string
	Destino string
}


func ParseCopy(tokens []string) (*COPY, string, error) {

	cmd := &COPY{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`(?i)-path="[^"]+"|(?i)-path=[^\s]+|(?i)-destino="[^"]+"|(?i)-destino=[^\s]+`)

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
			
			case "-destino":
				if value == "" {
					return nil, "ERROR: name es obligatorio", errors.New("el destino es obligatorio")
				}
				cmd.Destino = value

			default: 
				return nil, "ERROR: parámetro no reconocido", fmt.Errorf("parámetro no reconocido: %s", key)
		}
	}

	if cmd.Path == "" {
		return nil, "ERROR: el path es obligatorio", errors.New("el path es obligatorio")
	}	

	if cmd.Destino == "" {
		return nil, "ERROR: destino es obligatorio", errors.New("el destino es obligatorio")
	}

	msg, err := CommandCopy(cmd)
	if err != nil {
		return nil, "ERROR: no se pudo ejecutar el comando COPY", err
	}

	return cmd, msg, nil
}


func CommandCopy(cmd *COPY)(string, error){

	/*
	parenDirsPath, destiDirPath := utils.GetParentDirectories(cmd.Path)

	parenDirsDestino, destiDirDestino := utils.GetParentDirectories(cmd.Destino)

	idParticion := global.GetIDSession()

	//verificar si voy a copiar una carpeta o un archivo
	if strings.Contains(destiDirPath, ".txt"){
		
	}
	
	//verficar si el destino es una carpeta o un archivo
	if strings.Contains(destiDirDestino, ".txt"){

	}


	*/


	return "Comando COPY: realizado correctamente", nil
}