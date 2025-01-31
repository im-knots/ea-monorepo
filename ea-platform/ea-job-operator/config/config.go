package config

import "os"

// Config holds application configuration.
type Config struct {
	Port    string
	GinMode string
}

// LoadConfig initializes the configuration from environment variables.
func LoadConfig() Config {
	return Config{
		Port:    getEnv("PORT", "8080"),
		GinMode: getEnv("GIN_MODE", "release"),
	}
}

// getEnv retrieves environment variables or defaults to a fallback value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
