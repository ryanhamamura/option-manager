package main

import (
	"html/template"
	"log"
	"net/http"

	"option-manager/internal/config"
	"option-manager/internal/handler"
	"option-manager/internal/repository"
	"option-manager/internal/service"

	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	repo := repository.New(cfg.DSN())
	svc := service.New(repo)

	// Parse templates
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("failed to parse partials: %v", err)
	}
	templates, err = templates.ParseGlob("templates/partials/*.html")
	if err != nil {
		log.Fatalf("failed to parse partials: %v", err)
	}
	templates, err = templates.ParseGlob("templates/layouts/*.html")
	if err != nil {
		log.Fatalf("failed to parse layouts: %v", err)
	}

	h := handler.New(svc, templates)
	http.HandleFunc("/positions", h.GetPositions)
	http.HandleFunc("/register", h.RegisterUser)
	http.HandleFunc("/login", h.LoginUser)

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
