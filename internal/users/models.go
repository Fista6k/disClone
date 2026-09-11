package users

import "time"

type User struct {
	Id           int64
	Username     string
	Email        string
	PasswordHash string
	Created_at   time.Time
}

type RefreshToken struct {
	Id               int64
	RefreshTokenHash string
	UserId           int64
	ExpiredAt        time.Time
	CreatedAt        time.Time
	IsRevoked        bool
}
