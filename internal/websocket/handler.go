package websocket

import (
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var origin = os.Getenv("OriginWebSocket")

type WebSocketHandler struct {
}

// TODO : Implement origin logic
func (h *WebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Connection") != "Upgrade" {
		http.Error(w, "Not a websocket request", http.StatusBadRequest)
		return
	}

	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(w, r, nil)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	for {
		mesageType, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		err = conn.WriteMessage(mesageType, message)
		if err != nil {
			return
		}
	}
}
