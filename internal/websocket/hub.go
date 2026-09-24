package websocket

import "github.com/Fista6k/disClone/internal/messages"

type Hub struct {
	register   chan *Client
	unregister chan *Client
	clients    map[int64]*Client
	broadcast  chan []byte

	service *messages.MessageService
}

func NewHub(messageService *messages.MessageService) *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[int64]*Client),
		broadcast:  make(chan []byte),
		service:    messageService,
	}
}

func (h *Hub) Run() {
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

func (h *Hub) SendMessageToUser(message []byte, recipientID int64) {
	client, ok := h.clients[recipientID]
	if !ok {
		return
	}

	client.send <- message
}
