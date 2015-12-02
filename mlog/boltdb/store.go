// Package boltdb provides disk-backed storage for pipeviz's append-only message log
// via boltdb.
package boltdb

import (
	"encoding/binary"
	"errors"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pipeviz/pipeviz/mlog"
)

const (
	fileMode = 0600
)

var (
	// The name of the bucket to use.
	bucketName = []byte("mlog")
)

// BoltStore represents a single BoltDB storage backend for the append-only log.
// Bolt, a pure Go implementation inspired by LMDB, is a k/v store and thus
// provides more than we actually need, but it's an easy starting point.
type BoltStore struct {
	conn *bolt.DB
	path string
}

// NewBoltStore creates a handle to a BoltDB-backed log store
func NewBoltStore(path string) (mlog.Store, error) {
	// Allow 1s timeout on obtaining a file lock
	b, err := bolt.Open(path, fileMode, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}

	store := &BoltStore{
		conn: b,
		path: path,
	}

	// initialize the one bucket we use
	if err := store.init(); err != nil {
		// an err here doesn't matter
		_ = store.conn.Close()
		return nil, err
	}

	return store, nil
}

// init sets up the mlog bucket in the boltdb backend.
func (b *BoltStore) init() error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}

	// Create all the buckets
	if _, err := tx.CreateBucketIfNotExists(bucketName); err != nil {
		// already failed, error doesn't matter
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Get returns the item associated with the given index.
func (b *BoltStore) Get(idx uint64) (*mlog.Record, error) {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(bucketName)
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, idx)
	val := bucket.Get(key)

	if val == nil {
		return nil, errors.New("index not found")
	}

	l := &mlog.Record{}
	if _, err := l.UnmarshalMsg(val); err != nil {
		return nil, err
	}
	return l, nil
}

// NewEntry creates a record from the provided data, appends that record onto
// the end of the mlog, then returns the created record.
func (b *BoltStore) NewEntry(message []byte, remoteAddr string) (*mlog.Record, error) {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// no need to sync b/c the conn.Begin(true) call will block
	bucket := tx.Bucket(bucketName)

	record := mlog.NewRecord(message, remoteAddr)
	record.Index, err = bucket.NextSequence()
	if err != nil {
		return nil, err
	}

	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, record.Index)
	val, err := record.MarshalMsg(nil) // TODO nil will alloc for us; keep this zero-alloc
	if err != nil {
		return nil, err
	}

	if err = bucket.Put(key, val); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return record, nil
}

// Count reports the number of items in the mlog by opening a db cursor to
// grab the last item from the bucket. Because we're append-only, this is
// guaranteed to be the last one, and thus its index is the count.
func (b *BoltStore) Count() (uint64, error) {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	curs := tx.Bucket(bucketName).Cursor()
	var last []byte
	if last, _ = curs.Last(); last == nil {
		return 0, nil
	}

	return binary.BigEndian.Uint64(last), nil
}
