package repository

import (
	"errors"
	"option-manager/internal/types"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/lib/pq"
)

func TestSaveUser_Postgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()
	repo := &postgresRepo{db: db}
	user := types.User{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		PasswordHash: "hashedpass",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Expect the INSERT query
	mock.ExpectExec(`INSERT INTO users \(id, email, first_name, last_name, password_hash, created_at, updated_at\)`).
		WithArgs(user.ID, user.Email, user.FirstName, user.LastName, user.PasswordHash, user.CreatedAt, user.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SaveUser(user)
	if err != nil {
		t.Errorf("SaveUser() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}

	// Test error case
	mock.ExpectExec(`INSERT INTO users \(id, email, first_name, last_name, password_hash, created_at, updated_at\)`).
		WithArgs(user.ID, user.Email, user.FirstName, user.LastName, user.PasswordHash, user.CreatedAt, user.UpdatedAt).
		WillReturnError(errors.New("duplicate key violation"))

	err = repo.SaveUser(user)
	if err == nil {
		t.Errorf("SaveUser() expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}

}
