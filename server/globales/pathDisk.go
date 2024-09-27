package globales

var (
	PathDisks = make(map[string]string)
)


func GetPathDisk(id string) string {
	return PathDisks[id]
}

func SetPathDisk(id string, path string) {
	PathDisks[id] = path
}

// hay que pasarle el map al frontend
func GetPathDisks() map[string]string {
	return PathDisks
}

//borrar un disco en especifico
func DeletePathDisk(id string) {
	delete(PathDisks, id)
}
