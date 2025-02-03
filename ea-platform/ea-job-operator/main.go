package main

import (
	"ea-job-operator/config"
	"ea-job-operator/logger"
	"ea-job-operator/operator"
	"ea-job-operator/routes"
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
