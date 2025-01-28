package main

import (
	"log"
	"net/http"

	"option-manager/internal/database"
	"option-manager/internal/handlers"

	_ "github.com/lib/pq"
)

func main() {
	// Initialize database
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize handlers
	authHandler, err := handlers.NewAuthHandler(db)
	if err != nil {
		log.Fatalf("Failed to initialize auth handler: %v", err)
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

	// Protected routes
	http.HandleFunc("/dashboard", handlers.RequireAuth(db, func(w http.ResponseWriter, r *http.Request) {
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
