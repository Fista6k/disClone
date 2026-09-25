package messages

type MessageResponse struct {
	Messages []*Message `json:"messages"`
}
