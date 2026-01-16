package main

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	DBEngine          string
	DBUsername        string
	DBPassword        string
	DBName            string
	DBHost            string
	DBPort            int
	ServerPort        string
	RedirectPrefixURL string
	APIURL            string // Remote API URL for CLI mode
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() Config {
	return Config{
		DBEngine:          getEnv("DB_ENGINE", "postgres"),
		DBUsername:        getEnv("DB_USERNAME", "marin"),
		DBPassword:        getEnv("DB_PASSWORD", "devel"),
		DBName:            getEnv("DB_NAME", "shorturl"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnvInt("DB_PORT", 5432),
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		RedirectPrefixURL: getEnv("REDIRECT_PREFIX_URL", "http://127.0.0.1:8080/r/"),
		APIURL:            getEnv("API_URL", ""), // Empty means local mode
	}
}

func getEnv(key, defaultValue string) string {
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
