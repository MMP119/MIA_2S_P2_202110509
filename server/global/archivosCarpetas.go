package global

import (
	"server/structures"
	"strings"
)

var (
	// ArchivosCarpetas : Estructura que almacena los archivos y carpetas de una partición
	ArchivosCarpetas = make(map[string]string)
)


func ObtenerArchivosCarpetas(idParticion string, path string) map[string]string {


	partitionSuperblock, _, partitionPath, err := GetMountedPartitionSuperblock(idParticion)
	if err != nil {
		return nil
	}


	inode := &structures.Inode{}
	err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(0*partitionSuperblock.S_inode_size)))
	if err != nil {
		return nil
	}

	for _, block := range inode.I_block{
		if block != -1 {
			// significa que hay un bloque, entonces hay que leerlo y guardar los archivos y carpetas que contiene en ArchivosCarpetas
			FolderBlock := &structures.FolderBlock{}
			err = FolderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
			if err != nil {
				return nil
			}

			for _, content := range FolderBlock.B_content{
				
				if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".."{

					//verificar si es carpeta o archivo, los archivos tienen un txt al final
					if strings.Contains(strings.Trim(string(content.B_name[:]), "\x00"), ".txt") {
						//guardar el nombre del archivo
						ArchivosCarpetas[strings.Trim(string(content.B_name[:]), "\x00")] = "archivo"
					}else{
						//guardar el nombre de la carpeta
						ArchivosCarpetas[strings.Trim(string(content.B_name[:]), "\x00")] = "carpeta"
					}

					//moverse a los inodos que apuntan los bloques de carpetas con recursividad


				}
			}

		}
	}

	return ArchivosCarpetas
}

