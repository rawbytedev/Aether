package storage

import (
	"fmt"

	"github.com/rawbytedev/zerokv"
	"github.com/rawbytedev/zerokv/badgerdb"
	"github.com/rawbytedev/zerokv/pebbledb"
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
	Tiers      []Tier // only requires DBname
}
// TierManager manages multiple storage tiers
type TierManager struct {
	dbs map[string]*Tier
}
// NewTierManager creates a new TierManager with the given configuration
func NewTierManager(config TierConfigs) (*TierManager, error) {
	fmt.Printf("Using %s config", config.ConfigName)
	Manager := &TierManager{}
	for _, tier := range config.Tiers {
		Manager.dbs[string(tier.IType)] = &tier
	}
	return Manager, nil
}

// Initialize initializes the storage tiers by creating the appropriate storage instances based on the configuration
func (tm *TierManager) Initialize() (bool, error) {
	for _, db := range tm.dbs {
		tmp, err := NewStorage(db.DBName, db.Configs)
		if err != nil {
			return false, err
		}
		db.IStore = tmp
	}
	return true, nil
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
