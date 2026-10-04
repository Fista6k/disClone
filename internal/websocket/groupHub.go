package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Fista6k/disClone/internal/groups"
)

type GroupHub struct {
	register              chan *Client
	unregister            chan *Client
	broadcast             chan GroupBroadcast
	removeClientFromGroup chan GroupMembershipChange
	addClientToGroup      chan GroupMembershipChange

	groups  map[int64]map[int64]*Client
	clients map[int64]*Client

	service *groups.GroupService
}

type GroupMembershipChange struct {
	UserID  int64
	GroupID int64
}

type GroupBroadcast struct {
	GroupID int64
	Message []byte
}

func NewGroupHub(service *groups.GroupService) *GroupHub {
	return &GroupHub{
		register:              make(chan *Client),
		unregister:            make(chan *Client),
		removeClientFromGroup: make(chan GroupMembershipChange),
		addClientToGroup:      make(chan GroupMembershipChange),
		broadcast:             make(chan GroupBroadcast),
		groups:                make(map[int64]map[int64]*Client),
		clients:               make(map[int64]*Client),
		service:               service,
	}
}

func (h *GroupHub) Run() {
	for {
		select {
		case client := <-h.register:
			fmt.Println("register:", client.UserID, len(h.clients))
			h.clients[client.UserID] = client

			fmt.Println("online", len(h.clients))

			for groupID := range client.groups {
				fmt.Println("INITIAL GROUP:", groupID)
				group := h.groups[groupID]

				if group == nil {
					group = make(map[int64]*Client)
					h.groups[groupID] = group
				}

				group[client.UserID] = client
			}
		case client := <-h.unregister:
			delete(h.clients, client.UserID)

			for groupID := range client.groups {
				if _, ok := h.groups[groupID]; ok {
					delete(h.groups[groupID], client.UserID)

					if len(h.groups[groupID]) == 0 {
						delete(h.groups, groupID)
					}
				}
			}
		case subscription := <-h.removeClientFromGroup:
			if group, ok := h.groups[subscription.GroupID]; ok {
				delete(h.groups[subscription.GroupID], subscription.UserID)

				if len(group) == 0 {
					delete(h.groups, subscription.GroupID)
				}
			}

			if client, ok := h.clients[subscription.UserID]; ok {
				delete(client.groups, subscription.GroupID)
			}
		case subscription := <-h.addClientToGroup:
			fmt.Println(
				"ADD EVENT:",
				subscription.UserID,
				subscription.GroupID,
			)
			client, ok := h.clients[subscription.UserID]
			if !ok {
				fmt.Println("USER IS NOT ONLINE:", subscription.UserID)
				continue
			}

			group := h.groups[subscription.GroupID]

			if group == nil {
				group = make(map[int64]*Client)
				h.groups[subscription.GroupID] = group
			}

			group[subscription.UserID] = h.clients[subscription.UserID]
			client.groups[subscription.GroupID] = struct{}{}

			fmt.Println(
				"GROUP", subscription.GroupID,
				"SIZE:", len(group),
			)
		case message := <-h.broadcast:
			if group, ok := h.groups[message.GroupID]; ok {
				for _, client := range group {
					fmt.Println("SEND TO:", client.UserID)
					client.send <- message.Message
				}

				fmt.Println("GROUP SIZE:", len(group))
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

func (h *GroupHub) RemoveClientFromGroup(userID int64, groupID int64) {
	fmt.Println("REMOVE CLIENT FROM GROUP:", userID, groupID)
	h.removeClientFromGroup <- GroupMembershipChange{
		UserID:  userID,
		GroupID: groupID,
	}
}

func (h *GroupHub) AddClientToGroup(userID int64, groupID int64) {
	fmt.Println("ADD CLIENT TO GROUP:", userID, groupID)
	h.addClientToGroup <- GroupMembershipChange{
		UserID:  userID,
		GroupID: groupID,
	}
}
