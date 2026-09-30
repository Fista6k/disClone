package groups

import "time"

type CreateGroupRequest struct {
	Name string `json:"name"`
}

type AddMembersRequest struct {
	MembersIDs []int64 `json:"members_ids"`
}

type GroupInfoResponse struct {
	ID        int64                 `json:"id"`
	Name      string                `json:"name"`
	OwnerID   int64                 `json:"owner_id"`
	CreatedAt time.Time             `json:"created_at"`
	Members   []GroupMemberResponse `json:"members"`
}

type GroupMemberResponse struct {
	UserID   int64     `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
}

type GroupMessageResponse struct {
	ID        int64     `json:"id"`
	AuthorID  int64     `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
