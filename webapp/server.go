package webapp

import (
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

var (
	assetDir = filepath.Join(defaultBase("github.com/sdboyer/pipeviz/webapp"), "assets")
	jsDir    = filepath.Join(defaultBase("github.com/sdboyer/pipeviz/webapp"), "src")
	tmplDir  = filepath.Join(defaultBase("github.com/sdboyer/pipeviz/webapp"), "tmpl")
)

// Creates a Goji *web.Mux that can act as the http muxer for the frontend app.
func NewMux() *web.Mux {
	m := web.New()

	m.Use(middleware.Logger)
	m.Get("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetDir))))
	m.Get("/js/*", http.StripPrefix("/js/", http.FileServer(http.Dir(jsDir))))
	m.Get("/", RootEntry)

	return m
}

func RootEntry(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		Title string
	}{
		Title: "pipeviz",
	}

	t, err := template.ParseFiles(filepath.Join(tmplDir, "index.html"))
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, vars)
}
