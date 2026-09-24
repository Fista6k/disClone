package auth

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type KeyUserID string

const KeyUserId KeyUserID = "userId"

func CheckToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearerStr := r.Header.Get("Authorization")

		strs := strings.Fields(bearerStr)
		if len(strs) != 2 || !strings.EqualFold(strs[0], "Bearer") {
			http.Error(w, "invalid auth header", http.StatusUnauthorized)
			log.Printf("auth header: %q", bearerStr)
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

		ctx := context.WithValue(r.Context(), KeyUserId, strconv.Itoa(int(sub)))

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
