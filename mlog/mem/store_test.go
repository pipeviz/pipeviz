package mem

import (
	"bytes"
	"testing"

	"github.com/tag1consulting/pipeviz/mlog"
)

// Test that appending messages, then subsequently getting them, works as expected
func TestNewEntryGetCount(t *testing.T) {
	store := NewMemStore()

	m1 := []byte("msg1")
	a1 := "127.0.0.1"
	m2 := []byte("msg2")
	a2 := "127.0.0.1"

	var item1, item2 *mlog.Record

	item1, err := store.NewEntry(m1, a1)
	if err != nil {
		t.Errorf("Failed to complete first NewEntry() due to err: %s", err)
	}
	if item1.Index != 1 {
		t.Errorf("First log item should have been assigned index 1, got %d", item1.Index)
	}

	item2, err = store.NewEntry(m2, a2)
	if err != nil {
		t.Errorf("Failed to complete second NewEntry() due to err: %s", err)
	}
	if item2.Index != 2 {
		t.Errorf("Second log item should have been assigned index 2, got %d", item2.Index)
	}

	// Test Count()
	count, err := store.Count()
	if err != nil {
		t.Errorf("Failed to complete Count() due to err: %s", err)
	}
	if count != 2 {
		t.Errorf("After two appends Count() should report two items; reported %d", count)
	}

	// Test Get(), pulling keys in reverse order
	get1, err := store.Get(2)
	if err != nil {
		t.Errorf("Failed to complete Get() on second item due to err: %s", err)
	}
	if !bytes.Equal([]byte("msg2"), get1.Message) {
		t.Errorf("Second persisted message was incorrect, expected %q got %q", "msg2", get1.Message)
	}

	get2, err := store.Get(1)
	if err != nil {
		t.Errorf("Failed to complete Get() on first item due to err: %s", err)
	}
	if !bytes.Equal([]byte("msg1"), get2.Message) {
		t.Errorf("Second persisted message was incorrect, expected %q got %q", "msg1", get2.Message)
	}
}
