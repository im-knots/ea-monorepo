package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"ea-credentials-manager/logger"
	"ea-credentials-manager/metrics"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Kubernetes namespace where secrets are stored
const KUBERNETES_NAMESPACE = "ea-platform"

// HandleAddCredential updates a user's Kubernetes secret using PATCH
func HandleAddCredential(c *gin.Context) {
	userId := c.GetString("AuthenticatedUserID")
	if userId == "" {
		logger.Slog.Error("Authenticated user ID missing in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	logger.Slog.Info("Processing credential update", "userId", userId)

	// Step 2: Parse the request body to get new credentials
	var requestBody map[string]string
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Slog.Error("Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request body"})
		return
	}

	// Step 3: Encode credentials in Base64 (Kubernetes Secrets require base64 encoding)
	patchData := []map[string]string{}
	for key, value := range requestBody {
		encodedValue := base64.StdEncoding.EncodeToString([]byte(value))

		// JSON Patch format for adding/updating a key
		patchData = append(patchData, map[string]string{
			"op":    "add",
			"path":  fmt.Sprintf("/data/%s", key),
			"value": encodedValue,
		})
	}

	// Step 4: Marshal patch data to JSON
	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		logger.Slog.Error("Failed to marshal patch data", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	// Step 5: Initialize Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Slog.Error("Failed to create Kubernetes config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Slog.Error("Failed to create Kubernetes client", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Step 6: Patch the Kubernetes Secret
	secretName := fmt.Sprintf("third-party-user-creds-%s", userId)
	_, err = clientset.CoreV1().Secrets(KUBERNETES_NAMESPACE).Patch(
		context.TODO(),
		secretName,
		types.JSONPatchType,
		patchBytes,
		metav1.PatchOptions{},
	)

	if err != nil {
		logger.Slog.Error("Failed to patch Kubernetes secret", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update secret"})
		return
	}

	// Step 7: Return success response
	logger.Slog.Info("Credential update successful", "userId", userId)
	metrics.StepCounter.WithLabelValues(c.Request.URL.Path, "patch-secret", "success").Inc()
	c.JSON(http.StatusOK, gin.H{"message": "✅ Credentials updated successfully!"})
}
