package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"ea-ainu-manager/logger"
	"ea-ainu-manager/metrics"
	"ea-ainu-manager/mongo"

	"github.com/google/uuid"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	ID             string        `json:"id" bson:"id"`
	Name           string        `json:"name" bson:"name"`
	ComputeCredits int           `json:"compute_credits" bson:"compute_credits"`
	ComputeDevices []ComputeNode `json:"compute_devices" bson:"compute_devices"`
	Jobs           []AgentJob    `json:"jobs" bson:"jobs"`
	CreatedTime    time.Time     `json:"created_time" bson:"created_time"`
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
	CreatedTime time.Time `json:"created_time" bson:"created_time"`
}

// AgentJob represents an agent or job managed by the user
type AgentJob struct {
	JobName     string    `json:"job_name" bson:"job_name"`
	JobType     string    `json:"job_type" bson:"job_type"`
	Status      string    `json:"status" bson:"status"`
	LastActive  time.Time `json:"last_active" bson:"last_active"`
	ID          string    `json:"id" bson:"id"`
	CreatedTime time.Time `json:"created_time" bson:"created_time"`
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
	}

	input.ID = uuid.New().String()
	input.CreatedTime = time.Now()
	for i := range input.ComputeDevices {
		input.ComputeDevices[i].ID = uuid.New().String()
		input.ComputeDevices[i].CreatedTime = time.Now()
	}
	for i := range input.Jobs {
		input.Jobs[i].ID = uuid.New().String()
		input.Jobs[i].CreatedTime = time.Now()
	}

	result, err := dbClient.InsertRecord("ainuUsers", "users", input)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_insertion_error", "error").Inc()
		logger.Slog.Error("Failed to insert user into database", "error", err)
		http.Error(w, "Failed to insert user into database", http.StatusInternalServerError)
		return
	}

	metrics.StepCounter.WithLabelValues(path, "create_success", "success").Inc()
	logger.Slog.Info("User inserted successfully", "ID", result.InsertedID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "User created successfully",
		"user_id":    result.InsertedID,
		"user":       input.Name,
		"creat_time": input.CreatedTime,
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

// HandleAddComputeDevice adds a new compute device to an existing user
func HandleAddComputeDevice(w http.ResponseWriter, r *http.Request) {
	path := "/api/v1/users/"

	if r.Method != http.MethodPost {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from the URL
	segments := strings.Split(strings.TrimPrefix(r.URL.Path, path), "/")
	if len(segments) < 2 || segments[1] != "devices" {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	userID := segments[0]

	var newDevice ComputeNode
	if err := json.NewDecoder(r.Body).Decode(&newDevice); err != nil {
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Assign new UUID and CreatedTime to the device
	newDevice.ID = uuid.New().String()
	newDevice.CreatedTime = time.Now()

	// Convert user ID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "invalid_id", "error").Inc()
		logger.Slog.Error("Invalid user ID format", "error", err)
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Update user record with the new device
	filter := bson.M{"_id": objectID}
	update := bson.M{"$push": bson.M{"compute_devices": newDevice}}

	result, err := dbClient.UpdateRecord("ainuUsers", "users", filter, update)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_update_error", "error").Inc()
		logger.Slog.Error("Failed to update user record", "error", err)
		http.Error(w, "Failed to update user record", http.StatusInternalServerError)
		return
	}

	if result.ModifiedCount == 0 {
		metrics.StepCounter.WithLabelValues(path, "no_update", "warning").Inc()
		logger.Slog.Warn("No user found with given ID", "user_id", userID)
		http.Error(w, "No user found with given ID", http.StatusNotFound)
		return
	}

	metrics.StepCounter.WithLabelValues(path, "update_success", "success").Inc()
	logger.Slog.Info("User compute device added successfully", "user_id", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Compute device added successfully",
		"user_id": userID,
		"device":  newDevice,
	})
}

// HandleUpdateComputeCredits updates a user's compute credits
func HandleUpdateComputeCredits(w http.ResponseWriter, r *http.Request) {
	path := "/api/v1/users/"

	if r.Method != http.MethodPut {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from the URL
	segments := strings.Split(strings.TrimPrefix(r.URL.Path, path), "/")
	if len(segments) < 1 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	userID := segments[0]

	var requestBody struct {
		ComputeCredits int `json:"compute_credits"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Convert user ID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "invalid_id", "error").Inc()
		logger.Slog.Error("Invalid user ID format", "error", err)
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Update user record with the new compute credits
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"compute_credits": requestBody.ComputeCredits}}

	result, err := dbClient.UpdateRecord("ainuUsers", "users", filter, update)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_update_error", "error").Inc()
		logger.Slog.Error("Failed to update user compute credits", "error", err)
		http.Error(w, "Failed to update user compute credits", http.StatusInternalServerError)
		return
	}

	if result.ModifiedCount == 0 {
		metrics.StepCounter.WithLabelValues(path, "no_update", "warning").Inc()
		logger.Slog.Warn("No user found with given ID", "user_id", userID)
		http.Error(w, "No user found with given ID", http.StatusNotFound)
		return
	}

	metrics.StepCounter.WithLabelValues(path, "update_success", "success").Inc()
	logger.Slog.Info("User compute credits updated successfully", "user_id", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":         "Compute credits updated successfully",
		"user_id":         userID,
		"compute_credits": requestBody.ComputeCredits,
	})
}

// HandleDeleteComputeDevice removes a compute device from a user's record
func HandleDeleteComputeDevice(w http.ResponseWriter, r *http.Request) {
	path := "/api/v1/users/"

	if r.Method != http.MethodDelete {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID and device ID from the URL
	segments := strings.Split(strings.TrimPrefix(r.URL.Path, path), "/")
	if len(segments) < 3 || segments[1] != "devices" {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	userID := segments[0]
	deviceID := segments[2]

	// Convert user ID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "invalid_id", "error").Inc()
		logger.Slog.Error("Invalid user ID format", "error", err)
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Find user record
	user, err := dbClient.FindRecordByID("ainuUsers", "users", userID)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve user", "error", err)
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}

	// Extract compute_devices list safely and unmarshal it into a slice of ComputeNode
	computeDevicesRaw, exists := user["compute_devices"]
	if !exists {
		http.Error(w, "Compute devices field missing from user record", http.StatusInternalServerError)
		logger.Slog.Error("Compute devices field missing", "user_id", userID)
		return
	}

	computeDevicesData, err := json.Marshal(computeDevicesRaw)
	if err != nil {
		http.Error(w, "Failed to process compute devices", http.StatusInternalServerError)
		logger.Slog.Error("Failed to marshal compute devices", "error", err, "user_id", userID)
		return
	}

	var computeDevices []ComputeNode
	if err := json.Unmarshal(computeDevicesData, &computeDevices); err != nil {
		http.Error(w, "Failed to parse compute devices", http.StatusInternalServerError)
		logger.Slog.Error("Failed to unmarshal compute devices", "error", err, "user_id", userID)
		return
	}

	// Extract device name
	var deviceName string
	for _, dev := range computeDevices {
		if dev.ID == deviceID {
			deviceName = dev.DeviceName
			break
		}
	}

	if deviceName == "" {
		http.Error(w, "Device not found", http.StatusNotFound)
		logger.Slog.Error("Failed to find device ID", "user_id", userID, "device_id", deviceID)
		return
	}

	// Remove the compute device from the user's record
	filter := bson.M{"_id": objectID}
	update := bson.M{"$pull": bson.M{"compute_devices": bson.M{"id": deviceID}}}

	result, err := dbClient.UpdateRecord("ainuUsers", "users", filter, update)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_update_error", "error").Inc()
		logger.Slog.Error("Failed to remove compute device", "error", err)
		http.Error(w, "Failed to remove compute device", http.StatusInternalServerError)
		return
	}

	if result.ModifiedCount == 0 {
		metrics.StepCounter.WithLabelValues(path, "no_update", "warning").Inc()
		logger.Slog.Warn("No device found with given ID", "device_id", deviceID)
		http.Error(w, "No device found with given ID", http.StatusNotFound)
		return
	}

	metrics.StepCounter.WithLabelValues(path, "delete_success", "success").Inc()
	logger.Slog.Info("Compute device removed successfully", "user_id", userID, "device_id", deviceID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Compute device removed successfully",
		"user_id":     userID,
		"device_id":   deviceID,
		"device_name": deviceName,
	})
}

// HandleAddJob adds a new user job to an existing user
func HandleAddJob(w http.ResponseWriter, r *http.Request) {
	path := "/api/v1/users/"

	if r.Method != http.MethodPost {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from the URL
	segments := strings.Split(strings.TrimPrefix(r.URL.Path, path), "/")
	if len(segments) < 2 || segments[1] != "jobs" {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	userID := segments[0]

	var newJob AgentJob
	if err := json.NewDecoder(r.Body).Decode(&newJob); err != nil {
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Assign new UUID and CreatedTime to the device
	newJob.ID = uuid.New().String()
	newJob.CreatedTime = time.Now()

	// Convert user ID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "invalid_id", "error").Inc()
		logger.Slog.Error("Invalid user ID format", "error", err)
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Update user record with the new device
	filter := bson.M{"_id": objectID}
	update := bson.M{"$push": bson.M{"jobs": newJob}}

	result, err := dbClient.UpdateRecord("ainuUsers", "users", filter, update)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_update_error", "error").Inc()
		logger.Slog.Error("Failed to update user record", "error", err)
		http.Error(w, "Failed to update user record", http.StatusInternalServerError)
		return
	}

	if result.ModifiedCount == 0 {
		metrics.StepCounter.WithLabelValues(path, "no_update", "warning").Inc()
		logger.Slog.Warn("No user found with given ID", "user_id", userID)
		http.Error(w, "No user found with given ID", http.StatusNotFound)
		return
	}

	metrics.StepCounter.WithLabelValues(path, "update_success", "success").Inc()
	logger.Slog.Info("User job added successfully", "user_id", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User job added successfully",
		"user_id": userID,
		"job":     newJob,
	})
}

// HandleDeleteJob removes a user Job from a user's record
func HandleDeleteJob(w http.ResponseWriter, r *http.Request) {
	path := "/api/v1/users/"

	if r.Method != http.MethodDelete {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID and device ID from the URL
	segments := strings.Split(strings.TrimPrefix(r.URL.Path, path), "/")
	if len(segments) < 3 || segments[1] != "jobs" {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	userID := segments[0]
	jobID := segments[2]

	// Convert user ID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "invalid_id", "error").Inc()
		logger.Slog.Error("Invalid user ID format", "error", err)
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Find user record
	user, err := dbClient.FindRecordByID("ainuUsers", "users", userID)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve user", "error", err)
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}

	// Extract jobs list safely and unmarshal it into a slice of ComputeNode
	jobsRaw, exists := user["jobs"]
	if !exists {
		http.Error(w, "Jobs field missing from user record", http.StatusInternalServerError)
		logger.Slog.Error("Jobs field missing", "user_id", userID)
		return
	}

	jobsData, err := json.Marshal(jobsRaw)
	if err != nil {
		http.Error(w, "Failed to process user jobs", http.StatusInternalServerError)
		logger.Slog.Error("Failed to marshal user jobs", "error", err, "user_id", userID)
		return
	}

	var userJobs []AgentJob
	if err := json.Unmarshal(jobsData, &userJobs); err != nil {
		http.Error(w, "Failed to parse user jobs", http.StatusInternalServerError)
		logger.Slog.Error("Failed to unmarshal user jobs", "error", err, "user_id", userID)
		return
	}

	// Extract job name
	var jobName string
	for _, job := range userJobs {
		if job.ID == jobID {
			jobName = job.JobName
			break
		}
	}

	if jobName == "" {
		http.Error(w, "Job not found", http.StatusNotFound)
		logger.Slog.Error("Failed to find Job ID", "user_id", userID, "job_id", jobID)
		return
	}

	// Remove the user job from the user's record
	filter := bson.M{"_id": objectID}
	update := bson.M{"$pull": bson.M{"jobs": bson.M{"id": jobID}}}

	result, err := dbClient.UpdateRecord("ainuUsers", "users", filter, update)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_update_error", "error").Inc()
		logger.Slog.Error("Failed to remove user job", "error", err)
		http.Error(w, "Failed to remove user job", http.StatusInternalServerError)
		return
	}

	if result.ModifiedCount == 0 {
		metrics.StepCounter.WithLabelValues(path, "no_update", "warning").Inc()
		logger.Slog.Warn("No job found with given ID", "job_id", jobID)
		http.Error(w, "No job found with given ID", http.StatusNotFound)
		return
	}

	metrics.StepCounter.WithLabelValues(path, "delete_success", "success").Inc()
	logger.Slog.Info("User job removed successfully", "user_id", userID, "job_id", jobID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "User job removed successfully",
		"user_id":  userID,
		"job_id":   jobID,
		"job_name": jobName,
	})
}
