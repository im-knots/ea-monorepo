package main

import (
	"log"

	"ea-job-operator/config"
	"ea-job-operator/logger"
	"ea-job-operator/routes"
)

func main() {
	// Set up the logger
	logger.Slog.Info("Starting the application")

	// Load configuration
	config := config.LoadConfig()

	// Initialize Gin router for metrics
	router := routes.RegisterRoutes()

	// Start the server
	serverAddr := "0.0.0.0:" + config.Port
	logger.Slog.Info("Server starting", "address", serverAddr)
	log.Fatal(router.Run(serverAddr))

	// Probably start the operator goroutine

	// Probably start the operator CR cleanup goroutine
}
