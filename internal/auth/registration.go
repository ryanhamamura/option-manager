package auth

import (
	"database/sql"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type RegistrationInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

func ValidateRegistration(input RegistrationInput) error {
	if !strings.Contains(input.Email, "@") {
		return errors.New("invalid email address")
	}
	if len(input.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if strings.TrimSpace(input.FirstName) == "" {
		return errors.New("first name is required")
	}
	if strings.TrimSpace(input.LastName) == "" {
		return errors.New("last name is required")
	}
	return nil
}

func RegisterUser(db *sql.DB, input RegistrationInput) (*User, error) {
	// Check if email already exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", input.Email).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Insert new user
	var user User
	err = db.QueryRow(`
        INSERT INTO users (email, password_hash, first_name, last_name)
        VALUES ($1, $2, $3, $4)
        RETURNING id, email, password_hash, first_name, last_name
    `, input.Email, string(hashedPassword), input.FirstName, input.LastName).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
