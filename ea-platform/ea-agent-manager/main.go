package main

import (
	"log"

	"ea-agent-manager/config"
	"ea-agent-manager/handlers"
	"ea-agent-manager/logger"
	"ea-agent-manager/mongo"
	"ea-agent-manager/routes"
)

func main() {
	// Set up the logger
	logger.Slog.Info("Starting the application")

	// Load configuration
	config := config.LoadConfig()

	// Initialize MongoDB client
	dbClient, err := mongo.NewMongoClient(config.DBURL)
	if err != nil {
		logger.Slog.Error("Failed to connect to MongoDB", "error", err)
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer dbClient.Disconnect()

	// Set the initialized MongoDB client in handlers
	handlers.SetDBClient(dbClient)

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
