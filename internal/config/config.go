// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	AWS      AWSConfig
	Email    EmailConfig
	Security SecurityConfig
}

type AppConfig struct {
	Environment string
	BaseURL     string
	LogLevel    string
	LogFile     string
}

type ServerConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

type EmailConfig struct {
	SenderAddress string
	TemplatesDir  string
}

type SecurityConfig struct {
	AllowedHosts    []string
	SSLRedirect     bool
	IsDevelopment   bool
	SessionSecret   string
	CSRFAuthKey     string
	PasswordSaltLen int
	JWTSecret       string
}

// Load returns a Config struct populated with all configuration values
func Load() (*Config, error) {
	cfg := &Config{}

	// Load App config
	cfg.App = AppConfig{
		Environment: getEnvString("ENV", "development"),
		BaseURL:     getEnvString("BASE_URL", "http://localhost:8080"),
		LogLevel:    getEnvString("LOG_LEVEL", "info"),
		LogFile:     getEnvString("LOG_FILE", "app.log"),
	}

	// Load Server config
	cfg.Server = ServerConfig{
		Port:            getEnvInt("SERVER_PORT", 8080),
		ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 5*time.Second),
		WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
	}

	// Load Database config
	cfg.Database = DatabaseConfig{
		Host:     getEnvString("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnvString("DB_USER", "postgres"),
		Password: getEnvString("DB_PASSWORD", ""),
		DBName:   getEnvString("DB_NAME", "options_manager"),
		SSLMode:  getEnvString("DB_SSLMODE", "disable"),
		MaxConns: getEnvInt("DB_MAX_CONNS", 25),
	}

	// Load AWS config
	cfg.AWS = AWSConfig{
		Region:          getEnvString("AWS_REGION", ""),
		AccessKeyID:     getEnvString("AWS_ACCESS_KEY_ID", ""),
		SecretAccessKey: getEnvString("AWS_SECRET_ACCESS_KEY", ""),
	}

	// Load Email config
	cfg.Email = EmailConfig{
		SenderAddress: getEnvString("EMAIL_SENDER", ""),
		TemplatesDir:  getEnvString("EMAIL_TEMPLATES_DIR", "templates/email"),
	}

	// Load Security config
	cfg.Security = SecurityConfig{
		AllowedHosts:    getEnvStringSlice("ALLOWED_HOSTS", []string{}),
		SSLRedirect:     getEnvBool("SSL_REDIRECT", true),
		IsDevelopment:   cfg.App.Environment != "production",
		SessionSecret:   getEnvString("SESSION_SECRET", ""),
		CSRFAuthKey:     getEnvString("CSRF_AUTH_KEY", ""),
		PasswordSaltLen: getEnvInt("PASSWORD_SALT_LENGTH", 16),
		JWTSecret:       getEnvString("JWT_SECRET", ""),
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Check required values in production
	if c.App.Environment == "production" {
		if c.Security.SessionSecret == "" {
			return fmt.Errorf("SESSION_SECRET is required in production")
		}
		if c.Security.CSRFAuthKey == "" {
			return fmt.Errorf("CSRF_AUTH_KEY is required in production")
		}
		if c.Security.JWTSecret == "" {
			return fmt.Errorf("JWT_SECRET is required in production")
		}
		if len(c.Security.AllowedHosts) == 0 {
			return fmt.Errorf("ALLOWED_HOSTS is required in production")
		}
		if c.Database.Password == "" {
			return fmt.Errorf("DB_PASSWORD is required in production")
		}
	}

	// Validate AWS credentials if email sending is required
	if c.Email.SenderAddress != "" {
		if c.AWS.Region == "" {
			return fmt.Errorf("AWS_REGION is required when EMAIL_SENDER is set")
		}
		if c.AWS.AccessKeyID == "" {
			return fmt.Errorf("AWS_ACCESS_KEY_ID is required when EMAIL_SENDER is set")
		}
		if c.AWS.SecretAccessKey == "" {
			return fmt.Errorf("AWS_SECRET_ACCESS_KEY is required when EMAIL_SENDER is set")
		}
	}

	return nil
}

// Helper functions for environment variables
func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.Split(value, ",")
	}
	return defaultValue
}
