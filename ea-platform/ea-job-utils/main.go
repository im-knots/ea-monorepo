package main

import (
	"log"

	"ea-job-utils/config"
	"ea-job-utils/logger"
	"ea-job-utils/routes"
)

func main() {
	// Set up the logger
	logger.Slog.Info("Starting the application")

	// Load configuration
	config := config.LoadConfig()

	// Initialize Gin router
	router := routes.RegisterRoutes()

	// DEBUG: Print all registered routes
	for _, r := range router.Routes() {
		log.Printf("Registered Route: %s %s\n", r.Method, r.Path)
	}

	// Start the server
	serverAddr := "0.0.0.0:" + config.Port
	logger.Slog.Info("Server starting", "address", serverAddr)
	log.Fatal(router.Run(serverAddr))
}
