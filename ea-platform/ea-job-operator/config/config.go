package config

import "os"

// Config holds application configuration.
type Config struct {
	Port                      string
	GinMode                   string
	FeatureNewAgentJobs       string
	FeatureInactiveAgentJobs  string
	FeatureCompletedJobs      string
	FeatureCompletedAgentJobs string
	FeatureCleanOrphans       string
}

// LoadConfig initializes the configuration from environment variables.
func LoadConfig() Config {
	return Config{
		Port:                      getEnv("PORT", "8080"),
		GinMode:                   getEnv("GIN_MODE", "release"),
		FeatureNewAgentJobs:       getEnv("FEATURE_NEW_AGENT_JOBS", "true"),
		FeatureInactiveAgentJobs:  getEnv("FEATURE_INACTIVE_AGENT_JOBS", "true"),
		FeatureCompletedJobs:      getEnv("FEATURE_COMPLETED_JOBS", "true"),
		FeatureCompletedAgentJobs: getEnv("FEATURE_COMPLETED_AGENT_JOBS", "true"),
		FeatureCleanOrphans:       getEnv("FEATURE_CLEAN_ORPHANS", "false"),
	}
}

// getEnv retrieves environment variables or defaults to a fallback value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
