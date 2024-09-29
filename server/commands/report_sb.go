package commands

import (
	"fmt"
	"os"
	"os/exec"
	structures "server/structures"
	utils "server/util"
	"time"
)


func ReportSB(sb *structures.SuperBlock, path string, diskPath string) error {
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
				<tr><td colspan="2" bgcolor="palegreen3"> REPORTE SUPERBLOQUE </td></tr>
				<tr><td bgcolor ="palegreen">Filesystem Type</td><td>%d</td></tr>
				<tr><td bgcolor ="palegreen">Inodes Count</td><td>%d</td></tr>
				<tr><td bgcolor ="palegreen">Blocks Count</td><td>%d</td></tr>
				<tr><td bgcolor ="palegreen">Free Inodes Count</td><td>%d</td></tr>
				<tr><td bgcolor ="palegreen">Free Blocks Count</td><td>%d</td></tr>
				<tr><td bgcolor ="palegreen">Mount Time</td><td>%s</td></tr>
				<tr><td bgcolor ="palegreen">Unmount Time</td><td>%s</td></tr>
				<tr><td bgcolor ="palegreen">Mount Count</td><td>%d</td></tr>
				<tr><td bgcolor ="palegreen">Magic</td><td>%d</td></tr>
				<tr><td bgcolor ="palegreen">Inode Size</td><td>%d</td></tr>
				<tr><td bgcolor ="palegreen">Block Size</td><td>%d</td></tr>
				<tr><td bgcolor ="palegreen">First Inode</td><td>%d</td></tr>
				<tr><td bgcolor ="palegreen">First Block</td><td>%d</td></tr>
				<tr><td bgcolor ="palegreen">Bitmap Inode Start</td><td>%d</td></tr>
				<tr><td bgcolor ="palegreen">Bitmap Block Start</td><td>%d</td></tr>
				<tr><td bgcolor ="palegreen">Inode Start</td><td>%d</td></tr>
				<tr><td bgcolor ="palegreen">Block Start</td><td>%d</td></tr>
			</table>
		>]}`, sb.S_filesystem_type, sb.S_inodes_count, sb.S_blocks_count, sb.S_free_inodes_count, sb.S_free_blocks_count, time.Unix(int64(sb.S_mtime), 0).Format(time.RFC3339), time.Unix(int64(sb.S_umtime), 0).Format(time.RFC3339), sb.S_mnt_count, sb.S_magic, sb.S_inode_size, sb.S_block_size, sb.S_first_ino, sb.S_first_blo, sb.S_bm_inode_start, sb.S_bm_block_start, sb.S_inode_start, sb.S_block_start)

	// Crear el archivo DOT
	file, err := os.Create(dotFileName)
	if err != nil {
		return err
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

	return nil
}