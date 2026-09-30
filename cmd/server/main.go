package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Fista6k/disClone/internal"
	"github.com/Fista6k/disClone/internal/auth"
	"github.com/Fista6k/disClone/internal/groups"
	"github.com/Fista6k/disClone/internal/messages"
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

	userRepo := users.NewUserRepository(storage)
	userService := users.NewUserService(userRepo)
	authService := auth.NewAuthService(userRepo)
	authHandler := auth.NewAuthHandler(authService)

	messageRepo := messages.NewMessageRepo(storage)
	messageService := messages.NewMessageService(messageRepo)
	messagesHandler := messages.NewMessageHandler(messageService, userService)

	groupRepo := groups.NewGroupRepository(storage)
	groupService := groups.NewGroupService(groupRepo)
	groupHandler := groups.NewGroupHandler(groupService)

	hub := websocket.NewHub(messageService)
	go hub.Run()

	websoketHandler := websocket.NewHandler(hub)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/login", authHandler.Login)
	mux.Handle("GET /api/v1/me", auth.CheckToken(http.HandlerFunc(authHandler.GetMe)))
	mux.HandleFunc("POST /api/v1/refresh", authHandler.Refresh)
	mux.Handle("GET /api/v1/ws", auth.CheckToken(http.HandlerFunc(websoketHandler.HandleConn)))
	mux.Handle("GET /api/v1/messages/{user_id}", auth.CheckToken(http.HandlerFunc(messagesHandler.GetConversation)))
	mux.Handle("POST /api/v1/groups", auth.CheckToken(http.HandlerFunc(groupHandler.CreateGroup)))
	mux.Handle("GET /api/v1/groups", auth.CheckToken(http.HandlerFunc(groupHandler.GetMyGroups)))
	mux.Handle("POST /api/v1/groups/{group_id}/members", auth.CheckToken(http.HandlerFunc(groupHandler.AddNewMembers)))
	mux.Handle("GET /api/v1/groups/{group_id}", auth.CheckToken(http.HandlerFunc(groupHandler.GetGroupInfo)))
	mux.Handle("DELETE /api/v1/groups/{group_id}/members/{member_id}", auth.CheckToken(http.HandlerFunc(groupHandler.DeleteMember)))
	mux.Handle("GET /apiv1/groups/{group_id}/messages", auth.CheckToken(http.HandlerFunc(groupHandler.GetGroupHistory)))

	go func() {
		if err := http.ListenAndServe(":8080", mux); err != nil && err == http.ErrServerClosed {
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	_ = storage.DB.Close()
}
