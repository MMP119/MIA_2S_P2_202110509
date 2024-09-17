package structures

import (
	//"encoding/binary"
	"encoding/binary"
	"fmt"
	"os"
	util "server/util"
)

// FDISK estructura que representa el comando fdisk con sus parámetros
type FDISK struct {
	Size int    // Tamaño de la partición
	Unit string // Unidad de medida del tamaño (B, K o M); por defecto K
	Path string // Ruta del archivo del disco
	TypE  string // Tipo de partición (P, E, L)
	Fit  string // Tipo de ajuste (BF, FF, WF); por defecto WF
	Name string // Nombre de la partición
}

func CommandFdisk(fdisk *FDISK) (string, error) {
	
	// Convertir el tamaño a bytes
	sizeBytes, err := util.ConvertToBytes(fdisk.Size, fdisk.Unit)
	if err != nil {
		fmt.Println("Error converting size:", err)
		return "Error converting size en Fdisk", err
	}

	var msg string

	// para crear la particion primaria
	if(fdisk.TypE == "P"){
		msg, err = CreatePrimaryPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error creating primary partition:", err)
			return msg, err
		}
	}else if(fdisk.TypE == "E"){

		msg, err = CreateExtendPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error creating extended partition:", err)
			return msg, err
		}


	}else if(fdisk.TypE == "L"){

		msg, err = CreateLogicalPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error creating logical partition:", err)
			return msg, err
		}

	}
	return "",nil
}


func CreatePrimaryPartition(fdisk *FDISK, sizeBytes int)(string, error){
	
	var mbr MBR
	
	msg, err := mbr.DeserializeMBR(fdisk.Path)
	if err != nil {
		return msg, fmt.Errorf("error leyendo el MBR del disco, el disco no existe: %s", err)
	}

	// Contar el número de particiones primarias 
	primaryCount := 0
	for _, partition := range mbr.Mbr_partitions {
		if partition.Part_status[0] != '2' {
			if partition.Part_type[0] == 'P' {
				primaryCount++
			} 
		}
	}

	// Verificar que no se exceda el límite de 4 particiones
	if primaryCount >4 {
		return "ERROR: No se pueden crear más particiones primarias", fmt.Errorf("límite de particiones primarias alcanzado")
	}

	// verificar si hay espacio suficiente en el disco
	if sizeBytes > int(mbr.Mbr_size) {
		return "ERROR: No hay espacio suficiente en el disco", fmt.Errorf("tamaño de la partición excede el tamaño del disco")
	}

	// se obtiene la primera particion libre
	particionDisponible, inicioParticion, indexParticion, msg:= mbr.GetFirstPartitionAvailable()
	if particionDisponible == nil {
		return msg, fmt.Errorf("no hay particiones disponibles")
	}

	// crear la particion con los parámetros proporcionados 
	particionDisponible.CreatePartition(inicioParticion, sizeBytes, fdisk.TypE, fdisk.Fit, fdisk.Name)

	// montar la particion
	mbr.Mbr_partitions[indexParticion] = *particionDisponible //asignar la particion al MBR

	// Serialiazar el MBR modificado
	msg, err = mbr.SerializeMBR(fdisk.Path)
	if err != nil {
		return msg, fmt.Errorf("error escribiendo el MBR al disco: %s", err)
	}	
	return "",nil
}


func CreateExtendPartition(fdisk *FDISK, sizeBytes int)(string, error){
	
	var mbr MBR
	
	msg, err := mbr.DeserializeMBR(fdisk.Path)
	if err != nil {
		return msg, fmt.Errorf("error leyendo el MBR del disco: %s", err)
	}

	// Contar el número de particiones extendidas
	extendedExists := 0
	for _, partition := range mbr.Mbr_partitions {
		if partition.Part_status[0] != '2' {
			if partition.Part_type[0] == 'E' {
				extendedExists++
			}
		}
	}

	if extendedExists >=2 {
		return "ERROR: No se pueden crear más particiones extendidas, ya existe una en el disco", fmt.Errorf("ya existe una partición extendida")
	}

	// verificar si hay espacio suficiente en el disco
	if sizeBytes > int(mbr.Mbr_size) {
		return "ERROR: No hay espacio suficiente en el disco", fmt.Errorf("tamaño de la partición excede el tamaño del disco")
	}

	// se obtiene la primera particion libre
	particionDisponible, inicioParticion, indexParticion, msg:= mbr.GetFirstPartitionAvailable()
	if particionDisponible == nil {
		return msg, fmt.Errorf("no hay particiones disponibles")
	}

	// crear la particion con los parámetros proporcionados 
	particionDisponible.CreatePartition(inicioParticion, sizeBytes, fdisk.TypE, fdisk.Fit, fdisk.Name)

	// montar la particion
	mbr.Mbr_partitions[indexParticion] = *particionDisponible //asignar la particion al MBR

	// Serialiazar el MBR modificado
	msg, err = mbr.SerializeMBR(fdisk.Path)
	if err != nil {
		return msg, fmt.Errorf("error escribiendo el MBR al disco: %s", err)
	}	
	return "",nil
}


func CreateLogicalPartition(fdisk *FDISK, sizeBytes int) (string, error) {
	var mbr MBR

	msg, err := mbr.DeserializeMBR(fdisk.Path)
	if err != nil {
		return msg, fmt.Errorf("error leyendo el MBR del disco: %s", err)
	}

	// Buscar la partición extendida
	var extendedPartition *PARTITION
	for _, partition := range mbr.Mbr_partitions {
		if partition.Part_type[0] == 'E' {
			extendedPartition = &partition
			break
		}
	}

	if extendedPartition == nil {
		return "No se encontró una partición extendida", nil
	}

	// Verificar que la partición lógica no exceda el tamaño de la partición extendida
	if sizeBytes > int(extendedPartition.Part_size) {
		return "ERROR: Tamaño de la partición lógica excede el tamaño de la partición extendida", fmt.Errorf("tamaño de la partición lógica excede el tamaño de la partición extendida")
	}

	// Moverme al inicio de la partición extendida
	file, err := os.OpenFile(fdisk.Path, os.O_RDWR, 0644)
	if err != nil {
		return "Error al abrir el archivo del disco", err
	}
	defer file.Close()

	_, err = file.Seek(int64(extendedPartition.Part_start), 0)
	if err != nil {
		return "Error al moverse al inicio de la partición extendida", err
	}

	// Leer el primer EBR
	var ebr EBR
	err = binary.Read(file, binary.LittleEndian, &ebr)

	if err != nil || ebr.Part_size == 0 {
		// Si no se encuentra un EBR válido, crear el primero
		//fmt.Println("No se encontró un EBR. Creando el primero.")
		
		ebr = EBR{
			Part_mount: [1]byte{'0'},
			Part_fit:   [1]byte{fdisk.Fit[0]},
			Part_start: extendedPartition.Part_start, // El primer EBR comienza en el inicio de la partición extendida
			Part_size:  int32(sizeBytes),             // Tamaño de la partición lógica
			Part_next:  -1,                           // No hay más EBRs
		}
		copy(ebr.Part_name[:], []byte(fdisk.Name))

		// Moverme al inicio de la partición extendida para escribir el EBR
		_, err = file.Seek(int64(extendedPartition.Part_start), 0)
		if err != nil {
			return "Error al moverse al inicio de la partición extendida", err
		}

		// Escribir el primer EBR
		err = binary.Write(file, binary.LittleEndian, &ebr)
		if err != nil {
			return "Error al escribir el EBR", err
		}

		// Crear la partición lógica después del EBR (tomando en cuenta el tamaño del EBR)
		logicalStart := extendedPartition.Part_start + int32(binary.Size(ebr))

		var logicalPartition PARTITION
		logicalPartition.CreatePartition(int(logicalStart), sizeBytes, fdisk.TypE, fdisk.Fit, fdisk.Name)
		logicalPartition.Part_id = extendedPartition.Part_id

		// escribir la particion logica en el disco
		_, err = file.Seek(int64(logicalStart), 0)
		if err != nil {
			return "Error al moverse al inicio de la partición lógica", err
		}

		err = binary.Write(file, binary.LittleEndian, &logicalPartition)
		if err != nil {
			return "Error al escribir la partición lógica", err
		}

		//logicalPartition.Print()

		// Serializar el MBR actualizado
		msg, err = mbr.SerializeMBR(fdisk.Path)
		if err != nil {
			return msg, err
		}

		//fmt.Println("Primer EBR y partición lógica creados exitosamente.")
		return "PRIMER EBR creado exitosamente", nil
	}

	// Si ya existe un EBR al inicio, recorrer hasta el último EBR
	//fmt.Println("Se encontró un EBR. Buscando el último EBR.")

	for ebr.Part_next != -1 {
		_, err = file.Seek(int64(ebr.Part_next), 0) 
		if err != nil {
			return "Error al moverse al siguiente EBR", err
		}
		err = binary.Read(file, binary.LittleEndian, &ebr)
		if err != nil {
			return "Error al leer el siguiente EBR", err
		}
	}

	// Crear un nuevo EBR después de la última partición lógica
	newEBRStart := ebr.Part_start + ebr.Part_size + int32(binary.Size(ebr))

	// Actualizar el EBR anterior para que apunte al nuevo EBR
	ebr.Part_next = newEBRStart

	// Moverme al inicio del EBR anterior para actualizarlo
	_, err = file.Seek(int64(ebr.Part_start), 0)
	if err != nil {
		return "Error al moverse al EBR anterior para actualizarlo", err
	}

	// Escribir el EBR anterior con Part_next actualizado
	err = binary.Write(file, binary.LittleEndian, &ebr)
	if err != nil {
		return "Error al escribir el EBR anterior con el nuevo Part_next", err
	}

	// Ahora escribir el nuevo EBR (ebr1)
	ebr1 := EBR{
		Part_mount: [1]byte{'0'},
		Part_fit:   [1]byte{fdisk.Fit[0]},
		Part_start: newEBRStart,
		Part_size:  int32(sizeBytes),
		Part_next:  -1,
	}
	copy(ebr1.Part_name[:], []byte(fdisk.Name))

	_, err = file.Seek(int64(ebr1.Part_start), 0)
	if err != nil {
		return "Error al moverse para escribir el nuevo EBR", err
	}

	// Escribir el nuevo EBR
	err = binary.Write(file, binary.LittleEndian, &ebr1)
	if err != nil {
		return "Error al escribir el nuevo EBR", err
	}

	// Crear la partición lógica después del nuevo EBR
	logicalStart := newEBRStart + int32(binary.Size(ebr1))
	var logicalPartition PARTITION
	logicalPartition.CreatePartition(int(logicalStart), sizeBytes, fdisk.TypE, fdisk.Fit, fdisk.Name)
	logicalPartition.Part_id = extendedPartition.Part_id

	// Escribir la partición lógica en el disco
	_, err = file.Seek(int64(logicalStart), 0)
	if err != nil {
		return "Error al moverse al inicio de la partición lógica", err
	}

	err = binary.Write(file, binary.LittleEndian, &logicalPartition)
	if err != nil {
		return "Error al escribir la partición lógica", err
	}

	//logicalPartition.Print()

	// Serializar el MBR actualizado
	msg, err = mbr.SerializeMBR(fdisk.Path)
	if err != nil {
		return msg, err
	}

	//PrintEBRs(fdisk)
	//fmt.Println("Nuevo EBR y partición lógica creados exitosamente.")

	//msg1, err := PrintEBRs(fdisk)
	// if err != nil {
	// 	fmt.Println("Error imprimiendo los EBRs:", err)
	// } else {
	// 	fmt.Println(msg1)
	// }

	return "EBR creado exitosamente", nil
}




func PrintEBRs(fdisk *FDISK) (string, error) {
	var mbr MBR

	// Deserializar el MBR
	msg, err := mbr.DeserializeMBR(fdisk.Path)
	if err != nil {
		return msg, fmt.Errorf("error leyendo el MBR del disco, disco no existe: %s", err)
	}

	// Buscar la partición extendida
	var extendedPartition *PARTITION
	for _, partition := range mbr.Mbr_partitions {
		if partition.Part_type[0] == 'E' {
			extendedPartition = &partition
			break
		}
	}

	if extendedPartition == nil {
		return "No se encontró una partición extendida", nil
	}

	// Abrir el archivo del disco
	file, err := os.OpenFile(fdisk.Path, os.O_RDWR, 0644)
	if err != nil {
		return "Error al abrir el archivo del disco", err
	}
	defer file.Close()

	// Moverme al inicio de la partición extendida
	_, err = file.Seek(int64(extendedPartition.Part_start), 0)
	if err != nil {
		return "Error al moverse al inicio de la partición extendida", err
	}

	// Leer y recorrer los EBRs
	//fmt.Println("EBRs y Particiones Lógicas:")
	for {
		// Leer el EBR
		var ebr EBR
		err = binary.Read(file, binary.LittleEndian, &ebr)
		if err != nil {
			return "Error al leer el EBR", err
		}

		// Si el EBR tiene tamaño 0, significa que no hay más EBRs
		if ebr.Part_size == 0 {
			break
		}

		// Imprimir información del EBR y la partición lógica asociada
		fmt.Printf("\nEBR:\n")
		fmt.Printf("Nombre: %s\n", string(ebr.Part_name[:]))
		fmt.Printf("Inicio: %d\n", ebr.Part_start)
		fmt.Printf("Tamaño: %d\n", ebr.Part_size)
		fmt.Printf("Siguiente EBR: %d\n", ebr.Part_next)

		// Imprimir la partición lógica asociada al EBR
		logicalStart := ebr.Part_start + int32(binary.Size(ebr))
		fmt.Printf("\n Partición Lógica:\n")
		fmt.Printf("Inicio: %d\n", logicalStart)
		fmt.Printf("Tamaño: %d\n", ebr.Part_size)

		// Si no hay más EBRs, detener el ciclo
		if ebr.Part_next == -1 {
			break
		}

		// Moverme al siguiente EBR
		_, err = file.Seek(int64(ebr.Part_next), 0)
		if err != nil {
			return "Error al moverse al siguiente EBR", err
		}
	}

	return "EBRs impresos correctamente", nil
}