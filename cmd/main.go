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
	"option-manager/internal/middleware"
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

	// Initialize middleware stack
	logMiddleware := middleware.Logger
	recovererMiddleware := middleware.Recoverer
	authMiddleware := middleware.RequireAuth(services)

	// Routes
	// Root handler with middleware
	rootHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.NotFound(w, r)
	})
	http.Handle("/", middleware.Chain(rootHandler, logMiddleware, recovererMiddleware))

	// Public routes with basic middleware
	loginHandler := http.HandlerFunc(authHandler.LoginPage)
	http.Handle("/login", middleware.Chain(loginHandler, logMiddleware, recovererMiddleware))

	registerHandler := http.HandlerFunc(registrationHandler.RegisterPage)
	http.Handle("/register", middleware.Chain(registerHandler, logMiddleware, recovererMiddleware))

	verifyHandler := http.HandlerFunc(verificationHandler.VerifyEmail)
	http.Handle("/verify", middleware.Chain(verifyHandler, logMiddleware, recovererMiddleware))

	// Verification pending page
	verificationPendingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/verification-pending.html")
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})
	http.Handle("/verification-pending", middleware.Chain(
		verificationPendingHandler,
		logMiddleware,
		recovererMiddleware,
	))

	// Protected routes
	// Protected routes with full middleware stack
	dashboardHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprintf("Dashboard for user %d - Coming soon!", userID)))
	})
	http.Handle("/dashboard", middleware.Chain(
		dashboardHandler,
		recovererMiddleware,
		logMiddleware,
		authMiddleware,
	))

	// Logout with basic middleware
	logoutHandler := http.HandlerFunc(authHandler.Logout)
	http.Handle("/logout", middleware.Chain(logoutHandler, logMiddleware, recovererMiddleware))

	// Health check endpoint
	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Database connection failed"))
			return
		}
		w.Write([]byte("OK"))
	})
	http.Handle("/health", middleware.Chain(healthHandler, logMiddleware))

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
