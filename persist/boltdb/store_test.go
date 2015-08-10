package boltdb

import (
	"os"
	"testing"
)

// Ensure initializing the bolt journal works as expected
func TestNewBoltStore(t *testing.T) {
	b, err := NewBoltStore("test.boltdb")
	defer func() {
		b.conn.Close()
		os.Remove("test.boltdb")
	}()

	if err != nil {
		t.Errorf("Failed to create bolt store with err %s", err)
	}

	// Ensure db is empty
	if b.next != 1 {
		t.Errorf("New bolt journal should start with next == 1, was %d", b.next)
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
