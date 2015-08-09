package persist

// The persist package contains the persistence layer for pipeviz's append-only journal.

import "github.com/tag1consulting/pipeviz/persist/item"

// LogStore describes a storage backend for Pipeviz's append-only log.
// Based largely on the LogStorage interface in github.com/hashicorp/raft.
type LogStore interface {
	// Returns the number of items in the log. Probably expensive, call with care.
	Count() (uint64, error)

	// Gets the log item at a given index.
	Get(index uint64) (*item.Log, error)

	// Appends a log item. If successful, an Index is assigned to the provided Log.
	Append(log *item.Log) error
}
