// internal/repository/interfaces.go
package repository

import (
	"context"
	"time"
)

// User represents the user model
type User struct {
	ID                 int
	Email              string
	PasswordHash       string
	FirstName          string
	LastName           string
	EmailVerified      bool
	VerificationToken  *string
	VerificationExpiry *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Session represents the session model
type Session struct {
	ID        string
	UserID    int
	ExpiresAt time.Time
	CreatedAt time.Time
}

// UserRepository defines all user-related database operations
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id int) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByVerificationToken(ctx context.Context, token string) (*User, error)
	UpdateVerificationStatus(ctx context.Context, userID int, verified bool) error
	SetVerificationToken(ctx context.Context, userID int, token string, expiry time.Time) error
}

// SessionRepository defines all session-related database operations
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	FindByID(ctx context.Context, id string) (*Session, error)
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
}

// Repository holds all repositories
type Repository struct {
	User    UserRepository
	Session SessionRepository
}
