package handler

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"option-manager/internal/service"
	"option-manager/internal/types"
)

// Handler manages interaction with the service.
type Handler struct {
	svc       service.Service
	templates *template.Template
}

// New creates a new handler.
func New(svc service.Service, tmpl *template.Template) *Handler {
	return &Handler{svc: svc, templates: tmpl}
}

// GetPositions
func (h *Handler) GetPositions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	portfolioID := r.URL.Query().Get("portfolio_id")
	if portfolioID == "" {
		portfolioID = "port1" // Hardcoded for now; replace with auth later
	}
	positions, err := h.svc.GetPositions(portfolioID)
	if err != nil {
		http.Error(w, "Failed to fetch positions: "+err.Error(), http.StatusInternalServerError)
		return
	}
	for i, pos := range positions {
		trades, err := h.svc.GetTrades(pos.ID)
		if err != nil {
			http.Error(w, "Failed to fetch trades: "+err.Error(), http.StatusInternalServerError)
			return
		}
		positions[i].Trades = trades
	}
	data := struct {
		Positions []types.Position
		IsPremium bool
	}{Positions: positions, IsPremium: true} // Hardcoded premium for now
	if err := h.templates.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Failed to render template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}

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
