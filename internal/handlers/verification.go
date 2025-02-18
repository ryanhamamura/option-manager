package handlers

import (
	"html/template"
	"net/http"
	"option-manager/internal/service"
)

type VerificationData struct {
	Error   string
	Success string
}

type VerificationHandler struct {
	services *service.Services
	template *template.Template
}

func NewVerificationHandler(services *service.Services) (*VerificationHandler, error) {
	tmpl, err := template.ParseFiles("templates/verify.html")
	if err != nil {
		return nil, err
	}

	return &VerificationHandler{
		services: services,
		template: tmpl,
	}, nil
}

func (h *VerificationHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		h.template.Execute(w, VerificationData{
			Error: "Invalid verification link",
		})
		return
	}

	err := h.services.User.VerifyEmail(r.Context(), token)
	if err != nil {
		h.template.Execute(w, VerificationData{
			Error: err.Error(),
		})
		return
	}

	h.template.Execute(w, VerificationData{
		Success: "Your email has been verified! You can now log in.",
	})
}

func RequireAuth(services *service.Services, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if _, err := services.Auth.GetSession(r.Context(), cookie.Value); err != nil {
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
