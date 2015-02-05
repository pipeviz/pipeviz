package main

import (
	"io/ioutil"
	"net/http"

	"github.com/sdboyer/pipeviz/persist"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
)

type Vertex struct {
	vtype    string
	changeId uint64
	scanId   uint64
}

type Property struct {
	value    interface{}
	changeId uint64
	scanId   uint64
}

type Message struct {
	Id  uint64
	Raw []byte
}

// Channel for handling persisted messages. 100 cap to allow some wiggle room
// if there's a sudden burst of messages
var interpretChan chan Message = make(chan Message, 100)

func main() {
	m := web.New()

	m.Put("/environment", handle)

	go interpret(interpretChan)
	graceful.ListenAndServe("127.0.0.1:9132", m)
}

func handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// FIXME limit this, prevent overflowing memory from a malicious message
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// FIXME send back an error
	}

	// super-sloppy write back to client, but does the trick
	// TODO not writing the id back at all might make things simpler
	w.WriteHeader(202) // use 202 because it's a little more correct
	w.Write(id)

	id := persist.Append(b)

	// FIXME passing directly from here means it's possible for messages to arrive
	// at the interpretation layer in a different order than they went into the log

	interpretChan <- Message{Id: id, Raw: b}
}

func interpret(c <-chan Message) {

}
