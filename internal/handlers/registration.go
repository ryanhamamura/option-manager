package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"option-manager/internal/auth"
)

type RegistrationPageData struct {
	Error string
}

type RegistrationHandler struct {
	db       *sql.DB
	template *template.Template
}

func NewRegistrationHandler(db *sql.DB) (*RegistrationHandler, error) {
	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		return nil, err
	}

	return &RegistrationHandler{
		db:       db,
		template: tmpl,
	}, nil
}

func (h *RegistrationHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.template.Execute(w, RegistrationPageData{})
		return
	}

	if r.Method == http.MethodPost {
		input := auth.RegistrationInput{
			Email:     r.FormValue("email"),
			Password:  r.FormValue("password"),
			FirstName: r.FormValue("first_name"),
			LastName:  r.FormValue("last_name"),
		}

		// Validate input
		if err := auth.ValidateRegistration(input); err != nil {
			h.template.Execute(w, RegistrationPageData{
				Error: err.Error(),
			})
			return
		}

		// Register user
		user, err := auth.RegisterUser(h.db, input)
		if err != nil {
			h.template.Execute(w, RegistrationPageData{
				Error: err.Error(),
			})
			return
		}

		// Create session for the new user
		session, err := auth.CreateSession(h.db, user.ID, false)
		if err != nil {
			h.template.Execute(w, RegistrationPageData{
				Error: "Registration successful but failed to create session. Please login.",
			})
			return
		}

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    session.ID,
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil,
			Expires:  session.ExpiresAt,
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect to dashboard
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
