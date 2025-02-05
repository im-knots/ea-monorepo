package config

import (
	"os"
)

// Config holds application configuration.
type Config struct {
	Port                      string
	GinMode                   string
	FeatureNewAgentJobs       string
	FeatureInactiveAgentJobs  string
	FeatureCompletedAgentJobs string
	FeatureErrorAgentJobs     string
	DBURL                     string
}

// LoadConfig initializes the configuration from environment variables.
func LoadConfig() Config {
	return Config{
		Port:                      getEnv("PORT", "8088"),
		GinMode:                   getEnv("GIN_MODE", "release"),
		FeatureNewAgentJobs:       getEnv("FEATURE_NEW_AGENT_JOBS", "true"),
		FeatureInactiveAgentJobs:  getEnv("FEATURE_INACTIVE_AGENT_JOBS", "true"),
		FeatureCompletedAgentJobs: getEnv("FEATURE_COMPLETED_AGENT_JOBS", "true"),
		FeatureErrorAgentJobs:     getEnv("FEATURE_ERROR_AGENT_JOBS", "true"),
		DBURL:                     getEnv("DB_URL", "mongodb://admin:password@localhost:27017"),
	}
}

// getEnv retrieves environment variables or defaults to a fallback value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
