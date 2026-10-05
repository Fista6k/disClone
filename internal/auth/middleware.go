package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Fista6k/disClone/internal/domain"
	"github.com/Fista6k/disClone/internal/users"
	"github.com/golang-jwt/jwt/v5"
)

type ctxKey int

const (
	keyUserID ctxKey = iota
	keyUser
)

type IUserGetter interface {
	GetUserById(ctx context.Context, userId int64) (*users.User, error)
}

func CheckToken(secret []byte) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearerStr := r.Header.Get("Authorization")

			strs := strings.Fields(bearerStr)
			if len(strs) != 2 || !strings.EqualFold(strs[0], "Bearer") {
				http.Error(w, "invalid auth header", http.StatusUnauthorized)
				return
			}
			tokenStr := strs[1]

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
				return []byte(secret), nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

			if err != nil {
				http.Error(w, "jwt parsing error", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			sub, ok := claims["sub"].(float64)
			if !ok {
				http.Error(w, "invalid token subject", http.StatusUnauthorized)
				return
			}

			userID := int64(sub)

			ctx := context.WithValue(r.Context(), keyUserID, userID)

			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireUser(users IUserGetter) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := UserIdFromContext(r.Context())
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			_, err = users.GetUserById(r.Context(), id)
			if err != nil {
				if errors.Is(err, domain.ErrUserNotFound) {
					http.Error(w, "user not found", http.StatusNotFound)
					return
				}
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
