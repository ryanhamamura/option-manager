package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"option-manager/internal/repository"
	"option-manager/internal/service"
	"option-manager/internal/types"
	"reflect"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestRegisterUser_Integration(t *testing.T) {
	// Start PostgreSQL container
	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	}()

	// Get connection string
	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Apply migrations
	m, err := migrate.New("file://../../../migrations", dsn)
	if err != nil {
		t.Fatalf("failed to initialize migrations: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	// Set up dependencies
	repo := repository.New(dsn)
	svc := service.New(repo)
	h := New(svc)

	// Test cases
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   map[string]interface{}
	}{
		{
			name: "successful registration",
			body: `{
                "email": "alice@example.com",
                "first_name": "Alice",
                "last_name": "Smith",
                "password": "secret123"
            }`,
			wantStatus: http.StatusOK,
			wantBody: map[string]interface{}{
				"email":     "alice@example.com",
				"firstName": "Alice",
				"lastName":  "Smith",
				"message":   "User registered successfully",
			},
		},
		{
			name: "missing field",
			body: `{
                "email": "bob@example.com",
                "first_name": "Bob",
                "last_name": ""
            }`,
			wantStatus: http.StatusBadRequest,
			wantBody: map[string]interface{}{
				"error": "all fields (email, first name, last name, password) are required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte(tt.body)))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			h.RegisterUser(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("RegisterUser() status = %v, want %v", rr.Code, tt.wantStatus)
			}

			var gotBody map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&gotBody); err != nil {
				t.Errorf("failed to decode JSON response: %v (body: %q)", err, rr.Body.String())
				return
			}

			for key, wantVal := range tt.wantBody {
				gotVal, exists := gotBody[key]
				if !exists {
					t.Errorf("response missing key %q (got: %+v)", key, gotBody)
					continue
				}
				if key == "createdAt" || key == "updatedAt" {
					gotTime, err := time.Parse(time.RFC3339, gotVal.(string))
					if err != nil {
						t.Errorf("failed to parse %s time %v: %v", key, gotVal, err)
						continue
					}
					if gotTime.IsZero() {
						t.Errorf("response %s is zero, expected non-zero from DB", key)
					}
				} else if !reflect.DeepEqual(gotVal, wantVal) {
					t.Errorf("response %s = %v, want %v", key, gotVal, wantVal)
				}
			}

			// Verify DB state for success case
			if tt.wantStatus == http.StatusOK {
				var dbUser types.User
				err = repo.(*postgresRepo).db.QueryRow(
					"SELECT id, email, first_name, last_name, password_hash, created_at, updated_at FROM users WHERE email = $1",
					"alice@example.com",
				).Scan(&dbUser.ID, &dbUser.Email, &dbUser.FirstName, &dbUser.LastName, &dbUser.PasswordHash, &dbUser.CreatedAt, &dbUser.UpdatedAt)
				if err != nil {
					t.Errorf("failed to query user from DB: %v", err)
					return
				}
				if dbUser.Email != "alice@example.com" {
					t.Errorf("DB email = %v, want %v", dbUser.Email, "alice@example.com")
				}
				if dbUser.CreatedAt.IsZero() || dbUser.UpdatedAt.IsZero() {
					t.Errorf("DB timestamps are zero, expected non-zero")
				}
			}
		})
	}
}
