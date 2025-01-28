package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"option-manager/internal/auth"
	"time"
)

type LoginPageData struct {
	Error string
}

type AuthHandler struct {
	db       *sql.DB
	template *template.Template
}

func NewAuthHandler(db *sql.DB) (*AuthHandler, error) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		return nil, err
	}

	return &AuthHandler{
		db:       db,
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

		user, err := auth.AuthenticateUser(h.db, email, password)
		if err != nil {
			h.template.Execute(w, LoginPageData{
				Error: err.Error(),
			})
			return
		}

		// Create session
		session, err := auth.CreateSession(h.db, user.ID, rememberMe)
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
		auth.DeleteSession(h.db, cookie.Value)

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

func RequireAuth(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if _, err := auth.GetSession(db, cookie.Value); err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// If you need the session info later, uncomment and use this:
		/*
			session, err := auth.GetSession(db, cookie.Value)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			// Store user info in request context
			ctx := context.WithValue(r.Context(), "user_id", session.UserID)
			r = r.WithContext(ctx)
		*/

		next(w, r)
	}
}
