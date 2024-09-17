package reports

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	structures "server/structures"
	utils "server/util"
	"strings"
)

func ReportDisk(mbr *structures.MBR, path string, diskPath string) error {
    err := utils.CreateParentDirs(path)
    if err != nil {
        return err
    }

    dotFileName, outputImage := utils.GetFileNames(path)

    const mbrSize = 153.0

    totalSize := float64(mbr.Mbr_size)
    usableSize := totalSize - mbrSize

    dotContent := `digraph G {
        labelloc="t";
        label = "Reporte de Disco";
        node [shape=plaintext];
        
        tabla [label=<
        <table border="1" cellborder="1" cellspacing="0" cellpadding="10" bgcolor="#F9F9F9">
        <tr>
            <td rowspan="2" bgcolor="#007ACC" border="1" color="white"><b>MBR</b></td>`

    var partitionRows string
    var logicalRows string
    extendedPartitionFound := false
    totalUsedSpace := mbrSize
    freeSpaceOutsidePartitions := usableSize
    var remainingInExtended float64

    // Recorrer las particiones primarias y extendidas
    for _, part := range mbr.Mbr_partitions {
        if part.Part_size == 0 {
            continue
        }

        partName := strings.TrimRight(string(part.Part_name[:]), "\x00")
        partType := rune(part.Part_type[0])
        partSize := float64(part.Part_size)
        partPercentage := (partSize / totalSize) * 100

        totalUsedSpace += partSize
        freeSpaceOutsidePartitions -= partSize

        if partType == 'E' { // Partición extendida
            extendedPartitionFound = true
            partitionRows += fmt.Sprintf(`
            <td colspan="6" bgcolor="#6F42C1" border="1" color="white"><b>Extendida<br/>%.2f%% del Disco</b></td>`, partPercentage)

            file, err := os.Open(diskPath)
            if err != nil {
                return fmt.Errorf("error al abrir el archivo del disco: %v", err)
            }
            defer file.Close()

            var ebr structures.EBR
            _, err = file.Seek(int64(part.Part_start), 0)
            if err != nil {
                return fmt.Errorf("error al moverse al inicio de la partición extendida: %v", err)
            }
            err = binary.Read(file, binary.LittleEndian, &ebr)
            if err != nil {
                return fmt.Errorf("error al leer el primer EBR: %v", err)
            }

            // Recorrer todas las particiones lógicas dentro de la partición extendida
            for {
                ebrName := strings.TrimRight(string(ebr.Part_name[:]), "\x00")
                ebrSize := float64(ebr.Part_size)
                ebrPercentage := (ebrSize / totalSize) * 100

                logicalRows += fmt.Sprintf(`
                <td bgcolor="#FFC107" border="1" color="black">EBR</td>
                <td bgcolor="#20C997" border="1" color="black">Lógica<br/>%s<br/>%.2f%% del Disco</td>`, ebrName, ebrPercentage)

                totalUsedSpace += ebrSize + ebrSize // Considerar EBR y su partición lógica

                if ebr.Part_next == -1 {
                    // Calcular espacio libre dentro de la partición extendida y agregarlo aquí, después de la última partición lógica
                    remainingInExtended = float64(part.Part_size) - float64(ebr.Part_start - part.Part_start) - ebrSize
                    if remainingInExtended > 0 {
                        logicalRows += fmt.Sprintf(`
                        <td bgcolor="#E0E0E0" border="1" color="black">Espacio Libre en Extendida<br/>%.2f%% del Disco</td>`, (remainingInExtended / totalSize) * 100)
                    }
                    break
                }

                _, err = file.Seek(int64(ebr.Part_next), 0)
                if err != nil {
                    return fmt.Errorf("error al moverse al siguiente EBR: %v", err)
                }
                err = binary.Read(file, binary.LittleEndian, &ebr)
                if err != nil {
                    return fmt.Errorf("error al leer el siguiente EBR: %v", err)
                }
            }
        } else {
            if(string(partType) != "0"){
                partitionRows += fmt.Sprintf(`
                <td rowspan="2" bgcolor="#FF5733" border="1" color="white">%s<br/>%.2f%% del Disco</td>`, partName, partPercentage)
            }
            // partitionRows += fmt.Sprintf(`
            // <td rowspan="2" bgcolor="#FF5733" border="1" color="white">%s<br/>%.2f%% del Disco</td>`, partName, partPercentage)
        }
    }

    // Espacio libre fuera de las particiones
    if freeSpaceOutsidePartitions > 0 {
        freeSpaceOutsidePercentage := (freeSpaceOutsidePartitions / totalSize) * 100
        partitionRows += fmt.Sprintf(`
        <td rowspan="2" bgcolor="#E0E0E0" border="1" color="black">Espacio Libre<br/>%.2f%% del Disco</td>`, freeSpaceOutsidePercentage)
    }

    if !extendedPartitionFound {
		dotContent += partitionRows + `</tr></table>>]; }`
    }else{
		dotContent += partitionRows + `</tr><tr>` + logicalRows + `</tr></table>>]; }`
	}


    //dotContent += partitionRows + `</tr><tr>` + logicalRows + `</tr></table>>]; }`

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
