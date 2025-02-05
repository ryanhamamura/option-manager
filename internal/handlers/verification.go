package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"option-manager/internal/auth"
)

type VerificationData struct {
	Error   string
	Success string
}

type VerificationHandler struct {
	db       *sql.DB
	template *template.Template
}

func NewVerificationHandler(db *sql.DB) (*VerificationHandler, error) {
	tmpl, err := template.ParseFiles("templates/verify.html")
	if err != nil {
		return nil, err
	}

	return &VerificationHandler{
		db:       db,
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

	err := auth.VerifyEmail(h.db, token)
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
