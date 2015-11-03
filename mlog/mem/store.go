// Package mem provides memory-backed storage for pipeviz's append-only message log.
//
// Clearly, this is only for development, testing, and debugging purposes;
// in general, a memory-backed mlog is about as useful as waterproof teabags.
package mem

import (
	"errors"
	"sync"

	"github.com/pipeviz/pipeviz/mlog"
)

type memMessageLog struct {
	j    []*mlog.Record
	lock sync.RWMutex
}

// NewMemStore initializes a new memory-backed mlog.
func NewMemStore() mlog.Store {
	s := &memMessageLog{
		j: make([]*mlog.Record, 0),
	}

	return s
}

// Count returns the number of items in the mlog.
func (s *memMessageLog) Count() (uint64, error) {
	return uint64(len(s.j)), nil
}

// Get returns the mlog entry at the provided index.
func (s *memMessageLog) Get(index uint64) (*mlog.Record, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	i := int(index) // could be super-wrong, but who cares this is toy code
	if i <= 0 || i-1 > len(s.j) {
		return nil, errors.New("index out of range")
	}

	return s.j[i-1], nil
}

// NewEntry creates a record from the provided data, appends that record onto
// the end of the mlog, then returns the created record.
func (s *memMessageLog) NewEntry(message []byte, remoteAddr string) (*mlog.Record, error) {
	s.lock.Lock()

	record := mlog.NewRecord(message, remoteAddr)
	record.Index = uint64(len(s.j) + 1)

	s.j = append(s.j, record)

	s.lock.Unlock()
	return record, nil
}
