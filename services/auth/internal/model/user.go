package model

import "time"

// User is the auth domain entity (Clean Architecture model layer).
type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	Nickname     string
	Role         string // user | author | admin
}
