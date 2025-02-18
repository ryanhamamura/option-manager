package handlers

import (
	"html/template"
	"net/http"
	"option-manager/internal/service"
	"time"
)

type LoginPageData struct {
	Error string
}

type AuthHandler struct {
	services *service.Services
	template *template.Template
}

func NewAuthHandler(services *service.Services) (*AuthHandler, error) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		return nil, err
	}

	return &AuthHandler{
		services: services,
		template: tmpl,
	}, nil
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.template.Execute(w, LoginPageData{})
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")
		rememberMe := r.FormValue("remember-me") == "on"

		authResp, err := h.services.Auth.Authenticate(r.Context(), email, password)
		if err != nil {
			h.template.Execute(w, LoginPageData{
				Error: err.Error(),
			})
			return
		}

		// Create session
		session, err := h.services.Auth.CreateSession(r.Context(), authResp.User.ID, rememberMe)
		if err != nil {
			h.template.Execute(w, LoginPageData{
				Error: "Error creating session. Please try again.",
			})
			return
		}

		// Set session cookie
		cookie := &http.Cookie{
			Name:     "session_id",
			Value:    session.ID,
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil,
			Expires:  session.ExpiresAt,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, cookie)

		// Redirect to dashboard
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session_id"); err == nil {
		h.services.Auth.DeleteSession(r.Context(), cookie.Value)

		// Clear the cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(-24 * time.Hour),
			MaxAge:   -1,
		})
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
