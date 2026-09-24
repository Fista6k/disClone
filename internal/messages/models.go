package messages

import "time"

type Message struct {
	ID          int64
	AuthorID    int64
	Content     string
	RecipientID int64
	CreatedAt   time.Time
}
