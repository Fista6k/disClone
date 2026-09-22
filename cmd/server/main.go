package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Fista6k/disClone/internal"
	"github.com/Fista6k/disClone/internal/auth"
	"github.com/Fista6k/disClone/internal/users"
	"github.com/Fista6k/disClone/internal/websocket"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Info("Can't found .env file")
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage, err := internal.ConnToStorage(ctx)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	usersRepo := users.NewUserRepository(storage)
	authService := auth.NewAuthService(usersRepo)
	authHandler := auth.NewAuthHandler(authService)
	websoketHandler := websocket.WebSocketHandler{}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/login", authHandler.Login)
	mux.Handle("GET /api/v1/me", auth.CheckToken(http.HandlerFunc(authHandler.GetMe)))
	mux.HandleFunc("POST /api/v1/refresh", authHandler.Refresh)
	mux.HandleFunc("GET /api/v1/ws", websoketHandler.Handle)

	go func() {
		if err := http.ListenAndServe(":8080", mux); err != nil && err == http.ErrServerClosed {
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	_ = storage.DB.Close()
}
