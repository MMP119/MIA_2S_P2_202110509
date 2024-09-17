package commands

import (
	"errors" 
	"fmt"    
	"regexp" 
	"strings" 
	structures "server/structures"
	global "server/global"
	util "server/util"
)

type MOUNT struct {
	Path string 
	Name string 
	List string
}

func ParserMount(tokens []string) (*MOUNT, string, error) {
	cmd := &MOUNT{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`(?i)-path="[^"]+"|(?i)-path=[^\s]+|(?i)-name="[^"]+"|(?i)-name=[^\s]+`)

	matches := re.FindAllString(args, -1)

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {

		// Divide cada parte en clave y valor usando "=" como delimitador
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
					return nil,"ERROR: el path no puede estar vacío", errors.New("el path no puede estar vacío")
				}
				cmd.Path = value
			case "-name":
				if value == "" {
					return nil, "ERROR: el nombre no puede estar vacío", errors.New("el nombre no puede estar vacío")
				}
				cmd.Name = value
			default:
				return nil, "ERROR: parámetro desconocido", fmt.Errorf("parámetro desconocido: %s", key)
			}
	}

	if cmd.Path == "" {
		return nil, "ERROR: faltan parámetros requeridos: -path", errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.Name == "" {
		return nil, "ERROR: faltan parámetros requeridos: -name", errors.New("faltan parámetros requeridos: -name")
	}

	// se monta la partición 
	msg, err := CommandMount(cmd)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, msg, err
	}

	return cmd, "Comando MOUNT: Montaje de particion realizado exitosamente"+msg, nil // Devuelve el comando MOUNT creado
}


func CommandMount(mount *MOUNT) (string, error) {

	var mbr structures.MBR

	msg, err := mbr.DeserializeMBR(mount.Path)
	if err != nil {
		return msg, fmt.Errorf("error leyendo el MBR del disco: %s", err)
	}

	// buscar la particion con el nombre proporcionado
	partition, indexPartition, msg := mbr.GetPartitionByName(mount.Name, mount.Path)
	if partition == nil {
		return msg, fmt.Errorf("no se encontró la partición con el nombre: %s, al montar la particion", mount.Name)
	}

	// verificar si es una partición extendida o lógica, no se puede montar
	if partition.Part_type[0] == 'E' || partition.Part_type[0] == 'L' {
		return "ERROR: no se puede montar una partición extendida o lógica", errors.New("no se puede montar una partición extendida o lógica")
	}

	// verificar si la partición ya está montada
	if partition.Part_status[0] == '1' {
		return "ERROR: la partición %s ya está montada", errors.New("la partición ya está montada")
	}

	// actualizar los correlativos de las particiones
	mbr.UpdatePartitionCorrelatives()

	// Después de actualizar los correlativos, obtener la partición actualizada
	partition = &mbr.Mbr_partitions[indexPartition]

	// generar un id único para la partición usando el correlativo actualizado
	id, msg, err := GenerateIdPartition(mount, int(partition.Part_correlative))
	if err != nil {
		return msg, fmt.Errorf("error generando id de partición: %s", err)
	}

	// guardar la particion montada en la lista de montajes globales
	global.MountedPartitions[id] = mount.Path

	// modificar la particion para indicar que está montada
	partition.MountPartition(indexPartition, id)

	// guardar la particion modificada en el MBR
	mbr.Mbr_partitions[indexPartition] = *partition

	// serializar el MBR
	msg, err = mbr.SerializeMBR(mount.Path)
	if err != nil {
		return msg, fmt.Errorf("error escribiendo el MBR en el disco: %s", err)
	}

	return "", nil
}




func GenerateIdPartition(mount *MOUNT, indexPartition int) (string, string, error) {
	// Asignar una letra a la partición
	letter, err := util.GetLetter(mount.Path)
	if err != nil {
		fmt.Println("Error obteniendo la letra:", err)
		return "", "Error obteniendo la letra en mount",err
	}

	// Crear id de partición
	idPartition := fmt.Sprintf("%s%d%s", util.Carnet, indexPartition, letter)

	return idPartition, "comando mount realizado correctamente", nil
}