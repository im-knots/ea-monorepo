package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

// Application represents the structure of a job application
type Application struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	LinkedIn    string `json:"linkedin"`
	Position    string `json:"position"`
	CoverLetter string `json:"coverLetter"`
	ResumePath  string `json:"resumePath"`
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

// Update the Waitlist struct to include Desired Tier
func handleWaitlist(w http.ResponseWriter, r *http.Request) {
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

	var waitlistEntry struct {
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		Email       string `json:"email"`
		Username    string `json:"username"`
		DesiredTier string `json:"desiredTier"`
	}

	if err := json.NewDecoder(r.Body).Decode(&waitlistEntry); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received waitlist entry: %+v\n", waitlistEntry)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleApply handles the POST request for job applications
func handleApply(w http.ResponseWriter, r *http.Request) {
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

	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Extract form data
	firstName := r.FormValue("firstName")
	lastName := r.FormValue("lastName")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	linkedin := r.FormValue("linkedin")
	position := r.FormValue("position")
	coverLetter := r.FormValue("coverLetter")

	// Handle file upload
	file, handler, err := r.FormFile("resume")
	if err != nil {
		http.Error(w, "Failed to upload resume", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save the file to a local directory
	resumePath := fmt.Sprintf("./uploads/%s", handler.Filename)
	out, err := os.Create(resumePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Create an Application object
	application := Application{
		FirstName:   firstName,
		LastName:    lastName,
		Email:       email,
		Phone:       phone,
		LinkedIn:    linkedin,
		Position:    position,
		CoverLetter: coverLetter,
		ResumePath:  resumePath,
	}

	// Log the application and simulate saving to the database
	log.Printf("Received application: %+v\n", application)
	if err := MockDB(application); err != nil {
		http.Error(w, "Failed to save application", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func main() {
	http.HandleFunc("/submit", handleFormSubmit)
	http.HandleFunc("/subscribe", handleSubscribe)
	http.HandleFunc("/waitlist", handleWaitlist)
	http.HandleFunc("/apply", handleApply)

	// Bind to 0.0.0.0 to allow external connections
	port := ":8080"
	log.Printf("Server running on http://0.0.0.0%s\n", port)
	log.Fatal(http.ListenAndServe("0.0.0.0"+port, nil))
}
