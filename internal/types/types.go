package types

import "time"

// User represents a user in the system
type User struct {
	ID                 string
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
