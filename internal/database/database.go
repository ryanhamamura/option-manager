package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Connect establishes a connection to the database with retry logic
func Connect() (*sql.DB, error) {
	dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	var db *sql.DB
	var err error

	// Retry logic for database connection
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", dbURL)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Println("Successfully connected to database")
				return db, nil
			}
		}
		log.Printf("Failed to connect to database, attempt %d/5: %v", i+1, err)
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after 5 attempts: %v", err)
}
