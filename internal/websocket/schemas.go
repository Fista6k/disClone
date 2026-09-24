package websocket

type SendMessageRequest struct {
	RecipientID int64  `json:"recipientId"`
	Content     string `json:"content"`
}
