package messages

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Fista6k/disClone/internal/auth"
	"github.com/Fista6k/disClone/internal/users"
)

type MessageHandler struct {
	messageService *MessageService
	userService    *users.UserService
}

func NewMessageHandler(service *MessageService, userService *users.UserService) *MessageHandler {
	return &MessageHandler{
		messageService: service,
		userService:    userService,
	}
}

func (h *MessageHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	authorID, err := auth.UserIdFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	recipientIDStr := r.PathValue("user_id")
	if recipientIDStr == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	recipientID, err := strconv.ParseInt(recipientIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user_id format", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUserById(r.Context(), recipientID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	messages, err := h.messageService.GetConversation(r.Context(), authorID, recipientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := MessagesHistoryResponse{
		Messages: messages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
