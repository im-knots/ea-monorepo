package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ea-agent-manager/logger"
	"ea-agent-manager/metrics"
	"ea-agent-manager/mongo"

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

//-----------------------------------------------------------------------------
// Structs for Node Definitions & Agents
//-----------------------------------------------------------------------------

// NodeAPI describes how to call an API (base URL, endpoint, etc.).
type NodeAPI struct {
	BaseURL  string            `json:"base_url"`
	Endpoint string            `json:"endpoint"`
	Method   string            `json:"method"`
	Headers  map[string]string `json:"headers,omitempty"`
}

// NodeParameter describes each parameter for a NodeDefinition.
type NodeParameter struct {
	Key         string        `json:"key"`
	Type        string        `json:"type"`
	Description string        `json:"description,omitempty"`
	Default     interface{}   `json:"default,omitempty"`
	Enum        []interface{} `json:"enum,omitempty"`
}

// NodeDefinitionMetadata holds metadata about the node definition.
type NodeDefinitionMetadata struct {
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Additional  map[string]interface{} `json:"additional,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// NodeDefinition represents the "template" for a node.
type NodeDefinition struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Alias      string                 `json:"alias"`
	Name       string                 `json:"name,omitempty"`
	Creator    string                 `json:"creator,omitempty"`
	API        *NodeAPI               `json:"api,omitempty"`
	Parameters []NodeParameter        `json:"parameters,omitempty"`
	Outputs    []NodeParameter        `json:"outputs,omitempty"`
	Metadata   NodeDefinitionMetadata `json:"metadata"`
}

// NodeInstance represents a reference to a node definition.
type NodeInstance struct {
	Alias      string                 `json:"alias,omitempty"`
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// Edge represents a connection between nodes in an agent workflow.
type Edge struct {
	From MultiString `json:"from"`
	To   MultiString `json:"to"`
}

// Metadata holds timestamps for Agents.
type Metadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Agent represents an AI workflow with interconnected nodes.
type Agent struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Creator     string         `json:"creator"`
	Description string         `json:"description"`
	Nodes       []NodeInstance `json:"nodes"`
	Edges       []Edge         `json:"edges"`
	Metadata    Metadata       `json:"metadata"`
}

//-----------------------------------------------------------------------------
// NodeDefinition Handlers
//-----------------------------------------------------------------------------

// HandleCreateNodeDef handles creating a node definition.
func HandleCreateNodeDef(c *gin.Context) {
	var input NodeDefinition
	path := c.FullPath()

	metrics.StepCounter.WithLabelValues(path, "api_request_start", "success").Inc()
	logger.Slog.Info("Node definition creation request received")

	// Extract the authenticated user from Kong's header
	authenticatedUserID := c.GetHeader("X-Consumer-Username")
	if authenticatedUserID == "" {
		logger.Slog.Error("Missing X-Consumer-Username header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse the request body
	if err := c.ShouldBindJSON(&input); err != nil {
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	// Ensure the creator field matches the authenticated user
	if input.Creator != authenticatedUserID {
		logger.Slog.Error("User ID mismatch", "authenticated", authenticatedUserID, "request_creator", input.Creator)
		metrics.StepCounter.WithLabelValues(path, "user_spoofing_attempt", "failure").Inc()
		c.JSON(http.StatusForbidden, gin.H{"error": "User ID does not match authenticated user"})
		return
	}

	// Assign a unique ID and timestamps
	input.ID = uuid.New().String()
	input.Metadata.CreatedAt = time.Now().UTC()
	input.Metadata.UpdatedAt = time.Now().UTC()

	metrics.StepCounter.WithLabelValues(path, "valid_request_body", "success").Inc()
	result, err := dbClient.InsertRecord("nodeDefs", "nodes", input)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_insertion_error", "error").Inc()
		logger.Slog.Error("Failed to insert node definition", "mongo_id", result.InsertedID, "input_id", input.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert node definition"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "create_success", "success").Inc()
	logger.Slog.Info("Node definition inserted successfully", "node_id", input.ID, "creator", input.Creator)
	c.JSON(http.StatusCreated, gin.H{"message": "Node definition created", "node_id": input.ID, "creator": input.Creator})
}

// HandleGetAllNodeDefs retrieves all node definitions for the authenticated user.
func HandleGetAllNodeDefs(c *gin.Context) {
	path := c.FullPath()
	metrics.StepCounter.WithLabelValues(path, "api_hit", "success").Inc()

	// Extract the authenticated user from Kong's header
	authenticatedUserID := c.GetHeader("X-Consumer-Username")
	if authenticatedUserID == "" {
		logger.Slog.Error("Missing X-Consumer-Username header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Extract optional creator ID from query params
	requestedCreatorID := c.Query("creator_id")

	// Apply filtering logic:
	// - If creator_id is provided, enforce that it matches the authenticated user.
	// - If creator_id is absent, default to fetching the authenticated user's node definitions.
	filter := bson.M{"creator": authenticatedUserID}
	if requestedCreatorID != "" && requestedCreatorID != authenticatedUserID {
		logger.Slog.Error("User spoofing attempt detected", "authenticated", authenticatedUserID, "requested", requestedCreatorID)
		metrics.StepCounter.WithLabelValues(path, "user_spoofing_attempt", "failure").Inc()
		c.JSON(http.StatusForbidden, gin.H{"error": "Creator ID does not match authenticated user"})
		return
	}

	projection := bson.M{"id": 1, "type": 1, "creator": 1, "_id": 0}
	nodeDefs, err := dbClient.FindRecordsWithProjection("nodeDefs", "nodes", filter, projection)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve node definitions", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve node definitions"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "retrieval_success", "success").Inc()
	logger.Slog.Info("Node definitions retrieved successfully", "user", authenticatedUserID, "count", len(nodeDefs))
	c.JSON(http.StatusOK, nodeDefs)
}

// HandleGetNodeDef retrieves a specific node definition by ID.
func HandleGetNodeDef(c *gin.Context) {
	path := c.FullPath()
	nodeID := c.Param("node_id")

	if nodeID == "" {
		metrics.StepCounter.WithLabelValues(path, "missing_id", "error").Inc()
		logger.Slog.Error("Missing node definition ID in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing node definition ID"})
		return
	}

	// Extract the authenticated user from Kong's header
	authenticatedUserID := c.GetHeader("X-Consumer-Username")
	if authenticatedUserID == "" {
		logger.Slog.Error("Missing X-Consumer-Username header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "api_request_start", "success").Inc()
	logger.Slog.Info("Fetching node definition", "node_id", nodeID, "user", authenticatedUserID)

	// Retrieve the node definition
	nodeDef, err := dbClient.FindRecordByID("nodeDefs", "nodes", nodeID)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve node definition", "node_id", nodeID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve node definition"})
		return
	}

	if nodeDef == nil {
		metrics.StepCounter.WithLabelValues(path, "node_not_found", "error").Inc()
		logger.Slog.Warn("Node definition not found", "node_id", nodeID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Node definition not found"})
		return
	}

	// Extract creator field from retrieved node definition
	creatorID, ok := nodeDef["creator"].(string)
	if !ok || creatorID == "" {
		logger.Slog.Error("Node definition missing creator field", "node_id", nodeID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Node definition missing creator field"})
		return
	}

	// Ensure the creator ID matches the authenticated user
	if creatorID != authenticatedUserID {
		logger.Slog.Error("User spoofing attempt detected", "authenticated", authenticatedUserID, "creator", creatorID)
		metrics.StepCounter.WithLabelValues(path, "user_spoofing_attempt", "failure").Inc()
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: You do not own this node definition"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "retrieval_success", "success").Inc()
	logger.Slog.Info("Node definition retrieved successfully", "node_id", nodeID, "user", authenticatedUserID)
	c.JSON(http.StatusOK, nodeDef)
}

// HandleUpdateNodeDef updates an existing node definition by ID.
func HandleUpdateNodeDef(c *gin.Context) {
	path := c.FullPath()
	nodeID := c.Param("node_id")

	if nodeID == "" {
		metrics.StepCounter.WithLabelValues(path, "missing_id", "error").Inc()
		logger.Slog.Error("Missing node definition ID in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing node definition ID"})
		return
	}

	// Extract the authenticated user from Kong's header
	authenticatedUserID := c.GetHeader("X-Consumer-Username")
	if authenticatedUserID == "" {
		logger.Slog.Error("Missing X-Consumer-Username header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request body
	var input NodeDefinition
	if err := c.ShouldBindJSON(&input); err != nil {
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	input.Metadata.UpdatedAt = time.Now().UTC()
	input.Creator = authenticatedUserID // Force creator field to match authenticated user

	// Ensure the node being updated belongs to the authenticated user
	filter := bson.M{"id": nodeID, "creator": authenticatedUserID}
	update := bson.M{"$set": input}

	// Attempt to update the record
	result, err := dbClient.UpdateRecord("nodeDefs", "nodes", filter, update)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_update_error", "error").Inc()
		logger.Slog.Error("Failed to update node definition", "node_id", nodeID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update node definition"})
		return
	}

	if result.MatchedCount == 0 {
		metrics.StepCounter.WithLabelValues(path, "node_not_found_or_unauthorized", "error").Inc()
		logger.Slog.Warn("Node definition not found or unauthorized update attempt", "node_id", nodeID, "user", authenticatedUserID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Node definition not found or unauthorized update attempt"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "update_success", "success").Inc()
	logger.Slog.Info("Node definition updated successfully", "node_id", nodeID, "user", authenticatedUserID)
	c.JSON(http.StatusOK, gin.H{"message": "Node definition updated successfully", "node_id": nodeID})
}

// HandleDeleteNodeDef deletes a node definition by ID.
func HandleDeleteNodeDef(c *gin.Context) {
	path := c.FullPath()
	nodeID := c.Param("node_id")

	if nodeID == "" {
		metrics.StepCounter.WithLabelValues(path, "missing_id", "error").Inc()
		logger.Slog.Error("Missing node definition ID in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing node definition ID"})
		return
	}

	// Extract the authenticated user from Kong's header
	authenticatedUserID := c.GetHeader("X-Consumer-Username")
	if authenticatedUserID == "" {
		logger.Slog.Error("Missing X-Consumer-Username header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "api_request_start", "success").Inc()
	logger.Slog.Info("Attempting to delete node definition", "node_id", nodeID, "user", authenticatedUserID)

	// Ensure the node being deleted belongs to the authenticated user
	filter := bson.M{"id": nodeID, "creator": authenticatedUserID}

	// Attempt to delete the record
	deleteResult, err := dbClient.DeleteRecord("nodeDefs", "nodes", filter)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_deletion_error", "error").Inc()
		logger.Slog.Error("Failed to delete node definition", "node_id", nodeID, "user", authenticatedUserID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete node definition"})
		return
	}

	if deleteResult.DeletedCount == 0 {
		metrics.StepCounter.WithLabelValues(path, "node_not_found_or_unauthorized", "error").Inc()
		logger.Slog.Warn("Node definition not found or unauthorized delete attempt", "node_id", nodeID, "user", authenticatedUserID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Node definition not found or unauthorized delete attempt"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "delete_success", "success").Inc()
	logger.Slog.Info("Node definition deleted successfully", "node_id", nodeID, "user", authenticatedUserID)
	c.JSON(http.StatusOK, gin.H{"message": "Node definition deleted successfully", "node_id": nodeID})
}

//-----------------------------------------------------------------------------
// Agent Handlers
//-----------------------------------------------------------------------------

// HandleCreateAgent handles creating an agent.
func HandleCreateAgent(c *gin.Context) {
	var input Agent
	path := c.FullPath()

	metrics.StepCounter.WithLabelValues(path, "api_request_start", "success").Inc()
	logger.Slog.Info("Agent creation request received")

	// Extract authenticated user ID from Kong's header
	authenticatedUserID := c.GetHeader("X-Consumer-Username")
	if authenticatedUserID == "" {
		logger.Slog.Error("Missing X-Consumer-Username header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	// Ensure the creator field matches the authenticated user
	if input.Creator != authenticatedUserID {
		logger.Slog.Error("User spoofing attempt detected", "authenticated", authenticatedUserID, "request_creator", input.Creator)
		metrics.StepCounter.WithLabelValues(path, "user_spoofing_attempt", "failure").Inc()
		c.JSON(http.StatusForbidden, gin.H{"error": "User ID does not match authenticated user"})
		return
	}

	// Ensure each node has an alias
	for i, node := range input.Nodes {
		if node.Alias == "" {
			logger.Slog.Warn("Missing alias in node, assigning default alias", "node_type", node.Type)
			input.Nodes[i].Alias = fmt.Sprintf("node-%d", i) // Assign a default alias if missing
		}
	}

	// Set metadata and generate ID
	input.ID = uuid.New().String()
	input.Metadata = Metadata{CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}

	// Insert the agent into the database
	result, err := dbClient.InsertRecord("userAgents", "agents", input)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_insertion_error", "error").Inc()
		logger.Slog.Error("Failed to insert agent", "mongo_id", result.InsertedID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert agent"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "create_success", "success").Inc()
	logger.Slog.Info("Agent inserted successfully", "ID", input.ID, "creator", input.Creator)
	c.JSON(http.StatusCreated, gin.H{"message": "Agent created", "agent_id": input.ID, "creator": input.Creator})
}

// HandleGetAllAgents retrieves all agents.
func HandleGetAllAgents(c *gin.Context) {
	path := c.FullPath()
	metrics.StepCounter.WithLabelValues(path, "api_hit", "success").Inc()

	// Extract authenticated user ID from Kong's header
	authenticatedUserID := c.GetHeader("X-Consumer-Username")
	if authenticatedUserID == "" {
		logger.Slog.Error("Missing X-Consumer-Username header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Extract optional creator_id from query param
	creatorID := c.Query("creator_id")

	// If creator_id is provided and does not match the authenticated user, log & block it
	if creatorID != "" && creatorID != authenticatedUserID {
		logger.Slog.Error("User spoofing attempt detected", "authenticated", authenticatedUserID, "query_creator", creatorID)
		metrics.StepCounter.WithLabelValues(path, "user_spoofing_attempt", "failure").Inc()
		c.JSON(http.StatusForbidden, gin.H{"error": "User ID does not match authenticated user"})
		return
	}

	// Always use the authenticated user ID to filter results
	filter := bson.M{"creator": authenticatedUserID}
	projection := bson.M{"creator": 1, "id": 1, "name": 1, "_id": 0}

	agents, err := dbClient.FindRecordsWithProjection("userAgents", "agents", filter, projection)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve agents", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve agents"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "retrieval_success", "success").Inc()
	c.JSON(http.StatusOK, agents)
}

// HandleGetAgent retrieves a specific agent by ID.
func HandleGetAgent(c *gin.Context) {
	path := c.FullPath()
	agentID := c.Param("agent_id")

	// Validate agent ID
	if agentID == "" {
		metrics.StepCounter.WithLabelValues(path, "missing_id", "error").Inc()
		logger.Slog.Error("Missing agent ID in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing agent ID"})
		return
	}

	// Extract authenticated user ID from Kong's header
	authenticatedUserID := c.GetHeader("X-Consumer-Username")
	if authenticatedUserID == "" {
		logger.Slog.Error("Missing X-Consumer-Username header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "api_request_start", "success").Inc()
	logger.Slog.Info("Fetching agent details", "agent_id", agentID)

	// Retrieve agent and filter by creator ID
	filter := bson.M{"id": agentID, "creator": authenticatedUserID}
	projection := bson.M{"id": 1, "name": 1, "creator": 1, "description": 1, "nodes": 1, "edges": 1, "_id": 0}

	agent, err := dbClient.FindRecordsWithProjection("userAgents", "agents", filter, projection)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve agent", "agent_id", agentID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve agent"})
		return
	}

	// If no agent is found, return 404
	if agent == nil {
		metrics.StepCounter.WithLabelValues(path, "agent_not_found", "error").Inc()
		logger.Slog.Warn("Agent not found or user does not have access", "agent_id", agentID, "user_id", authenticatedUserID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "retrieval_success", "success").Inc()
	logger.Slog.Info("Agent retrieved successfully", "agent_id", agentID, "creator", authenticatedUserID)
	c.JSON(http.StatusOK, agent)
}

// HandleUpdateAgent updates an existing agent by ID.
func HandleUpdateAgent(c *gin.Context) {
	path := c.FullPath()
	agentID := c.Param("agent_id")

	// Validate agent ID
	if agentID == "" {
		metrics.StepCounter.WithLabelValues(path, "missing_id", "error").Inc()
		logger.Slog.Error("Missing agent ID in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing agent ID"})
		return
	}

	// Extract authenticated user ID from Kong's header
	authenticatedUserID := c.GetHeader("X-Consumer-Username")
	if authenticatedUserID == "" {
		logger.Slog.Error("Missing X-Consumer-Username header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Verify the user is the creator of the agent
	filter := bson.M{"id": agentID, "creator": authenticatedUserID}
	existingAgent, err := dbClient.FindRecordByID("userAgents", "agents", agentID)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve agent for update", "agent_id", agentID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve agent"})
		return
	}

	// If no matching agent is found, return 404
	if existingAgent == nil || existingAgent["creator"] != authenticatedUserID {
		metrics.StepCounter.WithLabelValues(path, "agent_not_found_or_unauthorized", "error").Inc()
		logger.Slog.Warn("Agent not found or user does not have permission", "agent_id", agentID, "user_id", authenticatedUserID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Parse request body for update
	var input Agent
	if err := c.ShouldBindJSON(&input); err != nil {
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	// Ensure the updated agent maintains the correct creator
	input.Metadata.UpdatedAt = time.Now().UTC()
	input.Creator = authenticatedUserID

	update := bson.M{"$set": input}

	result, err := dbClient.UpdateRecord("userAgents", "agents", filter, update)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_update_error", "error").Inc()
		logger.Slog.Error("Failed to update agent", "agent_id", agentID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agent"})
		return
	}

	if result.MatchedCount == 0 {
		metrics.StepCounter.WithLabelValues(path, "agent_not_found", "error").Inc()
		logger.Slog.Warn("Agent not found or user is unauthorized", "agent_id", agentID, "user_id", authenticatedUserID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "update_success", "success").Inc()
	logger.Slog.Info("Agent updated successfully", "agent_id", agentID, "creator", authenticatedUserID)
	c.JSON(http.StatusOK, gin.H{"message": "Agent updated successfully", "agent_id": agentID})
}

// HandleDeleteAgent deletes an agent by ID.
func HandleDeleteAgent(c *gin.Context) {
	path := c.FullPath()
	agentID := c.Param("agent_id")

	// Validate agent ID
	if agentID == "" {
		metrics.StepCounter.WithLabelValues(path, "missing_id", "error").Inc()
		logger.Slog.Error("Missing agent ID in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing agent ID"})
		return
	}

	// Extract authenticated user ID from Kong's header
	authenticatedUserID := c.GetHeader("X-Consumer-Username")
	if authenticatedUserID == "" {
		logger.Slog.Error("Missing X-Consumer-Username header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Ensure the user is the creator of the agent before deleting
	filter := bson.M{"id": agentID, "creator": authenticatedUserID}
	existingAgent, err := dbClient.FindRecordByID("userAgents", "agents", agentID)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve agent for deletion", "agent_id", agentID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve agent"})
		return
	}

	// If no matching agent is found or user is not the creator, return 404
	if existingAgent == nil || existingAgent["creator"] != authenticatedUserID {
		metrics.StepCounter.WithLabelValues(path, "agent_not_found_or_unauthorized", "error").Inc()
		logger.Slog.Warn("Agent not found or user does not have permission", "agent_id", agentID, "user_id", authenticatedUserID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "api_request_start", "success").Inc()
	logger.Slog.Info("Deleting agent", "agent_id", agentID, "creator", authenticatedUserID)

	// Perform deletion
	deleteResult, err := dbClient.DeleteRecord("userAgents", "agents", filter)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_deletion_error", "error").Inc()
		logger.Slog.Error("Failed to delete agent", "agent_id", agentID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete agent"})
		return
	}

	if deleteResult.DeletedCount == 0 {
		metrics.StepCounter.WithLabelValues(path, "agent_not_found", "error").Inc()
		logger.Slog.Warn("Agent not found or user is unauthorized", "agent_id", agentID, "user_id", authenticatedUserID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "delete_success", "success").Inc()
	logger.Slog.Info("Agent deleted successfully", "agent_id", agentID, "creator", authenticatedUserID)
	c.JSON(http.StatusOK, gin.H{"message": "Agent deleted successfully", "agent_id": agentID})
}

// HELPER FUNCTIONS
// MultiString allows a field to be either a single string or an array of strings.
type MultiString []string

// UnmarshalJSON for MultiString allows handling single strings as arrays.
func (m *MultiString) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*m = []string{single}
		return nil
	}

	var multiple []string
	if err := json.Unmarshal(data, &multiple); err == nil {
		*m = multiple
		return nil
	}

	return json.Unmarshal(data, m)
}
