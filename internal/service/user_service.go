// internal/service/user_service.go
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"option-manager/internal/email"
	"option-manager/internal/repository"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo     repository.UserRepository
	emailService *email.EmailService
}

func NewUserService(userRepo repository.UserRepository, emailService *email.EmailService) *UserService {
	return &UserService{
		userRepo:     userRepo,
		emailService: emailService,
	}
}

type RegistrationInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

func (s *UserService) RegisterUser(ctx context.Context, input RegistrationInput) (*repository.User, error) {
	// Check if user exists
	existingUser, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &repository.User{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    input.FirstName,
		LastName:     input.LastName,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GenerateVerificationToken(ctx context.Context, userID int) (string, error) {
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

func (s *UserService) VerifyEmail(ctx context.Context, token string) error {
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
