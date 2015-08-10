//go:generate msgp
//msgp:tuple Log
package item

import "time"

// Log is a single log entry in the journal.
type Log struct {
	// The index of the log item in the journal.
	Index uint64 `msg:"index"`

	// A system-local timestamp indicating when this log was persisted.
	Time time.Time `msg:"time"`

	// The network address (in the form IP:port) from which the message came.
	RemoteAddr []byte `msg:"remoteaddr"`

	// The body of the message.
	Message []byte `msg:"message"`
}
