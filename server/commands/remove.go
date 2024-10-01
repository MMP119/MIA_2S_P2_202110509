package commands

import (
	"errors"
	"fmt"
	"regexp"
	"server/global"
	utils "server/util"
	"strings"
	structures "server/structures"
)


type REMOVE struct{
	Path string
}


func ParseRemove(tokens []string)(*REMOVE, string, error){

	cmd := &REMOVE{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`(?i)-path="[^"]+"|(?i)-path=[^\s]+`)

	matches := re.FindAllString(args, -1)

	for _, math := range matches{

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

			default: 
				return nil, "ERROR: parámetro no reconocido", fmt.Errorf("parámetro no reconocido: %s", key)
		}
	}

	if cmd.Path == "" {
		return nil, "ERROR: el path es obligatorio", errors.New("el path es obligatorio")
	}

	msg, err := CommnadRemove(cmd)
	if err != nil {
		return nil, msg, err
	}

	return cmd, "Comando REMOVE: realizado correctamente", nil
}


func CommnadRemove(cmd *REMOVE)(string, error){

	parentDirs, destDir := utils.GetParentDirectories(cmd.Path)
	
	idParticion := global.GetIDSession()

	
	//verificar si destDir es una carpeta o un archivo
	if(strings.Contains(destDir, ".txt")){
		//es un archivo
	}
	
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
							RemoveRecursiva(inode, partitionSuperblock, partitionPath, parentDirs, destDir)
							return "Comando REMOVE: realizado correctamente", nil
						}
					}

					if strings.Trim(string(content.B_name[:]), "\x00") == destDir {

						contenido := &FolderBlock.B_content[i]

						//eliminar la carpeta con todos sus archivos y subcarpetas
						contenido.B_inodo = -1
						contenido.B_name = [12]byte{'-'}

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

	return "Comando REMOVE: realizado correctamente", nil
}


func RemoveRecursiva(inode *structures.Inode, partitionSuperblock *structures.SuperBlock, partitionPath string, parentDirs []string, destDir string)(error){

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
							RemoveRecursiva(inode, partitionSuperblock, partitionPath, parentDirs, destDir)
							return nil
						}
					}
					if strings.Trim(string(content.B_name[:]), "\x00") == destDir {

						contenido := &FolderBlock.B_content[i]

						//eliminar la carpeta con todos sus archivos y subcarpetas
						contenido.B_inodo = -1
						contenido.B_name = [12]byte{'-'}

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