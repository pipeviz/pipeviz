package webapp

import (
	"net/http"

	"github.com/zenazn/goji/web"
)

// Creates a Goji *web.Mux that can act as the http muxer for the frontend app.
func NewMux() *web.Mux {
	m := web.New()

	m.Get("/", MainEntry)

	return m
}

func MainEntry(w http.ResponseWriter, r *http.Request) {

}
