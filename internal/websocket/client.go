package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/coder/websocket"
)

const (
	PrivateMessageType = "private_message"
	GroupMessageType   = "group_message"

	PingInterval = 30 * time.Second
	PingTimeout  = 5 * time.Second
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
			c.handleReadError(err)
			return
		}

		var request WSMessage

		err = json.Unmarshal(data, &request)
		if err != nil {
			c.sendError("invalid JSON")
			continue
		}

		switch request.Type {
		case PrivateMessageType:
			c.personalHub.SendMessageToUser(request, c)
		case GroupMessageType:
			c.groupHub.SendMessageToGroup(request, c)
		default:
			c.sendError("unknown mesage type")
			continue
		}
	}
}

func (c *Client) Write() {
	ticker := time.NewTicker(PingInterval)
	defer func() {
		c.conn.CloseNow()
		ticker.Stop()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			err := c.conn.Write(ctx, websocket.MessageText, message)

			cancel()

			if err != nil {
				return
			}
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), PingTimeout)

			err := c.conn.Ping(ctx)

			cancel()

			if err != nil {
				return
			}
		}
	}
}

func (c *Client) handleReadError(err error) {
	var closeErr *websocket.CloseError

	if errors.As(err, &closeErr) {
		switch closeErr.Code {
		case websocket.StatusNormalClosure,
			websocket.StatusGoingAway:
			return
		case websocket.StatusMessageTooBig:
			slog.Error(
				"message was too big",
				"user_id", c.UserID,
			)
		default:
			slog.Error(
				"websocket closed",
				"user_id", c.UserID,
				"code", closeErr.Code,
				"reason", closeErr.Reason,
			)
		}

		return
	}

	slog.Error(
		"websocket read error",
		"user_id", c.UserID,
		"err", err,
	)
}

func (c *Client) sendError(message string) {
	response := WSError{
		Type:    "error",
		Message: message,
	}

	data, err := json.Marshal(response)
	if err != nil {
		slog.Error(
			"failed to marshal websocket error",
			"err", err,
		)
	}

	select {
	case c.send <- data:
	default:
		slog.Error(
			"send buffer is full",
			"user_id", c.UserID,
		)
	}
}
