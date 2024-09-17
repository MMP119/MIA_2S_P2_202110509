package commands

import (
	"errors"
	"fmt"
	"regexp"
	global "server/global"
	"strings"
	reports "server/reports"

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
	// Crear una nueva estructura MBR
	// mbr := &structures.MBR{}

	// // Deserializar la estructura MBR desde el archivo binario
	// msg, err := mbr.DeserializeMBR(cmd.Path)
	// if err != nil {
	// 	return msg, err
	// }

	// // Imprimir la información del MBR
	// fmt.Println("\nMBR\n----------------")
	// mbr.Print()

	// // Imprimir la información de cada partición
	// fmt.Println("\nParticiones\n----------------")
	// mbr.PrintPartitions()

	// // Imprimir partidas montadas
	// fmt.Println("\nParticiones montadas\n----------------")

	// for id, path := range global.MountedPartitions {
	// 	fmt.Printf("ID: %s, PATH: %s\n", id, path)
	// }

	// // Imprimir el SuperBloque de cada partición montada
	// index := 0
	// // Iterar sobre cada partición montada
	// for id, path := range global.MountedPartitions {
	// 	// Crear una nueva estructura SuperBloque
	// 	sb := &structures.SuperBlock{}
	// 	// Deserializar la estructura SuperBloque desde el archivo binario
	// 	err := sb.Deserialize(path, int64(mbr.Mbr_partitions[index].Part_start))
	// 	if err != nil {
	// 		fmt.Printf("Error al leer el SuperBloque de la partición %s: %s\n", id, err)
	// 		continue
	// 	}
	// 	fmt.Printf("\nPartición %s\n----------------", id)

	// 	// Imprimir la información del SuperBloque
	// 	fmt.Println("\nSuperBloque:")
	// 	sb.Print()

	// 	// Imprimir los inodos
	// 	sb.PrintInodes(path)

	// 	// Imprimir los bloques
	// 	sb.PrintBlocks(path)

	// 	index++
	// }

	mountedMbr, mountedSb, mountedDiskPath, err := global.GetMountedPartitionRep(rep.id)
	if err != nil {
		return "", err
	}

	switch rep.name{

	case "mbr":
		err = reports.ReportMBR(mountedMbr, rep.Path, mountedDiskPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	case "disk":
		err = reports.ReportDisk(mountedMbr, rep.Path, mountedDiskPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	case "inode":
		err = reports.ReportInode(mountedSb, mountedDiskPath, rep.Path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "bm_inode":
		err = reports.ReportBMInode(mountedSb, mountedDiskPath, rep.Path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "sb":
		err = reports.ReportSB(mountedSb, rep.Path, mountedDiskPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "block":
		err = reports.ReportBlock(mountedSb, mountedDiskPath, rep.Path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "file":
		err = reports.ReportFile(mountedSb, rep.Path, rep.path_file_ls)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	}

	return "", nil
}