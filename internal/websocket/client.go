package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coder/websocket"
)

const (
	PrivateMessageType = "private_message"
	GroupMessageType   = "group_message"
)

type Client struct {
	UserID      int64
	conn        *websocket.Conn
	send        chan []byte
	personalHub *PersonalHub
	groupHub    *GroupHub

	groups map[int64]struct{}
}

func (c *Client) Read() {
	defer func() {
		c.personalHub.unregister <- c
		c.groupHub.unregister <- c
		c.conn.CloseNow()
	}()

	for {
		_, data, err := c.conn.Read(context.Background())
		if err != nil {
			return
		}

		var request WSMessage

		err = json.Unmarshal(data, &request)
		if err != nil {
			return
		}

		fmt.Println(request.Type)

		switch request.Type {
		case PrivateMessageType:
			c.personalHub.SendMessageToUser(request, c)
		case GroupMessageType:
			c.groupHub.SendMessageToGroup(request, c)
		default:
			return
		}
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
