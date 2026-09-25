package users

import "context"

type UserService struct {
	repo IUserRepository
}

func NewUserService(repo IUserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) GetUserById(ctx context.Context, userId int64) (*User, error) {
	return s.repo.GetUserById(ctx, userId)
}
