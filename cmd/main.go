package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"option-manager/internal/handlers"
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
	svc := service.New(repo)
	h := handlers.New(svc)

	http.HandleFunc("/register", h.RegisterUser)
	fmt.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
