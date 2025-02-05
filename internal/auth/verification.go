package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"
)

const (
	verificationTokenExpiry = 24 * time.Hour
	tokenLength             = 32
)

func GenerateVerificationToken(db *sql.DB, userID int) (string, error) {
	// Generate random token
	b := make([]byte, tokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(b)

	// Set expiration
	expiresAt := time.Now().Add(verificationTokenExpiry)

	// Update user record
	_, err := db.Exec(`
        UPDATE users 
        SET verification_token = $1, 
            verification_expires_at = $2
        WHERE id = $3
    `, token, expiresAt, userID)

	if err != nil {
		return "", err
	}

	return token, nil
}

func VerifyEmail(db *sql.DB, token string) error {
	// Find and verify token
	var userID int
	var expiresAt time.Time

	err := db.QueryRow(`
        SELECT id, verification_expires_at 
        FROM users 
        WHERE verification_token = $1 
        AND email_verified = false
    `, token).Scan(&userID, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("invalid or expired verification token")
		}
		return err
	}

	if time.Now().After(expiresAt) {
		return errors.New("verification token has expired")
	}

	// Mark email as verified
	_, err = db.Exec(`
        UPDATE users 
        SET email_verified = true,
            verification_token = NULL,
            verification_expires_at = NULL
        WHERE id = $1
    `, userID)

	return err
}
