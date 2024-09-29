package commands

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	structures "server/structures"
	utils "server/util"
	"strings"
	"time"
)

// ReportMBR genera un reporte del MBR y lo guarda en la ruta especificada
func ReportMBR(mbr *structures.MBR, path string, diskPath string) error {

	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	// Obtener el nombre base del archivo sin la extensión
	dotFileName, outputImage := utils.GetFileNames(path)

	// Definir el contenido DOT con una tabla
	dotContent := fmt.Sprintf(`digraph G {
        node [shape=plaintext]
        tabla [label=<
            <table border="0" cellborder="1" cellspacing="0">
                <tr><td colspan="2" bgcolor="palegreen3"> REPORTE MBR </td></tr>
                <tr><td bgcolor ="palegreen">mbr_tamano</td><td>%d</td></tr>
                <tr><td bgcolor ="palegreen">mrb_fecha_creacion</td><td>%s</td></tr>
                <tr><td bgcolor ="palegreen">mbr_disk_signature</td><td>%d</td></tr>
            `, mbr.Mbr_size, time.Unix(int64(mbr.Mbr_creation_date), 0), mbr.Mbr_disk_signature)

	// Agregar las particiones a la tabla
	for i, part := range mbr.Mbr_partitions {

		// Convertir Part_name a string y eliminar los caracteres nulos
		partName := strings.TrimRight(string(part.Part_name[:]), "\x00")
		// Convertir Part_status, Part_type y Part_fit a char
		partStatus := rune(part.Part_status[0])
		partType := rune(part.Part_type[0])
		partFit := rune(part.Part_fit[0])

		// Agregar la partición a la tabla
		dotContent += fmt.Sprintf(`
				<tr><td colspan="2" bgcolor = "lightblue"> PARTICIÓN %d </td></tr>
				<tr><td bgcolor = "lightblue1">part_status</td><td bgcolor = "lightcyan">%c</td></tr>
				<tr><td bgcolor = "lightblue1">part_type</td><td bgcolor = "lightcyan">%c</td></tr>
				<tr><td bgcolor = "lightblue1">part_fit</td><td bgcolor = "lightcyan">%c</td></tr>
				<tr><td bgcolor = "lightblue1">part_start</td><td bgcolor = "lightcyan">%d</td></tr>
				<tr><td bgcolor = "lightblue1">part_size</td><td bgcolor = "lightcyan">%d</td></tr>
				<tr><td bgcolor = "lightblue1">part_name</td><td bgcolor = "lightcyan">%s</td></tr>
			`, i+1, partStatus, partType, partFit, part.Part_start, part.Part_size, partName)

		// Si la partición es extendida, buscar las particiones lógicas (EBRs)
		if partType == 'E' {
			// Moverme al inicio de la partición extendida
			file, err := os.Open(diskPath)
			if err != nil {
				return fmt.Errorf("error al abrir el archivo del disco: %v", err)
			}
			defer file.Close()

			// Leer el primer EBR en la partición extendida
			var ebr EBR
			_, err = file.Seek(int64(part.Part_start), 0)
			if err != nil {
				return fmt.Errorf("error al moverse al inicio de la partición extendida: %v", err)
			}
			err = binary.Read(file, binary.LittleEndian, &ebr)
			if err != nil {
				return fmt.Errorf("error al leer el primer EBR: %v", err)
			}

			// Si el EBR no es válido, entonces no hay particiones lógicas
			if ebr.Part_size == 0 {
				continue
			}

			// Recorrer todos los EBRs
			for {
				// Convertir el nombre y otros campos del EBR
				ebrName := strings.TrimRight(string(ebr.Part_name[:]), "\x00")
				ebrStatus := rune(ebr.Part_mount[0])
				ebrFit := rune(ebr.Part_fit[0])

				// Agregar el EBR (partición lógica) a la tabla
				dotContent += fmt.Sprintf(`
					<tr><td colspan="2" bgcolor="khaki"> PARTICIÓN LÓGICA</td></tr>
					<tr><td bgcolor="khaki1">ebr_status</td><td bgcolor="lemonchiffon">%c</td></tr>
					<tr><td bgcolor="khaki1">ebr_fit</td><td bgcolor="lemonchiffon">%c</td></tr>
					<tr><td bgcolor="khaki1">ebr_start</td><td bgcolor="lemonchiffon">%d</td></tr>
					<tr><td bgcolor="khaki1">ebr_size</td><td bgcolor="lemonchiffon">%d</td></tr>
					<tr><td bgcolor="khaki1">ebr_name</td><td bgcolor="lemonchiffon">%s</td></tr>
				`, ebrStatus, ebrFit, ebr.Part_start, ebr.Part_size, ebrName)

				// Si no hay más particiones lógicas (Part_next == -1), salir del ciclo
				if ebr.Part_next == -1 {
					break
				}

				// Moverme al siguiente EBR
				_, err = file.Seek(int64(ebr.Part_next), 0)
				if err != nil {
					return fmt.Errorf("error al moverse al siguiente EBR: %v", err)
				}
				err = binary.Read(file, binary.LittleEndian, &ebr)
				if err != nil {
					return fmt.Errorf("error al leer el siguiente EBR: %v", err)
				}
			}
		}
	}

	// Cerrar la tabla y el contenido DOT
	dotContent += "</table>>] }"

	// Guardar el contenido DOT en un archivo
	file, err := os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(dotContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo: %v", err)
	}

	// Ejecutar el comando Graphviz para generar la imagen
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar el comando Graphviz: %v", err)
	}

	//fmt.Println("Imagen de la tabla generada:", outputImage)
	return nil
}
