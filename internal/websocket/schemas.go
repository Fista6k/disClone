package websocket

type WSMessage struct {
	Type        string `json:"type"`
	Content     string `json:"content"`
	RecipientID int64  `json:"recipient_id,omitempty"`
	GroupID     int64  `json:"group_id,omitempty"`
}

type WSError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
