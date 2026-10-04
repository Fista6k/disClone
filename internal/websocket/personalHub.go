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
	clients    map[int64]*Client
	broadcast  chan []byte

	service *messages.MessageService
}

func NewHub(messageService *messages.MessageService) *PersonalHub {
	return &PersonalHub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[int64]*Client),
		broadcast:  make(chan []byte),
		service:    messageService,
	}
}

func (h *PersonalHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.UserID] = client

		case client := <-h.unregister:
			delete(h.clients, client.UserID)
			close(client.send)

		case message := <-h.broadcast:
			for client := range h.clients {
				h.clients[client].send <- message
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

	client, ok := h.clients[request.RecipientID]
	if !ok {
		return
	}

	client.send <- encoded
}
