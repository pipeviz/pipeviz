package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	gjs "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/xeipuuv/gojsonschema"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/zenazn/goji/graceful"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/zenazn/goji/web/middleware"
	"github.com/tag1consulting/pipeviz/interpret"
	"github.com/tag1consulting/pipeviz/journal"
	"github.com/tag1consulting/pipeviz/represent"
)

type Ingestor struct {
	journal       journal.LogStore
	schema        *gjs.Schema
	interpretChan chan *journal.Record
	brokerChan    chan represent.CoreGraph
}

// RunHttpIngestor sets up and runs the http listener that receives messages, validates
// them against the provided schema, persists those that pass validation, then sends
// them along to the interpretation layer via the server's interpret channel.
//
// This blocks on the http listening loop, so it should typically be called in its own goroutine.
//
// Closes the provided interpretation channel if/when the http server terminates.
func (s *Ingestor) RunHttpIngestor(addr string) {
	// TODO receive returned error and log if non-nil
	graceful.ListenAndServe(addr, s.buildIngestorMux())
	close(s.interpretChan)
}

func (s *Ingestor) buildIngestorMux() *web.Mux {
	mb := web.New()
	// TODO use more appropriate logger
	mb.Use(middleware.Logger)
	// Add middleware limiting body length to MaxMessageSize
	mb.Use(func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, MaxMessageSize)
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	})

	mb.Post("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		// pulling this out of buffer is incorrect: http://jmoiron.net/blog/crossing-streams-a-love-letter-to-ioreader/
		// but for now gojsonschema leaves us no choice
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			// Too long, or otherwise malformed request body
			// TODO add a body
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
			record, err := s.journal.NewEntry(b, r.RemoteAddr)
			if err != nil {
				w.WriteHeader(500)
				// should we tell the client this?
				w.Write([]byte("Failed to persist message to journal"))
				return
			}

			// super-sloppy write back to client, but does the trick
			w.WriteHeader(202) // use 202 because it's a little more correct
			w.Write([]byte(strconv.FormatUint(record.Index, 10)))

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
	})

	return mb
}

// The main message interpret/merge loop. This receives messages that have been
// validated and persisted, merges them into the graph, then sends the new
// graph along to listeners, workers, etc.
//
// The provided CoreGraph operates as the initial state into which received
// messages will be successively merged.
//
// When the interpret channel is closed (and emptied), this function also closes
// the broker channel.
func (s *Ingestor) Interpret(g represent.CoreGraph) {
	for m := range s.interpretChan {
		// TODO msgid here should be strictly sequential; check, and add error handling if not
		im := interpret.Message{Id: m.Index}
		json.Unmarshal(m.Message, &im)
		g = g.Merge(im)

		s.brokerChan <- g
	}
	close(s.brokerChan)
}
