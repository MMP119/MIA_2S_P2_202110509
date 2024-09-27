package global

import (
	"server/structures"
	"server/util"
	"strings"
)

var (
	// ArchivosCarpetas : Estructura que almacena los archivos y carpetas de una partición
	ArchivosCarpetas = make(map[string]string)
)

//con esta fucncion se obtiene los archivos y carpetas de la Raiz de la particion, solamente /
func ObtenerArchivosCarpetasRaiz(idParticion string, path string) map[string]string {

	//limpiar la estructura
	ArchivosCarpetas = make(map[string]string)

	if(path != ""){
		//return nil
		//aquí moverse y buscar la carpeta o archivo y obtener los archivos y carpetas de esa carpeta, en caso de archivo mostrar el contenido
		//en caso de carpeta mostrar los archivos y carpetas que contiene
		ArchivosCarpetasR := mostrarContenidos(idParticion, path)
		//fmt.Println(ArchivosCarpetasR)
		return ArchivosCarpetasR

	}

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

				}
			}

		}
	}

	return ArchivosCarpetas
}


//funcion para obtener los archivos o carpetas dentro de la ruta
func mostrarContenidos(idParticion string, path string)map[string]string{

	//limpiar la estructura
	ArchivosCarpetas = make(map[string]string)

	parentDirs, destino := utils.GetParentDirectories(path)

	if(strings.Contains(destino, ".txt")){
		return nil
	}


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

			FolderBlock := &structures.FolderBlock{}
			err = FolderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
			if err != nil {
				return nil
			}

			for _, content := range FolderBlock.B_content{
				
				if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".."{

					for i:=0; i<len(parentDirs); i++{

						if strings.Trim(string(content.B_name[:]), "\x00") == parentDirs[i] {
							
							//fmt.Println(parentDirs[i])
							//acá la funcion recursiva para avanzar entre nodos para llegar a los bloques
							ArchivosCarpetas1 := CarpetasRecursiva(inode, partitionSuperblock, partitionPath, parentDirs, destino)
							return ArchivosCarpetas1
						
						}}

					if strings.Trim(string(content.B_name[:]), "\x00") == destino {
						//movernos al inodo de la carpeta destino
						err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
						if err != nil {
							return nil
						}

						for _, block := range inode.I_block{
							if block != -1 {

								FolderBlock := &structures.FolderBlock{}
								err = FolderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
								if err != nil {
									return nil
								}

								for _, content := range FolderBlock.B_content{
									
									if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".."{

										if strings.Contains(strings.Trim(string(content.B_name[:]), "\x00"), ".txt") {
											//guardar el nombre del archivo
											ArchivosCarpetas[strings.Trim(string(content.B_name[:]), "\x00")] = "archivo"
										}else{
											//guardar el nombre de la carpeta
											ArchivosCarpetas[strings.Trim(string(content.B_name[:]), "\x00")] = "carpeta"
										}
									
									}

								}

							}
						
						}
						
					}

				}
			}

		}
	}
	return ArchivosCarpetas
}


func CarpetasRecursiva(inode *structures.Inode, partitionSuperblock *structures.SuperBlock, partitionPath string, parentDirs []string, destDir string)map[string]string{
	//limpiar la estructura
	ArchivosCarpetas = make(map[string]string)

	folderBlock := &structures.FolderBlock{}

	for _, block := range inode.I_block{
		if block != -1{
			err := folderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
			if err != nil {
				return nil
			}

			for _, content := range folderBlock.B_content{
				
				if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".."{
					
					for i := 0; i<len(parentDirs); i++{
						
						if strings.Trim(string(content.B_name[:]), "\x00") == parentDirs[i] {

							//movernos al inodo de la carpeta destino
							err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
							if err != nil {
								return nil
							}

							//llamar a la funcion recursiva
							ArchivosCarpetas1 := CarpetasRecursiva(inode, partitionSuperblock, partitionPath, parentDirs, destDir)
							return ArchivosCarpetas1
						}

						if strings.Trim(string(content.B_name[:]), "\x00") == destDir {
							
							err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
							if err != nil {
								return nil
							}

							for _, block := range inode.I_block{
								if block != -1 {

									FolderBlock := &structures.FolderBlock{}
									err = FolderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
									if err != nil {
										return nil
									}

									for _, content := range FolderBlock.B_content{
										
										if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".."{

											if strings.Contains(strings.Trim(string(content.B_name[:]), "\x00"), ".txt") {
												//guardar el nombre del archivo
												ArchivosCarpetas[strings.Trim(string(content.B_name[:]), "\x00")] = "archivo"
											}else{
												//guardar el nombre de la carpeta
												ArchivosCarpetas[strings.Trim(string(content.B_name[:]), "\x00")] = "carpeta"
											}
										
										}

									}

								}
							
							}

						}
					}
				}
			}
		}
	}
	return ArchivosCarpetas
}


// func mostrarContenidoArchivo(idParticion string, path string) string{


// }