package storage_test

import (
	"context"
	"testing"

	"github.com/rawbytedev/aether/storage"
)

func TestPut(t *testing.T) {
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
	err = mm.Put(
		context.Background(), storage.TierNVMe,
		[]byte{0x01, 0x02}, []byte{0x05, 0x50, 0x05, 0x20},
	)
	if err != nil {
		t.Log(err)
	}
	val, err := mm.Get(context.Background(), storage.TierNVMe, []byte{0x01, 0x02})
	if err != nil {
		t.Log(err)
	}
	t.Log(val)
}
