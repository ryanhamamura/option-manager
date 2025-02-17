package handlers

import (
	"html/template"
	"net/http"
	"option-manager/internal/service"
)

type RegistrationPageData struct {
	Error string
}

type RegistrationHandler struct {
	services *service.Services
	template *template.Template
}

func NewRegistrationHandler(services *service.Services) (*RegistrationHandler, error) {
	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		return nil, err
	}

	return &RegistrationHandler{
		services: services,
		template: tmpl,
	}, nil
}

func (h *RegistrationHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
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

		input := service.RegistrationInput{
			Email:     r.FormValue("email"),
			Password:  password,
			FirstName: r.FormValue("first_name"),
			LastName:  r.FormValue("last_name"),
		}

		// Validate input
		if err := h.services.User.ValidateRegistration(input); err != nil {
			h.template.Execute(w, RegistrationPageData{
				Error: err.Error(),
			})
			return
		}

		// Register user
		user, err := h.services.User.RegisterUser(r.Context(), input)
		if err != nil {
			h.template.Execute(w, RegistrationPageData{
				Error: err.Error(),
			})
			return
		}

		// Generate verification token
		token, err := h.services.User.GenerateVerificationToken(r.Context(), user.ID)
		if err != nil {
			h.template.Execute(w, RegistrationPageData{
				Error: "Failed to generate verification token. Please try again.",
			})
			return
		}

		// Send verification email
		err = h.services.Email.SendVerificationEmail(user.Email, user.FirstName, token)
		if err != nil {
			h.template.Execute(w, RegistrationPageData{
				Error: "Failed to send verification email. Please try again.",
			})
		}

		// Redirect to verification pending page
		http.Redirect(w, r, "/verification-pending", http.StatusSeeOther)

		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
