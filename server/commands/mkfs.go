package commands

import(
	global "server/global"
	structures "server/structures"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

type MKFS struct {
	id  string // ID del disco
	typ string // Tipo de formato (full)
	fs string // Tipo de sistema de archivos (ext2, ext3)
}

func ParserMkfs(tokens []string) (*MKFS, string,error) {
	cmd := &MKFS{} // Crea una nueva instancia de MKFS

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando mkfs
	re := regexp.MustCompile(`(?i)-id=[^\s]+|(?i)-type=[^\s]+|-fs=[23]fs`)
	// Encuentra todas las coincidencias de la expresión regular en la cadena de argumentos
	matches := re.FindAllString(args, -1)

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return nil, "ERROR: formato de parámetro inválido" ,fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Remove quotes from value if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar diferentes parámetros
		switch key {
		case "-id":
			// Verifica que el id no esté vacío
			if value == "" {
				return nil, "ERROR: id vacío en mkfs" ,errors.New("el id no puede estar vacío")
			}
			cmd.id = value
		case "-type":
			// Verifica que el tipo sea "full"
			if value != "full" {
				return nil, "ERROR: el tipo debe ser full" ,errors.New("el tipo debe ser full")
			}
			cmd.typ = value
		case "-fs":
			// Verifica que el tipo de sistema de archivos sea ext2 o ext3
			if value != "2fs" && value != "3fs" {
				return nil, "ERROR: tipo de sistema de archivos no válido" ,errors.New("tipo de sistema de archivos no válido")
			}
			cmd.fs = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return nil, "ERROR: Parámetro desconocido" ,fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que el parámetro -id haya sido proporcionado
	if cmd.id == "" {
		return nil, "ERROR: faltan parámetros requeridos: -id" ,errors.New("faltan parámetros requeridos: -id")
	}

	// Si no se proporcionó el tipo, se establece por defecto a "full"
	if cmd.typ == "" {
		cmd.typ = "full"
	}

	if cmd.fs == "" {
		cmd.fs = "2fs"
	}

	// Aquí se puede agregar la lógica para ejecutar el comando mkfs con los parámetros proporcionados
	err := CommandMkfs(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return cmd, "COMANDO MKFS: realizado correctamente" ,nil // Devuelve el comando MKFS creado
}

func CommandMkfs(mkfs *MKFS) error {
	// Obtener la partición montada
	mountedPartition, partitionPath, err := global.GetMountedPartition(mkfs.id)
	if err != nil {
		return err
	}

	// Verificar la partición montada
	// fmt.Println("\nPatición montada:")
	// mountedPartition.Print()

	// Calcular el valor de n
	n := calculateN(mountedPartition, mkfs.fs)

	// Verificar el valor de n
	fmt.Println("\nValor de n:", n)

	// Inicializar un nuevo superbloque
	superBlock := createSuperBlock(mountedPartition, n, mkfs.fs)

	// Verificar el superbloque
	// fmt.Println("\nSuperBlock:")
	// superBlock.Print()

	// Crear los bitmaps
	err = superBlock.CreateBitMaps(partitionPath)
	if err != nil {
		return err
	}


	// validar que sistema de archivos es
	if superBlock.S_filesystem_type == 3{

		// crear el archivo user.txt ext3
		err = superBlock.CreateUsersFileExt3(partitionPath, int64(mountedPartition.Part_start+int32(binary.Size(structures.SuperBlock{}))))
		if err != nil {
			return err
		}

	}else{
		fmt.Println("Entra en ext2")
		// Crear archivo users.txt ext2
		err = superBlock.CreateUsersFile(partitionPath)
		if err != nil {
			return err
		}

	}


	// Crear archivo users.txt
	// err = superBlock.CreateUsersFile(partitionPath)
	// if err != nil {
	// 	return err
	// }

	// Verificar superbloque actualizado
	// fmt.Println("\nSuperBlock actualizado:")
	// superBlock.Print()

	// Serializar el superbloque
	err = superBlock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return err
	}

	return nil
}

/*func calculateN(partition *structures.PARTITION) int32 {
	/*
		numerador = (partition_montada.size - sizeof(Structs::Superblock)
		denrominador base = (4 + sizeof(Structs::Inodes) + 3 * sizeof(Structs::Fileblock))
		n = floor(numerador / denrominador)
	

	numerator := int(partition.Part_size) - binary.Size(structures.SuperBlock{})
	denominator := 4 + binary.Size(structures.Inode{}) + 3*binary.Size(structures.FileBlock{}) 
	n := math.Floor(float64(numerator) / float64(denominator))

	return int32(n)
}*/


func calculateN(partition *structures.PARTITION, fs string) int32 {
	// Numerador: tamaño de la partición menos el tamaño del superblock
	numerator := int(partition.Part_size) - binary.Size(structures.SuperBlock{})

	// Denominador base: 4 + tamaño de inodos + 3 * tamaño de bloques de archivo
	baseDenominator := 4 + binary.Size(structures.Inode{}) + 3*binary.Size(structures.FileBlock{})

	// Si el sistema de archivos es "3fs", se añade el tamaño del journaling al denominador
	temp := 0
	if fs == "3fs" {
		temp = binary.Size(structures.Journal{})
	}

	// Denominador final
	denominator := baseDenominator + temp

	// Calcular n
	n := math.Floor(float64(numerator) / float64(denominator))

	return int32(n)
}



/*func createSuperBlock(partition *structures.PARTITION, n int32) *structures.SuperBlock {
	// Calcular punteros de las estructuras
	// Bitmaps
	bm_inode_start := partition.Part_start + int32(binary.Size(structures.SuperBlock{}))
	bm_block_start := bm_inode_start + n // n indica la cantidad de inodos, solo la cantidad para ser representada en un bitmap
	// Inodos
	inode_start := bm_block_start + (3 * n) // 3*n indica la cantidad de bloques, se multiplica por 3 porque se tienen 3 tipos de bloques
	// Bloques
	block_start := inode_start + (int32(binary.Size(structures.Inode{})) * n) // n indica la cantidad de inodos, solo que aquí indica la cantidad de estructuras Inode

	// Crear un nuevo superbloque
	superBlock := &structures.SuperBlock{
		S_filesystem_type:   2,
		S_inodes_count:      0,
		S_blocks_count:      0,
		S_free_inodes_count: int32(n),
		S_free_blocks_count: int32(n * 3),
		S_mtime:             float32(time.Now().Unix()),
		S_umtime:            float32(time.Now().Unix()),
		S_mnt_count:         1,
		S_magic:             0xEF53,
		S_inode_size:        int32(binary.Size(structures.Inode{})),
		S_block_size:        int32(binary.Size(structures.FileBlock{})),
		S_first_ino:         inode_start,
		S_first_blo:         block_start,
		S_bm_inode_start:    bm_inode_start,
		S_bm_block_start:    bm_block_start,
		S_inode_start:       inode_start,
		S_block_start:       block_start,
	}
	return superBlock
}*/

func createSuperBlock(partition *structures.PARTITION, n int32, fs string) *structures.SuperBlock {
	// Calcular punteros de las estructuras
	journal_start, bm_inode_start, bm_block_start, inode_start, block_start := calculateStartPositions(partition, fs, n)

	fmt.Printf("Journal Start: %d\n", journal_start)
	fmt.Printf("Bitmap Inode Start: %d\n", bm_inode_start)
	fmt.Printf("Bitmap Block Start: %d\n", bm_block_start)
	fmt.Printf("Inode Start: %d\n", inode_start)
	fmt.Printf("Block Start: %d\n", block_start)

	// Tipo de sistema de archivos
	var fsType int32

	if fs == "2fs" {
		fsType = 2
	} else {
		fsType = 3
	}

	// Crear un nuevo superbloque
	superBlock := &structures.SuperBlock{
		S_filesystem_type:   fsType,
		S_inodes_count:      0,
		S_blocks_count:      0,
		S_free_inodes_count: int32(n),
		S_free_blocks_count: int32(n * 3),
		S_mtime:             float32(time.Now().Unix()),
		S_umtime:            float32(time.Now().Unix()),
		S_mnt_count:         1,
		S_magic:             0xEF53,
		S_inode_size:        int32(binary.Size(structures.Inode{})),
		S_block_size:        int32(binary.Size(structures.FileBlock{})),
		S_first_ino:         inode_start,
		S_first_blo:         block_start,
		S_bm_inode_start:    bm_inode_start,
		S_bm_block_start:    bm_block_start,
		S_inode_start:       inode_start,
		S_block_start:       block_start,
	}
	return superBlock
}


func calculateStartPositions(partition *structures.PARTITION, fs string, n int32) (int32, int32, int32, int32, int32) {
	superblockSize := int32(binary.Size(structures.SuperBlock{}))
	journalSize := int32(binary.Size(structures.Journal{}))
	inodeSize := int32(binary.Size(structures.Inode{}))

	// Inicializar posiciones
	// EXT2
	journalStart := int32(0)
	bmInodeStart := partition.Part_start + superblockSize
	bmBlockStart := bmInodeStart + n
	inodeStart := bmBlockStart + (3 * n)
	blockStart := inodeStart + (inodeSize * n)

	// Ajustar para EXT3
	if fs == "3fs" {
		journalStart = partition.Part_start + superblockSize
		bmInodeStart = journalStart + (journalSize * n)
		bmBlockStart = bmInodeStart + n
		inodeStart = bmBlockStart + (3 * n)
		blockStart = inodeStart + (inodeSize * n)
	}

	return journalStart, bmInodeStart, bmBlockStart, inodeStart, blockStart
}