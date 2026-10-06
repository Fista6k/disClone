package app

import (
	"github.com/Fista6k/disClone/internal"
	"github.com/Fista6k/disClone/internal/auth"
	"github.com/Fista6k/disClone/internal/groups"
	"github.com/Fista6k/disClone/internal/messages"
	"github.com/Fista6k/disClone/internal/users"
	"github.com/Fista6k/disClone/internal/websocket"
)

type App struct {
	Storage  *internal.Storage
	Auth     *auth.AuthHandler
	Users    *users.UserService
	Groups   *groups.GroupHandler
	WS       *websocket.WebSocketHandler
	Messages *messages.MessageHandler
	Secret   []byte
}

func New(storage *internal.Storage, secret []byte, originPatterns []string) *App {
	userRepo := users.NewUserRepository(storage)
	msgRepo := messages.NewMessageRepo(storage)
	groupRepo := groups.NewGroupRepository(storage)

	userService := users.NewUserService(userRepo)
	groupService := groups.NewGroupService(groupRepo)
	authService := auth.NewAuthService(userRepo, secret)
	messageService := messages.NewMessageService(msgRepo)

	personalHub := websocket.NewHub(messageService)
	groupHub := websocket.NewGroupHub(groupService)
	go personalHub.Run()
	go groupHub.Run()

	return &App{
		Storage:  storage,
		Auth:     auth.NewAuthHandler(authService),
		Users:    userService,
		Groups:   groups.NewGroupHandler(groupService, groupHub),
		WS:       websocket.NewHandler(personalHub, groupHub, originPatterns),
		Messages: messages.NewMessageHandler(messageService, userService),
		Secret:   secret,
	}
}

func (a *App) Close() error {
	return a.Storage.DB.Close()
}
