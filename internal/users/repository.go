package users

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Fista6k/disClone/internal"
	"github.com/Fista6k/disClone/internal/domain"
)

type IUserRepository interface {
	GetByUsername(username string) (*User, error)
	CreateUser(user *User) error
	GetByEmail(email string) (*User, error)
	CreateRefreshToken(tokenHash string, userId int64) error
	GetRefreshToken(hashedToken string) (*RefreshToken, error)
	RevokeRefreshToken(tokenId int64) error
}

func NewUserRepository(storage *internal.Storage) *UserRepository {
	return &UserRepository{
		storage: storage,
	}
}

type UserRepository struct {
	storage *internal.Storage
}

func (r *UserRepository) GetByUsername(username string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at
		FROM users
		WHERE username = $1;
	`

	var user User
	err := r.storage.DB.QueryRow(query, username).Scan(&user.Id, &user.Username, &user.Email, &user.PasswordHash, &user.Created_at)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		} else {
			return nil, err
		}
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at
		FROM users
		WHERE email = $1;
	`

	var user User
	err := r.storage.DB.QueryRow(query, email).Scan(&user.Id, &user.Username, &user.Email, &user.PasswordHash, &user.Created_at)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		} else {
			return nil, err
		}
	}

	return &user, err
}

func (r *UserRepository) CreateUser(user *User) error {
	query := `
		INSERT INTO users (username, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	err := r.storage.DB.QueryRow(query, user.Username, user.Email, user.PasswordHash, user.Created_at).Scan(&user.Id)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) CreateRefreshToken(token string, userId int64) error {
	query := `
		INSERT INTO refresh_tokens (refresh_token_hash, user_id, expired_at, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var tokenId int

	err := r.storage.DB.QueryRow(query, token, userId, time.Now().Add(time.Hour*720), time.Now()).Scan(&tokenId)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) GetRefreshToken(hashedToken string) (*RefreshToken, error) {
	query := `
		SELECT id, refresh_token_hash, user_id, expired_at, created_at, is_revoked
		FROM refresh_tokens
		WHERE refresh_token_hash = $1;
	`

	var refreshToken RefreshToken

	err := r.storage.DB.QueryRow(query, hashedToken).Scan(
		&refreshToken.Id,
		&refreshToken.RefreshTokenHash,
		&refreshToken.UserId,
		&refreshToken.ExpiredAt,
		&refreshToken.CreatedAt,
		&refreshToken.IsRevoked,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRefreshTokenNotFound
		}
		return nil, err
	}

	return &refreshToken, nil

}

func (r *UserRepository) RevokeRefreshToken(tokenId int64) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = TRUE
		WHERE id = $1;
	`

	res, err := r.storage.DB.Exec(query, tokenId)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrRefreshTokenNotFound
	}

	return nil
}
