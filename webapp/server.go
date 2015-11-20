package webapp

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/unrolled/secure"
	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/zenazn/goji/graceful"
	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/pipeviz/pipeviz/broker"
	"github.com/pipeviz/pipeviz/log"
	"github.com/pipeviz/pipeviz/mlog"
	"github.com/pipeviz/pipeviz/represent"
	"github.com/pipeviz/pipeviz/represent/q"
	"github.com/pipeviz/pipeviz/types/system"
	"github.com/pipeviz/pipeviz/version"
)

// Count of active websocket clients (for expvars)
var clientCount int64

const (
	// Time allowed to write data to the client.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second
	// Send pings to client with this period; less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// Server acts as an HTTP server, and contains the necessary state to serve the viz frontend.
type Server struct {
	latest     system.CoreGraph
	unsub      func(broker.GraphReceiver)
	receiver   broker.GraphReceiver
	cancel     <-chan struct{}
	mlogGetter mlog.RecordGetter
}

// New creates a new webapp server, ready to be kicked off.
func New(receiver broker.GraphReceiver, unsub func(broker.GraphReceiver), cancel <-chan struct{}, f mlog.RecordGetter) *Server {
	return &Server{
		receiver:   receiver,
		unsub:      unsub,
		latest:     represent.NewGraph(),
		cancel:     cancel,
		mlogGetter: f,
	}
}

// ListenAndServe initiates the webapp http listener.
//
// This blocks on the http listening loop, so it should typically be called in its own goroutine.
func (s *Server) ListenAndServe(addr, pubdir, key, cert string, showVersion bool) {
	mf := web.New()
	useTLS := key != "" && cert != ""

	mf.Use(log.NewHTTPLogger("webapp"))

	if useTLS {
		sec := secure.New(secure.Options{
			AllowedHosts:         nil,                                             // TODO allow a way to declare these
			SSLRedirect:          false,                                           // we have just one port to work with, so an internal redirect can't work
			SSLTemporaryRedirect: false,                                           // Use 301, not 302
			SSLProxyHeaders:      map[string]string{"X-Forwarded-Proto": "https"}, // list of headers that indicate we're using TLS (which would have been set by TLS-terminating proxy)
			STSSeconds:           315360000,                                       // 1yr HSTS time, as is generally recommended
			STSIncludeSubdomains: false,                                           // don't include subdomains; it may not be correct in general case TODO allow config
			STSPreload:           false,                                           // can't know if this is correct for general case TODO allow config
			FrameDeny:            false,                                           // pipeviz is exactly the kind of thing where embedding is appropriate
			ContentTypeNosniff:   true,                                            // shouldn't be an issue for pipeviz, but doesn't hurt...probably?
			BrowserXssFilter:     false,                                           // really shouldn't be necessary for pipeviz
		})

		mf.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err := sec.Process(w, r)

				// If there was an error, do not continue.
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"system": "webapp",
						"err":    err,
					}).Warn("Error from security middleware, dropping request")
					return
				}

				h.ServeHTTP(w, r)
			})
		})
	}

	// If showing version, add a middleware to do it automatically for everything
	if showVersion {
		mf.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Server", version.Version())
				h.ServeHTTP(w, r)
			})
		})
	}

	mf.Get("/sock", s.openSocket)
	mf.Get("/message/:mid", s.getMessage)
	mf.Get("/*", http.StripPrefix("/", http.FileServer(http.Dir(pubdir))))

	mf.Compile()

	// kick off a goroutine to grab the latest graph and listen for cancel
	go func() {
		var g system.CoreGraph
		for {
			select {
			case <-s.cancel:
				s.unsub(s.receiver)
				return

			case g = <-s.receiver:
				s.latest = g
			}
		}
	}()

	var err error
	if useTLS {
		err = graceful.ListenAndServeTLS(addr, cert, key, mf)
	} else {
		err = graceful.ListenAndServe(addr, mf)
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"system": "webapp",
			"err":    err,
		}).Fatal("ListenAndServe returned with an error")
	}

	// TODO allow returning err
}

func graphToJSON(g system.CoreGraph) ([]byte, error) {
	var vertices []interface{}
	for _, v := range g.VerticesWith(q.Qbv(system.VTypeNone)) {
		vertices = append(vertices, v.Flat())
	}

	// TODO use something that lets us write to a reusable byte buffer instead
	return json.Marshal(struct {
		ID       uint64        `json:"id"`
		Vertices []interface{} `json:"vertices"`
	}{
		ID:       g.MsgID(),
		Vertices: vertices,
	})
}

func (s *Server) getMessage(c web.C, w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(c.URLParams["mid"], 10, 64)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	if s.mlogGetter == nil {
		http.Error(w, "Could not access mlog storage", 500)
		return
	}

	rec, err := s.mlogGetter(id)
	if err != nil {
		// TODO Might be something other than not found, but oh well for now
		http.Error(w, http.StatusText(404), 404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, _ = w.Write(rec.Message)
}
