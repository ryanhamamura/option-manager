package repository

import (
	"database/sql"
	"option-manager/internal/types"
	"reflect"
	"testing"

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
