package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"ea-job-engine/config"
	"ea-job-engine/logger"
	"ea-job-engine/metrics"

	"github.com/gin-gonic/gin"
)

// Define structs to store agent definition after lookup
type Agent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	User        string `json:"user"`
	Nodes       []Node `json:"nodes"`
	Edges       []Edge `json:"edges"`
}

type Node struct {
	ID            string                 `json:"id"`
	DefinitionRef string                 `json:"definition_ref"`
	Parameters    map[string]interface{} `json:"parameters"`
}

type Edge struct {
	From []string `json:"from"`
	To   []string `json:"to"`
}

type CreateJobRequest struct {
	AgentID string `json:"agentID"`
}

// HandleCreateJob handles job creation requests.
func HandleCreateJob(c *gin.Context) {
	path := c.FullPath()

	metrics.StepCounter.WithLabelValues(path, "api_request_start", "success").Inc()
	logger.Slog.Info("Job creation request received")

	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		metrics.StepCounter.WithLabelValues(path, "invalid_request_body", "error").Inc()
		logger.Slog.Error("Failed to decode request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "valid_request_body", "success").Inc()
	logger.Slog.Info("Received valid request body", "agentID", req.AgentID)

	if req.AgentID == "" {
		metrics.StepCounter.WithLabelValues(path, "missing_agent_id", "error").Inc()
		logger.Slog.Error("Missing agent ID in request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing agent ID"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "agent_id_present", "success").Inc()
	cfg := config.LoadConfig()
	agentURL := cfg.AgentManagerUrl + req.AgentID
	resp, err := http.Get(agentURL)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "agent_manager_http_error", "error").Inc()
		logger.Slog.Error("Failed to reach agent manager", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve agent"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		metrics.StepCounter.WithLabelValues(path, "agent_manager_non_200", "error").Inc()
		logger.Slog.Error("Agent manager returned non-200 response", "status", resp.StatusCode)
		c.JSON(resp.StatusCode, gin.H{"error": "Failed to retrieve agent"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "agent_manager_success", "success").Inc()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "agent_manager_read_error", "error").Inc()
		logger.Slog.Error("Failed to read response body", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse agent data"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "agent_manager_read_success", "success").Inc()
	var agent Agent
	if err := json.Unmarshal(body, &agent); err != nil {
		metrics.StepCounter.WithLabelValues(path, "agent_manager_unmarshal_error", "error").Inc()
		logger.Slog.Error("Failed to unmarshal agent JSON", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid agent data format"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "agent_manager_unmarshal_success", "success").Inc()
	logger.Slog.Info("Loaded agent from agent manager", "agent", agent)

	// TODO: Create a Kubernetes CRD based on the loaded agent
	c.JSON(http.StatusAccepted, gin.H{"status": "job accepted"})
}

// HandleGetAllJobs handles retrieving all jobs.
func HandleGetAllJobs(c *gin.Context) {
	path := c.FullPath()
	metrics.StepCounter.WithLabelValues(path, "api_hit", "success").Inc()
	logger.Slog.Info("Get all jobs request received")
	c.JSON(http.StatusOK, gin.H{"message": "List of all jobs"})
}

// HandleGetJob handles retrieving a specific job by ID.
func HandleGetJob(c *gin.Context) {
	path := c.FullPath()
	metrics.StepCounter.WithLabelValues(path, "api_hit", "success").Inc()
	logger.Slog.Info("Get job request received")
	c.JSON(http.StatusOK, gin.H{"message": "Job details"})
}
