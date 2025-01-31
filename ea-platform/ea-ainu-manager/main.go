package main

import (
	"log"

	"ea-ainu-manager/config"
	"ea-ainu-manager/handlers"
	"ea-ainu-manager/logger"
	"ea-ainu-manager/mongo"
	"ea-ainu-manager/routes"
)

func main() {
	// Set up logger
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

	// Pass MongoDB client to handlers
	handlers.SetDBClient(dbClient)
	logger.Slog.Info("MongoDB client successfully passed to handlers")

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
