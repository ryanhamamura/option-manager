// internal/service/user_service.go
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"option-manager/internal/repository"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo     repository.UserRepository
	emailService *EmailService
}

func NewUserService(userRepo repository.UserRepository, emailService *EmailService) (*UserService, error) {
	if userRepo == nil {
		return nil, fmt.Errorf("user repository is required")
	}
	if emailService == nil {
		return nil, fmt.Errorf("email service is required")
	}
	return &UserService{
		userRepo:     userRepo,
		emailService: emailService,
	}, nil
}

type RegistrationInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

func (s *UserService) RegisterUser(ctx context.Context, input RegistrationInput) (*repository.User, error) {
	if err := s.ValidateRegistration(input); err != nil {
		return nil, err
	}

	// Check if user exists
	existingUser, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking existing user: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create user
	user := &repository.User{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    input.FirstName,
		LastName:     input.LastName,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return user, nil
}

func (s *UserService) GenerateVerificationToken(ctx context.Context, userID int) (string, error) {
	if userID <= 0 {
		return "", errors.New("invalid user ID")
	}

	// Generate random token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(b)

	// Set expiration
	expiry := time.Now().Add(24 * time.Hour)

	// Save token
	if err := s.userRepo.SetVerificationToken(ctx, userID, token, expiry); err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) ValidateRegistration(input RegistrationInput) error {
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

func (s *UserService) VerifyEmail(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("verification token is required")
	}
	// Find user by verification token
	user, err := s.userRepo.FindByVerificationToken(ctx, token)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("invalid or expired verification token")
	}

	// Check if token is expired
	if user.VerificationExpiry != nil && time.Now().After(*user.VerificationExpiry) {
		return errors.New("verification token has expired")
	}

	// Update verification status
	return s.userRepo.UpdateVerificationStatus(ctx, user.ID, true)
}
