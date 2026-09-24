package websocket

type SendMessageRequest struct {
	RecipientID int64  `json:"recipient_id"`
	Content     string `json:"content"`
}
