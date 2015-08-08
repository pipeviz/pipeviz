package main

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	gjs "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/xeipuuv/gojsonschema"
	"github.com/tag1consulting/pipeviz/persist"
	"github.com/tag1consulting/pipeviz/persist/item"
	"github.com/tag1consulting/pipeviz/represent"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

type Server struct {
	journal       persist.LogStore
	schema        gjs.Schema
	interpretChan chan *item.Log
	brokerChan    chan represent.CoreGraph
}

func (s *Server) RunHttpIngestor(addr string) {
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
			item := &item.Log{RemoteAddr: []byte(r.RemoteAddr), Message: b}
			// Index of message gets written by the LogStore
			err := s.journal.Append(item)

			// super-sloppy write back to client, but does the trick
			w.WriteHeader(202) // use 202 because it's a little more correct
			w.Write([]byte(strconv.Itoa(id)))

			// FIXME passing directly from here means it's possible for messages to arrive
			// at the interpretation layer in a different order than they went into the log
			// ...especially if go scheduler changes become less cooperative https://groups.google.com/forum/#!topic/golang-nuts/DbmqfDlAR0U (...?)

			s.interpretChan <- message{Id: item.Index, Raw: b}
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

	graceful.ListenAndServe(addr, mb)
	close(s.interpretChan)
}
