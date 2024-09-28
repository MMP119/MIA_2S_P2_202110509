package global

import (
	"server/structures"
	"strings"
	"sync"
)

// Mapa global para almacenar el estado de las sesiones por partición y usuario
var sessionMap = make(map[string]map[string]bool)
var mutex sync.Mutex // Para sincronización

//variable para guardar el idParticion, usuario y password
var (
    credenciales = make(map[string]map[string]string)
)


// Función para iniciar sesión en una partición (marcar sesión activa para un usuario en esa partición)
func ActivateSession(partitionID string, user string) {
    mutex.Lock()
    defer mutex.Unlock()

    // Si la partición no existe en el mapa, crear un nuevo mapa de usuarios
    if _, exists := sessionMap[partitionID]; !exists {
        sessionMap[partitionID] = make(map[string]bool)
    }
    
    // Marcar el usuario como activo en la partición
    sessionMap[partitionID][user] = true
}

// Función para cerrar todas las sesiones (desactivar todas las particiones)
func DeactivateSession() {
    mutex.Lock()
    defer mutex.Unlock()

    for partitionID := range sessionMap {
        for user := range sessionMap[partitionID] {
            sessionMap[partitionID][user] = false
        }
    }
}

// Función para verificar si un usuario tiene sesión activa en una partición específica
func IsSessionActive(partitionID string) bool {
    mutex.Lock()
    defer mutex.Unlock()

    if users, exists := sessionMap[partitionID]; exists {
        for _, active := range users {
            if active {
                return true
            }
        }
    }
    return false
}

// Función para verificar si hay alguna sesión activa en cualquier partición
func IsAnySessionActive() bool {
    mutex.Lock()
    defer mutex.Unlock()

    for _, users := range sessionMap {
        for _, active := range users {
            if active {
                return true
            }
        }
    }
    return false
}

// Función para obtener el ID de la partición con alguna sesión activa
func GetIDSession() string {
    mutex.Lock()
    defer mutex.Unlock()

    for partitionID, users := range sessionMap {
        for _, active := range users {
            if active {
                return partitionID
            }
        }
    }
    return ""
}

// Funcion para obtener el user que está activo en una partición
func GetUserActive(partitionID string) string {
    mutex.Lock()
    defer mutex.Unlock()

    if users, exists := sessionMap[partitionID]; exists {
        for user, active := range users {
            if active {
                return user
            }
        }
    }
    return ""
}


// Función para obtener una lista de usuarios activos en una partición específica
func GetActiveUsers(partitionID string) []string {
    mutex.Lock()
    defer mutex.Unlock()

    var activeUsers []string
    if users, exists := sessionMap[partitionID]; exists {
        for user, active := range users {
            if active {
                activeUsers = append(activeUsers, user)
            }
        }
    }
    return activeUsers
}


func IniciarSesion(idParticion string, usuario string, password string){
    //guardar el id, usuario y password, si ya existen, no se sobreescriben
    if _, ok := credenciales[idParticion]; !ok{
        credenciales[idParticion] = make(map[string]string)
    }
    credenciales[idParticion][usuario] = password
}


func ComprobarCredenciales(idParticion string, usuario string, password string)bool{
    //verificar si el id, usuario y password son correctos
    if credenciales[idParticion][usuario] == password{
        ActivateSession(idParticion, usuario)
        return true
    }else{
        return false
    }
}



func VerificarSesion(idParticion string, usuario string, password string)(string, error){
    
    partitionSuperblock, _, partitionPath, err := GetMountedPartitionSuperblock(idParticion)
	if err != nil {
		return "1 no se puede obtener el superbloque", err
	}

    inode := &structures.Inode{}
    
    err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(0*partitionSuperblock.S_inode_size)))
    if err != nil {
        return "2 no se puede deserealizar", err
    }
    salida := ""
    for _, block := range inode.I_block{
        if block != -1{
            folderBlock := &structures.FolderBlock{}
            err = folderBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
            if err != nil {
                return "3 no se puede deserealizar", err
            }
            for _, content := range folderBlock.B_content{
                name := strings.Trim(string(content.B_name[:]), "\x00")
                if name == "users.txt"{
                    inode = &structures.Inode{}
                    err = inode.Deserialize(partitionPath, int64(partitionSuperblock.S_inode_start+(content.B_inodo*partitionSuperblock.S_inode_size)))
                    if err != nil {
                        return "4 no se puede deserealizar", err
                    }

                    fileBlock := &structures.FileBlock{}
                    for _, block := range inode.I_block{
                        if block != -1{
                            err = fileBlock.Deserialize(partitionPath, int64(partitionSuperblock.S_block_start+(block*partitionSuperblock.S_block_size)))
                            if err != nil {
                                return "5 no se puede deserealizar", err
                            }
                            salida = strings.Trim(string(fileBlock.B_content[:]), "\x00")  
                            salida = strings.ReplaceAll(salida, "\r\n", "\n")
                            users := strings.Split(salida, "\n")
                            for _, user := range users{
                                values := strings.Split(user, ",")
                                if len(values)>=5 && values[1] == "U"{
                                    IniciarSesion(idParticion, values[3], values[4])
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    return "", nil
}