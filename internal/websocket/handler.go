package websocket

import (
	"net/http"
	"os"

	"github.com/Fista6k/disClone/internal/auth"
	"github.com/coder/websocket"
)

var origin = os.Getenv("OriginWebSocket")

type WebSocketHandler struct {
	PersonalHub *PersonalHub
	GroupHub    *GroupHub
}

func NewHandler(personalHub *PersonalHub, groupHub *GroupHub) *WebSocketHandler {
	return &WebSocketHandler{
		PersonalHub: personalHub,
		GroupHub:    groupHub,
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

	c.SetReadLimit(64 * 1024)

	client := &Client{
		UserID:      userId,
		conn:        c,
		send:        make(chan []byte, 256),
		personalHub: h.PersonalHub,
		groupHub:    h.GroupHub,
		groups:      make(map[int64]struct{}),
	}

	client.personalHub.register <- client

	groups, err := h.GroupHub.service.GetMyGroups(r.Context(), userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, group := range groups {
		client.groups[group.ID] = struct{}{}
	}

	client.groupHub.register <- client

	go client.Write()
	client.Read()
}
