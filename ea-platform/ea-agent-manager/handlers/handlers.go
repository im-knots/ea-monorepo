package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"ea-agent-manager/logger"
	"ea-agent-manager/metrics"
	"ea-agent-manager/mongo"

	"github.com/gin-gonic/gin"
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
}

// NodeDefinition represents the "template" for a node.
type NodeDefinition struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name,omitempty"`
	API        *NodeAPI               `json:"api,omitempty"`
	Parameters []NodeParameter        `json:"parameters,omitempty"`
	Outputs    []NodeParameter        `json:"outputs,omitempty"`
	Metadata   NodeDefinitionMetadata `json:"metadata,omitempty"`
}

// NodeInstance represents a reference to a node definition.
type NodeInstance struct {
	ID            string                 `json:"id"`
	DefinitionRef string                 `json:"definition_ref"`
	Parameters    map[string]interface{} `json:"parameters,omitempty"`
}

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
	Name        string         `json:"name"`
	User        string         `json:"user"`
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

	if err := c.ShouldBindJSON(&input); err != nil {
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "valid_request_body", "success").Inc()
	result, err := dbClient.InsertRecord("nodeDefs", "nodes", input)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_insertion_error", "error").Inc()
		logger.Slog.Error("Failed to insert node definition", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert node definition"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "create_success", "success").Inc()
	logger.Slog.Info("Node definition inserted successfully", "ID", result.InsertedID)
	c.JSON(http.StatusCreated, gin.H{"message": "Node definition created", "node_id": result.InsertedID})
}

// HandleGetAllNodeDefs retrieves all node definitions (only `id` and `name`).
func HandleGetAllNodeDefs(c *gin.Context) {
	path := c.FullPath()
	metrics.StepCounter.WithLabelValues(path, "api_hit", "success").Inc()

	projection := bson.M{"id": 1, "name": 1, "_id": 1}
	nodeDefs, err := dbClient.FindRecordsWithProjection("nodeDefs", "nodes", bson.M{}, projection)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve node definitions", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve node definitions"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "retrieval_success", "success").Inc()
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

	metrics.StepCounter.WithLabelValues(path, "api_request_start", "success").Inc()
	logger.Slog.Info("Fetching node definition", "node_id", nodeID)

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

	metrics.StepCounter.WithLabelValues(path, "retrieval_success", "success").Inc()
	logger.Slog.Info("Node definition retrieved successfully", "node_id", nodeID)
	c.JSON(http.StatusOK, nodeDef)
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

	if err := c.ShouldBindJSON(&input); err != nil {
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	input.Metadata = Metadata{CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	result, err := dbClient.InsertRecord("userAgents", "agents", input)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_insertion_error", "error").Inc()
		logger.Slog.Error("Failed to insert agent", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert agent"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "create_success", "success").Inc()
	logger.Slog.Info("Agent inserted successfully", "ID", result.InsertedID)
	c.JSON(http.StatusCreated, gin.H{"message": "Agent created", "agent_id": result.InsertedID})
}

// HandleGetAllAgents retrieves all agents with `user`, `_id`, and `name` fields.
func HandleGetAllAgents(c *gin.Context) {
	path := c.FullPath()
	metrics.StepCounter.WithLabelValues(path, "api_hit", "success").Inc()

	projection := bson.M{"user": 1, "_id": 1, "name": 1}
	agents, err := dbClient.FindRecordsWithProjection("userAgents", "agents", bson.M{}, projection)
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

	if agentID == "" {
		metrics.StepCounter.WithLabelValues(path, "missing_id", "error").Inc()
		logger.Slog.Error("Missing agent ID in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing agent ID"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "api_request_start", "success").Inc()
	logger.Slog.Info("Fetching agent details", "agent_id", agentID)

	agent, err := dbClient.FindRecordByID("userAgents", "agents", agentID)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve agent", "agent_id", agentID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve agent"})
		return
	}

	if agent == nil {
		metrics.StepCounter.WithLabelValues(path, "agent_not_found", "error").Inc()
		logger.Slog.Warn("Agent not found", "agent_id", agentID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	metrics.StepCounter.WithLabelValues(path, "retrieval_success", "success").Inc()
	logger.Slog.Info("Agent retrieved successfully", "agent_id", agentID)
	c.JSON(http.StatusOK, agent)
}
