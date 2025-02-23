package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"option-manager/internal/service"
)

// Handler manages interaction with the service.
type Handler struct {
	svc service.Service
}

// New creates a new handler.
func New(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterUser handles the HTTP registration request.
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	// Set Content-Type for all responses
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		errResponse := map[string]string{"error": "Method not allowed"}
		if err := json.NewEncoder(w).Encode(errResponse); err != nil {
			log.Printf("Failed to encode error response: %v", err)
		}
		return
	}

	var input struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Password  string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errResponse := map[string]string{"error": "Invalid request body"}
		if err := json.NewEncoder(w).Encode(errResponse); err != nil {
			log.Printf("Failed to encode error response: %v", err)
		}
		return
	}

	user, err := h.svc.RegisterUser(input.Email, input.FirstName, input.LastName, input.Password)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errResponse := map[string]string{"error": err.Error()}
		if err := json.NewEncoder(w).Encode(errResponse); err != nil {
			log.Printf("Failed to encode error response: %v", err)
		}
		return
	}

	response := map[string]interface{}{
		"id":        user.ID,
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"createdAt": user.CreatedAt,
		"updatedAt": user.UpdatedAt,
		"message":   "User registered successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// LoginUser handles the HTTP login request.
func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		errResponse := map[string]string{"error": "Method not allowed"}
		if err := json.NewEncoder(w).Encode(errResponse); err != nil {
			log.Printf("Failed to encode error response: %v", err)
		}
		return
	}

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errResponse := map[string]string{"error": "Invalid request body"}
		if err := json.NewEncoder(w).Encode(errResponse); err != nil {
			log.Printf("Failed to encode error response: %v", err)
		}
	}

	user, err := h.svc.LoginUser(input.Email, input.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errResponse := map[string]string{"error": err.Error()}
		if err := json.NewEncoder(w).Encode(errResponse); err != nil {
			log.Printf("Failed to encode error response: %v", err)
		}
		return
	}

	response := map[string]interface{}{
		"id":        user.ID,
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"message":   "Login successful",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
