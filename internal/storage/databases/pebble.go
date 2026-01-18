// Copyright (C) Enviora
// This file is part of go-enviora
//
// PebbleDB implementation for StorageDB interface.
// Provides efficient key-value storage for blocks, transactions, and other data.
// Supports batch operations and is optimized for concurrent access.
//
// Usage:
//   db, err := NewPebbledb(cfg)
//   err = db.Put(key, value)
//   value, err := db.Get(key)
//   err = db.Del(key)
//   err = db.Close()
//
// Batch operations:
//   err = db.BatchPut(key, value)   // enqueue
//   err = db.BatchPut(nil, nil)     // flush
//   err = db.BatchDel(key)          // enqueue delete
//   err = db.BatchDel(nil)          // flush

package databases

import (
	dbconfig "crosstrace/internal/configs"
	"errors"

	"github.com/cockroachdb/pebble"
)

// pebbledb manages Database Insert/Deletion/Batch Operations for Pebble.
// It only handles []byte keys and values.
type pebbledb struct {
	batch *pebble.Batch
	db    *pebble.DB
}

// NewPebbledb creates a new PebbleDB instance with the given config.
// Returns a StorageDB interface or an error if initialization fails.
func NewPebbledb(cfg dbconfig.JournalConfig) (StorageDB, error) {
	opts := &pebble.Options{
		MaxOpenFiles:    5000,
		BytesPerSync:    1 << 20,
		WALBytesPerSync: 1 << 20,
	}
	if cfg.CacheSize != "" {
		size, err := dbconfig.ParseSize(cfg.CacheSize)
		if err != nil {
			return nil, err
		}
		opts.Cache = pebble.NewCache(size)
	}
	db, err := pebble.Open(cfg.DBPath, opts)
	if err != nil {
		return nil, err
	}
	batch := db.NewBatch()
	if batch == nil {
		db.Close()
		return nil, err
	}
	return &pebbledb{db: db, batch: batch}, nil
}

// Put inserts or updates a key-value pair in the database.
func (p *pebbledb) Put(key []byte, data []byte) error {
	if p.db == nil {
		return pebble.ErrClosed
	}
	if key == nil {
		return ErrEmptydbKey
	}
	if data == nil {
		return ErrEmptydbValue
	}
	return p.db.Set(key, data, pebble.Sync)
}

// Get retrieves the value for a given key. Returns an error if not found.
func (p *pebbledb) Get(key []byte) ([]byte, error) {
	if p.db == nil {
		return nil, pebble.ErrClosed
	}
	if key == nil {
		return nil, ErrEmptydbKey
	}
	val, closer, err := p.db.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	return val, nil
}

// BatchPut adds a key-value pair to the current batch. If key and data == nil commits
// since we use key = nil and data = nil before committing we'll have to make checks
func (p *pebbledb) BatchPut(key []byte, data []byte) error {
	if p.batch == nil {
		return pebble.ErrClosed
	}

	if key != nil && data != nil {
		if err := p.batch.Set(key, data, pebble.NoSync); err != nil {
			p.batch.Close()
			return err
		}
	}
	// both key and data must be nil to commit
	if key == nil && data == nil {
		if err := p.FlushBatch(); err != nil {
			return err
		}
		return nil // important to avoid ErrEmptydbKey below
	}
	if key == nil {
		return ErrEmptydbKey
	}
	if data == nil {
		return ErrEmptydbValue
	}
	return nil
}

// BatchDel adds a delete operation to the current batch.
func (p *pebbledb) BatchDel(key []byte) error {
	if p.batch == nil {
		return pebble.ErrClosed
	}
	if key != nil {
		if err := p.batch.Delete(key, pebble.NoSync); err != nil {
			p.batch.Close()
			return err
		}
		return nil
	} else {
		if err := p.FlushBatch(); err != nil {
			return err
		}
		return nil
	}
}

// Del deletes a key-value pair from the database.
func (p *pebbledb) Del(key []byte) error {
	if p.db == nil {
		return pebble.ErrClosed
	}
	if key == nil {
		return ErrEmptydbKey
	}
	return p.db.Delete(key, pebble.Sync)
}

// flushBatch flushes any pending batch operations.
func (p *pebbledb) FlushBatch() error {
	if p.batch == nil {
		return pebble.ErrClosed
	}
	return p.batch.Commit(pebble.Sync)
}

// Close closes the database and releases all resources.
func (p *pebbledb) Close() error {
	var errs []error
	if p.batch != nil {
		if err := p.batch.Commit(pebble.Sync); err != nil {
			errs = append(errs, err)
		}
		p.batch.Close()
	}
	if p.db != nil {
		if err := p.db.Close(); err != nil {
			errs = append(errs, err)
		}

	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
