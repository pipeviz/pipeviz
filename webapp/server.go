package webapp

import (
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/sdboyer/pipeviz/broker"
	"github.com/sdboyer/pipeviz/represent"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

var (
	assetDir = filepath.Join(defaultBase("github.com/sdboyer/pipeviz/webapp"), "assets")
	jsDir    = filepath.Join(defaultBase("github.com/sdboyer/pipeviz/webapp"), "src")
	tmplDir  = filepath.Join(defaultBase("github.com/sdboyer/pipeviz/webapp"), "tmpl")
)

var (
	// TODO crappily hardcoded, for now
	brokerListen broker.GraphReceiver
	latestGraph  represent.CoreGraph
)

func init() {
	// Subscribe to the master broker and store latest locally as it comes
	brokerListen = broker.Get().Subscribe()
	// FIXME spawning a goroutine in init() used to be crappy, is it still?
	go func() {
		for g := range brokerListen {
			latestGraph = g
		}
	}()
}

// Creates a Goji *web.Mux that can act as the http muxer for the frontend app.
func NewMux() *web.Mux {
	m := web.New()

	m.Use(middleware.Logger)
	m.Get("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetDir))))
	m.Get("/js/*", http.StripPrefix("/js/", http.FileServer(http.Dir(jsDir))))
	m.Get("/", WebRoot)

	return m
}

func WebRoot(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		Title string
	}{
		Title: "pipeviz",
	}

	// TODO we start by kitchen sink-ing - just sending back everything.
	//vertices := latestGraph.VerticesWith()

	t, err := template.ParseFiles(filepath.Join(tmplDir, "index.html"))
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, vars)
}
