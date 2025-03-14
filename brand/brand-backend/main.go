package main

import (
	"log"
	"net/http"

	"brand-backend/config"
	"brand-backend/handlers"
	"brand-backend/mongo"
	"brand-backend/routes"
)

func main() {
	// Load configuration
	config := config.LoadConfig()

	// Initialize MongoDB client
	dbClient, err := mongo.NewMongoClient(config.DBURL)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer dbClient.Disconnect()

	// Pass MongoDB client to handlers
	handlers.SetDBClient(dbClient)
	log.Println("MongoDB client successfully passed to handlers")

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux)

	log.Printf("Server running on http://0.0.0.0:%s\n", config.Port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+config.Port, mux))
}
