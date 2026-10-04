package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Fista6k/disClone/internal/groups"
)

type GroupHub struct {
	register   chan GroupSubscription
	unregister chan GroupSubscription
	broadcast  chan GroupBroadcast

	groups map[int64]map[int64]*Client

	service *groups.GroupService
}

type GroupSubscription struct {
	Client  *Client
	GroupID int64
}

type GroupBroadcast struct {
	GroupID int64
	Message []byte
}

func NewGroupHub(service *groups.GroupService) *GroupHub {
	return &GroupHub{
		register:   make(chan GroupSubscription),
		unregister: make(chan GroupSubscription),
		broadcast:  make(chan GroupBroadcast),
		groups:     make(map[int64]map[int64]*Client),

		service: service,
	}
}

func (h *GroupHub) Run() {
	for {
		select {
		case subscription := <-h.register:
			group := h.groups[subscription.GroupID]

			if group == nil {
				group = make(map[int64]*Client)
				h.groups[subscription.GroupID] = group
			}

			group[subscription.Client.UserID] = subscription.Client
		case subscription := <-h.unregister:
			if group, ok := h.groups[subscription.GroupID]; ok {
				delete(h.groups[subscription.GroupID], subscription.Client.UserID)

				if len(group) == 0 {
					delete(h.groups, subscription.GroupID)
				}
			}
		case message := <-h.broadcast:
			if group, ok := h.groups[message.GroupID]; ok {
				for _, client := range group {
					client.send <- message.Message
				}
			}
		}
	}
}

func (h *GroupHub) SendMessageToGroup(request WSMessage, c *Client) {
	message, err := h.service.CreateGroupMessage(
		context.Background(),
		request.GroupID,
		c.UserID,
		request.Content,
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	response := groups.GroupMessageResponse{
		ID:        message.ID,
		AuthorID:  message.AuthorID,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}

	encoded, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	h.broadcast <- GroupBroadcast{
		GroupID: request.GroupID,
		Message: encoded,
	}
}
