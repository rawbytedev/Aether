package index

import (
	"time"

	"github.com/rawbytedev/aether/storage"
)

type AccessEvent struct {
	Key      []byte
	Score    float64
	Tier     storage.TierType
	Accessed time.Time
}
type MigrationEvent struct {
}
type IndexConfig struct {
	Dir string
}
type IndexService struct {
	Migration chan *MigrationEvent
}

func NewIndexService(config IndexConfig) {

}
