package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/Fista6k/disClone/internal"
	"github.com/Fista6k/disClone/internal/auth"
	"github.com/Fista6k/disClone/internal/users"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Info("Can't found .env file")
	}
}

func main() {
	storage, err := internal.ConnToStorage()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	usersRepo := users.NewUserRepository(storage)
	authService := auth.NewAuthService(usersRepo)
	authHandler := auth.NewAuthHandler(authService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/login", authHandler.Login)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
