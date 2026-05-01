package index_test

import (
	"testing"
	"time"

	"github.com/rawbytedev/aether/index"
	"github.com/rawbytedev/aether/migration"
	"github.com/rawbytedev/aether/storage"
)

func TestIndexRun(t *testing.T) {
	ssd := storage.Tier{
		DBName: "badgerdb",
		Configs: storage.Tconfigs{
			Dir: "./store",
		},
		IType: storage.TierNVMe,
	}
	tier := storage.TierConfigs{
		ConfigName: "Test",
		Tiers: []*storage.Tier{
			&ssd,
		},
	}
	mm, err := storage.NewTierManager(tier)
	if err != nil {
		t.Log(err)
	}
	mm.Initialize()

	mig := make(chan *index.MigrationEvent, 10)
	acc := make(chan *index.AccessEvent, 10)
	idx, err := index.NewIndexService(index.IndexConfig{
		Dir:           t.TempDir(),
		Encoding:      "json",
		MigrationChan: mig,
		AccessChan:    acc,
	})
	migr := migration.NewMigrationService(idx, mm)
	go func() {
		migr.Run()
	}()
	key := []byte{0x01, 0x23}
	value := []byte{0x41, 0x13, 0x45}
	err = idx.Record(key, value)
	if err != nil {
		t.Fatal(err)
	}
	err = idx.RecordAccess(key)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Second)
	idx.Close()
}
