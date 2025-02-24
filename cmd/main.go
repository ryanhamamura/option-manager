package main

import (
	"fmt"
	"log"
	"net/http"

	"option-manager/internal/config"
	"option-manager/internal/handler"
	"option-manager/internal/repository"
	"option-manager/internal/service"

	_ "github.com/lib/pq"
)

func main() {

	// dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
	// 	os.Getenv("DB_HOST"),
	// 	os.Getenv("DB_PORT"),
	// 	os.Getenv("DB_USER"),
	// 	os.Getenv("DB_PASSWORD"),
	// 	os.Getenv("DB_NAME"),
	// )
	//
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	repo := repository.New(cfg.DSN())
	svc := service.New(repo)
	h := handler.New(svc)

	http.HandleFunc("/register", h.RegisterUser)
	http.HandleFunc("/login", h.LoginUser)
	fmt.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
