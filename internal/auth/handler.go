package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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

	err = h.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
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

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	accessToken, err := h.authService.Login(req.Username, req.Password)
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
		"access_token": accessToken,
	})

	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.Header.Get("X-User-Id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, "invalid user id format", http.StatusBadRequest)
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

func validatePassword(password string) error {
	if password == "" {
		return errors.New("Password can't ve empty")
	}

	if len(password) < 8 {
		return errors.New("Password nust contain at least 8 characters")
	}

	return nil
}
