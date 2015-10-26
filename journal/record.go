//go:generate msgp
//msgp:tuple Record

package journal

import (
	"net"
	"time"
)

// Record is a single entry in the journal.
type Record struct {
	// The index of the log item in the journal.
	Index uint64 `msg:"index"`

	// A system-local timestamp, split into seconds and nanoseconds, indicating
	// when this record was persisted.
	TimeSec  int64 `msg:"ts"`
	TimeNSec int64 `msg:"tns"`

	// The network address (in the form IP:port) from which the message came.
	RemoteAddr []byte `msg:"remoteaddr"`

	// The body of the message.
	Message []byte `msg:"message"`
}

// NewRecord creates a new Record struct with a current timestamp. The
// expectation is that it will be immediately persisted to disk.
func NewRecord(message []byte, RemoteAddr string) *Record {
	t := time.Now()
	return &Record{
		Index:      0,
		TimeSec:    t.Unix(),
		TimeNSec:   t.UnixNano(),
		RemoteAddr: net.ParseIP(RemoteAddr),
		Message:    message,
	}
}

// Time returns a standard Go time.Time object composed from the timestamp
// indicating when the record was persisted to the journal.
func (r Record) Time() time.Time {
	return time.Unix(r.TimeSec, r.TimeNSec)
}
