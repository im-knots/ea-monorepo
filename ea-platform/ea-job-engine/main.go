package main

import (
	"log"
	"net/http"

	"ea-job-engine/config"
	"ea-job-engine/logger"
	"ea-job-engine/routes"
)

func main() {
	// Set up the logger
	logger.Slog.Info("Starting the application")

	// Load configuration
	config := config.LoadConfig()

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux)

	// Start watching for custom resource definitions (CRDs)
	// go operator.WatchCRDs()

	logger.Slog.Info("Server starting", "address", "http://0.0.0.0:"+config.Port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+config.Port, mux))
}
