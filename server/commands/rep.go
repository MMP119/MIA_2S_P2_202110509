package commands

import (
	"errors"
	"fmt"
	"regexp"
	global "server/global"
	"strings"
	//"/reports"

)

type REP struct {
	name string
	Path string 
	id string
	path_file_ls string
}

func ParseRep(tokens []string)(*REP, string, error){
	
	cmd := &REP{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`(?i)-id=[^\s]+|(?i)-path="[^"]+"|(?i)-path=[^\s]+|(?i)-name=[^\s]+|(?i)-path_file_ls="[^"]+"|(?i)-path_file_ls=[^\s]+`)

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

			case "-name":
				validNames := []string{"mbr", "disk", "inode", "block", "bm_inode", "bm_block", "sb", "file", "ls"}
				if !contains(validNames, value) {
					return nil, "nombre inválido, debe ser uno de los siguientes: mbr, disk, inode, block, bm_inode, bm_block, sb, file, ls", errors.New("nombre inválido, debe ser uno de los siguientes: mbr, disk, inode, block, bm_inode, bm_block, sb, file, ls")
				}
				cmd.name = value


			case "-path":
				if value == "" {
					return nil, "ERROR: el path es obligatorio", errors.New("el path es obligatorio")
				}
				cmd.Path = value

			
			case "-id":
				if value == "" {
					return nil, "ERROR: el id es obligatorio", errors.New("el id es obligatorio")
				}
				cmd.id = value

			case "-path_file_ls": //este es opcional
				cmd.path_file_ls = value


			default: 
				return nil, "ERROR: parámetro no reconocido", fmt.Errorf("parámetro no reconocido: %s", key)
		}
	}

	if cmd.id == "" || cmd.Path == "" || cmd.name == "" {
		return nil, "faltan parámetros requeridos: -id, -path, -name", errors.New("faltan parámetros requeridos: -id, -path, -name")
	}

	msg, err := CommandRep(cmd)
	if err != nil {
		return nil, msg, err
	}

	
	return cmd, "Comando REP: Reporte realizado correctamente", nil
}


func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}


func CommandRep(rep *REP) (string, error) {

	mountedMbr, mountedSb, mountedDiskPath, err := global.GetMountedPartitionRep(rep.id)
	if err != nil {
		return "", err
	}

	switch rep.name{

	case "mbr":
		err = ReportMBR(mountedMbr, rep.Path, mountedDiskPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	case "disk":
		err = ReportDisk(mountedMbr, rep.Path, mountedDiskPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	case "inode":
		err = ReportInode(mountedSb, mountedDiskPath, rep.Path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "bm_inode":
		err = ReportBMInode(mountedSb, mountedDiskPath, rep.Path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "sb":
		err = ReportSB(mountedSb, rep.Path, mountedDiskPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "block":
		err = ReportBlock(mountedSb, mountedDiskPath, rep.Path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "file":
		err = ReportFile(mountedSb, rep.Path, rep.path_file_ls)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	}

	return "", nil
}