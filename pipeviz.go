package main

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/sdboyer/pipeviz/persist"
	gjs "github.com/xeipuuv/gojsonschema"
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

var masterSchema *gjs.Schema

func main() {
	src, err := ioutil.ReadFile("./schema.json")
	if err != nil {
		panic(err.Error())
	}

	masterSchema, err = gjs.NewSchema(gjs.NewStringLoader(string(src)))
	if err != nil {
		panic(err.Error())
	}

	go interpret(interpretChan)

	m := web.New()
	m.Put("/environment", handle)

	graceful.ListenAndServe("127.0.0.1:9132", m)
}

func handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// FIXME limit this, prevent overflowing memory from a malicious message
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// FIXME send back an error
	}

	result, err := masterSchema.Validate(gjs.NewStringLoader(string(b)))
	if err != nil {
		// FIXME send back an error
	}

	if result.Valid() {
		id := persist.Append(b, r.RemoteAddr)

		// super-sloppy write back to client, but does the trick
		// TODO not writing the id back at all might make things simpler
		w.WriteHeader(202) // use 202 because it's a little more correct
		w.Write([]byte(strconv.FormatUint(id, 10)))

		// FIXME passing directly from here means it's possible for messages to arrive
		// at the interpretation layer in a different order than they went into the log

		interpretChan <- Message{Id: id, Raw: b}
	} else {
		w.WriteHeader(422)
		var resp []string
		for _, desc := range result.Errors() {
			resp = append(resp, desc.String())
		}
		w.Write([]byte(strings.Join(resp, "\n")))
	}
}

func interpret(c <-chan Message) {

}
