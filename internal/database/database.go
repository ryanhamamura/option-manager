package database

import (
	"database/sql"
	"fmt"
	"log"
	"option-manager/internal/config"
	"time"

	_ "github.com/lib/pq"
)

// Options holds configuration for the database connection pool
type Options struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultOptions returns sensible defaults for database connection pool settings
func DefaultOptions() Options {
	return Options{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// Connect establishes a connection to the database with retry logic
func Connect(cfg config.DatabaseConfig) (*sql.DB, error) {
	return ConnectWithOptions(cfg, DefaultOptions())
}

// ConnectWithOptions establishes a connection to the database with custom pool options
func ConnectWithOptions(cfg config.DatabaseConfig, opts Options) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	var db *sql.DB
	var err error

	maxRetries := 5
	retryDelay := 5 * time.Second

	// Retry logic for database connection
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Printf("Successfully connected to database on attempt %d", i+1)

				// Configure connection pool
				db.SetMaxOpenConns(opts.MaxOpenConns)
				db.SetMaxIdleConns(opts.MaxIdleConns)
				db.SetConnMaxLifetime(opts.ConnMaxLifetime)
				db.SetConnMaxIdleTime(opts.ConnMaxIdleTime)

				return db, nil
			}
		}
		log.Printf("Failed to connect to database, attempt %d/%d: %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %v", maxRetries, err)
}

// ValidateConnection performs a thorough connection test
func ValidateConnection(db *sql.DB) error {
	// Test basic connectivity
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping failed: %v", err)
	}

	// Test query execution
	var now time.Time
	err := db.QueryRow("SELECT NOW()").Scan(&now)
	if err != nil {
		return fmt.Errorf("test query failed: %v", err)
	}

	// Get connection stats
	stats := db.Stats()
	log.Printf("Database connection pool stats:")
	log.Printf("- Open connections: %d", stats.OpenConnections)
	log.Printf("- In use connections: %d", stats.InUse)
	log.Printf("- Idle connections: %d", stats.Idle)
	log.Printf("- Queries running: %d", stats.InUse)
	log.Printf("- Connections waited for: %d", stats.WaitCount)
	log.Printf("- Total wait time: %v", stats.WaitDuration)
	log.Printf("- Max idle closed: %d", stats.MaxIdleClosed)
	log.Printf("- Max lifetime closed: %d", stats.MaxLifetimeClosed)

	return nil
}

// Close gracefully closes the database connection
func Close(db *sql.DB) error {
	if db != nil {
		log.Println("Closing database connection...")
		return db.Close()
	}
	return nil
}
