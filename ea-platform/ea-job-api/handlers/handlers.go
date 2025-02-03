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
	ID          string `json:"_id"`
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

	if req.AgentID == "" {
		logger.Slog.Error("Missing agent ID in request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing agent ID"})
		return
	}

	cfg := config.LoadConfig()
	agentURL := fmt.Sprintf("%s%s", cfg.AgentManagerUrl, req.AgentID)
	resp, err := http.Get(agentURL)
	if err != nil {
		logger.Slog.Error("Failed to reach agent manager", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve agent"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Slog.Error("Agent manager returned non-200 response", "status", resp.StatusCode)
		c.JSON(resp.StatusCode, gin.H{"error": "Failed to retrieve agent"})
		return
	}

	body, err := io.ReadAll(resp.Body)
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
				"user":    agent.User,
				"nodes":   agent.Nodes,
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
	c.JSON(http.StatusAccepted, gin.H{"status": "job created", "jobName": jobName})
}
