package groups

import (
	"context"
	"time"
)

type GroupService struct {
	repo IGroupRepository
}

func NewGroupService(groupRepo IGroupRepository) *GroupService {
	return &GroupService{
		repo: groupRepo,
	}
}

func (s *GroupService) CreateGroup(ctx context.Context, name string, owner_id int64) error {
	group := &Group{
		Name:      name,
		OwnerID:   owner_id,
		CreatedAt: time.Now(),
	}

	return s.repo.CreateGroup(ctx, group)
}

func (s *GroupService) GetGroups(ctx context.Context, owner_id int64) ([]Group, error) {
	return s.repo.GetGroups(ctx, owner_id)
}

func (s *GroupService) AddNewMembers(ctx context.Context, group_id int64, members_ids []int64) error {
	return s.repo.AddNewMembers(ctx, group_id, members_ids)
}
