package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"option-manager/internal/auth"
	"option-manager/internal/email"
)

type RegistrationPageData struct {
	Error string
}

type RegistrationHandler struct {
	db           *sql.DB
	template     *template.Template
	emailService *email.EmailService
}

func NewRegistrationHandler(db *sql.DB, emailService *email.EmailService) (*RegistrationHandler, error) {
	log.Printf("Creating new registration handler...")
	if emailService == nil {
		return nil, fmt.Errorf("email service cannot be nil")
	}

	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		return nil, err
	}

	handler := &RegistrationHandler{
		db:           db,
		template:     tmpl,
		emailService: emailService,
	}

	log.Printf("Registration handler created with email service: %v", emailService != nil)
	return handler, nil
}

func (h *RegistrationHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("RegisterPage called with email service: %v", h.emailService != nil)

	if h.emailService == nil {
		log.Printf("Email service is nil in RegisterPage handler")
		h.template.Execute(w, RegistrationPageData{
			Error: "Registration is temporarily unavailable. Please try again later.",
		})
	}

	if r.Method == http.MethodGet {
		h.template.Execute(w, RegistrationPageData{})
		return
	}

	if r.Method == http.MethodPost {
		// Verify passwords match
		password := r.FormValue("password")
		passwordConfirm := r.FormValue("password_confirm")
		if password != passwordConfirm {
			h.template.Execute(w, RegistrationPageData{
				Error: "Passwords do not match",
			})
			return
		}

		input := auth.RegistrationInput{
			Email:     r.FormValue("email"),
			Password:  password,
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

		// Generate verification token
		verificationToken, err := auth.GenerateVerificationToken(h.db, user.ID)
		if err != nil {
			h.template.Execute(w, RegistrationPageData{
				Error: "Failed to generate verification token. Please try again.",
			})
			return
		}

		// Send verification email
		err = h.emailService.SendVerificationEmail(
			user.Email,
			user.FirstName,
			verificationToken,
			"http://localhost:8080", // Replace with your actual domain in production
		)
		if err != nil {
			h.template.Execute(w, RegistrationPageData{
				Error: "Failed to send verification email. Please try again.",
			})
		}

		// Redirect to verification pending page
		http.Redirect(w, r, "/verification-pending", http.StatusSeeOther)

		// Redirect to dashboard
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
