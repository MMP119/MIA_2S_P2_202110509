package reports

import (
	"fmt"
	"os"
	"os/exec"
	structures "server/structures"
	utils "server/util"
	"strings"
)

// ReportBlock genera un reporte de los bloques y los guarda en la ruta especificada
func ReportBlock(sb *structures.SuperBlock, diskPath string, path string) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	// Obtener el nombre base del archivo sin la extensión
	dotFileName, outputImage := utils.GetFileNames(path)

	// Iniciar el contenido DOT
	dotContent := `digraph G {
		node [shape=plaintext]
	`

	var prevBlockIndex int32 = -1 // Variable para enlazar los bloques

	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &structures.Inode{}
		err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return err
		}

		// Iterar sobre cada bloque del inodo
		for _, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				break
			}

			// Verificar si es un bloque de carpeta
			if inode.I_type[0] == '0' {
				block := &structures.FolderBlock{}
				err := block.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return err
				}

				bContentStr := ""
				for _, content := range block.B_content {
					// Eliminar los caracteres nulos de B_name
					contentName := strings.Trim(string(content.B_name[:]), "\x00")
					bContentStr += fmt.Sprintf("%s (Inodo: %d) ", contentName, content.B_inodo)
				}

				// Crear el bloque en Graphviz
				dotContent += fmt.Sprintf(`block%d [label=<
					<table border="0" cellborder="1" cellspacing="0">
						<tr><td colspan="2" bgcolor = "lightblue"> BLOQUE %d </td></tr>
						<tr><td bgcolor = "lightblue1">b_content</td><td>%s</td></tr>
					</table>>];`, blockIndex, blockIndex, bContentStr)

			// Verificar si es un bloque de archivo
			} else if inode.I_type[0] == '1' {
				block := &structures.FileBlock{}
				err := block.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return err
				}

				dotContent += fmt.Sprintf(`block%d [label=<
					<table border="0" cellborder="1" cellspacing="0">
						<tr><td colspan="2" bgcolor = "lightblue"> BLOQUE %d </td></tr>
						<tr><td bgcolor = "lightblue1">b_content</td><td>%s</td></tr>
					</table>>];`, blockIndex, blockIndex, strings.Trim(string(block.B_content[:]), "\x00"))
			}

			// Enlazar los bloques con flechas
			if prevBlockIndex != -1 {
				dotContent += fmt.Sprintf("block%d -> block%d;\n", prevBlockIndex, blockIndex)
			}
			prevBlockIndex = blockIndex
		}
	}

	dotContent += `}`

	// Crear el archivo DOT
	dotFile, err := os.Create(dotFileName)
	if err != nil {
		return err
	}
	defer dotFile.Close()

	_, err = dotFile.WriteString(dotContent)
	if err != nil {
		return err
	}

	// Generar la imagen
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
