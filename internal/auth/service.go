package auth

import (
	"os"
	"time"

	"github.com/Fista6k/disClone/internal/domain"
	"github.com/Fista6k/disClone/internal/users"
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
)

var secret = os.Getenv("JWT_SECRET")

func NewAuthService(repo users.IUserRepository) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

type AuthService struct {
	repo users.IUserRepository
}

func (s *AuthService) Register(username, email, password string) error {
	_, err := s.repo.GetByUsername(username)
	if err == nil {
		return domain.ErrUserExists
	} else {
		if err != domain.ErrUserNotFound {
			return err
		}
	}

	_, err = s.repo.GetByEmail(username)
	if err == nil {
		return domain.ErrUserExists
	} else {
		if err != domain.ErrUserNotFound {
			return err
		}
	}

	password_hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return err
	}

	user := &users.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(password_hash),
		Created_at:   time.Now(),
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return "", err
	}

	match, err := argon2id.ComparePasswordAndHash(password, user.PasswordHash)
	if err != nil {
		return "", err
	}

	if !match {
		return "", domain.ErrIncorrectPassword
	}

	claims := jwt.MapClaims{
		"sub": user.Id,
		"exp": time.Now().Add(time.Minute * 15).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}
