package item

// Log is a single log entry in the journal.
type Log struct {
	// The index of the log item in the journal.
	Index uint64

	// The network address (in the form IP:port) from which the message came.
	RemoteAddr []byte

	// The body of the message.
	Message []byte
}
