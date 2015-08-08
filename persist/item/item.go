package item

// Log is a single log entry in the append-only log.
type Log struct {
	Index   uint64
	Message []byte
}
