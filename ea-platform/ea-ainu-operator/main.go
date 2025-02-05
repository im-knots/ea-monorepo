package main

import (
	"ea-ainu-operator/config"
	"ea-ainu-operator/logger"
	"ea-ainu-operator/mongo"
	"ea-ainu-operator/operator"
	"ea-ainu-operator/routes"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	// Pass MongoDB client to operators
	operator.SetDBClient(dbClient)
	logger.Slog.Info("MongoDB client successfully passed to operators")

	// Initialize Gin router for metrics
	router := routes.RegisterRoutes()

	// Create a stop channel for informers
	stopCh := make(chan struct{})
	defer close(stopCh)

	go operator.StartOperators(stopCh)

	// Handle OS signals for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		logger.Slog.Info("Received termination signal, shutting down gracefully...")
		close(stopCh)
		os.Exit(0)
	}()

	// Start the server
	serverAddr := "0.0.0.0:" + config.Port
	logger.Slog.Info("Server starting", "address", serverAddr)
	log.Fatal(router.Run(serverAddr))
}
