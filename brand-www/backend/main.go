package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// FormSubmission represents the structure of form data
type FormSubmission struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Country   string `json:"country"`
	Message   string `json:"message"`
}

// Subscription represents the structure of a subscription
type Subscription struct {
	Email string `json:"email"`
}

// MockDB is a simple mock function simulating a database insert
func MockDB(data interface{}) error {
	// Simulate a successful database operation
	log.Printf("MockDB: Received data: %+v\n", data)
	return nil
}

// handleFormSubmit handles the POST request for form submission
func handleFormSubmit(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for both preflight and actual requests
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ensure it's a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse and decode the JSON body
	var submission FormSubmission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log the submission
	log.Printf("Received submission: %+v\n", submission)

	// Simulate saving to the database
	if err := MockDB(submission); err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleSubscribe handles the POST request for email subscription
func handleSubscribe(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for both preflight and actual requests
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ensure it's a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse and decode the JSON body
	var subscription Subscription
	if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log the subscription
	log.Printf("Received subscription: %+v\n", subscription)

	// Simulate saving to the database
	if err := MockDB(subscription); err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "subscribed"})
}

func main() {
	http.HandleFunc("/submit", handleFormSubmit)  // Endpoint for form submissions
	http.HandleFunc("/subscribe", handleSubscribe) // Endpoint for subscriptions

	// Start the server
	port := ":8082"
	log.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
