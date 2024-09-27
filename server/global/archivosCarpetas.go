package global

import (
	"regexp"
	"server/structures"
	"server/util"
	"strings"
)

var (
	// ArchivosCarpetas : Estructura que almacena los archivos y carpetas de una partición
	ArchivosCarpetas = make(map[string]string)
)

//con esta fucncion se obtiene los archivos y carpetas de la Raiz de la particion, solamente /
func ObtenerArchivosCarpetasRaiz(idParticion string, path string) (map[string]string, error) {

	//limpiar la estructura
	ArchivosCarpetas = make(map[string]string)

	if(path != ""){
		//return nil
		//aquí moverse y buscar la carpeta o archivo y obtener los archivos y carpetas de esa carpeta, en caso de archivo mostrar el contenido
		//en caso de carpeta mostrar los archivos y carpetas que contiene
		ArchivosCarpetasR, err := mostrarContenidos(idParticion, path)
		//fmt.Println(ArchivosCarpetasR)
		return ArchivosCarpetasR, err

	}

	partitionSuperblock, _, partitionPath, err := GetMountedPartitionSuperblock(idParticion)
	if err != nil {
		return nil, err
	}


	inode := &structures.Inode{}
	err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(0*partitionSuperblock.S_inode_size)))
	if err != nil {
		return nil, err
	}

	for _, block := range inode.I_block{
		if block != -1 {
			// significa que hay un bloque, entonces hay que leerlo y guardar los archivos y carpetas que contiene en ArchivosCarpetas
			FolderBlock := &structures.FolderBlock{}
			err = FolderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
			if err != nil {
				return nil, err
			}

			for _, content := range FolderBlock.B_content{
				
				if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".."{
					raiz := ""
					if(strings.Contains(string(content.B_name[:]), "\x00")){
						raiz = strings.Trim(string(content.B_name[:]), "\x00")
						if strings.Contains(raiz, "\x00"){
							raiz = strings.Trim(raiz, "\x00")
						}
					}

					//si continene un caracter nulo o especial no se guarda
					var valid = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)
					//valid.MatchString(raiz)
					if(!valid.MatchString(raiz)){
						continue
					}


					mutex.Lock() // Mutex para asegurar acceso exclusivo al mapa

					//verificar si es carpeta o archivo, los archivos tienen un txt al final
					if strings.Contains(raiz, ".txt") {
						//guardar el nombre del archivo
						ArchivosCarpetas[raiz] = "archivo"
					}else{
						//guardar el nombre de la carpeta
						ArchivosCarpetas[raiz] = "carpeta"
					}
					mutex.Unlock() // Liberar el mutex después de escribir en el mapa

				}
			}

		}
	}

	return ArchivosCarpetas, nil
}


//funcion para obtener los archivos o carpetas dentro de la ruta
func mostrarContenidos(idParticion string, path string)(map[string]string, error){

	//limpiar la estructura
	ArchivosCarpetas = make(map[string]string)

	parentDirs, destino := utils.GetParentDirectories(path)

	err := error(nil)

	if(strings.Contains(destino, ".txt")){
		ArchivosCarpetas, err = mostrarArchivos(idParticion, path)
		return ArchivosCarpetas, err
	}


	partitionSuperblock, _, partitionPath, err := GetMountedPartitionSuperblock(idParticion)
	if err != nil {
		return nil, err
	}

	inode := &structures.Inode{}
	err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(0*partitionSuperblock.S_inode_size)))
	if err != nil {
		return nil, err
	}

	for _, block := range inode.I_block{
		if block != -1 {

			FolderBlock := &structures.FolderBlock{}
			err = FolderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
			if err != nil {
				return nil, err
			}

			for _, content := range FolderBlock.B_content{
				
				if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".."{

					for i:=0; i<len(parentDirs); i++{

						if strings.Trim(string(content.B_name[:]), "\x00") == parentDirs[i] {

							//acá la funcion recursiva para avanzar entre nodos para llegar a los bloques
							ArchivosCarpetas1, err := CarpetasRecursiva(inode, partitionSuperblock, partitionPath, parentDirs, destino)
							return ArchivosCarpetas1, err
						
						}}

					if strings.Trim(string(content.B_name[:]), "\x00") == destino {
						//movernos al inodo de la carpeta destino
						err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
						if err != nil {
							return nil, err
						}

						for _, block := range inode.I_block{
							if block != -1 {

								FolderBlock := &structures.FolderBlock{}
								err = FolderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
								if err != nil {
									return nil, err
								}

								for _, content := range FolderBlock.B_content{
									
									if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".."{

										
										mutex.Lock() // Mutex para asegurar acceso exclusivo al mapa

										if strings.Contains(strings.Trim(string(content.B_name[:]), "\x00"), ".txt") {
											//guardar el nombre del archivo
											ArchivosCarpetas[strings.Trim(string(content.B_name[:]), "\x00")] = "archivo"
										}else{
											//guardar el nombre de la carpeta
											ArchivosCarpetas[strings.Trim(string(content.B_name[:]), "\x00")] = "carpeta"
										}
										
										mutex.Unlock() // Liberar el mutex después de escribir en el mapa
									}

								}

							}
						
						}
						
					}

				}
			}

		}
	}
	return ArchivosCarpetas, nil
}


func CarpetasRecursiva(inode *structures.Inode, partitionSuperblock *structures.SuperBlock, partitionPath string, parentDirs []string, destDir string)(map[string]string, error){
	//limpiar la estructura
	ArchivosCarpetas = make(map[string]string)

	folderBlock := &structures.FolderBlock{}

	for _, block := range inode.I_block{
		if block != -1{
			err := folderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
			if err != nil {
				return nil, err
			}

			for _, content := range folderBlock.B_content{
				
				if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".."{
					
					for i := 0; i<len(parentDirs); i++{
						
						if strings.Trim(string(content.B_name[:]), "\x00") == parentDirs[i] {

							//movernos al inodo de la carpeta destino
							err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
							if err != nil {
								return nil, err
							}

							//llamar a la funcion recursiva
							ArchivosCarpetas1, err:= CarpetasRecursiva(inode, partitionSuperblock, partitionPath, parentDirs, destDir)
							return ArchivosCarpetas1, err
						}

						if strings.Trim(string(content.B_name[:]), "\x00") == destDir {
							
							err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
							if err != nil {
								return nil, err
							}

							for _, block := range inode.I_block{
								if block != -1 {

									FolderBlock := &structures.FolderBlock{}
									err = FolderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
									if err != nil {
										return nil, err
									}

									for _, content := range FolderBlock.B_content{
										
										if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".."{

											mutex.Lock() // Mutex para asegurar acceso exclusivo al mapa

											if strings.Contains(strings.Trim(string(content.B_name[:]), "\x00"), ".txt") {
												//guardar el nombre del archivo
												ArchivosCarpetas[strings.Trim(string(content.B_name[:]), "\x00")] = "archivo"
											}else{
												//guardar el nombre de la carpeta
												ArchivosCarpetas[strings.Trim(string(content.B_name[:]), "\x00")] = "carpeta"
											}

											mutex.Unlock() // Liberar el mutex después de escribir en el mapa
										
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
	return ArchivosCarpetas, nil
}



// para mostrar archivos
func mostrarArchivos(idParticion string, path string)(map[string]string, error){
	//limpiar la estructura
	ArchivosCarpetas = make(map[string]string)

	parentDirs, destino := utils.GetParentDirectories(path)

	partitionSuperblock, _, partitionPath, err := GetMountedPartitionSuperblock(idParticion)
	if err != nil {
		return nil, err
	}

	inode := &structures.Inode{}
	err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(0*partitionSuperblock.S_inode_size)))
	if err != nil {
		return nil, err
	}

	for _, block := range inode.I_block{
		if block != -1 {

			FolderBlock := &structures.FolderBlock{}
			err = FolderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
			if err != nil {
				return nil, err
			}

			for _, content := range FolderBlock.B_content{
				
				if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".."{

					for i:=0; i<len(parentDirs); i++{

						if strings.Trim(string(content.B_name[:]), "\x00") == parentDirs[i] {

							//acá la funcion recursiva para avanzar entre nodos para llegar a los bloques
							ArchivosCarpetas1, err := ArchivosRecursiva(inode, partitionSuperblock, partitionPath, parentDirs, destino)
							return ArchivosCarpetas1, err
						
						}}

					if strings.Trim(string(content.B_name[:]), "\x00") == destino {
						//movernos al inodo de la carpeta destino
						err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
						if err != nil {
							return nil, err
						}

						for _, block := range inode.I_block{
							if block != -1 {
								FolderBlock := &structures.FolderBlock{}
								err = FolderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
								if err != nil {
									return nil, err
								}

								fileBlock := &structures.FileBlock{}
								salida := ""
								for _, block := range inode.I_block{
									if block != -1 {
										err = fileBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
										if err != nil {
											return nil, err
										}
										salida += strings.Trim(string(fileBlock.B_content[:]), "\x00")
									}
								}
								mutex.Lock() // Mutex para asegurar acceso exclusivo al mapa
								ArchivosCarpetas[salida] = "salida"			
								mutex.Unlock() // Liberar el mutex después de escribir en el mapa
							}
						}
					}
				}
			}
		}
	}
	return ArchivosCarpetas, nil
}



func ArchivosRecursiva(inode *structures.Inode, partitionSuperblock *structures.SuperBlock, partitionPath string, parentDirs []string, destDir string)(map[string]string, error){
	//limpiar la estructura
	ArchivosCarpetas = make(map[string]string)

	folderBlock := &structures.FolderBlock{}

	for _, block := range inode.I_block{
		if block != -1{
			err := folderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
			if err != nil {
				return nil, err
			}

			for _, content := range folderBlock.B_content{
				
				if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".."{
					
					for i := 0; i<len(parentDirs); i++{
						
						if strings.Trim(string(content.B_name[:]), "\x00") == parentDirs[i] {

							//movernos al inodo de la carpeta destino
							err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
							if err != nil {
								return nil, err
							}

							//llamar a la funcion recursiva
							ArchivosCarpetas1, err := ArchivosRecursiva(inode, partitionSuperblock, partitionPath, parentDirs, destDir)
							return ArchivosCarpetas1, err
						}

						if strings.Trim(string(content.B_name[:]), "\x00") == destDir {
							
							err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
							if err != nil {
								return nil, err
							}

							for _, block := range inode.I_block{
								if block != -1 {

									FolderBlock := &structures.FolderBlock{}
									err = FolderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
									if err != nil {
										return nil, err
									}

									fileBlock := &structures.FileBlock{}
									salida := ""
									for _, block := range inode.I_block{
										if block != -1 {
											err = fileBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
											if err != nil {
												return nil, err
											}
											salida += strings.Trim(string(fileBlock.B_content[:]), "\x00")
										}
									}
									mutex.Lock()
									ArchivosCarpetas[salida] = "salida"
									mutex.Unlock()
								}							
							}
						}
					}
				}
			}
		}
	}
	return ArchivosCarpetas, nil
}