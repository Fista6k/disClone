package websocket

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/Fista6k/disClone/internal/auth"
	"github.com/Fista6k/disClone/internal/httpapi"
	"github.com/coder/websocket"
)

type WebSocketHandler struct {
	PersonalHub    *PersonalHub
	GroupHub       *GroupHub
	CallHub        *CallHub
	originPatterns []string
	ticketMu       sync.Mutex
	tickets        map[string]wsTicket
	userTickets    map[int64]string
}

const wsTicketLifetime = 30 * time.Second

type wsTicket struct {
	userID    int64
	expiresAt time.Time
}

func NewHandler(personalHub *PersonalHub, groupHub *GroupHub, callHub *CallHub, originPatterns []string) *WebSocketHandler {
	return &WebSocketHandler{
		PersonalHub:    personalHub,
		GroupHub:       groupHub,
		CallHub:        callHub,
		originPatterns: originPatterns,
		tickets:        make(map[string]wsTicket),
		userTickets:    make(map[int64]string),
	}
}

func (h *WebSocketHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserIdFromContext(r.Context())
	if err != nil {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		httpapi.WriteInternalError(w, err)
		return
	}
	ticket := base64.RawURLEncoding.EncodeToString(random)
	expiresAt := time.Now().Add(wsTicketLifetime)

	h.ticketMu.Lock()
	now := time.Now()
	for oldTicket, entry := range h.tickets {
		if now.After(entry.expiresAt) || entry.userID == userID {
			delete(h.tickets, oldTicket)
			if h.userTickets[entry.userID] == oldTicket {
				delete(h.userTickets, entry.userID)
			}
		}
	}
	h.tickets[ticket] = wsTicket{userID: userID, expiresAt: expiresAt}
	h.userTickets[userID] = ticket
	h.ticketMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"ticket":     ticket,
		"expires_in": int(wsTicketLifetime.Seconds()),
	}); err != nil {
		httpapi.WriteInternalError(w, err)
	}
}

func (h *WebSocketHandler) HandleConn(w http.ResponseWriter, r *http.Request) {
	userId, ok := h.consumeTicket(r.URL.Query().Get("ticket"))
	if !ok {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid WebSocket ticket required")
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
		callHub:     h.CallHub,
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

func (h *WebSocketHandler) consumeTicket(ticket string) (int64, bool) {
	h.ticketMu.Lock()
	defer h.ticketMu.Unlock()

	entry, ok := h.tickets[ticket]
	if !ok {
		return 0, false
	}
	delete(h.tickets, ticket)
	if h.userTickets[entry.userID] == ticket {
		delete(h.userTickets, entry.userID)
	}
	if time.Now().After(entry.expiresAt) {
		return 0, false
	}
	return entry.userID, true
}
