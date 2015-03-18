package handlers

import (
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/server"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
    "sync/atomic"
)

type websocketHandler struct {
	messages  <-chan []byte
	keepAlive time.Duration
	logger    *gosteno.Logger
    totalMessagesSent int64
}

func NewWebsocketHandler(m <-chan []byte, keepAlive time.Duration, logger *gosteno.Logger) *websocketHandler {
	return &websocketHandler{messages: m, keepAlive: keepAlive, logger: logger}
}

func (h *websocketHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("websocket handler: ServeHTTP entered with request %v", r)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(*http.Request) bool { return true },
	}
	ws, err := upgrader.Upgrade(rw, r, nil)

	if err != nil {
		h.logger.Debugf("websocket handler: Not a websocket handshake: %s", err.Error())
		return
	}
	defer ws.Close()
	defer ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Time{})
	keepAliveExpired := make(chan struct{})
	clientWentAway := make(chan struct{})

	// TODO: remove this loop (but keep ws.ReadMessage()) once we retire support in the cli for old style keep alives
	go func() {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				close(clientWentAway)
				return
			}
		}
	}()

	go func() {
		server.NewKeepAlive(ws, h.keepAlive).Run()
		close(keepAliveExpired)
		h.logger.Debugf("websocket handler: Connection from %s timed out", r.RemoteAddr)
	}()

	for {
		select {
		case <-clientWentAway:
			return
		case <-keepAliveExpired:
			return
		case message, ok := <-h.messages:
			if !ok {
				return
			}
            atomic.AddInt64(&h.totalMessagesSent, 1)
			err = ws.WriteMessage(websocket.BinaryMessage, message)
			if err != nil {
				h.logger.Debugf("websocket handler: Error writing to websocket: %s", err.Error())
				return
			}
		}
	}
}

func (h *websocketHandler) GetTotalMessagesSent() int64{
    return h.totalMessagesSent
}
