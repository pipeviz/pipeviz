package persist

import (
	"github.com/sdboyer/pipeviz/persist/item"
	"github.com/tag1consulting/pipeviz/persist/boltdb"
)

// The persist package contains the persistence layer for pipeviz's append-only log.
//
// The initial/prototype implementation is simply a slice of strings. That will be
// swapped out eventually, but the basic interface should remain the same-ish.

var msgs [][]byte

// LogStore describes a storage backend for Pipeviz's append-only log.
// Based largely on the LogStorage interface in github.com/hashicorp/raft.
type LogStore interface {
	// Returns the number of items in the log. Probably expensive, call with care.
	Count() (uint64, error)

	// Gets the log item at a given index.
	Get(index uint64) (*LogItem, error)

	// Appends a log item.
	Append(log *item.Log) error
}

// Appends a new message to the log, returning the id of the message.
func Append(msg []byte, remoteAddr string) int { // TODO uint64 for ids
	// TODO allow err
	msgs = append(msgs, msg)
	return len(msgs)
}

// Returns the message associated with the given id.
func Get(id int) []byte { // TODO uint64 for ids
	// TODO allow err
	return msgs[id-1]
}

// NewLog take a file path and returns a LogStorage.
func NewLog(path string) (LogStore, error) {
	return boltdb.NewBoltStore(path)
}
