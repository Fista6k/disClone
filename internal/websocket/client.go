package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coder/websocket"
)

type Client struct {
	UserID int64
	conn   *websocket.Conn
	send   chan []byte
	hub    *Hub
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

		message, err := c.hub.service.CreateMessage(
			context.Background(),
			c.UserID,
			request.RecipientID,
			request.Content,
		)
		if err != nil {
			fmt.Println(err.Error())
			return
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
