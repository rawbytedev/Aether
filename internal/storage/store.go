package storage

type tier int

const (
	NVMe tier = iota
	SSD
	HDD
)

// This is a handler specific to storage
type StorageHandler struct {
	available map[tier]*Stores
}

type Stores struct {
	name string
}

// Create a new storageHandler Instance takes as parameter a list of stores to use
func NewStorageHandler(store []string) *StorageHandler {
	var available map[tier]*Stores
	for i := 0; i < len(store); i++ {
		switch store[i] {
		case "NVMe":
			available[NVMe] = &Stores{name: "NVMe"}
		case "SSD":
			available[SSD] = &Stores{name: "SSD"}
		case "HDD":
			available[HDD] = &Stores{name: "HDD"}
		default:

		}
	}
	return &StorageHandler{available: available}
}
