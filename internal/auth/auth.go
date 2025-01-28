package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
}

type Session struct {
	ID        string
	UserID    int
	ExpiresAt time.Time
}

func AuthenticateUser(db *sql.DB, email, password string) (*User, error) {
	user := &User{}
	query := `SELECT id, email, password_hash, first_name, last_name 
			  FROM users WHERE email = $1`

	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

func CreateSession(db *sql.DB, userID int, rememberMe bool) (*Session, error) {
	// Generate random session ID
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	sessionID := base64.URLEncoding.EncodeToString(b)

	// Set expiration based on remember-me
	var expiresAt time.Time
	if rememberMe {
		expiresAt = time.Now().Add(30 * 24 * time.Hour) // 30 days
	} else {
		expiresAt = time.Now().Add(24 * time.Hour) // 24 hours
	}

	// Store session in database
	_, err := db.Exec(`
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, sessionID, userID, expiresAt)

	if err != nil {
		return nil, err
	}

	return &Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}, nil
}

func GetSession(db *sql.DB, sessionID string) (*Session, error) {
	session := &Session{}
	err := db.QueryRow(`
		SELECT id, user_id, expires_at 
		FROM sessions 
		WHERE id = $1 AND expires_at > NOW()
	`, sessionID).Scan(&session.ID, &session.UserID, &session.ExpiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("session not found or expired")
		}
		return nil, err
	}

	return session, nil
}

func DeleteSession(db *sql.DB, sessionID string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE id = $1", sessionID)
	return err
}
