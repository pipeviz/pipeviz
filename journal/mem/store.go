package mem

// Provides a memory-backed store for pipeviz's append-only journal.
//
// Clearly, this is only for testing and debugging purposes. A memory-backed
// journal is about as useful as waterproof teabags.

import (
	"errors"
	"sync"

	"github.com/tag1consulting/pipeviz/journal"
)

type MemJournal struct {
	j    []*journal.Record
	lock sync.RWMutex
}

// NewMemStore initializes a new memory-backed journal.
func NewMemStore() journal.LogStore {
	s := &MemJournal{
		j: make([]*journal.Record, 0),
	}

	return s
}

// Count returns the number of items in the journal.
func (s *MemJournal) Count() (uint64, error) {
	return uint64(len(s.j)), nil
}

// Get returns the journal entry at the provided index.
func (s *MemJournal) Get(index uint64) (*journal.Record, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	i := int(index) // could be super-wrong, but who cares this is toy code
	if i <= 0 || i-1 > len(s.j) {
		return nil, errors.New("index out of range")
	}

	return s.j[i-1], nil
}

// NewEntry creates a record from the provided data, appends that record onto
// the end of the journal, then returns the created record.
func (s *MemJournal) NewEntry(message []byte, remoteAddr string) (*journal.Record, error) {
	s.lock.Lock()

	record := journal.NewRecord(message, remoteAddr)
	record.Index = uint64(len(s.j) + 1)

	s.j = append(s.j, record)

	s.lock.Unlock()
	return record, nil
}
