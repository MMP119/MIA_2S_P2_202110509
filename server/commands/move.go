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

type MOVE struct{
	Path string
	Destino string
}

func ParseMove(tokens []string) (*MOVE, string, error) {

	cmd := &MOVE{}

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

	msg, err := CommandMove(cmd)
	if err != nil {
		return nil, msg, err
	}


	return cmd, msg, nil
}

func CommandMove(cmd *MOVE)(string, error){

	parenDirsPath, destiDirPath := utils.GetParentDirectories(cmd.Path)

	parenDirsDestino, destiDirDestino := utils.GetParentDirectories(cmd.Destino)

	idParticion := global.GetIDSession()

	partitionSuperblock, _, partitionPath, err := global.GetMountedPartitionSuperblock(idParticion)
	if err != nil {
		return "Error al obtener la partición montada en el comando login", fmt.Errorf("error al obtener la partición montada: %v", err)
	}

	textPath := false
	textDestino := false
	CarpetaPath := false

	//verificar si voy a copiar una carpeta o un archivo
	if strings.Contains(destiDirPath, ".txt"){
		textPath = true
	}else{
		CarpetaPath = true
	}
	
	//verficar si el destino es una carpeta o un archivo
	if strings.Contains(destiDirDestino, ".txt"){
		textDestino = true
	}

	if textPath && textDestino {
		return "ERROR: no se puede mover un archivo a otro archivo, el destino debe ser una carpeta", errors.New("no se puede mover un archivo a otro archivo")
	}

	if CarpetaPath && textDestino {
		return "ERROR: no se puede mover una carpeta a un archivo", errors.New("no se puede mover una carpeta a un archivo")
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

					for i:=0; i<len(parenDirsPath); i++{

						if strings.Trim(string(content.B_name[:]), "\x00") == parenDirsPath[i] {
							//recorro hasta llegar a la carpeta destino
							err := moveRecursiva(inode, partitionSuperblock, partitionPath, parenDirsPath, destiDirPath, parenDirsDestino, destiDirDestino)
							if err != nil {
								return "Error: no se pudo renombrar la carpeta", err
							}
							return "Comando MOVE: realizado correctamente", nil
						}
					}

					if strings.Trim(string(content.B_name[:]), "\x00") == destiDirPath {

						contenido := &FolderBlock.B_content[i]
						
						//cambiar las referencias del inodo path, dejarlo en -1 y mover esa referencia al inodo destino	
						refInodo := contenido.B_inodo //el * es para obtener el valor de la referencia
						contenido.B_inodo = -1 //cambiar la referencia del inodo path a -1
						
						//irse a la funcion que recorre los inodos y cambia las referencias del inodo destino
						inode1 := &structures.Inode{}
						err = inode1.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(0*partitionSuperblock.S_inode_size)))
						if err != nil {
							return "no se pudo deserializar el inodo", err
						}
						err = MoverDestino(inode1, partitionSuperblock, partitionPath, parenDirsDestino, destiDirDestino, refInodo)
						if err != nil {
							return "Error: no se pudo mover la carpeta", err
						}

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

	return "Comando MOVE: realizado correctamente", nil
}


func moveRecursiva(inode *structures.Inode, partitionSuperblock *structures.SuperBlock, partitionPath string, parentDirs []string, destDir string, parentDirsDestino []string, destiDirDestino string)(error){

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
							moveRecursiva(inode, partitionSuperblock, partitionPath, parentDirs, destDir, parentDirsDestino, destiDirDestino)
							return nil
						}
					}
					if strings.Trim(string(content.B_name[:]), "\x00") == destDir {

						contenido := &FolderBlock.B_content[i]
						
						//cambiar las referencias del inodo path, dejarlo en -1 y mover esa referencia al inodo destino	
						refInodo := contenido.B_inodo //el * es para obtener el valor de la referencia
						contenido.B_inodo = -1 //cambiar la referencia del inodo path a -1
						
						//irse a la funcion que recorre los inodos y cambia las referencias del inodo destino
						
						inode1 := &structures.Inode{}
						err = inode1.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(0*partitionSuperblock.S_inode_size)))
						if err != nil {
							return err
						}
						
						err = MoverDestino(inode1, partitionSuperblock, partitionPath, parentDirsDestino, destiDirDestino, refInodo)
						if err != nil {
							return err
						}

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


func MoverDestino(inode *structures.Inode, partitionSuperblock *structures.SuperBlock, partitionPath string, parentDirs []string, destDir string, refCambiar int32)(error){


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
							err = MoverDestino(inode, partitionSuperblock, partitionPath, parentDirs, destDir, refCambiar)
							if err != nil {
								return err
							}
						}
					}
					if strings.Trim(string(content.B_name[:]), "\x00") == destDir {

						contenido := &FolderBlock.B_content[i]
						
						//cambiar las referencias del inodo path, dejarlo en -1 y mover esa referencia al inodo destino	
						//refInodo := contenido.B_inodo //el * es para obtener el valor de la referencia
						contenido.B_inodo = refCambiar //cambiar la referencia del inodo path a -1
						
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