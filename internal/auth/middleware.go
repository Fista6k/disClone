package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Fista6k/disClone/internal/domain"
	"github.com/Fista6k/disClone/internal/httpapi"
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
				httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
				return
			}
			tokenStr := strs[1]

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
				return []byte(secret), nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

			if err != nil {
				httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid or expired access token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid access token")
				return
			}

			sub, ok := claims["sub"].(float64)
			if !ok {
				httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid access token")
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
				httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
				return
			}

			_, err = users.GetUserById(r.Context(), id)
			if err != nil {
				if errors.Is(err, domain.ErrUserNotFound) {
					httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "user is no longer available")
					return
				}
				httpapi.WriteInternalError(w, err)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
