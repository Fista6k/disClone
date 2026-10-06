package messages

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Fista6k/disClone/internal/auth"
	"github.com/Fista6k/disClone/internal/domain"
	"github.com/Fista6k/disClone/internal/httpapi"
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
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	recipientIDStr := r.PathValue("user_id")
	if recipientIDStr == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "user_id is required")
		return
	}

	recipientID, err := strconv.ParseInt(recipientIDStr, 10, 64)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "user_id must be a valid integer")
		return
	}

	user, err := h.userService.GetUserById(r.Context(), recipientID)
	if errors.Is(err, domain.ErrUserNotFound) || user == nil {
		httpapi.WriteError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	if err != nil {
		httpapi.WriteInternalError(w, err)
		return
	}

	messages, err := h.messageService.GetConversation(r.Context(), authorID, recipientID)
	if err != nil {
		httpapi.WriteInternalError(w, err)
		return
	}

	data := MessagesHistoryResponse{
		Messages: messages,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		httpapi.WriteInternalError(w, err)
	}
}
