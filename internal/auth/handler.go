package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Fista6k/disClone/internal/domain"
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
		http.Error(w, "request body is empty", http.StatusBadRequest)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Username cant be empty", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email cant be empty", http.StatusBadRequest)
		return
	}

	err := validatePassword(req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.authService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			http.Error(w, "user with this username already exists", http.StatusBadRequest)
			return
		}
		http.Error(w, "can't save this user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered completely",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	ctx := r.Context()

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		if err == domain.ErrIncorrectPassword {
			http.Error(w, "invalid password or username", http.StatusBadRequest)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})

	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userId, err := UserIdFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]any{
		"user_id": userId,
	})
	if err != nil {
		http.Error(w, "cant encode json ofr responce", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json format", http.StatusBadRequest)
		return
	}

	dbToken, err := h.authService.repo.GetRefreshToken(ctx, HashRefreshToken(req.RefreshToken))
	if err != nil {
		if err == domain.ErrRefreshTokenNotFound {
			http.Error(w, "n such refresh token", http.StatusUnauthorized)
			return
		}

		log.Println("ogo")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	hashedToken := HashRefreshToken(req.RefreshToken)

	if strings.Compare(hashedToken, dbToken.RefreshTokenHash) != 0 {
		http.Error(w, "invalid auth", http.StatusUnauthorized)
		return
	}

	if dbToken.IsRevoked {
		http.Error(w, "expired", http.StatusUnauthorized)
		return
	}

	if dbToken.ExpiredAt.Before(time.Now()) {

		http.Error(w, "expired", http.StatusUnauthorized)
		return
	}

	accessToken, refreshToken, err := h.authService.Refresh(ctx, dbToken.UserId, dbToken.Id)
	if err != nil {
		log.Println("ogo1")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

func validatePassword(password string) error {
	if password == "" {
		return errors.New("Password can't ve empty")
	}

	if len(password) < 8 {
		return errors.New("Password nust contain at least 8 characters")
	}

	return nil
}

func UserIdFromContext(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value(keyUserID).(int64)
	if !ok {
		return 0, errors.New("user id not found in context")
	}

	return userID, nil
}
