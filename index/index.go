package index

import (
	"context"
	"time"

	"github.com/rawbytedev/aether/storage"
	"github.com/rawbytedev/zerokv"
	"github.com/rawbytedev/zerokv/badgerdb"
	"github.com/rawbytedev/zerokv/encoders"
)

type AccessEvent struct {
	Key      []byte
	Score    float64
	Tier     storage.TierType
	Prev     storage.TierType
	Accessed time.Time
}
type MigrationEvent struct {
	Key       []byte
	Value     []byte
	status    string
	Prev_Tier storage.TierType
	Tier      storage.TierType
}

type IndexConfig struct {
	Dir           string
	Encoding      string
	MigrationChan chan *MigrationEvent
	AccessChan    chan *AccessEvent
}
type IndexService struct {
	Migration chan *MigrationEvent
	Access    chan<- *AccessEvent
	Config    IndexConfig
	IStore    zerokv.Core
	IBatch    zerokv.Batch
	IEncode   encoders.Encoder
}

func NewIndexService(config IndexConfig) (*IndexService, error) {
	store, err := badgerdb.NewBadgerDB(badgerdb.Config{Dir: config.Dir})
	if err != nil {
		return nil, err
	}
	return &IndexService{
		config.MigrationChan,
		config.AccessChan,
		config,
		store,
		store.Batch(),
		encoders.NewJsonEncoder(),
	}, nil
}

func (idx *IndexService) RecordAccess(key []byte) error {
	val, err := idx.IStore.Get(context.Background(), key)
	if err != nil {
		return err
	}
	Event := &AccessEvent{}
	err = idx.IEncode.Decode(val, Event)
	if err != nil {
		return err
	}
	Event.Increment()
	select {
	case idx.Access <- Event:
	default:
		// channel full – drop event, it's fine. Policy will catch up via periodic scan.
	}
	return nil
}

func (idx *IndexService) Record(key []byte, value []byte) error {
	Event := NewAccessEvent(key, 0, storage.TierNVMe)
	val, err := idx.IEncode.Encode(Event)
	if err != nil {
		return err
	}
	err = idx.IStore.Put(context.Background(), key, val)
	if err != nil {
		return err
	}
	idx.ProposeMigration(storage.TierNVMe, Event, value)
	return nil
}

func (idx *IndexService) Edit(key []byte, value []byte) error {
	val, err := idx.IStore.Get(context.Background(), key)
	if err != nil {
		return err
	}
	Event := &AccessEvent{}
	err = idx.IEncode.Decode(val, Event)
	if err != nil {
		return err
	}
	return nil
}

func (idx *IndexService) Retrieve(key []byte) ([]byte, error) {
	val, err := idx.IStore.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}
	Event := &AccessEvent{}
	err = idx.IEncode.Decode(val, Event)
	if err != nil {
		return nil, err
	}
	err = idx.ProposeMigration(Event.Tier, Event, nil)
	if err != nil {
		return nil, err
	}
	panic("Not implemented entirely")
}
func (idx *IndexService) ProposeMigration(tier storage.TierType, ac *AccessEvent, value []byte) error {
	if ac.Tier != tier {
		ac.UpdateTier(tier)
		val, err := idx.IEncode.Encode(ac)
		if err != nil {
			ac.RollBack()
			return err
		}
		err = idx.IStore.Put(context.Background(), ac.Key, val)
		if err != nil {
			ac.RollBack()
			return err
		}
	}
	idx.Migration <- NewMigrationEvent(ac.Key, value, ac.Prev, ac.Tier)
	return nil
}

func (idx *IndexService) ClaimMigrationJob() *MigrationEvent {
	select {
	case data, ok := <-idx.Migration:
		if ok {
			return data
		}
		return nil
	default:
		return nil
	}
}

func (idx *IndexService) CompleteMigration(key []byte, value []byte, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	if key != nil {
		return nil, nil
	}
	if value != nil {
		return value, nil
	}
	return nil, nil
}

func (idx *IndexService) RollBackTier(key []byte) error {
	ac := &AccessEvent{}
	val, err := idx.IStore.Get(context.Background(), key)
	if err != nil {
		return err
	}
	err = idx.IEncode.Decode(val, ac)
	if err != nil {
		return err
	}
	ac.RollBack()
	enc, err := idx.IEncode.Encode(ac)
	if err != nil {
		return err
	}
	err = idx.IStore.Put(context.Background(), ac.Key, enc)
	if err != nil {
		return err
	}
	return nil
}
func (idx *IndexService) Close() error {
	return idx.IStore.Close()
}

func NewMigrationEvent(key []byte, value []byte, prevTier storage.TierType, Tier storage.TierType) *MigrationEvent {
	return &MigrationEvent{
		Key:       key,
		Value:     value,
		status:    "migrating",
		Prev_Tier: prevTier,
		Tier:      Tier,
	}
}

func (mig *MigrationEvent) HasValue() bool {
	if mig.Value != nil {
		return true
	}
	return false
}

func (mig *MigrationEvent) Done() {
	mig.status = "complete"
}

func (mig *MigrationEvent) Fail() {
	mig.status = "fail"
}

func NewAccessEvent(key []byte, score float64, tier storage.TierType) *AccessEvent {
	return &AccessEvent{
		Key:      key,
		Score:    0,
		Tier:     tier,
		Accessed: time.Now(),
	}
}

func (ac *AccessEvent) UpdateTier(tier storage.TierType) {
	ac.Prev = ac.Tier
	ac.Tier = tier
}

func (ac *AccessEvent) Increment() {
	ac.Score += 1
	ac.Accessed = time.Now()
}

func (ac *AccessEvent) RollBack() {
	ac.Tier = ac.Prev
	ac.Prev = storage.TierType("")
}
