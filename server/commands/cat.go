package commands

import (
	"errors"
	"fmt"
	"regexp"
	global "server/global"
	structures "server/structures"
	utils "server/util"
	"strings"
)

type CAT struct {
	Filen string
}

func ParseCat(tokens []string) (*CAT, string, error) {
	cmd := &CAT{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`(?i)-file[1-9][0-9]*="[^"]+"|(?i)-file[1-9][0-9]*=[^\s]+`)

	matches := re.FindAllString(args, -1)

	var allContent strings.Builder // Acumulador para el contenido de todos los archivos

	for _, match := range matches {

		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return nil, "ERROR: formato de parámetro inválido", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		if strings.HasPrefix(key, "-file") {
			if value == "" {
				return nil, "ERROR: la ruta del archivo no puede estar vacía", errors.New("el nombre del archivo no puede estar vacío")
			}
			cmd.Filen = value
		} else {
			return nil, "ERROR: parámetro no reconocido", fmt.Errorf("parámetro no reconocido: %s", key)
		}

		if cmd.Filen == "" {
			return nil, "ERROR: el nombre del archivo es obligatorio", errors.New("el nombre del archivo es obligatorio")
		}

		// Leer el contenido del archivo
		content, err := CommandCAT(cmd)
		if err != nil {
			return nil, content, err
		}

		// Concatenar el contenido al acumulador
		allContent.WriteString(content+"\n")
	}
	// Retornar el contenido concatenado
	return cmd, "Comando CAT: realizado correctamente\n LECTURA: \n" + allContent.String(), nil
}

func CommandCAT(cmd *CAT) (string, error) {

	// leer un archivo que esté en la ruta especificada dentro del bloque
	// inodo -> bloque -> contenido

	//la ruta del archivo es cmd.Filen, donde está el inodo -> bloque -> contenido
	parentDirs, destDir := utils.GetParentDirectories(cmd.Filen)
	fmt.Println("\nDirectorios padres hacia el archivo:", parentDirs) //-> arreglo de strings de los directorios padres [home user docs ...]
	fmt.Println("Directorio destino (archivo):", destDir)             //-> nombre del archivo a.txt por ejemplo

	//obtener el id de la particion donde se está logueado
	idPartition := global.GetIDSession()

	//obtenemos primero el superbloque para obtener el inodo raíz y luego el inodo del archivo
	partitionSuperblock, partition, partitionPath, err := global.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return "Error al obtener la partición montada en el comando login", fmt.Errorf("error al obtener la partición montada: %v", err)
	}

	inode := &structures.Inode{}

	err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(0*partitionSuperblock.S_inode_size)))
	if err != nil {
		return "Error al obtener el inodo raíz", fmt.Errorf("error al obtener el inodo raíz: %v", err)
	}

	//recorrer los bloques del inodo raíz
	for _, block := range inode.I_block {

		if block != -1 {

			//verificar sobre los bloques del inodo, recorrerlos para encontar el bloque que contiene la ruta para llegar al archivo
			folderBlock := &structures.FolderBlock{}

			err = folderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
			if err != nil {
				return "Error al obtener el bloque", fmt.Errorf("error al obtener el bloque: %v", err)
			}

			//recorrer los contenidos del bloque
			for _, content := range folderBlock.B_content {
				//fmt.Println(strings.Trim(string(content.B_name[:]), "\x00"))

				if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".." && strings.Trim(string(content.B_name[:]), "\x00") != "users.txt" {
					//fmt.Println("Bloque encontrado:", content.B_inodo, string(content.B_name[:]))

					for i := 0; i < len(parentDirs); i++ {
						//fmt.Println("Directorio a buscar:", parentDirs[i])

						if strings.Trim(string(content.B_name[:]), "\x00") == parentDirs[i] {
							//fmt.Println("Directorio encontrado:", parentDirs[i])
							//vamos al inodo que apunte el bloque
							err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
							if err != nil {
								return "Error al obtener el inodo", fmt.Errorf("error al obtener el inodo: %v", err)
							}
							msg := ""
							msg, err = recursiveBlock(inode, partitionSuperblock, partitionPath, parentDirs, destDir)
							if err != nil {
								return msg, err
							}
							return msg, nil

						}
						if strings.Trim(string(content.B_name[:]), "\x00") == destDir {
							//fmt.Println("Archivo encontrado:", destDir)
							//moverse al inodo que apunta el bloque
							err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
							if err != nil {
								return "Error al obtener el inodo", fmt.Errorf("error al obtener el inodo: %v", err)
							}

							//recorrer los bloques del inodo para obtener el contenido del archivo
							fileBlock := &structures.FileBlock{}
							salida := ""
							for _, block := range inode.I_block {
								if block != -1 {
									err = fileBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
									if err != nil {
										return "Error al obtener el bloque", fmt.Errorf("error al obtener el bloque: %v", err)
									}
									//eliminar caracteres nulos
									salida += strings.Trim(string(fileBlock.B_content[:]), "\x00")
									// return salida, nil
								}
							}
							return salida, nil
						}

					}

					if strings.Trim(string(content.B_name[:]), "\x00") == destDir {
						//fmt.Println("Archivo encontrado:", destDir)
						//moverse al inodo que apunta el bloque
						err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
						if err != nil {
							return "Error al obtener el inodo", fmt.Errorf("error al obtener el inodo: %v", err)
						}

						//recorrer los bloques del inodo para obtener el contenido del archivo
						fileBlock := &structures.FileBlock{}
						salida := ""
						for _, block := range inode.I_block {
							if block != -1 {
								err = fileBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
								if err != nil {
									return "Error al obtener el bloque", fmt.Errorf("error al obtener el bloque: %v", err)
								}
								//eliminar caracteres nulos
								salida += strings.Trim(string(fileBlock.B_content[:]), "\x00")
								// return salida, nil
							}
						}
						return salida, nil
					}

				}
			}

		}
	}

	err = partitionSuperblock.Serialize(partitionPath, int64(partition.Part_start))
	if err != nil {
		return "error al serializar el superbloque de la partición", fmt.Errorf("error al serializar el superbloque de la partición: %v", err)
	}

	return "", nil

}

// funcion recursiva para analizar los bloques de un inodo y moverse al bloque
func recursiveBlock(inode *structures.Inode, partitionSuperblock *structures.SuperBlock, partitionPath string, parentDirs []string, destDir string) (string, error) {
	//verificar sobre los bloques del inodo, recorrerlos para encontar el bloque que contiene la ruta para llegar al archivo
	folderBlock := &structures.FolderBlock{}

	//recorrer los bloques del inodo raíz
	for _, block := range inode.I_block {
		if block != -1 {
			err := folderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
			if err != nil {
				return "Error al obtener el bloque", fmt.Errorf("error al obtener el bloque: %v", err)
			}

			//recorrer los contenidos del bloque
			for _, content := range folderBlock.B_content {

				if content.B_inodo != -1 && content.B_inodo != 0 {

					for i := 0; i < len(parentDirs); i++ {

						if strings.Trim(string(content.B_name[:]), "\x00") == parentDirs[i] {
							fmt.Println("Directorio encontrado RECURSIVA:", parentDirs[i])

							//vamos al inodo que apunte el bloque
							err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
							if err != nil {
								return "Error al obtener el inodo", fmt.Errorf("error al obtener el inodo: %v", err)
							}

							//si ya llegamos al último directorio padre, entonces inodo que apunta el bloque tiene el bloque que contiene el archivo
							if i == len(parentDirs)-1 {
								msg := ""
								msg, err = recursiveBlock(inode, partitionSuperblock, partitionPath, parentDirs, destDir)
								if err != nil {
									return msg, err
								}
								return msg, nil

							} else {
								//si no, entonces seguimos buscando en otro bloque del inodo
								msg := ""
								msg, err = recursiveBlock(inode, partitionSuperblock, partitionPath, parentDirs, destDir)
								if err != nil {
									return msg, err
								}
								return msg, nil
							}

						}
						if strings.Trim(string(content.B_name[:]), "\x00") == destDir {
							fmt.Println("Archivo encontrado RECURSIVA:", destDir)
							//moverse al inodo que apunta el bloque
							err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
							if err != nil {
								return "Error al obtener el inodo", fmt.Errorf("error al obtener el inodo: %v", err)
							}

							//recorrer los bloques del inodo para obtener el contenido del archivo
							fileBlock := &structures.FileBlock{}
							salida := ""
							for _, block := range inode.I_block {
								if block != -1 {
									err = fileBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
									if err != nil {
										return "Error al obtener el bloque", fmt.Errorf("error al obtener el bloque: %v", err)
									}
									//eliminar caracteres nulos
									salida += strings.Trim(string(fileBlock.B_content[:]), "\x00")
									//return salida, nil
								}
							}
							return salida, nil
						}
					}
				}
			}

		}
	}

	return "No se encontró el archivo", nil

}
