// internal/service/auth_service.go
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"option-manager/internal/repository"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewAuthService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

type AuthenticateResponse struct {
	User    *repository.User
	Session *repository.Session
}

func (s *AuthService) Authenticate(ctx context.Context, email, password string) (*AuthenticateResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	return &AuthenticateResponse{
		User: user,
	}, nil
}

func (s *AuthService) CreateSession(ctx context.Context, userID int, rememberMe bool) (*repository.Session, error) {
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

	session := &repository.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *AuthService) GetSession(ctx context.Context, sessionID string) (*repository.Session, error) {
	return s.sessionRepo.FindByID(ctx, sessionID)
}

func (s *AuthService) DeleteSession(ctx context.Context, sessionID string) error {
	return s.sessionRepo.Delete(ctx, sessionID)
}
