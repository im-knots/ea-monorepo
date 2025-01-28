package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"ea-job-engine/config"
	"ea-job-engine/logger"
	"ea-job-engine/metrics"
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
func HandleCreateJob(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if r.Method != http.MethodPost {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		logger.Slog.Error("Invalid request method", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "api_request_start", "success").Inc()
		logger.Slog.Info("method", r.Method, "Job creation request received")
	}

	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		metrics.StepCounter.WithLabelValues(path, "invalid_request_body", "error").Inc()
		logger.Slog.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "valid_request_body", "success").Inc()
		logger.Slog.Info("Received valid request body", "agentID", req.AgentID)
	}

	if req.AgentID == "" {
		metrics.StepCounter.WithLabelValues(path, "missing_agent_id", "error").Inc()
		logger.Slog.Error("Missing agent ID in request body")
		http.Error(w, "Missing agent ID", http.StatusBadRequest)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "agent_id_present", "success").Inc()
	}

	cfg := config.LoadConfig()
	agentURL := cfg.AgentManagerUrl + req.AgentID
	resp, err := http.Get(agentURL)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "agent_manager_http_error", "error").Inc()
		logger.Slog.Error("Failed to reach agent manager", "error", err)
		http.Error(w, "Failed to retrieve agent", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		metrics.StepCounter.WithLabelValues(path, "agent_manager_non_200", "error").Inc()
		logger.Slog.Error("Agent manager returned non-200 response", "status", resp.StatusCode)
		http.Error(w, "Failed to retrieve agent", resp.StatusCode)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "agent_manager_success", "success").Inc()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "agent_manager_read_error", "error").Inc()
		logger.Slog.Error("Failed to read response body", "error", err)
		http.Error(w, "Failed to parse agent data", http.StatusInternalServerError)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "agent_manager_read_success", "success").Inc()
	}

	var agent Agent
	if err := json.Unmarshal(body, &agent); err != nil {
		metrics.StepCounter.WithLabelValues(path, "agent_manager_unmarshal_error", "error").Inc()
		logger.Slog.Error("Failed to unmarshal agent JSON", "error", err)
		http.Error(w, "Invalid agent data format", http.StatusInternalServerError)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "agent_manager_unmarshal_success", "success").Inc()
		logger.Slog.Info("Loaded agent from agent manager", "agent", agent)
	}

	// TODO: Create a Kubernetes CRD based on the loaded agent
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "job accepted"})
}

// HandleGetAllJobs handles retrieving all jobs.
func HandleGetAllJobs(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if r.Method != http.MethodGet {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		logger.Slog.Error("Invalid request method", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	metrics.StepCounter.WithLabelValues(path, "api_hit", "success").Inc()
	logger.Slog.Info("Get all jobs request received")
}

// HandleGetJob handles retrieving a specific job by ID.
func HandleGetJob(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if r.Method != http.MethodGet {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		logger.Slog.Error("Invalid request method", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	metrics.StepCounter.WithLabelValues(path, "api_hit", "success").Inc()
	logger.Slog.Info("Get job request received")
}
