package structures

import (
	"fmt"
	"os"
	"path/filepath"
	util "server/util"
	guardarPath "server/globales"
)

type MKDISK struct {
	Size int    
	Unit string
	Fit  string 
	Path string 
}


func CommandMkdisk(mkdisk *MKDISK) (string, error) {

	sizeBytes, err := util.ConvertToBytes(mkdisk.Size, mkdisk.Unit)
	if err != nil {
		fmt.Println("Error converting size:", err)
		return "Error converting size MKDISK",err
	}

	var msg string

	msg, err = CreateDisk(mkdisk, sizeBytes)
	if err != nil {
		fmt.Println("Error creating disk:", err)
		return msg, err
	}

	msg, err = CreateMBR(mkdisk, sizeBytes)
	if err != nil {
		fmt.Println("Error creating MBR:", err)
		return msg,err
	}

	return "",nil
}

func CreateDisk(mkdisk *MKDISK, sizeBytes int) (string, error) {

	err := os.MkdirAll(filepath.Dir(mkdisk.Path), os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directories:", err)
		return "Error creating directories",err
	}

	file, err := os.Create(mkdisk.Path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return "Error creating file", err
	}
	defer file.Close()

	buffer := make([]byte, 1024*1024) // Crea un buffer de 1 MB
	for sizeBytes > 0 {
		writeSize := len(buffer)
		if sizeBytes < writeSize {
			writeSize = sizeBytes // Ajusta el tamaño de escritura si es menor que el buffer
		}
		if _, err := file.Write(buffer[:writeSize]); err != nil {
			return "Error, falló la escritura en el disco",err // Devuelve un error si la escritura falla
		}
		sizeBytes -= writeSize // Resta el tamaño escrito del tamaño total
	}
	
	//obtener el nombre del disco
	idDisco := filepath.Base(mkdisk.Path)

	guardarPath.SetPathDisk(idDisco, mkdisk.Path)

	return "",nil
}