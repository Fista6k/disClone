package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Fista6k/disClone/internal/domain"
	"github.com/Fista6k/disClone/internal/httpapi"
)

type AuthHandler struct {
	authService *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{
		authService: service,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Body == nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body is required")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	if req.Username == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "username is required")
		return
	}

	if req.Email == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "email is required")
		return
	}

	err := validatePassword(req.Password)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_password", err.Error())
		return
	}

	err = h.authService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		httpapi.WriteDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered completely",
	}); err != nil {
		httpapi.WriteInternalError(w, err)
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	ctx := r.Context()

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	accessToken, refreshToken, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		httpapi.WriteDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})

	if err != nil {
		httpapi.WriteInternalError(w, err)
		return
	}
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userId, err := UserIdFromContext(r.Context())
	if err != nil {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]any{
		"user_id": userId,
	})
	if err != nil {
		httpapi.WriteInternalError(w, err)
		return
	}
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	if req.RefreshToken == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "refresh_token is required")
		return
	}

	dbToken, err := h.authService.repo.GetRefreshToken(ctx, HashRefreshToken(req.RefreshToken))
	if err != nil {
		if errors.Is(err, domain.ErrRefreshTokenNotFound) {
			httpapi.WriteError(w, http.StatusUnauthorized, "invalid_refresh_token", "refresh token is invalid or expired")
			return
		}
		httpapi.WriteInternalError(w, err)
		return
	}

	hashedToken := HashRefreshToken(req.RefreshToken)

	if hashedToken != dbToken.RefreshTokenHash {
		httpapi.WriteError(w, http.StatusUnauthorized, "invalid_refresh_token", "refresh token is invalid or expired")
		return
	}

	if dbToken.IsRevoked {
		httpapi.WriteError(w, http.StatusUnauthorized, "invalid_refresh_token", "refresh token is invalid or expired")
		return
	}

	if dbToken.ExpiredAt.Before(time.Now()) {

		httpapi.WriteError(w, http.StatusUnauthorized, "invalid_refresh_token", "refresh token is invalid or expired")
		return
	}

	accessToken, refreshToken, err := h.authService.Refresh(ctx, dbToken.UserId, dbToken.Id)
	if err != nil {
		httpapi.WriteInternalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
	if err != nil {
		httpapi.WriteInternalError(w, err)
		return
	}
}

func validatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}

	if len(password) < 8 {
		return errors.New("password must contain at least 8 characters")
	}

	return nil
}

func UserIdFromContext(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value(keyUserID).(int64)
	if !ok {
		return 0, errors.New("user ID not found in context")
	}

	return userID, nil
}
