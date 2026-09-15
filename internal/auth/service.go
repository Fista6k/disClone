package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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

func (s *AuthService) Register(ctx context.Context, username, email, password string) error {
	_, err := s.repo.GetByUsername(ctx, username)
	if err == nil {
		return domain.ErrUserExists
	} else {
		if err != domain.ErrUserNotFound {
			return err
		}
	}

	_, err = s.repo.GetByEmail(ctx, username)
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

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, string, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return "", "", err
	}

	match, err := argon2id.ComparePasswordAndHash(password, user.PasswordHash)
	if err != nil {
		return "", "", err
	}

	if !match {
		return "", "", domain.ErrIncorrectPassword
	}

	claims := jwt.MapClaims{
		"sub": user.Id,
		"exp": time.Now().Add(time.Minute * 15).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	hashedRefreshToken := HashRefreshToken(refreshToken)

	err = s.repo.CreateRefreshToken(ctx, hashedRefreshToken, user.Id)
	if err != nil {
		return "", "", err
	}

	t, err := token.SignedString([]byte(secret))
	return t, refreshToken, err
}

func (s *AuthService) Refresh(ctx context.Context, userId, tokenId int64) (string, string, error) {
	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	err = s.repo.RotateRefreshToken(ctx, tokenId, HashRefreshToken(refreshToken), userId)
	if err != nil {
		return "", "", err
	}

	claims := jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(time.Minute * 15).Unix(),
		"iat": time.Now().Unix(),
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
