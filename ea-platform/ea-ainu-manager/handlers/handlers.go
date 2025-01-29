package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"ea-ainu-manager/logger"
	"ea-ainu-manager/metrics"
	"ea-ainu-manager/mongo"

	"go.mongodb.org/mongo-driver/bson"
)

// dbClient is the shared MongoDB client for handlers.
var dbClient *mongo.MongoClient

// SetDBClient sets the MongoDB client for handlers.
func SetDBClient(client *mongo.MongoClient) {
	if client == nil {
		logger.Slog.Error("SetDBClient called with nil client")
	}
	dbClient = client
	logger.Slog.Info("Database client successfully initialized in handlers")
}

// UserDefinition represents a user on the Ea platform
type UserDefinition struct {
	Name           string        `json:"name" bson:"name"`
	ComputeCredits int           `json:"compute_credits" bson:"compute_credits"`
	ComputeRate    float64       `json:"compute_rate" bson:"compute_rate"`
	InferenceJobs  int           `json:"inference_jobs" bson:"inference_jobs"`
	TrainingJobs   int           `json:"training_jobs" bson:"training_jobs"`
	Agents         int           `json:"agents" bson:"agents"`
	ComputeDevices []ComputeNode `json:"compute_devices" bson:"compute_devices"`
	Jobs           []AgentJob    `json:"jobs" bson:"jobs"`
}

// ComputeNode represents a user's compute device
type ComputeNode struct {
	DeviceName  string    `json:"device_name" bson:"device_name"`
	DeviceOS    string    `json:"device_os" bson:"device_os"`
	ComputeType string    `json:"compute_type" bson:"compute_type"`
	Status      string    `json:"status" bson:"status"`
	ComputeRate float64   `json:"compute_rate" bson:"compute_rate"`
	ID          string    `json:"id" bson:"id"`
	LastActive  time.Time `json:"last_active" bson:"last_active"`
}

// AgentJob represents an agent or job managed by the user
type AgentJob struct {
	JobName    string    `json:"job_name" bson:"job_name"`
	JobType    string    `json:"job_type" bson:"job_type"`
	Status     string    `json:"status" bson:"status"`
	LastActive time.Time `json:"last_active" bson:"last_active"`
}

// HandleCreateUser handles the creation of a User data entry
func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var input UserDefinition
	path := "/api/v1/users"

	if r.Method != http.MethodPost {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		logger.Slog.Error("Invalid request method", "method", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "decoding_request", "success").Inc()
	}

	// Insert UserDefinition into the "ainuUsers" database and "users" collection
	result, err := dbClient.InsertRecord("ainuUsers", "users", input)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_insertion_error", "error").Inc()
		logger.Slog.Error("Failed to insert node definition into database", "error", err)
		http.Error(w, "Failed to insert node definition into database", http.StatusInternalServerError)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "db_insertion", "success").Inc()
	}

	metrics.StepCounter.WithLabelValues(path, "create_success", "success").Inc()
	logger.Slog.Info("User inserted successfully", "ID", result.InsertedID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Ainulindale User created successfully",
		"user_id": result.InsertedID,
	})
}

// HandleGetAllUsers retrieves all ainulindale users from the database, but only their IDs and names.
func HandleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	path := "/api/v1/users"

	if r.Method != http.MethodGet {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Query all node definitions but only retrieve `id` and `name`
	projection := bson.M{
		"name": 1, // Include the `name` field
		"_id":  1, // Include the MongoDB internal `_id` field
	}
	users, err := dbClient.FindRecordsWithProjection("ainuUsers", "users", bson.M{}, projection)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve node definitions from database", "error", err)
		http.Error(w, "Failed to retrieve node definitions", http.StatusInternalServerError)
		return
	}

	metrics.StepCounter.WithLabelValues(path, "retrieval_success", "success").Inc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// HandleGetUser retrieves a specific User by ID.
func HandleGetUser(w http.ResponseWriter, r *http.Request) {
	path := "/api/v1/users/"

	if r.Method != http.MethodGet {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, path)
	if id == "" {
		metrics.StepCounter.WithLabelValues(path, "missing_id", "error").Inc()
		logger.Slog.Error("Missing User ID")
		http.Error(w, "Missing User ID", http.StatusBadRequest)
		return
	}

	user, err := dbClient.FindRecordByID("ainuUsers", "users", id)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve user from database", "error", err, "id", id)
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}

	metrics.StepCounter.WithLabelValues(path, "retrieval_success", "success").Inc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
