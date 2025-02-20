package config

import (
	"os"
	"strconv"
)

// Config holds application configuration.
type Config struct {
	Port                        string
	GinMode                     string
	FeatureNewAgentJobs         string
	FeatureInactiveAgentJobs    string
	FeatureCompletedJobs        string
	FeatureCompletedAgentJobs   string
	FeatureNodeStatusUpdates    string
	CompletedCleanupGracePeriod int
}

// LoadConfig initializes the configuration from environment variables.
func LoadConfig() Config {
	gracePeriod, err := strconv.Atoi(getEnv("CLEANUP_GRACE_PERIOD", "5000"))
	if err != nil {
		gracePeriod = 0 // Default to 0 if conversion fails
	}

	return Config{
		Port:                        getEnv("PORT", "8080"),
		GinMode:                     getEnv("GIN_MODE", "release"),
		FeatureNewAgentJobs:         getEnv("FEATURE_NEW_AGENT_JOBS", "true"),
		FeatureInactiveAgentJobs:    getEnv("FEATURE_INACTIVE_AGENT_JOBS", "true"),
		FeatureCompletedJobs:        getEnv("FEATURE_COMPLETED_JOBS", "true"),
		FeatureCompletedAgentJobs:   getEnv("FEATURE_COMPLETED_AGENT_JOBS", "true"),
		FeatureNodeStatusUpdates:    getEnv("FEATURE_NODE_STATUS_UPDATES", "true"),
		CompletedCleanupGracePeriod: gracePeriod,
	}
}

// getEnv retrieves environment variables or defaults to a fallback value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
