package websocket

import (
	"log/slog"
	"net/http"

	"github.com/Fista6k/disClone/internal/auth"
	"github.com/Fista6k/disClone/internal/httpapi"
	"github.com/coder/websocket"
)

type WebSocketHandler struct {
	PersonalHub    *PersonalHub
	GroupHub       *GroupHub
	originPatterns []string
}

func NewHandler(personalHub *PersonalHub, groupHub *GroupHub, originPatterns []string) *WebSocketHandler {
	return &WebSocketHandler{
		PersonalHub:    personalHub,
		GroupHub:       groupHub,
		originPatterns: originPatterns,
	}
}

func (h *WebSocketHandler) HandleConn(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r.Context())
	if err != nil {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	groups, err := h.GroupHub.service.GetMyGroups(r.Context(), userId)
	if err != nil {
		httpapi.WriteInternalError(w, err)
		return
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: h.originPatterns,
	})
	if err != nil {
		slog.Warn("websocket handshake failed", "user_id", userId, "err", err)
		return
	}

	c.SetReadLimit(64 * 1024)

	client := &Client{
		UserID:      userId,
		conn:        c,
		send:        make(chan []byte, 256),
		done:        make(chan struct{}),
		personalHub: h.PersonalHub,
		groupHub:    h.GroupHub,
		groups:      make(map[int64]struct{}),
	}

	for _, group := range groups {
		client.groups[group.ID] = struct{}{}
	}

	client.personalHub.register <- client
	client.groupHub.register <- client

	go client.Write()
	client.Read()
}
