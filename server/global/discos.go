package global

import (
	"path/filepath"
	"server/globales"
)


func GetPathDisk(id string)map[string]string{

	if(id == "root"){
		return globales.GetPathDisks(id)
	}
	_, partitionPath, _ := GetMountedPartition(id)

	//obtener el ultimo caracter del path
	idDisk := filepath.Base(partitionPath)
	return globales.GetPathDisks(idDisk)
}