// Package config handles application configuration from environment variables and .env files.
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadConfig loads environment variables from config/app.env if it exists.
func LoadConfig() {
	err := godotenv.Load("config/app.env")
	if err != nil {
		// It's okay if the file doesn't exist, we'll use system env vars or defaults.
		log.Println("Note: config/app.env not found, using system environment variables.")
	}
}

// GetEnv returns the value of an environment variable or a default value if not set.
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
