package repository

import (
	"database/sql"
	"errors"
	"option-manager/internal/types"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/lib/pq"
)

type inMemoryRepo struct {
	users map[string]types.User
}

func (r *inMemoryRepo) SaveUser(user types.User) (types.User, error) {
	r.users[user.ID] = user
	return user, nil
}

func (r *inMemoryRepo) GetUsersByEmail(email string) (types.User, error) {
	for _, u := range r.users {
		return u, nil
	}
	return types.User{}, sql.ErrNoRows
}

func TestSaveUser_InMemory(t *testing.T) {
	repo := &inMemoryRepo{users: make(map[string]types.User)}
	user := types.User{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		PasswordHash: "hashedpass",
	}

	savedUser, err := repo.SaveUser(user)
	if err != nil {
		t.Errorf("SaveUser() error = %v, want nil", err)
	}

	saved, exists := repo.users[user.ID]
	if !exists {
		t.Errorf("user not saved in repo")
	}
	if !reflect.DeepEqual(saved, savedUser) {
		t.Errorf("saved user = %+v, want %+v", saved, savedUser)
	}
}

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
	}

	// Expect the INSERT query
	mock.ExpectExec(`INSERT INTO users \(id, email, first_name, last_name, password_hash\)`).
		WithArgs(user.ID, user.Email, user.FirstName, user.LastName, user.PasswordHash).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = repo.SaveUser(user)
	if err != nil {
		t.Errorf("SaveUser() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}

	// Test error case
	mock.ExpectExec(`INSERT INTO users \(id, email, first_name, last_name, password_hash\)`).
		WithArgs(user.ID, user.Email, user.FirstName, user.LastName, user.PasswordHash).
		WillReturnError(errors.New("duplicate key violation"))

	_, err = repo.SaveUser(user)
	if err == nil {
		t.Errorf("SaveUser() expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}

}
