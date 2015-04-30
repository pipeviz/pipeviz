package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/sdboyer/pipeviz/interpret"
	"github.com/sdboyer/pipeviz/persist"
	"github.com/sdboyer/pipeviz/represent"
	"github.com/sdboyer/pipeviz/webapp"
	gjs "github.com/xeipuuv/gojsonschema"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
)

type message struct {
	Id  int
	Raw []byte
}

// Channel for handling persisted messages. 1000 cap to allow some wiggle room
// if there's a sudden burst of messages
var interpretChan chan message = make(chan message, 1000)

var masterSchema *gjs.Schema

var masterGraph represent.CoreGraph

func main() {
	src, err := ioutil.ReadFile("./schema.json")
	if err != nil {
		panic(err.Error())
	}

	masterSchema, err = gjs.NewSchema(gjs.NewStringLoader(string(src)))
	if err != nil {
		panic(err.Error())
	}

	// Initialize the central graph object. For now, we're just controlling this
	// by having it be an unexported pkg var, but this will need to change.
	masterGraph = represent.NewGraph()

	// TODO hardcoded 8008 for http frontend
	mf := webapp.NewMux()
	graceful.ListenAndServe("127.0.0.1:8008", mf)

	go interp(interpretChan)

	// Pipeviz has two fully separated HTTP ports - one for input into the logic
	// machine, and one for graph data consumption. This is done primarily
	// because security/firewall concerns are completely different, and having
	// separate ports makes it much easier to implement separate policies.
	// Differing semantics are a contributing, but lesser consideration.
	mb := web.New()
	mb.Post("/", handle)

	// 2309, because Cayte
	graceful.ListenAndServe("127.0.0.1:2309", mb)
	graceful.Wait()
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
		w.WriteHeader(202) // use 202 because it's a little more correct
		w.Write([]byte(strconv.Itoa(id)))

		// FIXME passing directly from here means it's possible for messages to arrive
		// at the interpretation layer in a different order than they went into the log
		// if go scheduler changes become less cooperative https://groups.google.com/forum/#!topic/golang-nuts/DbmqfDlAR0U (...?)

		interpretChan <- message{Id: id, Raw: b}
	} else {
		w.WriteHeader(422)
		var resp []string
		for _, desc := range result.Errors() {
			resp = append(resp, desc.String())
		}
		w.Write([]byte(strings.Join(resp, "\n")))
	}
}

// The main message-interpretation loop. This receives messages that have been
// validated and persisted, merges them into the graph, then sends the new
// graph along to listeners, workers, etc.
func interp(c <-chan message) {
	for m := range c {
		im := interpret.Message{Id: m.Id}
		json.Unmarshal(m.Raw, &im)
		masterGraph = masterGraph.Merge(im)

		go dispatchLatest(masterGraph)
	}
}

// Dispatches the latest version of the graph out to all listeners.
func dispatchLatest(g represent.CoreGraph) {
	// TODO so hacke much import
	webapp.GraphListen <- masterGraph
}
