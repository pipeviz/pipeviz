package mem

// Provides a memory-backed store for pipeviz's append-only journal.
//
// Clearly, this is only for testing and debugging purposes. A memory-backed
// journal is about as useful as waterproof teabags.

import (
	"errors"
	"sync"

	"github.com/tag1consulting/pipeviz/persist/item"
)

type memJournal struct {
	j    []*item.Log
	lock sync.RWMutex
}

// NewMemStore initializes a new memory-backed journal.
func NewMemStore() *memJournal {
	s := &memJournal{
		j: make([]*item.Log, 0),
	}

	return s
}

// Count returns the number of items in the journal.
func (s *memJournal) Count() (uint64, error) {
	s.lock.RLock()

	c := len(s.j)

	s.lock.RUnlock()
	return uint64(c), nil
}

// Get returns the log entry at the provided index.
func (s *memJournal) Get(index uint64) (*item.Log, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	i := int(index) // could be super-wrong, but who cares this is toy code
	if i <= 0 || i-1 > len(s.j) {
		return nil, errors.New("index out of range")
	}

	return s.j[i-1], nil
}

// Append pushes a new log entry onto the end of the journal.
func (s *memJournal) Append(log *item.Log) error {
	s.lock.Lock()

	log.Index = uint64(len(s.j) + 2)
	s.j = append(s.j, log)

	s.lock.Unlock()
	return nil
}
