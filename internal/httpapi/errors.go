package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Fista6k/disClone/internal/domain"
)

type ErrorResponse struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: APIError{Code: code, Message: message}})
}

func WriteDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrGroupNotFound), errors.Is(err, domain.ErrRefreshTokenNotFound):
		WriteError(w, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, domain.ErrUserExists), errors.Is(err, domain.ErrUserAlreadyMember):
		WriteError(w, http.StatusConflict, "conflict", "resource already exists")
	case errors.Is(err, domain.ErrIncorrectPassword):
		WriteError(w, http.StatusUnauthorized, "invalid_credentials", "invalid username or password")
	case errors.Is(err, domain.ErrNotGroupOwner):
		WriteError(w, http.StatusForbidden, "forbidden", "you do not have permission to do this")
	case errors.Is(err, domain.ErrNotGroupMember):
		WriteError(w, http.StatusNotFound, "not_found", "group membership not found")
	default:
		slog.Error("request failed", "err", err)
		WriteError(w, http.StatusInternalServerError, "internal", "internal server error")
	}
}

func WriteInternalError(w http.ResponseWriter, err error) {
	slog.Error("request failed", "err", err)
	WriteError(w, http.StatusInternalServerError, "internal", "internal server error")
}
