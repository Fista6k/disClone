package groups

import "time"

type Group struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	OwnerID   int64     `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type GroupMember struct {
	groupID  int64
	userID   int64
	joinedAt time.Time
}

type GroupMessage struct {
	ID        int64
	GroupID   int64
	AuthorID  int64
	Content   string
	CreatedAt time.Time
}
