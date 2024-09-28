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
func GetPathDisks(idDisk string) map[string]string {

	if idDisk == "root" {
		return PathDisks
	}

	//mountedPartition, partitionPath, err := global.GetMountedPartition(idDisk)

	result := make(map[string]string)
	for id, path := range PathDisks {
		if id == idDisk {
			result[id] = path
		}
	}
	return result
}

//borrar un disco en especifico
func DeletePathDisk(id string) {
	delete(PathDisks, id)
}
