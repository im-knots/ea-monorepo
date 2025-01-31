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

	// Start the blank > Inactive job watcher operator in a goroutine
	go func() {
		operator.WatchNewAgentJobs()
	}()

	// Start the inactive > executing job watcher operator in a goroutine
	go func() {
		operator.WatchInactiveAgentJobs()
	}()

	// Start the executing > complete job watcher operator in a goroutine
	go func() {
		operator.WatchCompletedJobs()
	}()

	// Start the server
	serverAddr := "0.0.0.0:" + config.Port
	logger.Slog.Info("Server starting", "address", serverAddr)
	log.Fatal(router.Run(serverAddr))
}
