// Package mlog contains interfaces and shared components for pipeviz's
// append-only message log.
package mlog

// Store describes a storage backend for Pipeviz's append-only log.
// Based largely on the LogStorage interface in github.com/hashicorp/raft.
type Store interface {
	// Returns the number of items in the log. Probably expensive, call with care.
	Count() (uint64, error)

	// Gets the log item at a given index.
	Get(index uint64) (*Record, error)

	// NewEntry creates a record from the provided data, appends it onto the
	// end of the mlog, and returns the created record.
	NewEntry(message []byte, remoteAddr string) (*Record, error)
}

// RecordGetter is a function type that gets records out of a mlog.
type RecordGetter func(index uint64) (*Record, error)
