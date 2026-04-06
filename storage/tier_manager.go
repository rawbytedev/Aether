package storage

import (
	"context"
	"fmt"

	"github.com/rawbytedev/zerokv"
	"github.com/rawbytedev/zerokv/badgerdb"
	"github.com/rawbytedev/zerokv/pebbledb"
	//"github.com/rawbytedev/zerokv/memdb"
)

// TierType represents the type of storage tier
type TierType string

const (
	TierNVMe TierType = "nvme"
	TierSSD  TierType = "ssd"
	TierHDD  TierType = "hdd"
)

// Tier represents a storage tier with its configuration and state
type Tier struct {
	DBName  string
	Enabled bool
	IType   TierType
	IStore  zerokv.Core
	IBatch  zerokv.Batch
	Configs Tconfigs
}

// Tconfigs represents the configuration for a storage tier
type Tconfigs struct {
	Dir string
}

// TierConfigs represents the configuration for all storage tiers
type TierConfigs struct {
	ConfigName string
	Tiers      []*Tier // only requires DBname
}

// TierManager manages multiple storage tiers
type TierManager struct {
	dbs map[TierType]*Tier
}

// NewTierManager creates a new TierManager with the given configuration
func NewTierManager(config TierConfigs) (*TierManager, error) {
	fmt.Printf("Using %s config", config.ConfigName)
	Manager := &TierManager{}
	Manager.dbs = make(map[TierType]*Tier, 10)
	for _, tier := range config.Tiers {
		Manager.dbs[tier.IType] = tier
	}
	return Manager, nil
}

// Initialize initializes the storage tiers by creating the appropriate storage instances based on the configuration
func (tm *TierManager) Initialize() error {
	for _, db := range tm.dbs {
		tmp, err := NewStorage(db.DBName, db.Configs)
		if err != nil {
			return err
		}
		db.IStore = tmp
	}
	return nil
}

func (tm *TierManager) Put(ctx context.Context, tier TierType, key []byte, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return tm.dbs[tier].IStore.Put(ctx, key, value)
}
func (tm *TierManager) Get(ctx context.Context, tier TierType, key []byte) ([]byte, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return tm.dbs[tier].IStore.Get(ctx, key)
}
func (tm *TierManager) Delete(ctx context.Context, tier TierType, key []byte) error {

	if err := ctx.Err(); err != nil {
		return err
	}
	return tm.dbs[tier].IStore.Delete(ctx, key)
}
func (tm *TierManager) BatchPut(ctx context.Context, tier TierType, key []byte, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if tm.dbs[tier].IBatch == nil {
		tm.dbs[tier].IBatch = tm.dbs[tier].IStore.Batch()
	}
	return tm.dbs[tier].IBatch.Put(key, value)
}
func (tm *TierManager) BatchDelete(ctx context.Context, tier TierType, key []byte, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if tm.dbs[tier].IBatch == nil {
		tm.dbs[tier].IBatch = tm.dbs[tier].IStore.Batch()
	}
	return tm.dbs[tier].IBatch.Delete(key)
}
func (tm *TierManager) Commit(ctx context.Context, tier TierType) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if tm.dbs[tier].IBatch == nil {
		tm.dbs[tier].IBatch = tm.dbs[tier].IStore.Batch()
		return nil
	}
	err := tm.dbs[tier].IBatch.Commit(ctx)
	if err != nil {
		return err
	}
	tm.dbs[tier].IBatch = nil
	return nil
}

// NewStorage creates a new storage instance based on the given database name and configuration
func NewStorage(dbname string, config Tconfigs) (zerokv.Core, error) {
	switch dbname {
	case "pebbledb":
		return pebbledb.NewPebbleDB(pebbledb.Config{Dir: config.Dir})
	case "badgerdb":
		return badgerdb.NewBadgerDB(badgerdb.Config{Dir: config.Dir})
	default:
		return nil, fmt.Errorf("")
	}
}
