package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"option-manager/internal/types"
	"reflect"
	"strings"
	"testing"
	"time"
)

type mockService struct {
	registerFunc func(email, firstName, lastName, password string) (types.User, error)
}

func (m *mockService) RegisterUser(email, firstName, lastName, password string) (types.User, error) {
	return m.registerFunc(email, firstName, lastName, password)
}

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		registerResult types.User
		registerErr    error
		wantStatus     int
		wantBody       map[string]interface{}
	}{
		{
			name:   "successful registration",
			method: http.MethodPost,
			body: `{
                "email": "alice@example.com",
                "first_name": "Alice",
                "last_name": "Smith",
                "password": "secret123"
            }`,
			registerResult: types.User{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				Email:        "alice@example.com",
				FirstName:    "Alice",
				LastName:     "Smith",
				PasswordHash: "hashedpass",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			registerErr: nil,
			wantStatus:  http.StatusOK,
			wantBody: map[string]interface{}{
				"id":        "550e8400-e29b-41d4-a716-446655440000",
				"email":     "alice@example.com",
				"firstName": "Alice",
				"lastName":  "Smith",
				"message":   "User registered successfully",
			},
		},
		{
			name:       "invalid JSON",
			method:     http.MethodPost,
			body:       `{"email": "bob@example.com", "first_name": "Bob", "last_name": "Jones", "password": "pass456"`,
			wantStatus: http.StatusBadRequest,
			wantBody: map[string]interface{}{
				"error": "Invalid request body",
			},
		},
		{
			name:   "missing field",
			method: http.MethodPost,
			body: `{
                "email": "bob@example.com",
                "first_name": "Bob",
                "last_name": ""
            }`,
			registerErr: errors.New("all fields (email, first name, last name, password) are required"),
			wantStatus:  http.StatusBadRequest,
			wantBody: map[string]interface{}{
				"error": "all fields (email, first name, last name, password) are required",
			},
		},
		{
			name:       "method not allowed",
			method:     http.MethodGet,
			body:       "",
			wantStatus: http.StatusMethodNotAllowed,
			wantBody: map[string]interface{}{
				"error": "Method not allowed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{
				registerFunc: func(email, firstName, lastName, password string) (types.User, error) {
					if tt.registerErr != nil {
						return types.User{}, tt.registerErr
					}
					return tt.registerResult, nil
				},
			}
			h := New(mockSvc)

			req, err := http.NewRequest(tt.method, "/register", bytes.NewBuffer([]byte(tt.body)))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			h.RegisterUser(rr, req)

			// Log raw response for debugging
			t.Logf("Status: %d, Body: %q", rr.Code, rr.Body.String())

			if rr.Code != tt.wantStatus {
				t.Errorf("RegisterUser() status = %v, want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
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
						if diff := gotTime.Sub(tt.registerResult.CreatedAt); diff > time.Second || diff < -time.Second {
							t.Errorf("response %s = %v, want approximately %v (diff: %v)", key, gotTime, tt.registerResult.CreatedAt, diff)
						}
					} else if !reflect.DeepEqual(gotVal, wantVal) {
						t.Errorf("response %s = %v, want %v", key, gotVal, wantVal)
					}
				}
			} else {
				gotBody := strings.TrimSpace(rr.Body.String())
				wantBody := tt.wantBody["error"].(string)
				if gotBody != wantBody {
					t.Errorf("response body = %q, want %q", gotBody, wantBody)
				}
			}
		})
	}
}
