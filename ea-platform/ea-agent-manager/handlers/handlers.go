package handlers

import (
	"net/http"
)

// HandleRoot checks MongoDB connection and returns a 200 status if successful.
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	// w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// if r.Method == http.MethodOptions {
	// 	w.WriteHeader(http.StatusOK)
	// 	return
	// }

	// if r.Method != http.MethodGet {
	// 	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	// 	return
	// }

	// if dbClient == nil {
	// 	http.Error(w, "Database client is not initialized", http.StatusInternalServerError)
	// 	log.Println("HandleRoot error: dbClient is nil")
	// 	return
	// }

	// // Test the connection to the MongoDB collection
	// collectionName := "yourCollectionName" // Replace with your actual collection name
	// if err := dbClient.TestConnection("ea-www-agent-manager", collectionName); err != nil {
	// 	http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
	// 	log.Printf("Failed to connect to MongoDB collection '%s': %v\n", collectionName, err)
	// 	return
	// }

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("MongoDB connection successful"))
}
