package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ea-job-api/config"
	"ea-job-api/logger"
	"ea-job-api/metrics"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// Define structs to store agent definition after lookup
type Agent struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Creator     string        `json:"creator"`
	Description string        `json:"description"`
	Nodes       []Node        `json:"nodes"`
	Edges       []Edge        `json:"edges"`
	Metadata    AgentMetadata `json:"metadata"`
}

type Node struct {
	Alias      string                 `json:"alias"`
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

type Edge struct {
	From []string `json:"from"`
	To   []string `json:"to"`
}

type CreateJobRequest struct {
	AgentID string `json:"agent_id"`
	UserID  string `json:"user_id"`
}

// Metadata holds timestamps for Agents.
type AgentMetadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HandleCreateJob handles job creation requests.

func HandleCreateJob(c *gin.Context) {
	path := c.FullPath()

	metrics.StepCounter.WithLabelValues(path, "api_request_start", "success").Inc()
	logger.Slog.Info("Job creation request received")

	authenticatedUserID := c.GetString("AuthenticatedUserID")
	if authenticatedUserID == "" {
		logger.Slog.Error("Authenticated user ID missing in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		metrics.StepCounter.WithLabelValues(path, "invalid_request_body", "error").Inc()
		logger.Slog.Error("Failed to decode request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 🔹 Ensure non-internal users can only create jobs for themselves
	if authenticatedUserID != "internal" {
		if req.UserID != authenticatedUserID {
			logger.Slog.Error("User ID mismatch", "authenticated", authenticatedUserID, "request", req.UserID)
			metrics.StepCounter.WithLabelValues(path, "user_spoofing_attempt", "failure").Inc()
			c.JSON(http.StatusForbidden, gin.H{"error": "User ID does not match authenticated user"})
			return
		}
	}

	// Ensure agent ID is provided
	if req.AgentID == "" {
		logger.Slog.Error("Missing agent ID in request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing agent ID"})
		return
	}

	// Fetch the agent details from the Agent Manager **using the user's ID**
	cfg := config.LoadConfig()
	agentURL := fmt.Sprintf("%s%s", cfg.AgentManagerUrl, req.AgentID)

	// Create the request to the Agent Manager
	agentReq, err := http.NewRequest("GET", agentURL, nil)
	if err != nil {
		logger.Slog.Error("Failed to create request to Agent Manager", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Pass along the user's authorization token (so the agent manager applies user-based access controls)
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		agentReq.Header.Set("Authorization", authHeader)
	}

	// Send the request
	agentClient := &http.Client{}
	resp, err := agentClient.Do(agentReq)
	if err != nil {
		logger.Slog.Error("Failed to reach agent manager", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve agent"})
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		logger.Slog.Error("Agent manager returned non-200 response", "status", resp.StatusCode)
		c.JSON(resp.StatusCode, gin.H{"error": "Failed to retrieve agent"})
		return
	}

	// Parse the response body
	body, err := io.ReadAll(resp.Body)
	logger.Slog.Info("agent manager response body", "body", resp.Body)
	if err != nil {
		logger.Slog.Error("Failed to read response body", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse agent data"})
		return
	}

	var agent Agent
	if err := json.Unmarshal(body, &agent); err != nil {
		logger.Slog.Error("Failed to unmarshal agent JSON", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid agent data format"})
		return
	}

	if agent.ID == "" {
		logger.Slog.Error("Agent ID is missing from agent manager response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Agent data missing ID"})
		return
	}

	// Generate a unique job name
	hash := generateRandomHash()
	jobName := fmt.Sprintf("agentjob-%s-%s", agent.ID, hash)

	// Create Kubernetes client configuration
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Slog.Error("Failed to create in-cluster Kubernetes config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Kubernetes client config"})
		return
	}

	// Create a dynamic Kubernetes client
	dynamicClient, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		logger.Slog.Error("Failed to create dynamic Kubernetes client", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Kubernetes dynamic client"})
		return
	}

	// Define the GroupVersionResource (GVR) for the AgentJob CRD
	agentJobGVR := schema.GroupVersionResource{
		Group:    "ea.erulabs.ai",
		Version:  "v1",
		Resource: "agentjobs",
	}

	// Ensure parameters field retains complex structure
	var nodes []map[string]interface{}
	for _, node := range agent.Nodes {
		// Explicitly treat parameters as a deeply nested map[string]interface{}
		parametersMap := make(map[string]interface{})

		// Convert Parameters into JSON and back to preserve structure
		parametersJSON, err := json.Marshal(node.Parameters)
		if err != nil {
			logger.Slog.Error("Failed to marshal node parameters", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process node parameters"})
			return
		}
		if err := json.Unmarshal(parametersJSON, &parametersMap); err != nil {
			logger.Slog.Error("Failed to unmarshal node parameters into map", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process node parameters"})
			return
		}

		// Construct the node map with properly formatted parameters
		nodeMap := map[string]interface{}{
			"alias":      node.Alias,
			"type":       node.Type,
			"parameters": parametersMap, // Ensures structured handling
		}
		nodes = append(nodes, nodeMap)
	}

	// Define the AgentJob Custom Resource
	agentJob := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "ea.erulabs.ai/v1",
			"kind":       "AgentJob",
			"metadata": map[string]interface{}{
				"name":      jobName,
				"namespace": "ea-platform",
			},
			"spec": map[string]interface{}{
				"agentID": agent.ID,
				"name":    agent.Name,
				"user":    req.UserID,
				"creator": agent.Creator,
				"nodes":   nodes, // Now properly formatted
				"edges":   agent.Edges,
				"metadata": map[string]interface{}{
					"created_at": time.Now().Format(time.RFC3339),
					"updated_at": time.Now().Format(time.RFC3339),
				},
			},
		},
	}

	// Create the AgentJob CR in Kubernetes
	_, err = dynamicClient.Resource(agentJobGVR).
		Namespace("ea-platform").
		Create(context.TODO(), agentJob, metav1.CreateOptions{})

	if err != nil {
		logger.Slog.Error("Failed to create AgentJob custom resource", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job in Kubernetes"})
		return
	}

	logger.Slog.Info("Successfully created AgentJob CR", "jobName", jobName)
	c.JSON(http.StatusAccepted, gin.H{"status": "job created", "job_name": jobName, "user_id": req.UserID})
}

// Helper Functions
// generateRandomHash creates a random 6-character hexadecimal string
func generateRandomHash() string {
	b := make([]byte, 3) // 3 bytes = 6 hex characters
	_, err := rand.Read(b)
	if err != nil {
		logger.Slog.Error("Failed to generate random hash", "error", err)
		return "000000" // Fallback in case of error
	}
	return hex.EncodeToString(b)
}
