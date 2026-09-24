package websocket

import (
	"net/http"
	"os"

	"github.com/Fista6k/disClone/internal/auth"
	"github.com/coder/websocket"
)

var origin = os.Getenv("OriginWebSocket")

type WebSocketHandler struct {
	Hub *Hub
}

func NewHandler(hub *Hub) *WebSocketHandler {
	return &WebSocketHandler{
		Hub: hub,
	}
}

// TODO : Implement origin logic
func (h *WebSocketHandler) HandleConn(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		UserID: userId,
		conn:   c,
		send:   make(chan []byte, 256),
		hub:    h.Hub,
	}

	client.hub.register <- client

	go client.Write()
	client.Read()
}
