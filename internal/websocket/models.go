package websocket

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
	hub  *Hub
}

type Hub struct {
	register   chan *Client
	unregister chan *Client
	clients    map[*Client]bool
	broadcast  chan []byte
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			delete(h.clients, client)

		case message := <-h.broadcast:
			for client := range h.clients {
				client.send <- message
			}
		}
	}
}

func (c *Client) Read() {
	defer func() {
		c.conn.Close()
		c.hub.unregister <- c
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		c.hub.broadcast <- message
	}
}

func (c *Client) Write() {
	defer func() {
		c.conn.Close()
		c.hub.unregister <- c
	}()

	for {
		message, ok := <-c.send
		if !ok {
			return
		}

		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}
