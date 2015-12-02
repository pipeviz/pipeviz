package ingest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/unrolled/secure"
	gjs "github.com/xeipuuv/gojsonschema"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/pipeviz/pipeviz/log"
	"github.com/pipeviz/pipeviz/mlog"
	"github.com/pipeviz/pipeviz/types/system"
)

// Ingestor brings together the required components to run a pipeviz ingestion HTTP server.
type Ingestor struct {
	mlog           mlog.Store
	schema         *gjs.Schema
	interpretChan  chan *mlog.Record
	brokerChan     chan system.CoreGraph
	maxMessageSize int64
}

// New creates a new pipeviz ingestor mux, ready to be kicked off.
func New(j mlog.Store, s *gjs.Schema, ic chan *mlog.Record, bc chan system.CoreGraph, max int64) *Ingestor {
	return &Ingestor{
		mlog:           j,
		schema:         s,
		interpretChan:  ic,
		brokerChan:     bc,
		maxMessageSize: max,
	}
}

// RunHTTPIngestor sets up and runs the http listener that receives messages, validates
// them against the provided schema, persists those that pass validation, then sends
// them along to the interpretation layer via the server's interpret channel.
//
// This blocks on the http listening loop, so it should typically be called in its own goroutine.
//
// Closes the provided interpretation channel if/when the http server terminates.
func (s *Ingestor) RunHTTPIngestor(addr, key, cert string) error {
	var err error
	mb := web.New()
	useTLS := key != "" && cert != ""

	// TODO use more appropriate logger
	mb.Use(log.NewHTTPLogger("ingestor"))

	if useTLS {
		sec := secure.New(secure.Options{
			AllowedHosts:         nil,                                             // TODO allow a way to declare these
			SSLRedirect:          false,                                           // we have just one port to work with, so an internal redirect can't work
			SSLTemporaryRedirect: false,                                           // Use 301, not 302
			SSLHost:              "",                                              // use the same host to redirect from HTTP to HTTPS
			SSLProxyHeaders:      map[string]string{"X-Forwarded-Proto": "https"}, // list of headers that indicate we're using TLS (which would have been set by TLS-terminating proxy)
			STSSeconds:           315360000,                                       // 1yr HSTS time, as is generally recommended
			STSIncludeSubdomains: false,                                           // don't include subdomains; it may not be correct in general case TODO allow config
			STSPreload:           false,                                           // can't know if this is correct for general case TODO allow config
			FrameDeny:            true,                                            // could never make sense on the ingestion side
			ContentTypeNosniff:   true,                                            // shouldn't be an issue for pipeviz, but doesn't hurt...probably?
			BrowserXssFilter:     false,                                           // really shouldn't be necessary for pipeviz
		})

		// TODO consider using a custom handler in order to log errors
		mb.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err := sec.Process(w, r)

				// If there was an error, do not continue.
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"system": "webapp",
						"err":    err,
					}).Warn("Error from security middleware, dropping message")
					return
				}

				h.ServeHTTP(w, r)
			})
		})
	}

	// Middleware to limit body length to MaxMessageSize
	mb.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, s.maxMessageSize)
			h.ServeHTTP(w, r)
		})
	})

	mb.Post("/", s.handleMessage)

	if useTLS {
		err = graceful.ListenAndServeTLS(addr, cert, key, mb)
	} else {
		err = graceful.ListenAndServe(addr, mb)
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"system": "ingestor",
			"err":    err,
		}).Error("Ingestion httpd failed to start")
	}

	close(s.interpretChan)
	return err
}

func (s *Ingestor) handleMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// pulling this out of buffer is incorrect: http://jmoiron.net/blog/crossing-streams-a-love-letter-to-ioreader/
	// but for now gojsonschema leaves us no choice
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// Too long, or otherwise malformed request body
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	result, err := s.schema.Validate(gjs.NewStringLoader(string(b)))
	if err != nil {
		// Malformed JSON, likely
		// TODO add a body
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	if result.Valid() {
		// Index of message gets written by the LogStore
		record, err := s.mlog.NewEntry(b, r.RemoteAddr)
		if err != nil {
			w.WriteHeader(500)
			// should we tell the client this?
			w.Write([]byte("Failed to persist message to mlog"))
			return
		}

		// super-sloppy write back to client, but does the trick
		w.WriteHeader(202) // use 202 because it's a little more correct
		_, err = w.Write([]byte(strconv.FormatUint(record.Index, 10)))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"system": "ingestor",
				"err":    err,
			}).Warn("Failed to write msgid back to client; continuing anyway.")
		}

		// FIXME passing directly from here means it's possible for messages to arrive
		// at the interpretation layer in a different order than they went into the log
		// ...especially if go scheduler changes become less cooperative https://groups.google.com/forum/#!topic/golang-nuts/DbmqfDlAR0U (...?)

		s.interpretChan <- record
	} else {
		// Invalid results, so write back 422 for malformed entity
		w.WriteHeader(422)
		var resp []string
		for _, desc := range result.Errors() {
			resp = append(resp, desc.String())
		}
		w.Write([]byte(strings.Join(resp, "\n")))
	}
}

// Interpret is the main message interpret/merge loop. It receives messages that
// have been validated and persisted, merges them into the graph, then sends the
// new graph along to listeners, workers, etc.
//
// The provided CoreGraph operates as the initial state into which received
// messages will be successively merged.
//
// When the interpret channel is closed (and emptied), this function also closes
// the broker channel.
func (s *Ingestor) Interpret(g system.CoreGraph) {
	for m := range s.interpretChan {
		// TODO msgid here should be strictly sequential; check, and add error handling if not
		im := Message{}
		json.Unmarshal(m.Message, &im)
		g = g.Merge(m.Index, im.UnificationForm())

		s.brokerChan <- g
	}
	close(s.brokerChan)
}
