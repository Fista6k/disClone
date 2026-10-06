package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Fista6k/disClone/internal/messages"
)

type PersonalHub struct {
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	events     chan recipientRequest

	clients map[int64]*Client

	service *messages.MessageService
}

type recipientRequest struct {
	userID  int64
	message []byte
}

func NewHub(messageService *messages.MessageService) *PersonalHub {
	return &PersonalHub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		events:     make(chan recipientRequest),

		clients: make(map[int64]*Client),

		service: messageService,
	}
}

func (h *PersonalHub) Run() {
	for {
		select {
		case client := <-h.register:
			if oldClient, ok := h.clients[client.UserID]; ok {
				oldClient.Close()
				h.removeClient(oldClient)
			}

			h.clients[client.UserID] = client

		case client := <-h.unregister:
			h.removeClient(client)

		case message := <-h.broadcast:
			for client := range h.clients {
				h.clients[client].enqueue(message)
			}
		case request := <-h.events:
			client, ok := h.clients[request.userID]
			if ok {
				client.enqueue(request.message)
			}
		}
	}
}

func (h *PersonalHub) SendMessageToUser(request WSMessage, c *Client) {
	message, err := c.personalHub.service.CreateMessage(
		context.Background(),
		c.UserID,
		request.RecipientID,
		request.Content,
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	response := messages.MessageResponse{
		ID:          message.ID,
		AuthorID:    message.AuthorID,
		Content:     message.Content,
		RecipientID: message.RecipientID,
		CreatedAt:   message.CreatedAt,
	}

	encoded, err := json.Marshal(response)
	if err != nil {
		return
	}

	c.personalHub.events <- recipientRequest{
		userID:  request.RecipientID,
		message: encoded,
	}
}

func (h *PersonalHub) removeClient(client *Client) {
	if h.clients[client.UserID] == client {
		delete(h.clients, client.UserID)
	}
}
