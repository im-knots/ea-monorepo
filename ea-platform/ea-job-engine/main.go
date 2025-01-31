package main

import (
	"log"

	"ea-job-engine/config"
	"ea-job-engine/logger"
	"ea-job-engine/routes"
)

func main() {
	// Set up the logger
	logger.Slog.Info("Starting the application")

	// Load configuration
	config := config.LoadConfig()

	// Initialize Gin router
	router := routes.RegisterRoutes()

	// Start the server
	serverAddr := "0.0.0.0:" + config.Port
	logger.Slog.Info("Server starting", "address", serverAddr)
	log.Fatal(router.Run(serverAddr))
}
