package handlers

import (
	"encoding/base64"
	"net/http"

	"ea-job-utils/logger"
	"ea-job-utils/metrics"

	"github.com/gin-gonic/gin"
)

// Base64DecodeRequest represents the expected JSON payload
type Base64DecodeRequest struct {
	Data string `json:"data"`
}

// HandleBase64Decode handles decoding of a base64 string from a JSON payload
func HandleBase64Decode(c *gin.Context) {
	var req Base64DecodeRequest

	// Bind JSON input
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Slog.Error("Invalid request payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if req.Data == "" {
		logger.Slog.Error("Missing 'data' field in request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'data' field in request payload"})
		return
	}

	// Decode base64 string
	decodedBytes, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		logger.Slog.Error("Failed to decode base64 string", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 string"})
		return
	}

	// Convert bytes to string
	decodedStr := string(decodedBytes)

	// Log the successful decoding
	logger.Slog.Info("Successfully decoded base64 string", "decoded_value", decodedStr)

	// Increment metrics
	metrics.StepCounter.WithLabelValues("/api/v1/base64decode", "decode", "success").Inc()

	// Return JSON response
	c.JSON(http.StatusOK, gin.H{
		"decoded": decodedStr,
	})
}
