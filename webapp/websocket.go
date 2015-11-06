package webapp

import (
	"net/http"
	"time"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/gorilla/websocket"
	"github.com/pipeviz/pipeviz/broker"
	"github.com/pipeviz/pipeviz/types/system"
)

var (
	// gorilla websocket upgrader
	upgrader = websocket.Upgrader{ReadBufferSize: 256, WriteBufferSize: 8192}
)

func (s *WebAppServer) openSocket(w http.ResponseWriter, r *http.Request) {
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
	go s.wsWriter(ws)
	s.wsReader(ws)
	clientCount--
}

func (s *WebAppServer) wsReader(ws *websocket.Conn) {
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

func (s *WebAppServer) wsWriter(ws *websocket.Conn) {
	graphIn := broker.Get().Subscribe()
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		broker.Get().Unsubscribe(graphIn)
		ws.Close()
	}()

	// write the current graph state first, before entering loop
	graphToSock(ws, s.latest)
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
