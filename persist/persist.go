package persist

// The persist package contains the persistence layer for pipeviz's append-only log.
//
// The initial/prototype implementation is simply a slice of strings. That will be
// swapped out eventually, but the basic interface should remain the same-ish.

var msgs [][]byte

// Appends a new message to the log, returning the id of the message.
func Append(msg []byte, remoteAddr string) int { // TODO uint64 for ids
	// TODO allow err
	msgs = append(msgs, msg)
	return len(msgs)
}

// Returns the message associated with the given id.
func Get(id int) []byte { // TODO uint64 for ids
	// TODO allow err
	return msgs[id-1]
}
