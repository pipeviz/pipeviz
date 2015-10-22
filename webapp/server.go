package webapp

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/gorilla/websocket"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/tag1consulting/pipeviz/broker"
	"github.com/tag1consulting/pipeviz/journal"
	"github.com/tag1consulting/pipeviz/log"
	"github.com/tag1consulting/pipeviz/represent"
	"github.com/tag1consulting/pipeviz/represent/q"
	"github.com/tag1consulting/pipeviz/types/system"
)

var (
	publicDir = filepath.Join(defaultBase("github.com/tag1consulting/pipeviz/webapp/"), "public")
)

var (
	// Subscribe to the master broker and store latest locally as it comes
	brokerListen = broker.Get().Subscribe()
	// Initially set the latestGraph to a new, empty one to avoid nil pointer
	latestGraph = represent.NewGraph()
	// Count of active websocket clients (for expvars)
	clientCount int64
)

const (
	// Time allowed to write data to the client.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second
	// Send pings to client with this period; less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	// gorilla websocket upgrader
	upgrader = websocket.Upgrader{ReadBufferSize: 256, WriteBufferSize: 8192}
)

func init() {
	// Kick off goroutine to listen on the graph broker and keep our local pointer up to date
	go func() {
		for g := range brokerListen {
			latestGraph = g
		}
	}()
}

// Creates a Goji *web.Mux that can act as the http muxer for the frontend app.
func NewMux() *web.Mux {
	m := web.New()

	m.Use(log.NewHTTPLogger("webapp"))
	m.Get("/sock", openSocket)
	m.Get("/message/:mid", getMessage)
	m.Get("/*", http.StripPrefix("/", http.FileServer(http.Dir(publicDir))))

	return m
}

func graphToJSON(g system.CoreGraph) ([]byte, error) {
	var vertices []interface{}
	for _, v := range g.VerticesWith(q.Qbv(system.VTypeNone)) {
		vertices = append(vertices, v.Flat())
	}

	// TODO use something that lets us write to a reusable byte buffer instead
	return json.Marshal(struct {
		Id       uint64        `json:"id"`
		Vertices []interface{} `json:"vertices"`
	}{
		Id:       g.MsgID(),
		Vertices: vertices,
	})
}

func getMessage(c web.C, w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(c.URLParams["mid"], 10, 64)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	var getter journal.RecordGetter
	var ok bool
	if fun, exists := c.Env["journalGet"]; !exists {
		http.Error(w, "Could not access log store", 500)
		return
	} else if getter, ok = fun.(journal.RecordGetter); !ok {
		http.Error(w, "Could not access log store", 500)
		return
	}

	rec, err := getter(id)
	if err != nil {
		// TODO Might be something other than not found, but oh well for now
		http.Error(w, http.StatusText(404), 404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(rec.Message)
}

func openSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		entry := logrus.WithFields(logrus.Fields{
			"system": "webapp",
			"err":    err,
		})

		if _, ok := err.(websocket.HandshakeError); !ok {
			entry.Error("Error on attempting upgrade to websocket")
		} else {
			entry.Warn("Handshake error on websocket upgrade")
		}
		return
	}

	clientCount++
	go wsWriter(ws)
	wsReader(ws)
	clientCount--
}

func wsReader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	// for now, we ignore all messages from the frontend
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func wsWriter(ws *websocket.Conn) {
	graphIn := broker.Get().Subscribe()
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		broker.Get().Unsubscribe(graphIn)
		ws.Close()
	}()

	// write the current graph state first, before entering loop
	graphToSock(ws, latestGraph)
	var g system.CoreGraph
	for {
		select {
		case <-pingTicker.C:
			// ensure client connection is healthy
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case g = <-graphIn:
			graphToSock(ws, g)
		}
	}
}

func graphToSock(ws *websocket.Conn, g system.CoreGraph) {
	j, err := graphToJSON(g)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"system": "webapp",
			"err":    err,
		}).Error("Error while marshaling graph into JSON for transmission over websocket")
	}

	if j != nil {
		ws.SetWriteDeadline(time.Now().Add(writeWait))
		if err := ws.WriteMessage(websocket.TextMessage, j); err != nil {
			logrus.WithFields(logrus.Fields{
				"system": "webapp",
				"err":    err,
			}).Error("Error while writing graph data to websocket")
			return
		}
	}
}
