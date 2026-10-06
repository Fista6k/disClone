package groups

import (
	"context"
	"time"

	"github.com/Fista6k/disClone/internal/domain"
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

func (s *GroupService) GetMyGroups(ctx context.Context, userID int64) ([]Group, error) {
	return s.repo.GetMyGroups(ctx, userID)
}

func (s *GroupService) AddNewMembers(ctx context.Context, group_id int64, members_ids []int64) error {
	return s.repo.AddNewMembers(ctx, group_id, members_ids)
}

func (s *GroupService) GetGroupInfo(ctx context.Context, groupID int64) (*GroupInfoResponse, error) {
	group, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	members, err := s.repo.GetMembersByGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	response := &GroupInfoResponse{
		ID:        group.ID,
		Name:      group.Name,
		OwnerID:   group.OwnerID,
		CreatedAt: group.CreatedAt,
		Members:   make([]GroupMemberResponse, 0, len(members)),
	}

	for _, member := range members {
		response.Members = append(response.Members, GroupMemberResponse{
			UserID:   member.userID,
			JoinedAt: member.joinedAt,
		})
	}

	return response, nil
}

func (s *GroupService) DeleteMember(ctx context.Context, groupID int64, memberID int64, userID int64) error {
	group, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return err
	}

	if group.OwnerID != userID {
		return domain.ErrNotGroupOwner
	}

	return s.repo.DeleteMember(ctx, groupID, memberID)
}

func (s *GroupService) GetGroupHistory(ctx context.Context, groupID int64) ([]GroupMessageResponse, error) {
	messages, err := s.repo.GetMessageHistory(ctx, groupID)
	if err != nil {
		return nil, err
	}

	response := make([]GroupMessageResponse, 0, len(messages))
	for _, message := range messages {
		response = append(response, GroupMessageResponse{
			ID:        message.ID,
			AuthorID:  message.AuthorID,
			Content:   message.Content,
			CreatedAt: message.CreatedAt,
		})
	}

	return response, nil
}

func (s *GroupService) CreateGroupMessage(ctx context.Context, groupID int64, authorID int64, content string) (*GroupMessage, error) {
	group, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group == nil {
		return nil, domain.ErrGroupNotFound
	}

	message := &GroupMessage{
		GroupID:   groupID,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: time.Now(),
	}
	return s.repo.CreateGroupMessage(ctx, message)
}
