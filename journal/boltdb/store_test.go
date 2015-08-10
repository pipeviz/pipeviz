package boltdb

import (
	"bytes"
	"os"
	"testing"

	"github.com/tag1consulting/pipeviz/journal"
)

// Ensure initializing the bolt journal works as expected
func TestNewBoltStore(t *testing.T) {
	ls, err := NewBoltStore("test.boltdb")
	b := ls.(*BoltStore)
	defer func() {
		b.conn.Close()
		os.Remove("test.boltdb")
	}()

	if err != nil {
		t.Errorf("Failed to create bolt store with err %s", err)
	}

	// Ensure expected bucket exists
	tx, err := b.conn.Begin(false)
	defer tx.Rollback()
	if err != nil {
		t.Errorf("Could not create read-only txn on bolt database, errored with %s", err)
	}

	if bucket := tx.Bucket(bucketName); bucket == nil {
		t.Errorf("Expected bucket with name %s does not exist after init", bucketName)
	}
}

// Test that appending messages, then subsequently getting them, works as expected
func TestNewEntryGetCount(t *testing.T) {
	ls, err := NewBoltStore("test.boltdb")
	b := ls.(*BoltStore)
	defer func() {
		b.conn.Close()
		os.Remove("test.boltdb")
	}()

	if err != nil {
		t.Errorf("Failed to create bolt store with err %s", err)
	}

	m1 := []byte("msg1")
	a1 := "127.0.0.1"
	m2 := []byte("msg2")
	a2 := "127.0.0.1"

	var item1, item2 *journal.Record

	item1, err = b.NewEntry(m1, a1)
	if err != nil {
		t.Errorf("Failed to complete first NewEntry() due to err: %s", err)
	}
	if item1.Index != 1 {
		t.Errorf("First log item should have been assigned index 1, got %d", item1.Index)
	}

	item2, err = b.NewEntry(m2, a2)
	if err != nil {
		t.Errorf("Failed to complete second NewEntry() due to err: %s", err)
	}
	if item2.Index != 2 {
		t.Errorf("Second log item should have been assigned index 2, got %d", item2.Index)
	}

	// Test Count()
	count, err := b.Count()
	if err != nil {
		t.Errorf("Failed to complete Count() due to err: %s", err)
	}
	if count != 2 {
		t.Errorf("After two appends Count() should report two items; reported %d", count)
	}

	// Test Get(), pulling keys in reverse order
	get1, err := b.Get(2)
	if err != nil {
		t.Errorf("Failed to complete Get() on second item due to err: %s", err)
	}
	if !bytes.Equal([]byte("msg2"), get1.Message) {
		t.Errorf("Second persisted message was incorrect, expected %q got %q", "msg2", get1.Message)
	}

	get2, err := b.Get(1)
	if err != nil {
		t.Errorf("Failed to complete Get() on first item due to err: %s", err)
	}
	if !bytes.Equal([]byte("msg1"), get2.Message) {
		t.Errorf("Second persisted message was incorrect, expected %q got %q", "msg1", get2.Message)
	}
}
