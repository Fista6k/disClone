package groups

type CreateGroupRequest struct {
	Name string `json:"name"`
}

type AddMembersRequest struct {
	MembersIDs []int64 `json:"members_ids"`
}
