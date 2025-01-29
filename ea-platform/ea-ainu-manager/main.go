package main

import (
	"log"
	"net/http"

	"ea-ainu-manager/config"
	"ea-ainu-manager/handlers"
	"ea-ainu-manager/logger"
	"ea-ainu-manager/mongo"
	"ea-ainu-manager/routes"
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

	// Pass MongoDB client to handlers
	handlers.SetDBClient(dbClient)
	logger.Slog.Info("MongoDB client successfully passed to handlers")

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux)

	logger.Slog.Info("Server starting", "address", "http://0.0.0.0:"+config.Port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+config.Port, mux))
}
