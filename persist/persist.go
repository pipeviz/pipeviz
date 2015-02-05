package persist

// The persist package contains the persistence layer for pipeviz's append-only log.
//
// The initial/prototype implementation is simply a slice of strings. That will be
// swapped out eventually, but the basic interface should remain the same-ish.

var msgs [][]byte

// Appends a new message to the log, returning the id of the message.
func Append(msg []byte) uint64 {
	// TODO allow err
	msgs = append(msgs, msg)
	return uint64(len(msgs) + 1)
}

// Returns the message associated with the given id.
func Get(id uint64) []byte {
	// TODO allow err
	return msgs[id-1]
}
