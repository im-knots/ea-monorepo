package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"brand-backend/mongo"
)

// Subscription represents a subscription request.
type Subscription struct {
	Email string `json:"email"`
}

type FormSubmission struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Country   string `json:"country"`
	Message   string `json:"message"`
}

// WaitlistEntry represents a waitlist entry.
type WaitlistEntry struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	DesiredTier string `json:"desiredTier"`
}

// dbClient is the shared MongoDB client for handlers.
var dbClient *mongo.MongoClient

// SetDBClient sets the MongoDB client for handlers.
func SetDBClient(client *mongo.MongoClient) {
	if client == nil {
		log.Fatal("SetDBClient called with nil client")
	}
	dbClient = client
	log.Println("Database client successfully initialized in handlers")
}

// HandleSubscribe processes subscription requests.
func HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if dbClient == nil {
		http.Error(w, "Database client is not initialized", http.StatusInternalServerError)
		log.Println("HandleSubscribe error: dbClient is nil")
		return
	}

	var subscription Subscription
	if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Insert the subscription into the MongoDB collection
	_, err := dbClient.InsertRecord("myDatabase", "emailSubscriptions", subscription)
	if err != nil {
		log.Printf("Failed to save subscription: %v\n", err)
		http.Error(w, "Failed to save subscription", http.StatusInternalServerError)
		return
	}

	log.Printf("Received subscription: %+v\n", subscription)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "subscribed"})
}

// HandleContact processes form submission requests.
func HandleContact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if dbClient == nil {
		http.Error(w, "Database client is not initialized", http.StatusInternalServerError)
		log.Println("HandleContact error: dbClient is nil")
		return
	}

	var submission FormSubmission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Insert the form submission into the MongoDB collection
	_, err := dbClient.InsertRecord("myDatabase", "contactForms", submission)
	if err != nil {
		log.Printf("Failed to save contact form: %v\n", err)
		http.Error(w, "Failed to save contact form", http.StatusInternalServerError)
		return
	}

	log.Printf("Received submission: %+v\n", submission)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// HandleWaitlist processes waitlist entries.
func HandleWaitlist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if dbClient == nil {
		http.Error(w, "Database client is not initialized", http.StatusInternalServerError)
		log.Println("HandleWaitlist error: dbClient is nil")
		return
	}

	var entry WaitlistEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Insert the waitlist entry into the MongoDB collection
	_, err := dbClient.InsertRecord("myDatabase", "waitlistEntries", entry)
	if err != nil {
		log.Printf("Failed to save waitlist entry: %v\n", err)
		http.Error(w, "Failed to save waitlist entry", http.StatusInternalServerError)
		return
	}

	log.Printf("Received waitlist entry: %+v\n", entry)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
