package persist

import "github.com/boltdb/bolt"

// The persist package contains the persistence layer for pipeviz's append-only log.
//
// The initial/prototype implementation is simply a slice of strings. That will be
// swapped out eventually, but the basic interface should remain the same-ish.

var msgs [][]byte

// Represents a single BoltDB file storage backend for the append-only log.
// Bolt, a pure Go implementation inspired by LMDB, is a k/v store and thus
// provides more than we actually need, but is an easy place to start from.
type BoltLog struct {
	conn *bolt.DB
	path string
}

// LogItem is a single log entry in the append-only log.
type LogItem struct {
	Index   uint64
	Message []byte
}

// LogStore describes a storage backend for Pipeviz's append-only log.
// Based largely on the LogStore interface in github.com/hashicorp/raft.
type LogStore interface {
	// Returns the first index written, or 0 for no entries.
	FirstIndex() (uint64, error)

	// Returns the last index written, or 0 for no entries.
	LastIndex() (uint64, error)

	// Gets the log item at a given index, writing it into the provided
	// LogItem struct.
	GetLog(index uint64, log *LogItem) error

	// Stores a log item.
	StoreLog(log *LogItem) error
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
