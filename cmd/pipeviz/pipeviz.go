package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	log "github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/spf13/pflag"
	gjs "github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/xeipuuv/gojsonschema"
	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/zenazn/goji/graceful"
	"github.com/pipeviz/pipeviz/broker"
	"github.com/pipeviz/pipeviz/ingest"
	"github.com/pipeviz/pipeviz/mlog"
	"github.com/pipeviz/pipeviz/mlog/boltdb"
	"github.com/pipeviz/pipeviz/mlog/mem"
	"github.com/pipeviz/pipeviz/represent"
	"github.com/pipeviz/pipeviz/schema"
	"github.com/pipeviz/pipeviz/types/system"
	"github.com/pipeviz/pipeviz/webapp"
)

// Pipeviz uses two separate HTTP ports - one for input into the logic
// machine, and one for graph data consumption. This is done primarily
// because security/firewall concerns are completely different, and having
// separate ports makes it much easier to implement separate policies.
// Differing semantics are a contributing, but lesser consideration.
const (
	DefaultIngestionPort int = 2309 // 2309, because Cayte
	DefaultAppPort           = 8008
	MaxMessageSize           = 5 << 20 // Max input message size is 5MB
)

var (
	version    = "dev"
	vflag      = pflag.BoolP("version", "v", false, "Print version")
	bindAll    = pflag.BoolP("bind-all", "b", false, "Listen on all interfaces. Applies both to ingestor and webapp.")
	dbPath     = pflag.StringP("data-dir", "d", ".", "The base directory to use for all persistent storage.")
	useSyslog  = pflag.Bool("syslog", false, "Write log output to syslog.")
	ingestKey  = pflag.String("ingest-key", "", "Path to an x509 key to use for TLS on the ingestion port. If no cert is provided, unsecured HTTP will be used.")
	ingestCert = pflag.String("ingest-cert", "", "Path to an x509 certificate to use for TLS on the ingestion port. If key is provided, will try to find a certificate of the same name plus .crt extension.")
	webappKey  = pflag.String("webapp-key", "", "Path to an x509 key to use for TLS on the webapp port. If no cert is provided, unsecured HTTP will be used.")
	webappCert = pflag.String("webapp-cert", "", "Path to an x509 certificate to use for TLS on the webapp port. If key is provided, will try to find a certificate of the same name plus .crt extension.")
	mlstore    = pflag.String("mlog-storage", "bolt", "Storage backend to use for the message log. Valid options: 'memory' or 'bolt'. Defaults to bolt.")
	publicDir  = pflag.String("webapp-dir", "webapp/public", "Path to the 'public' directory containing javascript application files.")
)

func main() {
	pflag.Parse()
	if *vflag {
		fmt.Println("pipeviz version", version)
		return
	}

	setUpLogging()

	src, err := schema.Master()
	if err != nil {
		log.WithFields(log.Fields{
			"system": "main",
			"err":    err,
		}).Fatal("Could not locate master schema file, exiting")
	}

	// The master JSON schema used for validating all incoming messages
	masterSchema, err := gjs.NewSchema(gjs.NewStringLoader(string(src)))
	if err != nil {
		log.WithFields(log.Fields{
			"system": "main",
			"err":    err,
		}).Fatal("Error while creating a schema object from the master schema file, exiting")
	}

	// Channel to receive persisted messages from HTTP workers. 1000 cap to allow
	// some wiggle room if there's a sudden burst of messages and the interpreter
	// gets behind.
	interpretChan := make(chan *mlog.Record, 1000)

	var listenAt string
	if *bindAll == false {
		listenAt = "127.0.0.1:"
	} else {
		listenAt = ":"
	}

	var j mlog.Store
	switch *mlstore {
	case "bolt":
		j, err = boltdb.NewBoltStore(*dbPath + "/mlog.bolt")
		if err != nil {
			log.WithFields(log.Fields{
				"system": "main",
				"err":    err,
			}).Fatal("Error while setting up bolt mlog storage, exiting")
		}
	case "memory":
		j = mem.NewMemStore()
	default:
		log.WithFields(log.Fields{
			"system":  "main",
			"storage": *mlstore,
		}).Fatal("Invalid storage type requested for mlog, exiting")
	}

	// Restore the graph from the mlog (or start from nothing if mlog is empty)
	// TODO move this down to after ingestor is started
	g, err := restoreGraph(j)
	if err != nil {
		log.WithFields(log.Fields{
			"system": "main",
			"err":    err,
		}).Fatal("Error while rebuilding the graph from the mlog")
	}

	// Kick off fanout on the master/singleton graph broker. This will bridge between
	// the state machine and the listeners interested in the machine's state.
	brokerChan := make(chan system.CoreGraph, 0)
	broker.Get().Fanout(brokerChan)
	brokerChan <- g

	srv := ingest.New(j, masterSchema, interpretChan, brokerChan, MaxMessageSize)

	// Kick off the http message ingestor.
	// TODO let config/params control address
	go func() {
		if *ingestKey != "" && *ingestCert == "" {
			*ingestCert = *ingestKey + ".crt"
		}
		err := srv.RunHTTPIngestor(listenAt+strconv.Itoa(DefaultIngestionPort), *ingestKey, *ingestCert)
		if err != nil {
			log.WithFields(log.Fields{
				"system": "main",
				"err":    err,
			}).Fatal("Error while starting the ingestion http server")
		}
	}()

	// Kick off the intermediary interpretation goroutine that receives persisted
	// messages from the ingestor, merges them into the state graph, then passes
	// them along to the graph broker.
	go srv.Interpret(g)

	// And finally, kick off the webapp.
	frontend := webapp.New(broker.Get().Subscribe(), broker.Get().Unsubscribe, make(chan struct{}), j.Get, version)
	// TODO let config/params control address
	if *webappKey != "" && *webappCert == "" {
		*webappCert = *webappKey + ".crt"
	}
	go frontend.ListenAndServe(listenAt+strconv.Itoa(DefaultAppPort), *publicDir, *webappKey, *webappCert)

	// Block on goji's graceful waiter, allowing the http connections to shut down nicely.
	// FIXME using this should be unnecessary if we're crash-only
	graceful.Wait()
}

// Rebuilds the graph from the extant entries in a mlog.
func restoreGraph(j mlog.Store) (system.CoreGraph, error) {
	g := represent.NewGraph()

	var item *mlog.Record
	tot, err := j.Count()
	if err != nil {
		// mlog failed to report a count for some reason, bail out
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
			msg := &ingest.Message{}
			json.Unmarshal(item.Message, msg)
			g = g.Merge(item.Index, msg.UnificationForm())
		}
	}

	return g, nil
}
