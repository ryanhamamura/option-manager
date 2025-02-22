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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Password  string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.svc.RegisterUser(input.Email, input.FirstName, input.LastName, input.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Send success response with timestamps
	w.Header().Set("Content-Type", "application/json")
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
