package users

import (
	"database/sql"

	"github.com/Fista6k/disClone/internal"
	"github.com/Fista6k/disClone/internal/domain"
)

type IUserRepository interface {
	GetByUsername(username string) (*User, error)
	CreateUser(user *User) error
	GetByEmail(email string) (*User, error)
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
