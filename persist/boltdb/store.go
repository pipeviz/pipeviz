package boltdb

import (
	"github.com/boltdb/bolt"
	"github.com/tag1consulting/pipeviz/persist/item"
)

// Represents a single BoltDB file storage backend for the append-only log.
// Bolt, a pure Go implementation inspired by LMDB, is a k/v store and thus
// provides more than we actually need, but is an easy place to start from.
type BoltStore struct {
	conn *bolt.DB
	path string
}

// NewBoltStore creates a handle to a BoltDB-backed log store
func NewBoltStore(path string) (*BoltStore, error) {
	b, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	store := &BoltStore{
		conn: b,
		path: path,
	}

	// initialize the one bucket we use
	if err := store.init(); err != nil {
		store.conn.Close()
		return nil, err
	}

	return store, nil
}

func (b *BoltStore) init() error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}

	// Create all the buckets
	if _, err := tx.CreateBucketIfNotExists([]byte("aol")); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Get returns the item associated with the given index.
func (b *BoltStore) Get(idx uint64, item *item.Log) error {

}
