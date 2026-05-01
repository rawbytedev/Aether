package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/rawbytedev/aether/index"
	"github.com/rawbytedev/aether/storage"
)

type Migration struct {
	index       *index.IndexService
	tierStorage *storage.TierManager
}

func NewMigrationService(index *index.IndexService, TierStorage *storage.TierManager) *Migration {
	return &Migration{
		index:       index,
		tierStorage: TierStorage,
	}
}

func (m *Migration) Run() {
	for {
		j := m.index.ClaimMigrationJob()
		if j != nil {
			var val []byte
			var err error
			fmt.Print("New Migration starting\n")

			if j.Value != nil {
				m.Put(j)
			} else {
				val, err = m.Get(j)
			}
			m.index.CompleteMigration(j.Key, val, err)
		}
		time.Sleep(1 * time.Second)
	}
}

func (m *Migration) Put(data *index.MigrationEvent) error {
	return m.tierStorage.Put(context.Background(), data.Tier, data.Key, data.Value)
}

func (m *Migration) Get(data *index.MigrationEvent) ([]byte, error) {
	return m.tierStorage.Get(context.Background(), data.Tier, data.Key)
}
