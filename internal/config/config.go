package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Slack    SlackConfig
	Template TemplateConfig
	Logging  LoggingConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port int
	Host string
	Env  string
}

// SlackConfig holds Slack API configuration
type SlackConfig struct {
	APIToken       string
	ChannelID      string
	RealtimeEnable bool
}

// TemplateConfig holds template-related configuration
type TemplateConfig struct {
	CacheEnabled bool
	TemplateDir  string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

var (
	config Config
	once   sync.Once
)

// Get returns the singleton config instance
func Get() *Config {
	once.Do(func() {
		// Load .env file if it exists
		_ = godotenv.Load()

		config = Config{
			Server: ServerConfig{
				Port: getEnvAsInt("SERVER_PORT", 8080),
				Host: getEnvAsString("SERVER_HOST", "0.0.0.0"),
				Env:  getEnvAsString("ENV", "development"),
			},
			Slack: SlackConfig{
				APIToken:       getEnvAsString("SLACK_API_TOKEN", ""),
				ChannelID:      getEnvAsString("SLACK_CHANNEL_ID", ""),
				RealtimeEnable: getEnvAsBool("SLACK_REALTIME_ENABLED", true),
			},
			Template: TemplateConfig{
				CacheEnabled: getEnvAsBool("TEMPLATE_CACHE_ENABLED", true),
				TemplateDir:  getEnvAsString("TEMPLATE_DIR", "./web/templates"),
			},
			Logging: LoggingConfig{
				Level:  getEnvAsString("LOG_LEVEL", "info"),
				Format: getEnvAsString("LOG_FORMAT", "json"),
			},
		}
	})

	return &config
}

// Validate ensures all required configuration is present
func (c *Config) Validate() error {
	if c.Slack.APIToken == "" {
		return fmt.Errorf("SLACK_API_TOKEN is required")
	}
	if c.Slack.ChannelID == "" {
		return fmt.Errorf("SLACK_CHANNEL_ID is required")
	}
	return nil
}

// Helper functions to get environment variables with type conversion
func getEnvAsString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}
