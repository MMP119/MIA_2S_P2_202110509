package commands


import(
	global "server/global"
	structures "server/structures"
	utils "server/util"
	"errors"
	"fmt"
	"regexp"
	"strings"
)


type MKDIR struct{
	Path string
	P bool			// -p para crear los directorios que no existe
}


func ParseMkdir(tokens []string) (*MKDIR,string, error) {
	
	cmd := &MKDIR{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`(?i)-path="[^"]+"|(?i)-path=[^\s]+|-p`)

	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return nil, "ERROR: formato de parámetro inválido", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
			case "-path":
				if value == "" {
					return nil, "ERROR: Path no especificado en MKDIR", errors.New("path no especificado")
				}
				cmd.Path = value

			case "-p":
				cmd.P = true
			
			default:
				return nil, "ERROR: parámetro desconocido en MKDIR", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.Path == "" {
		return nil, "ERROR: Path no especificado en MKDIR", errors.New("path no especificado")
	}

	
	msg, err:= CommandMkdir(cmd)
	if err != nil {
		return nil, msg, err
	}

	salida := fmt.Sprintf("Comando MKDIR: realizado correctamente, directorio %s creado", cmd.Path)
	return cmd,salida, nil
}


func CommandMkdir(mkdir *MKDIR)(string, error){

	//verificar si existe una sesión activa y si es así, obtener el id enlazado
	idPartition := global.GetIDSession()
	if idPartition == "" {
		return "ERROR: No hay ninguna sesión activa", errors.New("mkdir: no hay ninguna sesión activa")
	}

	partitionSuperblock, mountedPartition, partitionPath, err := global.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return "ERROR: al obtener el sb de la particion montada",fmt.Errorf("al obtener el sb de la particion montada: %w", err)
	}

	//se crea el directorio
	msg := ""
	msg, err = createDirectory(mkdir.Path, partitionSuperblock, partitionPath, mountedPartition)
	if err != nil {
		return msg, fmt.Errorf("al crear el directorio: %w", err)
	}

	return "", nil
}

func createDirectory(dirPath string, sb *structures.SuperBlock, partitionPath string, mountedPartition *structures.PARTITION)(string, error){
	
	parentDirs, destDir := utils.GetParentDirectories(dirPath)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("Directorio destino:", destDir)

	//crear directorio segun el path
	err := sb.CreateFolder(partitionPath, parentDirs, destDir)
	if err != nil {
		return "ERROR: al crear el directorio", fmt.Errorf("al crear el directorio: %w", err)
	}

	// imprimir inodos y bloques
	sb.PrintInodes(partitionPath)
	sb.PrintBlocks(partitionPath)

	err = sb.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return "ERROR: al serializar el superbloque", fmt.Errorf("al serializar el superbloque: %w", err)
	}

	
	return "", nil
}