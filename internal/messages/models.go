package messages

import "time"

type Message struct {
	ID          int64     `json:"id"`
	AuthorID    int64     `json:"author_id"`
	Content     string    `json:"content"`
	RecipientID int64     `json:"recipient_id"`
	CreatedAt   time.Time `json:"created_at"`
}
