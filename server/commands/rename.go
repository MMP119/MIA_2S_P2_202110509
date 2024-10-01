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


type RENAME struct{
	Path string
	Name string
}


func ParseRename(tokens []string) (*RENAME, string, error) {

	cmd := &RENAME{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`(?i)-path="[^"]+"|(?i)-path=[^\s]+|(?i)-name="[^"]+"|(?i)-name=[^\s]+`)

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
			
			case "-name":
				if value == "" {
					return nil, "ERROR: name es obligatorio", errors.New("el contenido es obligatorio")
				}
				cmd.Name = value

			default: 
				return nil, "ERROR: parámetro no reconocido", fmt.Errorf("parámetro no reconocido: %s", key)
		}
	}

	if cmd.Path == "" {
		return nil, "ERROR: el path es obligatorio", errors.New("el path es obligatorio")
	}	

	if cmd.Name == "" {
		return nil, "ERROR: name es obligatorio", errors.New("el contenido es obligatorio")
	}

	msg, err := CommandRename(cmd)
	if err != nil {
		return nil, msg, err
	}

	return cmd, msg, nil
}


func CommandRename(cmd *RENAME) (string, error) {

	parentDirs, destDir := utils.GetParentDirectories(cmd.Path)
	
	idParticion := global.GetIDSession()
	
	partitionSuperblock, _, partitionPath, err := global.GetMountedPartitionSuperblock(idParticion)
	if err != nil {
		return "Error al obtener la partición montada en el comando login", fmt.Errorf("error al obtener la partición montada: %v", err)
	}

	inode := &structures.Inode{}
	err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(0*partitionSuperblock.S_inode_size)))
	if err != nil {
		return "no se pudo deserializar el inodo", err
	}

	for _, block := range inode.I_block{
		if block != -1 {

			FolderBlock := &structures.FolderBlock{}
			err = FolderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
			if err != nil {
				return "Error: no se pudo deserializar el bloque", err
			}

			for i, content := range FolderBlock.B_content{
				
				if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".."{

					for i:=0; i<len(parentDirs); i++{

						if strings.Trim(string(content.B_name[:]), "\x00") == parentDirs[i] {
							//recorro hasta llegar a la carpeta destino
							err := RenameRecursiva(inode, partitionSuperblock, partitionPath, parentDirs, destDir, cmd.Name)
							if err != nil {
								return "Error: no se pudo renombrar la carpeta", err
							}
							return "Comando RENAME: realizado correctamente", nil
						}
					}

					if strings.Trim(string(content.B_name[:]), "\x00") == destDir {

						if len(cmd.Name) > 12 {
							return "Error: el nombre es mayor a 12 bytes", fmt.Errorf("el nombre '%s' es mayor a 12 bytes", cmd.Name)
						}

						contenido := &FolderBlock.B_content[i]

						var nameArray [12]byte
						copy(nameArray[:], cmd.Name)
						//contenido.B_name = nameArray
						copy(contenido.B_name[:], nameArray[:])

						err = FolderBlock.Serialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
						if err != nil {
							return "Error: no se pudo serializar el bloque", err
						}
						break
					}
				}
			}
		}
	}

	return "Comando RENAME: realizado correctamente", nil
}


func RenameRecursiva(inode *structures.Inode, partitionSuperblock *structures.SuperBlock, partitionPath string, parentDirs []string, destDir string, Name string)(error){

	FolderBlock := &structures.FolderBlock{}

	for _, block := range inode.I_block{
		if block != -1{
			err := FolderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
			if err != nil {
				return err
			}

			for i, content := range FolderBlock.B_content{
				
				if content.B_inodo != -1 && content.B_inodo != 0 && strings.Trim(string(content.B_name[:]), "\x00") != "." && strings.Trim(string(content.B_name[:]), "\x00") != ".."{
					
					for i := 0; i<len(parentDirs); i++{
						
						if strings.Trim(string(content.B_name[:]), "\x00") == parentDirs[i] {

							//movernos al inodo de la carpeta destino
							err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
							if err != nil {
								return err
							}

							//llamar a la funcion recursiva
							RenameRecursiva(inode, partitionSuperblock, partitionPath, parentDirs, destDir, Name)
							return nil
						}
					}
					if strings.Trim(string(content.B_name[:]), "\x00") == destDir {

						if len(Name) > 12 {
							return fmt.Errorf("el nombre '%s' es mayor a 12 bytes", Name)
						}

						contenido := &FolderBlock.B_content[i]

						var nameArray [12]byte
						copy(nameArray[:], Name)
						//contenido.B_name = nameArray
						copy(contenido.B_name[:], nameArray[:])

						err = FolderBlock.Serialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
						if err != nil {
							return err
						}
						break
					}
				}
			}
		}
	}
	return nil
}