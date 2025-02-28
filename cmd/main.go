package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"option-manager/internal/config"
	"option-manager/internal/handler"
	"option-manager/internal/repository"
	"option-manager/internal/service"

	_ "github.com/lib/pq"
)

type Server struct {
	Handler *handler.Handler
	Config  *config.Config
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
		os.Exit(1)
	}

	repo := repository.New(cfg.DSN())
	svc := service.New(repo)

	// Parse templates
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("failed to parse partials: %v", err)
		os.Exit(1)
	}
	templates, err = templates.ParseGlob("templates/partials/*.html")
	if err != nil {
		log.Fatalf("failed to parse partials: %v", err)
		os.Exit(1)
	}
	templates, err = templates.ParseGlob("templates/layouts/*.html")
	if err != nil {
		log.Fatalf("failed to parse layouts: %v", err)
		os.Exit(1)
	}

	logger := log.New(os.Stdout, "options-manager: ", log.LstdFlags)
	h := handler.New(svc, templates, logger)
	server := &Server{
		Handler: h,
		Config:  cfg,
	}

	http.HandleFunc("/positions", server.Handler.GetPositions)
	http.HandleFunc("/register", server.Handler.RegisterUser)
	http.HandleFunc("/login", server.Handler.LoginUser)

	httpServer := &http.Server{
		Addr: ":8080",
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on :8080...")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server failed: %v", err)
			os.Exit(1)
		}
	}()

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	}
	log.Printf("Server shut down gracefully")
	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
