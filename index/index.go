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
	Prev_Tier storage.TierType
	Tier      storage.TierType
}
type IndexConfig struct {
	Dir      string
	Encoding string
}
type IndexService struct {
	Migration chan<- *MigrationEvent
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
		make(chan *MigrationEvent),
		make(chan *AccessEvent),
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
	//idx.Access <- Event
	return nil
}

func (idx *IndexService) Record(key []byte, value []byte) error {
	Event := NewAccessEvent(key, 0, storage.TierNVMe)
	val, err := idx.IEncode.Encode(Event)
	if err != nil {
		return err
	}
	return idx.IStore.Put(context.Background(), key, val)
}

func (idx *IndexService) UpdateTier(tier storage.TierType, ac *AccessEvent) error {
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
	//idx.Migration <- &MigrationEvent{ac.Key, ac.Tier, tier}
	return nil
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
