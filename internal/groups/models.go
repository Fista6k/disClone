package groups

import "time"

type Group struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	OwnerID   int64     `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type GroupMember struct {
	group_id  int64
	user_id   int64
	joined_at int64
}
