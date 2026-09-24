package websocket

import (
	"context"
	"encoding/json"

	"github.com/coder/websocket"
)

type Client struct {
	UserID int64
	conn   *websocket.Conn
	send   chan []byte
	hub    *Hub
}

type Hub struct {
	register   chan *Client
	unregister chan *Client
	clients    map[int64]*Client
	broadcast  chan []byte
}

type Message struct {
	AuthorID    int64
	Content     string
	RecipientID int64
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[int64]*Client),
		broadcast:  make(chan []byte),
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

func (c *Client) Read() {
	defer func() {
		c.hub.unregister <- c
		c.conn.CloseNow()
	}()

	for {
		_, data, err := c.conn.Read(context.Background())
		if err != nil {
			return
		}

		var request SendMessageRequest

		err = json.Unmarshal(data, &request)
		if err != nil {
			return
		}

		message := Message{
			AuthorID:    c.UserID,
			Content:     request.Content,
			RecipientID: request.RecipientID,
		}

		encoded, err := json.Marshal(message)
		if err != nil {
			return
		}

		c.hub.SendMessageToUser(encoded, request.RecipientID)
	}
}

func (c *Client) Write() {
	defer c.conn.CloseNow()

	for {
		message, ok := <-c.send
		if !ok {
			return
		}

		err := c.conn.Write(context.Background(), websocket.MessageText, message)
		if err != nil {
			return
		}
	}
}
