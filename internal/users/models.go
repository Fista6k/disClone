package users

import "time"

type User struct {
	Id           int64
	Username     string
	Email        string
	PasswordHash string
	Created_at   time.Time
}
