package commands

import (
	"errors"
	"fmt"
	"regexp"
	global "server/global"
	structures "server/structures"
	"strings"
)

type CHGRP struct {
	User string
	Group string

}


//este comando crea un grupo para los usuarios de la particion
func ParseChgrp(tokens []string)(*CHGRP, string, error){

	cmd := &CHGRP{}

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando mkfile
	re := regexp.MustCompile(`(?i)-user="[^"]+"|(?i)-user=[^\s]+|(?i)-grp="[^"]+"|(?i)-grp=[^\s]+`)
	// Encuentra todas las coincidencias de la expresión regular en la cadena de argumentos
	matches := re.FindAllString(args, -1)

	for _, match := range matches{
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return nil, "ERROR: formato de parámetro inválido", fmt.Errorf("formato de parámetro inválido: %s", match)
		}

		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
			case "-user":
				if cmd.User != "" {
					return nil, "ERROR: user no puede estar vacío", errors.New("user no puede estar vacío")
				}
				cmd.User = value
			case "-grp":
				if cmd.Group != "" {
					return nil, "ERROR: grupo no puede estar vacío", errors.New("grupo no puede estar vacío")
				}
				cmd.Group = value
			default:
				return nil, "ERROR: parámetro inválido", fmt.Errorf("parámetro inválido: %s", key)
		}
	}

	if cmd.User == "" {
		return nil, "ERROR: nombre no puede estar vacío", errors.New("nombre no puede estar vacío")
	}
	if cmd.Group == "" {
		return nil, "ERROR: grupo no puede estar vacío", errors.New("grupo no puede estar vacío")
	}

	msg, err := CommandChgrp(cmd)
	if err != nil {
		return nil, msg, err
	}

	return cmd, msg, nil

}


func CommandChgrp(cmd *CHGRP) (string, error) {
	idPartition := global.GetIDSession()

	usuario := global.GetUserActive(idPartition)

	//verificar que el usuario sea el root
	if usuario != "root" {
		return "Error: el usuario no es root", errors.New("el usuario no es root")
	}


	// Obtener la partición con el id en donde se realizará el login
	partitionSuperblock, _, partitionPath, err := global.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return "Error al obtener la partición montada en el comando chgrp", fmt.Errorf("error al obtener la partición montada: %v", err)
	}

	inode := &structures.Inode{}

	// Deserializar el inodo raíz
	err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(0*partitionSuperblock.S_inode_size)))
	if err != nil {
		return "Error al obtener el inodo raíz", fmt.Errorf("error al obtener el inodo raíz: %v", err)
	}

	// Verificar que el primer i-nodo esté en cero
	if inode.I_block[0] == 0 {
		// Moverme al bloque 0
		folderBlock := &structures.FolderBlock{}

		err = folderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(inode.I_block[0]*partitionSuperblock.S_block_size)))
		if err != nil {
			return "Error al obtener el bloque 0", fmt.Errorf("error al obtener el bloque 0: %v", err)
		}

		// Recorrer los contenidos del bloque 0 para buscar el archivo users.txt
		for _, contenido := range folderBlock.B_content {
			name := strings.Trim(string(contenido.B_name[:]), "\x00") // Elimina caracteres nulos
			apuntador := contenido.B_inodo
			if name == "users.txt" {

				// Moverme al inodo que apunta el contenido
				err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(apuntador*partitionSuperblock.S_inode_size)))
				if err != nil {
					return "Error al obtener el inodo del archivo users.txt", fmt.Errorf("error al obtener el inodo del archivo users.txt: %v", err)
				}

				// Verificar que el primer i-nodo esté en 1
				if inode.I_block[0] == 1 {
					// Moverme al bloque 1
					fileBlock := &structures.FileBlock{}

					err = fileBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(inode.I_block[0]*partitionSuperblock.S_block_size)))
					if err != nil {
						return "Error al obtener el bloque 1 del archivo users.txt", fmt.Errorf("error al obtener el bloque 1 del archivo users.txt: %v", err)
					}

					// Obtener el contenido del archivo users.txt
					contenido := strings.Trim(string(fileBlock.B_content[:]), "\x00") // Elimina caracteres nulos

					// Reemplazar \r\n con \n para asegurar saltos de línea uniformes
					contenido = strings.ReplaceAll(contenido, "\r\n", "\n")

					// Dividir en líneas para obtener cada usuario o grupo
					lines := strings.Split(contenido, "\n")

					//variable para indicar si el grupo existe
					groupFound := false

					// Variable para indicar si el usuario fue encontrado
					userFound := false

					for _, line := range lines {
						if strings.TrimSpace(line) == "" {
							continue // Ignorar líneas vacías
						}

						values := strings.Split(line, ",")

						// Verificar si es un usuario y si coincide con el usuario que queremos cambiar
						if len(values) >= 3 && values[1] == "G" {
							if values[2] == cmd.Group {
								groupFound = true
							}
						}
					}

					if !groupFound {
						return fmt.Sprintf("Error: Grupo %s no encontrado", cmd.Group), nil
					}

					// Recorrer cada línea del archivo users.txt
					for i, line := range lines {
						if strings.TrimSpace(line) == "" {
							continue // Ignorar líneas vacías
						}

						values := strings.Split(line, ",")

						// Verificar si es un usuario y si coincide con el usuario que queremos cambiar
						if len(values) >= 5 && values[1] == "U" && values[3] == cmd.User {
							// Actualizar el grupo del usuario
							values[2] = cmd.Group
							lines[i] = strings.Join(values, ",")
							userFound = true
						}

					}

					if !userFound {
						return fmt.Sprintf("Error: Usuario %s no encontrado", cmd.User), nil
					}

					// Volver a unir las líneas en un único string
					newContent := strings.Join(lines, "\n")

					// Escribir el contenido actualizado en el bloque del archivo
					copy(fileBlock.B_content[:], newContent)

					// Guardar los cambios en el archivo
					err = fileBlock.Serialize(partitionPath, int64(partitionSuperblock.S_block_start+(inode.I_block[0]*partitionSuperblock.S_block_size)))
					if err != nil {
						return "Error al escribir el bloque 1 del archivo users.txt", fmt.Errorf("error al escribir el bloque 1 del archivo users.txt: %v", err)
					}
					fmt.Println("-----------MODIFICAR GRUPO USUARIO----------------")
					fileBlock.Print()
					fmt.Println("---------------------------")
					return fmt.Sprintf("Comando CHGRP: El grupo del usuario %s ha sido cambiado exitosamente a %s", cmd.User, cmd.Group), nil
				}
			}
		}
	}

	return "Error: No se encontró el archivo users.txt", nil
}