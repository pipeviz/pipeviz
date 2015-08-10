package main

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/pflag"
	gjs "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/xeipuuv/gojsonschema"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/zenazn/goji/graceful"
	"github.com/tag1consulting/pipeviz/broker"
	"github.com/tag1consulting/pipeviz/interpret"
	"github.com/tag1consulting/pipeviz/journal"
	"github.com/tag1consulting/pipeviz/journal/boltdb"
	"github.com/tag1consulting/pipeviz/represent"
	"github.com/tag1consulting/pipeviz/webapp"
)

// Pipeviz has two fully separated HTTP ports - one for input into the logic
// machine, and one for graph data consumption. This is done primarily
// because security/firewall concerns are completely different, and having
// separate ports makes it much easier to implement separate policies.
// Differing semantics are a contributing, but lesser consideration.
const (
	DefaultIngestionPort = 2309 // 2309, because Cayte
	DefaultAppPort       = 8008
	MaxMessageSize       = 5 << 20 // Max input message size is 5MB
)

var (
	bindAll *bool   = pflag.BoolP("bind-all", "b", false, "Listen on all interfaces. Applies both to ingestor and webapp.")
	dbPath  *string = pflag.StringP("data-dir", "d", ".", "The base directory to use for persistent storage.")
)

func main() {
	src, err := ioutil.ReadFile("./schema.json")
	if err != nil {
		panic(err.Error())
	}

	// The master JSON schema used for validating all incoming messages
	masterSchema, err := gjs.NewSchema(gjs.NewStringLoader(string(src)))
	if err != nil {
		panic(err.Error())
	}

	// Channel to receive persisted messages from HTTP workers. 1000 cap to allow
	// some wiggle room if there's a sudden burst of messages and the interpreter
	// gets behind.
	interpretChan := make(chan *journal.Record, 1000)

	pflag.Parse()
	var listenAt string
	if *bindAll == false {
		listenAt = "127.0.0.1:"
	} else {
		listenAt = ":"
	}

	j, err := boltdb.NewBoltStore(*dbPath + "/journal.bolt")
	if err != nil {
		panic(err.Error())
	}

	// Restore the graph from the journal (or start from nothing if journal is empty)
	// TODO move this down to after ingestor is started
	g, err := restoreGraph(j)
	if err != nil {
		panic(err.Error())
	}

	// Kick off fanout on the master/singleton graph broker. This will bridge between
	// the state machine and the listeners interested in the machine's state.
	brokerChan := make(chan represent.CoreGraph, 0)
	broker.Get().Fanout(brokerChan)
	brokerChan <- g

	srv := &Ingestor{
		journal:       j,
		schema:        masterSchema,
		interpretChan: interpretChan,
		brokerChan:    brokerChan,
	}

	// Kick off the http message ingestor.
	// TODO let config/params control address
	go srv.RunHttpIngestor(listenAt + strconv.Itoa(DefaultIngestionPort))

	// Kick off the intermediary interpretation goroutine that receives persisted
	// messages from the ingestor, merges them into the state graph, then passes
	// them along to the graph broker.
	go srv.Interpret(g)

	// And finally, kick off the webapp.
	// TODO let config/params control address
	go RunWebapp(listenAt + strconv.Itoa(DefaultAppPort))

	// Block on goji's graceful waiter, allowing the http connections to shut down nicely.
	// FIXME using this should be unnecessary if we're crash-only
	graceful.Wait()
}

// RunWebapp runs the pipeviz http frontend webapp on the specified address.
//
// This blocks on the http listening loop, so it should typically be called in its own goroutine.
func RunWebapp(addr string) {
	mf := webapp.NewMux()
	graceful.ListenAndServe(addr, mf)
}

// Rebuilds the graph from the extant entries in a journal.
func restoreGraph(j journal.LogStore) (represent.CoreGraph, error) {
	g := represent.NewGraph()

	var item *journal.Record
	tot, err := j.Count()
	if err != nil {
		// journal failed to report a count for some reason, bail out
		return g, err
	} else if tot > 0 {
		// we manually iterate to the count because we assume that any messages
		// that come in while we do this processing will be queued elsewhere.
		for i := uint64(1); i-1 < tot; i++ {
			item, err = j.Get(i)
			if err != nil {
				// TODO returning out here could end us up somwehere weird
				return g, err
			}
			msg := interpret.Message{Id: item.Index}
			json.Unmarshal(item.Message, &msg)
			g = g.Merge(msg)
		}
	}

	return g, nil
}
