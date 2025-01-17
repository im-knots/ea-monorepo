package main

import (
	"log"
	"net/http"

	"ea-agent-manager/config"
	"ea-agent-manager/routes"
)

func main() {
	// Load configuration
	config := config.LoadConfig()

	log.Println("MongoDB client successfully passed to handlers")

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux)

	log.Printf("Server running on http://0.0.0.0:%s\n", config.Port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+config.Port, mux))
}
