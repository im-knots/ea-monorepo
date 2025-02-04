package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"ea-ainu-manager/logger"
	"ea-ainu-manager/metrics"
	"ea-ainu-manager/mongo"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// ---- MODELS ----

type UserDefinition struct {
	ID             string        `json:"id" bson:"id"`
	Name           string        `json:"name" bson:"name"`
	ComputeCredits int           `json:"compute_credits" bson:"compute_credits"`
	ComputeDevices []ComputeNode `json:"compute_devices" bson:"compute_devices"`
	Jobs           []AgentJob    `json:"jobs" bson:"jobs"`
	CreatedTime    time.Time     `json:"created_time" bson:"created_time"`
}

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

type AgentJob struct {
	JobName     string    `json:"job_name" bson:"job_name"`
	JobType     string    `json:"job_type" bson:"job_type"`
	Status      string    `json:"status" bson:"status"`
	LastActive  time.Time `json:"last_active" bson:"last_active"`
	ID          string    `json:"id" bson:"id"`
	CreatedTime time.Time `json:"created_time" bson:"created_time"`
}

// ---- USER HANDLERS ----

func HandleCreateUser(c *gin.Context) {
	metrics.StepCounter.WithLabelValues("/api/v1/users", "create_user", "request").Inc()

	var input UserDefinition
	if err := c.ShouldBindJSON(&input); err != nil {
		metrics.StepCounter.WithLabelValues("/api/v1/users", "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
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
		metrics.StepCounter.WithLabelValues("/api/v1/users", "db_insertion_error", "error").Inc()
		logger.Slog.Error("Failed to insert user into database", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert user"})
		return
	}

	metrics.StepCounter.WithLabelValues("/api/v1/users", "create_success", "success").Inc()
	logger.Slog.Info("User inserted successfully", "ID", result.InsertedID)
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "id": input.ID, "user": input.Name, "creat_time": input.CreatedTime})
}

func HandleGetAllUsers(c *gin.Context) {
	metrics.StepCounter.WithLabelValues("/api/v1/users", "get_all_users", "request").Inc()

	projection := bson.M{"name": 1, "id": 1, "_id": 0}
	users, err := dbClient.FindRecordsWithProjection("ainuUsers", "users", bson.M{}, projection)
	if err != nil {
		metrics.StepCounter.WithLabelValues("/api/v1/users", "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve users", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	metrics.StepCounter.WithLabelValues("/api/v1/users", "retrieval_success", "success").Inc()
	c.JSON(http.StatusOK, users)
}

func HandleGetUser(c *gin.Context) {
	userID := c.Param("user_id")
	metrics.StepCounter.WithLabelValues("/api/v1/users/:user_id", "get_user", "request").Inc()

	user, err := dbClient.FindRecordByID("ainuUsers", "users", userID)
	if err != nil {
		metrics.StepCounter.WithLabelValues("/api/v1/users/:user_id", "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve user", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	metrics.StepCounter.WithLabelValues("/api/v1/users/:user_id", "retrieval_success", "success").Inc()
	c.JSON(http.StatusOK, user)
}

// ---- COMPUTE DEVICE HANDLERS ----

func HandleAddComputeDevice(c *gin.Context) {
	path := "/api/v1/users/:user_id/devices"
	userID := c.Param("user_id")
	metrics.StepCounter.WithLabelValues(path, "add_device", "request").Inc()

	// Parse request body
	var newDevice ComputeNode
	if err := c.ShouldBindJSON(&newDevice); err != nil {
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Assign ID and timestamp
	newDevice.ID = uuid.New().String()
	newDevice.CreatedTime = time.Now()

	// Update user record in MongoDB
	update := bson.M{"$push": bson.M{"compute_devices": newDevice}}
	result, err := dbClient.UpdateRecord("ainuUsers", "users", bson.M{"id": userID}, update)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_update_error", "error").Inc()
		logger.Slog.Error("Failed to add compute device", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add compute device"})
		return
	}

	if result.ModifiedCount == 0 {
		metrics.StepCounter.WithLabelValues(path, "no_update", "warning").Inc()
		logger.Slog.Warn("No user found to update", "user_id", userID)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Success response
	metrics.StepCounter.WithLabelValues(path, "update_success", "success").Inc()
	logger.Slog.Info("Compute device added successfully", "user_id", userID, "device", newDevice)
	c.JSON(http.StatusOK, gin.H{"message": "Compute device added successfully", "device": newDevice})
}

func HandleDeleteComputeDevice(c *gin.Context) {
	path := "/api/v1/users/:user_id/devices/:device_id"
	userID := c.Param("user_id")
	deviceID := c.Param("device_id")

	metrics.StepCounter.WithLabelValues(path, "delete_device", "request").Inc()

	// Retrieve user record
	user, err := dbClient.FindRecordByID("ainuUsers", "users", userID)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve user", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	// Ensure compute_devices exists and is an array
	computeDevicesRaw, exists := user["compute_devices"]
	if !exists {
		logger.Slog.Error("Compute devices field missing", "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Compute devices field missing"})
		return
	}

	// Convert compute_devices to a JSON-compatible structure and re-decode
	computeDevicesBytes, err := json.Marshal(computeDevicesRaw)
	if err != nil {
		logger.Slog.Error("Failed to marshal compute devices", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process compute devices"})
		return
	}

	var computeDevices []map[string]interface{}
	if err := json.Unmarshal(computeDevicesBytes, &computeDevices); err != nil {
		logger.Slog.Error("Failed to unmarshal compute devices", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse compute devices"})
		return
	}

	// Locate the device name
	var deviceName string
	for _, dev := range computeDevices {
		if id, ok := dev["id"].(string); ok && id == deviceID {
			if name, exists := dev["device_name"].(string); exists {
				deviceName = name
			}
			break
		}
	}

	if deviceName == "" {
		metrics.StepCounter.WithLabelValues(path, "no_device_found", "warning").Inc()
		logger.Slog.Warn("No device found with given ID", "user_id", userID, "device_id", deviceID)
		c.JSON(http.StatusNotFound, gin.H{"error": "No device found with given ID"})
		return
	}

	filter := bson.M{"id": userID}
	update := bson.M{"$pull": bson.M{"compute_devices": bson.M{"id": deviceID}}}

	result, err := dbClient.UpdateRecord("ainuUsers", "users", filter, update)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_update_error", "error").Inc()
		logger.Slog.Error("Failed to remove compute device", "error", err, "user_id", userID, "device_id", deviceID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove compute device"})
		return
	}

	if result.ModifiedCount == 0 {
		metrics.StepCounter.WithLabelValues(path, "no_update", "warning").Inc()
		logger.Slog.Warn("No matching device found in database update", "user_id", userID, "device_id", deviceID)
		c.JSON(http.StatusNotFound, gin.H{"error": "No device found with given ID"})
		return
	}

	// Success response
	metrics.StepCounter.WithLabelValues(path, "delete_success", "success").Inc()
	logger.Slog.Info("Compute device removed successfully", "user_id", userID, "device_id", deviceID)
	c.JSON(http.StatusOK, gin.H{
		"message":     "Compute device removed successfully",
		"user_id":     userID,
		"device_id":   deviceID,
		"device_name": deviceName,
	})
}

// ---- JOB HANDLERS ----

func HandleAddJob(c *gin.Context) {
	path := "/api/v1/users/:user_id/jobs"
	userID := c.Param("user_id")
	metrics.StepCounter.WithLabelValues(path, "add_job", "request").Inc()

	// Parse request body
	var newJob AgentJob
	if err := c.ShouldBindJSON(&newJob); err != nil {
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Assign ID and timestamp
	newJob.ID = uuid.New().String()
	newJob.CreatedTime = time.Now()

	// Update user record in MongoDB
	update := bson.M{"$push": bson.M{"jobs": newJob}}
	result, err := dbClient.UpdateRecord("ainuUsers", "users", bson.M{"id": userID}, update)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_update_error", "error").Inc()
		logger.Slog.Error("Failed to add user Job", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user Job"})
		return
	}

	if result.ModifiedCount == 0 {
		metrics.StepCounter.WithLabelValues(path, "no_update", "warning").Inc()
		logger.Slog.Warn("No user found to update", "user_id", userID)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Success response
	metrics.StepCounter.WithLabelValues(path, "update_success", "success").Inc()
	logger.Slog.Info("User Job added successfully", "user_id", userID, "device", newJob)
	c.JSON(http.StatusOK, gin.H{
		"message": "User Job added successfully",
		"device":  newJob,
		"user_id": userID,
	})
}

// HandleDeleteJob removes a job from a user's record
func HandleDeleteJob(c *gin.Context) {
	path := "/api/v1/users/:user_id/jobs/:job_id"
	userID := c.Param("user_id")
	jobID := c.Param("job_id")

	metrics.StepCounter.WithLabelValues(path, "delete_job", "request").Inc()

	// Retrieve user record
	user, err := dbClient.FindRecordByID("ainuUsers", "users", userID)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve user", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	// Ensure jobs exist and are an array
	jobsRaw, exists := user["jobs"]
	if !exists {
		logger.Slog.Error("Jobs field missing", "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Jobs field missing"})
		return
	}

	// Convert jobs to a JSON-compatible structure and re-decode
	jobsBytes, err := json.Marshal(jobsRaw)
	if err != nil {
		logger.Slog.Error("Failed to marshal jobs", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process jobs"})
		return
	}

	var userJobs []map[string]interface{}
	if err := json.Unmarshal(jobsBytes, &userJobs); err != nil {
		logger.Slog.Error("Failed to unmarshal jobs", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse jobs"})
		return
	}

	// Locate the job name
	var jobName string
	for _, job := range userJobs {
		if id, ok := job["id"].(string); ok && id == jobID {
			if name, exists := job["job_name"].(string); exists {
				jobName = name
			}
			break
		}
	}

	if jobName == "" {
		metrics.StepCounter.WithLabelValues(path, "no_job_found", "warning").Inc()
		logger.Slog.Warn("No job found with given ID", "user_id", userID, "job_id", jobID)
		c.JSON(http.StatusNotFound, gin.H{"error": "No job found with given ID"})
		return
	}

	// Remove the job from the jobs array
	filter := bson.M{"id": userID}
	update := bson.M{"$pull": bson.M{"jobs": bson.M{"id": jobID}}}

	result, err := dbClient.UpdateRecord("ainuUsers", "users", filter, update)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_update_error", "error").Inc()
		logger.Slog.Error("Failed to remove job", "error", err, "user_id", userID, "job_id", jobID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove job"})
		return
	}

	if result.ModifiedCount == 0 {
		metrics.StepCounter.WithLabelValues(path, "no_update", "warning").Inc()
		logger.Slog.Warn("No matching job found in database update", "user_id", userID, "job_id", jobID)
		c.JSON(http.StatusNotFound, gin.H{"error": "No job found with given ID"})
		return
	}

	// Success response
	metrics.StepCounter.WithLabelValues(path, "delete_success", "success").Inc()
	logger.Slog.Info("Job removed successfully", "user_id", userID, "job_id", jobID)
	c.JSON(http.StatusOK, gin.H{
		"message":  "Job removed successfully",
		"user_id":  userID,
		"job_id":   jobID,
		"job_name": jobName,
	})
}

// HandleUpdateComputeCredits updates a user's compute credits
func HandleUpdateComputeCredits(c *gin.Context) {
	path := "/api/v1/users/:user_id/credits"
	userID := c.Param("user_id")

	metrics.StepCounter.WithLabelValues(path, "update_credits", "request").Inc()

	// Parse JSON request body
	var requestBody struct {
		ComputeCredits int `json:"compute_credits"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Slog.Error("Failed to parse request body", "error", err)
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate input (ensure compute credits are not negative)
	if requestBody.ComputeCredits < 0 {
		logger.Slog.Warn("Invalid compute credits value", "user_id", userID, "compute_credits", requestBody.ComputeCredits)
		metrics.StepCounter.WithLabelValues(path, "invalid_value", "warning").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Compute credits must be a non-negative integer"})
		return
	}

	// Update user record with the new compute credits
	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{"compute_credits": requestBody.ComputeCredits}}

	result, err := dbClient.UpdateRecord("ainuUsers", "users", filter, update)
	if err != nil {
		logger.Slog.Error("Failed to update user compute credits", "error", err, "user_id", userID)
		metrics.StepCounter.WithLabelValues(path, "db_update_error", "error").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update compute credits"})
		return
	}

	if result.ModifiedCount == 0 {
		logger.Slog.Warn("No user found to update", "user_id", userID)
		metrics.StepCounter.WithLabelValues(path, "no_update", "warning").Inc()
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Success response
	metrics.StepCounter.WithLabelValues(path, "update_success", "success").Inc()
	logger.Slog.Info("Compute credits updated successfully", "user_id", userID, "compute_credits", requestBody.ComputeCredits)

	c.JSON(http.StatusOK, gin.H{
		"message":         "Compute credits updated successfully",
		"user_id":         userID,
		"compute_credits": requestBody.ComputeCredits,
	})
}
