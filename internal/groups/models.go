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
	groupID   int64
	authorID  int64
	content   string
	createdAt time.Time
}
