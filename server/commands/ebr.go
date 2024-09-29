package commands

import (
	"encoding/binary"
	"os"
)

//structura del EBR

type EBR struct {
	Part_mount	[1]byte
	Part_fit	[1]byte
	Part_start	int32
	Part_size	int32
	Part_next	int32
	Part_name	[16]byte
}

func CreateEBR(path string, fdisk *FDISK, startEBR  int32) (string, error) {

    ebr := &EBR{
        Part_mount: [1]byte{'0'},
        Part_fit: [1]byte{fdisk.Fit[0]},
        Part_start: startEBR,
        Part_size: int32(fdisk.Size),  
        Part_next: -1,
    }
    copy(ebr.Part_name[:], []byte(fdisk.Name))

    file, err :=  os.Open(path)
    if err != nil {
        return "Error al abrir el archivo del disco", err
    }
    defer file.Close()

    _, err = file.Seek(int64(startEBR), 0)
    if err != nil {
        return "Error al moverse al inicio del EBR", err
    }

    err = binary.Write(file, binary.LittleEndian, ebr)
    if err != nil {
        return "Error al escribir el EBR", err
    }

    return "", nil
}

// funciones para agregar a la particion extendida