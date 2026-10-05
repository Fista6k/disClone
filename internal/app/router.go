package app

import (
	"net/http"

	"github.com/Fista6k/disClone/internal/auth"
)

func (a *App) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/login", a.Auth.Login)
	mux.HandleFunc("POST /api/v1/auth/register", a.Auth.Register)
	mux.HandleFunc("POST /api/v1/refresh", a.Auth.Refresh)

	needAuth := func(h http.HandlerFunc) http.Handler {
		return auth.CheckToken(a.Secret)(h)
	}

	//TODO find usage for this func or just delete
	//requireUser := func(h http.HandlerFunc) http.Handler {
	//	return auth.RequireUser(a.Users)(h)
	//}

	mux.Handle("GET /api/v1/me", needAuth(a.Auth.GetMe))
	mux.Handle("GET /api/v1/ws", needAuth(a.WS.HandleConn))
	mux.Handle("GET /api/v1/messages/{user_id}", needAuth(a.Messages.GetConversation))
	mux.Handle("POST /api/v1/groups", needAuth(a.Groups.CreateGroup))
	mux.Handle("GET /api/v1/groups", needAuth(a.Groups.GetMyGroups))
	mux.Handle("POST /api/v1/groups/{group_id}/members", needAuth(a.Groups.AddNewMembers))
	mux.Handle("GET /api/v1/groups/{group_id}", needAuth(a.Groups.GetGroupInfo))
	mux.Handle("DELETE /api/v1/groups/{group_id}/members/{member_id}", needAuth(a.Groups.DeleteMember))
	mux.Handle("GET /api/v1/groups/{group_id}/messages", needAuth(a.Groups.GetGroupHistory))

	return mux
}
