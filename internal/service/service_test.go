package service

import (
	"errors"
	"option-manager/internal/types"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// mockRepository simulates the Repository interface for testing
type mockRepository struct {
	saveFunc func(user types.User) (types.User, error)
	getFunc  func(email string) (types.User, error)
}

func (m *mockRepository) SaveUser(user types.User) (types.User, error) {
	return m.saveFunc(user)
}

func (m *mockRepository) GetUserByEmail(email string) (types.User, error) {
	if m.getFunc != nil {
		return m.getFunc(email)
	}
	return types.User{}, errors.New("not implemented")
}

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		firstName  string
		lastName   string
		password   string
		saveFunc   func(user types.User) (types.User, error)
		wantErr    bool
		wantErrMsg string
		checkUser  func(t *testing.T, user types.User)
	}{
		{
			name:      "valid registration",
			email:     "alice@example.com",
			firstName: "Alice",
			lastName:  "Smith",
			password:  "secret123",
			saveFunc: func(user types.User) (types.User, error) {
				// Simulate PostgreSQL setting timestamps
				now := time.Now()
				user.CreatedAt = now
				user.UpdatedAt = now
				return user, nil
			},
			wantErr: false,
			checkUser: func(t *testing.T, user types.User) {
				if user.ID == "" {
					t.Errorf("expected non-empty UUID, got empty")
				}
				if user.Email != "alice@example.com" {
					t.Errorf("expected email %q, got %q", "alice@example.com", user.Email)
				}
				if user.FirstName != "Alice" || user.LastName != "Smith" {
					t.Errorf("expected name Alice Smith, got %s %s", user.FirstName, user.LastName)
				}
				if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("secret123")); err != nil {
					t.Errorf("password not hashed correctly: %v", err)
				}
				if user.CreatedAt.IsZero() || user.UpdatedAt.IsZero() {
					t.Errorf("expected non-zero timestamps, got %v, %v", user.CreatedAt, user.UpdatedAt)
				}
				if !user.CreatedAt.Equal(user.UpdatedAt) {
					t.Errorf("expected CreatedAt to equal UpdatedAt on creation, got %v vs %v", user.CreatedAt, user.UpdatedAt)
				}
			},
		},
		{
			name:      "missing email",
			email:     "",
			firstName: "Alice",
			lastName:  "Smith",
			password:  "secret123",
			saveFunc: func(user types.User) (types.User, error) {
				return user, nil
			},
			wantErr:    true,
			wantErrMsg: "all fields (email, first name, last name, password) are required",
		},
		{
			name:      "missing password",
			email:     "bob@example.com",
			firstName: "Bob",
			lastName:  "Jones",
			password:  "",
			saveFunc: func(user types.User) (types.User, error) {
				return user, nil
			},
			wantErr:    true,
			wantErrMsg: "all fields (email, first name, last name, password) are required",
		},
		{
			name:      "repository failure",
			email:     "bob@example.com",
			firstName: "Bob",
			lastName:  "Jones",
			password:  "pass456",
			saveFunc: func(user types.User) (types.User, error) {
				return user, errors.New("database connection lost")
			},
			wantErr:    true,
			wantErrMsg: "failed to register user: database connection lost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				saveFunc: tt.saveFunc,
			}
			svc := New(repo)

			got, err := svc.RegisterUser(tt.email, tt.firstName, tt.lastName, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("RegisterUser() error msg = %q, want %q", err.Error(), tt.wantErrMsg)
			}
			if !tt.wantErr && tt.checkUser != nil {
				tt.checkUser(t, got)
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	// Pre-generate a bcrypt hash for "secret123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)

	tests := []struct {
		name       string
		email      string
		password   string
		getFunc    func(email string) (types.User, error)
		wantErr    bool
		wantErrMsg string
		checkUser  func(t *testing.T, user types.User)
	}{
		{
			name:     "successful login",
			email:    "alice@example.com",
			password: "secret123",
			getFunc: func(email string) (types.User, error) {
				return types.User{
					ID:           "550e8400-e29b-41d4-a716-446655440000",
					Email:        "alice@example.com",
					FirstName:    "Alice",
					LastName:     "Smith",
					PasswordHash: string(hashedPassword),
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil
			},
			wantErr: false,
			checkUser: func(t *testing.T, user types.User) {
				if user.ID != "550e8400-e29b-41d4-a716-446655440000" {
					t.Errorf("expected ID %q, got %q", "550e8400-e29b-41d4-a716-446655440000", user.ID)
				}
				if user.Email != "alice@example.com" {
					t.Errorf("expected email %q, got %q", "alice@example.com", user.Email)
				}
			},
		},
		{
			name:     "invalid password",
			email:    "alice@example.com",
			password: "wrongpass",
			getFunc: func(email string) (types.User, error) {
				return types.User{
					ID:           "550e8400-e29b-41d4-a716-446655440000",
					Email:        "alice@example.com",
					FirstName:    "Alice",
					LastName:     "Smith",
					PasswordHash: string(hashedPassword),
				}, nil
			},
			wantErr:    true,
			wantErrMsg: "invalid email or password",
		},
		{
			name:     "user not found",
			email:    "bob@example.com",
			password: "secret123",
			getFunc: func(email string) (types.User, error) {
				return types.User{}, errors.New("user not found")
			},
			wantErr:    true,
			wantErrMsg: "invalid email or password",
		},
		{
			name:     "missing email",
			email:    "",
			password: "secret123",
			getFunc: func(email string) (types.User, error) {
				return types.User{}, errors.New("not implemented")
			},
			wantErr:    true,
			wantErrMsg: "email and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				getFunc: tt.getFunc,
			}
			svc := New(repo)

			got, err := svc.LoginUser(tt.email, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoginUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("LoginUser() error msg = %q, want %q", err.Error(), tt.wantErrMsg)
			}
			if !tt.wantErr && tt.checkUser != nil {
				tt.checkUser(t, got)
			}
		})
	}
}
