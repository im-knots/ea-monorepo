package config

import "os"

// Config holds application configuration.
type Config struct {
	Port            string
	AgentManagerUrl string
}

// LoadConfig initializes the configuration from environment variables.
func LoadConfig() Config {
	return Config{
		Port:            getEnv("PORT", "8080"),
		AgentManagerUrl: getEnv("AGENT_MANAGER_URL", "http://ea-agent-manager:8080/api/v1/agents/"),
	}
}

// getEnv retrieves environment variables or defaults to a fallback value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
