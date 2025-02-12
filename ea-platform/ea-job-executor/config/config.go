package config

import "os"

// Config holds application configuration.
type Config struct {
	Port             string
	AgentManagerUrl  string
	FeatureK8sEvents string
}

// LoadConfig initializes the configuration from environment variables.
func LoadConfig() Config {
	return Config{
		//AgentManagerUrl: getEnv("AGENT_MANAGER_URL", "http://agent-manager.ea.erulabs.local/api/v1"), //for local testing
		AgentManagerUrl:  getEnv("AGENT_MANAGER_URL", "http://ea-agent-manager.ea-platform.svc.cluster.local:8080/api/v1"),
		FeatureK8sEvents: getEnv("FEATURE_K8S_EVENTS", "true"),
	}
}

// getEnv retrieves environment variables or defaults to a fallback value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
