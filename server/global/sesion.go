package global

import "sync"

// Mapa global para almacenar el estado de las sesiones por partición y usuario
var sessionMap = make(map[string]map[string]bool)
var mutex sync.Mutex // Para sincronización

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
