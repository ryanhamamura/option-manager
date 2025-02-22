// internal/service/service.go
package service

import (
	"errors"
	"option-manager/internal/types"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// mockRepository simulates the Repository interface for testing
type mockRepository struct {
	saveFunc func(user types.User) error
}

func (m *mockRepository) SaveUser(user types.User) error {
	return m.saveFunc(user)
}

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		firstName  string
		lastName   string
		password   string
		saveErr    error
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
			saveErr:   nil,
			wantErr:   false,
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
			name:       "missing email",
			email:      "",
			firstName:  "Alice",
			lastName:   "Smith",
			password:   "secret123",
			saveErr:    nil,
			wantErr:    true,
			wantErrMsg: "all fields (email, first name, last name, password) are required",
		},
		{
			name:       "missing password",
			email:      "bob@example.com",
			firstName:  "Bob",
			lastName:   "Jones",
			password:   "",
			saveErr:    nil,
			wantErr:    true,
			wantErrMsg: "all fields (email, first name, last name, password) are required",
		},
		{
			name:       "repository failure",
			email:      "bob@example.com",
			firstName:  "Bob",
			lastName:   "Jones",
			password:   "pass456",
			saveErr:    errors.New("database connection lost"),
			wantErr:    true,
			wantErrMsg: "failed to register user: database connection lost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				saveFunc: func(user types.User) error {
					return tt.saveErr
				},
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
