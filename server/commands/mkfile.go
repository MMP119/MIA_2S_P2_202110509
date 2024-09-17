package commands


import (
	global "server/global"
	structures "server/structures"
	utils "server/util"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"io/ioutil"
)


type MKFILE struct {
	Path string // Ruta del archivo
	r    bool   // Opción recursiva
	Size int    // Tamaño del archivo
	Cont string // Contenido del archivo
}



func ParseMkfile(tokens []string) (*MKFILE, string, error) {

	cmd := &MKFILE{} // Crea una nueva instancia de MKFILE

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando mkfile
	re := regexp.MustCompile(`(?i)-path="[^"]+"|(?i)-path=[^\s]+|(?i)-r|-size=\d+|(?i)-cont="[^"]+"|(?i)-cont=[^\s]+`)
	// Encuentra todas las coincidencias de la expresión regular en la cadena de argumentos
	matches := re.FindAllString(args, -1)

	// Verificar que todos los tokens fueron reconocidos por la expresión regular
	if len(matches) != len(tokens) {
		// Identificar el parámetro inválido
		for _, token := range tokens {
			if !re.MatchString(token) {
				return nil, "", fmt.Errorf("parámetro inválido: %s", token)
			}
		}
	}

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		key := strings.ToLower(kv[0])
		var value string
		if len(kv) == 2 {
			value = kv[1]
		}

		// Remove quotes from value if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar diferentes parámetros
		switch key {
		case "-path":
			// Verifica que el path no esté vacío
			if value == "" {
				return nil,"", errors.New("el path no puede estar vacío")
			}
			cmd.Path = value
		case "-r":
			// Establece el valor de r a true
			cmd.r = true
		case "-size":
			// Convierte el valor del tamaño a un entero
			size, err := strconv.Atoi(value)
			if err != nil || size < 0 {
				return nil,"", errors.New("el tamaño debe ser un número entero no negativo")
			}
			cmd.Size = size
		case "-cont":
			// Verifica que el contenido no esté vacío
			if value == "" {
				return nil,"", errors.New("el contenido no puede estar vacío")
			}
			cmd.Cont = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return nil,"", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que el parámetro -path haya sido proporcionado
	if cmd.Path == "" {
		return nil,"", errors.New("faltan parámetros requeridos: -path")
	}

	// Si no se proporcionó el tamaño, se establece por defecto a 0
	if cmd.Size == 0 {
		cmd.Size = 0
	}

	// Si no se proporcionó el contenido, se establece por defecto a ""
	if cmd.Cont == "" {
		cmd.Cont = ""
	}

	// Crear el archivo con los parámetros proporcionados
	msg,err := commandMkfile(cmd)
	if err != nil {
		return nil,msg, err
	}

	salida := fmt.Sprintf("Comando MKFILE: realizado correctamente, archivo creado en %s", cmd.Path)
	return cmd, salida, nil
}

func commandMkfile(mkfile *MKFILE) (string, error) {

	idPartition := global.GetIDSession()
	if idPartition == "" {
		return "ERROR: No hay ninguna sesión activa", errors.New("mkdir: no hay ninguna sesión activa")
	}

	partitionSuperblock, mountedPartition, partitionPath, err := global.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return "ERROR: al obtener el sb de la particion montada",fmt.Errorf("al obtener el sb de la particion montada: %w", err)
	}

	// Generar el contenido del archivo si no se proporcionó
	if mkfile.Cont == "" {
		mkfile.Cont = generateContent(mkfile.Size)
	}

	// if mkfile.Cont != "" && mkfile.Size != 0 {
	// 	return "ERROR: no se puede proporcionar contenido y tamaño al mismo tiempo", errors.New("no se puede proporcionar contenido y tamaño al mismo tiempo")
	// }

	if mkfile.Cont != "" && mkfile.Size == 0 {

		//leer un archivo de mi pc, obtener el contenido y ponerlo en mkfile.Cont
		content, err := ioutil.ReadFile(mkfile.Cont)
		if err != nil {
			return "ERROR: al leer el archivo",fmt.Errorf("al leer el archivo: %w", err)
		}

		//asignar el contenido del archivo a mkfile.Cont
		mkfile.Cont = string(content)

		//asignar el tamaño del archivo
		mkfile.Size = len(mkfile.Cont)
	}

	fmt.Println("\nContenido del archivo:", mkfile.Cont)

	// Crear el archivo
	err = createFile(mkfile.Path, mkfile.Size, mkfile.Cont, partitionSuperblock, partitionPath, mountedPartition)
	if err != nil {
		err = fmt.Errorf("error al crear el archivo: %w", err)
	}

	return "",err

}


// generateContent genera una cadena de números del 0 al 9 hasta cumplir el tamaño ingresado
func generateContent(size int) string {
	content := ""
	for len(content) < size {
		content += "0123456789"
	}
	return content[:size] // Recorta la cadena al tamaño exacto
}

// Funcion para crear un archivo
func createFile(filePath string, size int, content string, sb *structures.SuperBlock, partitionPath string, mountedPartition *structures.PARTITION) error {
	fmt.Println("\nCreando archivo:", filePath)

	parentDirs, destDir := utils.GetParentDirectories(filePath)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("Directorio destino:", destDir)

	// Obtener contenido por chunks
	chunks := utils.SplitStringIntoChunks(content)
	fmt.Println("\nChunks del contenido:", chunks)

	// Crear el archivo
	err := sb.CreateFile(partitionPath, parentDirs, destDir, size, chunks)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %w", err)
	}

	// Imprimir inodos y bloques
	sb.PrintInodes(partitionPath)
	sb.PrintBlocks(partitionPath)

	// Serializar el superbloque
	err = sb.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}