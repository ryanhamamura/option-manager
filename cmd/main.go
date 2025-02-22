package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"option-manager/internal/email"
	"option-manager/internal/handlers"
	"option-manager/internal/middleware"
	"option-manager/internal/repository"
	"option-manager/internal/service"

	_ "github.com/lib/pq"
)

func main() {

	dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	repo := repository.New(dbURL)

	// Initialize email client
	emailClient, err := email.NewClient(
		os.Getenv("AWS_REGION"),
		os.Getenv("EMAIL_SENDER"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize email client: %v", err)
	}

	// Initialize services with email client
	services, err := service.NewServices(
		repo,
		emailClient,
		os.Getenv("BASE_URL"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}

	// Initialize handlers
	authHandler, err := handlers.NewAuthHandler(services)
	if err != nil {
		log.Fatalf("Failed to initialize auth handler: %v", err)
	}

	log.Printf("Creating registration handler with email service...")
	registrationHandler, err := handlers.NewRegistrationHandler(services)
	if err != nil {
		log.Fatalf("Failed to initialize registration handler: %v", err)
	}
	log.Printf("Registration handler created successfully")

	verificationHandler, err := handlers.NewVerificationHandler(services)
	if err != nil {
		log.Fatalf("Failed to initialize verification handler: %v", err)
	}

	// Create base middleware chain
	baseChain := []middleware.Middleware{
		middleware.Logger,    // Add logging first to capture everything
		middleware.Recoverer, // Recover from panics
	}

	authChain := append(baseChain, middleware.RequireAuth(services))

	// Routes
	// Public routes with basic middleware
	http.Handle("/", middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			http.NotFound(w, r)
		}),
		baseChain...,
	))

	// Base routes
	http.Handle("/login", middleware.Chain(
		http.HandlerFunc(authHandler.LoginPage),
		baseChain...,
	))

	http.Handle("/register", middleware.Chain(
		http.HandlerFunc(registrationHandler.RegisterPage),
		baseChain...,
	))

	http.Handle("/verify", middleware.Chain(
		http.HandlerFunc(verificationHandler.VerifyEmail),
		baseChain...,
	))

	http.Handle("/verification-pending", middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tmpl, err := template.ParseFiles("templates/verification-pending.html")
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			tmpl.Execute(w, nil)
		}),
		baseChain...,
	))

	// Protected routes with full middleware stack
	http.Handle("/dashboard", middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := middleware.GetUserID(r.Context())
			if !ok {
				http.Error(w, "User not found in context", http.StatusInternalServerError)
				return
			}
			w.Write([]byte(fmt.Sprintf("Dashboard for user %d - Coming soon!", userID)))
		}),
		authChain...,
	))

	http.Handle("/logout", middleware.Chain(
		http.HandlerFunc(authHandler.Logout),
		baseChain...,
	))

	// Health check endpoint with minimal middleware
	http.Handle("/health", middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := db.Ping(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Database connection failed"))
				return
			}
			w.Write([]byte("OK"))
		}),
		middleware.Logger, // Only use logger for health checks
	))
}

// Helper function to mask sensitive values
func maskString(s string) string {
	if len(s) == 0 {
		return "not set"
	}
	return "set (length: " + fmt.Sprint(len(s)) + ")"
}
