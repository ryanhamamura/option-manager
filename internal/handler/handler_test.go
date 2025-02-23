package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"option-manager/internal/types"
	"reflect"
	"testing"
	"time"
)

type mockService struct {
	registerFunc func(email, firstName, lastName, password string) (types.User, error)
	loginFunc    func(email, password string) (types.User, error)
}

func (m *mockService) RegisterUser(email, firstName, lastName, password string) (types.User, error) {
	return m.registerFunc(email, firstName, lastName, password)
}

func (m *mockService) LoginUser(email, password string) (types.User, error) {
	return m.loginFunc(email, password)
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

			// All responses are now JSON
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
		})
	}
}

func TestLoginUser(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		loginUser  types.User
		loginErr   error
		wantStatus int
		wantBody   map[string]interface{}
	}{
		{
			name:   "successful login",
			method: http.MethodPost,
			body: `{
                "email": "alice@example.com",
                "password": "secret123"
            }`,
			loginUser: types.User{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				Email:        "alice@example.com",
				FirstName:    "Alice",
				LastName:     "Smith",
				PasswordHash: "hashedpass",
			},
			loginErr:   nil,
			wantStatus: http.StatusOK,
			wantBody: map[string]interface{}{
				"id":        "550e8400-e29b-41d4-a716-446655440000",
				"email":     "alice@example.com",
				"firstName": "Alice",
				"lastName":  "Smith",
				"message":   "Login successful",
			},
		},
		{
			name:       "invalid credentials",
			method:     http.MethodPost,
			body:       `{"email": "alice@example.com", "password": "wrongpass"}`,
			loginErr:   errors.New("invalid email or password"),
			wantStatus: http.StatusUnauthorized,
			wantBody: map[string]interface{}{
				"error": "invalid email or password",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{
				loginFunc: func(email, password string) (types.User, error) {
					if tt.loginErr != nil {
						return types.User{}, tt.loginErr
					}
					return tt.loginUser, nil
				},
			}
			h := New(mockSvc)

			req, err := http.NewRequest(tt.method, "/login", bytes.NewBuffer([]byte(tt.body)))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			h.LoginUser(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("LoginUser() status = %v, want %v", rr.Code, tt.wantStatus)
			}

			var gotBody map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&gotBody); err != nil {
				t.Errorf("failed to decode JSON response: %v (body: %s)", err, rr.Body.String())
				return
			}

			for key, wantVal := range tt.wantBody {
				gotVal, exists := gotBody[key]
				if !exists {
					t.Errorf("response missing key %q", key)
					continue
				}
				if !reflect.DeepEqual(gotVal, wantVal) {
					t.Errorf("response %s = %v, want %v", key, gotVal, wantVal)
				}
			}
		})
	}
}
