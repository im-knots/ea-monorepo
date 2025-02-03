package main

import (
	"ea-job-operator/config"
	"ea-job-operator/logger"
	"ea-job-operator/operator"
	"ea-job-operator/routes"
	"log"
)

func main() {
	// Set up the logger
	logger.Slog.Info("Starting the application")

	// Load configuration
	config := config.LoadConfig()

	// Initialize Gin router for metrics
	router := routes.RegisterRoutes()

	// Start the blank > Inactive Agent job watcher operator in a goroutine
	if config.FeatureNewAgentJobs == "true" {
		go func() {
			operator.WatchNewAgentJobs()
		}()
	}

	// Start the inactive > executing Agent job watcher operator in a goroutine
	if config.FeatureInactiveAgentJobs == "true" {
		go func() {
			operator.WatchInactiveAgentJobs()
		}()
	}

	// Start the executing > complete k8s job watcher operator in a goroutine
	if config.FeatureCompletedJobs == "true" {
		go func() {
			operator.WatchCompletedJobs()
		}()
	}

	// Start the complete > deleted Agent job watcher operator in a goroutine
	if config.FeatureCompletedAgentJobs == "true" {
		go func() {
			operator.WatchCompletedAgentJobs()
		}()
	}

	// Start the orphan cleanup/unlock operator in a goroutine
	if config.FeatureCleanOrphans == "true" {
		go func() {
			operator.WatchCleanOrphans()
		}()
	}

	// Start the server
	serverAddr := "0.0.0.0:" + config.Port
	logger.Slog.Info("Server starting", "address", serverAddr)
	log.Fatal(router.Run(serverAddr))
}
