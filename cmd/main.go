package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"option-manager/internal/database"
	"option-manager/internal/email"
	"option-manager/internal/handlers"
	"option-manager/internal/repository/postgres"
	"option-manager/internal/service"

	_ "github.com/lib/pq"
)

func main() {
	// Log all relevant environment variables
	log.Printf("Environment variables:")
	log.Printf("AWS_REGION: %s", os.Getenv("AWS_REGION"))
	log.Printf("EMAIL_SENDER: %s", os.Getenv("EMAIL_SENDER"))
	log.Printf("AWS_ACCESS_KEY_ID: %s", maskString(os.Getenv("AWS_ACCESS_KEY_ID")))
	log.Printf("AWS_SECRET_ACCESS_KEY: %s", maskString(os.Getenv("AWS_SECRET_ACCESS_KEY")))

	// Initialize database
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Intialize repositories
	repo := postgres.NewRepository(db)

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

	// Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.NotFound(w, r)
	})

	http.HandleFunc("/login", authHandler.LoginPage)
	http.HandleFunc("/logout", authHandler.Logout)
	http.HandleFunc("/register", registrationHandler.RegisterPage)
	http.HandleFunc("/verify", verificationHandler.VerifyEmail)
	http.HandleFunc("/verification-pending", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/verification-pending.html")
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	// Protected routes
	http.HandleFunc("/dashboard", handlers.RequireAuth(services, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Dashboard - Coming soon!"))
	}))

	// Basic health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		err := db.Ping()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Database connection failed"))
			return
		}
		w.Write([]byte("OK"))
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// Helper function to mask sensitive values
func maskString(s string) string {
	if len(s) == 0 {
		return "not set"
	}
	return "set (length: " + fmt.Sprint(len(s)) + ")"
}
