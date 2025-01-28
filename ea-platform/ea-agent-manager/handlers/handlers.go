package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"ea-agent-manager/logger"
	"ea-agent-manager/metrics"
	"ea-agent-manager/mongo"

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
// 1. NodeDefinition (the "template") structs and handler
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
	Enum        []interface{} `json:"enum,omitempty"` // Could be []string if all enum values are strings
}

// NodeDefinitionMetadata holds metadata about the node definition.
type NodeDefinitionMetadata struct {
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Additional  map[string]interface{} `json:"additional,omitempty"`
}

// NodeDefinition is the "template" for a node, stored in the nodeDefs collection.
type NodeDefinition struct {
	ID         string                 `json:"id"`   // e.g. "worker.inference.llm.ollama"
	Type       string                 `json:"type"` // e.g. "worker.inference.llm"
	Name       string                 `json:"name,omitempty"`
	API        *NodeAPI               `json:"api,omitempty"`
	Parameters []NodeParameter        `json:"parameters,omitempty"`
	Metadata   NodeDefinitionMetadata `json:"metadata,omitempty"`
}

// HandleCreateNodeDef handles the creation of a node definition (template).
func HandleCreateNodeDef(w http.ResponseWriter, r *http.Request) {
	var input NodeDefinition
	path := "/api/v1/nodes"

	if r.Method != http.MethodPost {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		logger.Slog.Error("Invalid request method", "method", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "api_hit", "success").Inc()
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "decoding_request", "success").Inc()
	}

	// Insert NodeDefinition into the "nodeDefs" database and "nodes" collection
	result, err := dbClient.InsertRecord("nodeDefs", "nodes", input)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_insertion_error", "error").Inc()
		logger.Slog.Error("Failed to insert node definition into database", "error", err)
		http.Error(w, "Failed to insert node definition into database", http.StatusInternalServerError)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "db_insertion", "success").Inc()
	}

	metrics.StepCounter.WithLabelValues(path, "create_success", "success").Inc()
	logger.Slog.Info("Node definition inserted successfully", "ID", result.InsertedID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Node definition created successfully",
		"node_id": result.InsertedID,
	})
}

// HandleGetAllNodeDefs retrieves all node definitions from the database, but only their IDs and names.
func HandleGetAllNodeDefs(w http.ResponseWriter, r *http.Request) {
	path := "/api/v1/nodes"

	if r.Method != http.MethodGet {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Query all node definitions but only retrieve `id` and `name`
	projection := bson.M{
		"id":   1, // Include the `id` field
		"name": 1, // Include the `name` field
		"_id":  1, // Include the MongoDB internal `_id` field
	}
	nodeDefs, err := dbClient.FindRecordsWithProjection("nodeDefs", "nodes", bson.M{}, projection)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve node definitions from database", "error", err)
		http.Error(w, "Failed to retrieve node definitions", http.StatusInternalServerError)
		return
	}

	metrics.StepCounter.WithLabelValues(path, "retrieval_success", "success").Inc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodeDefs)
}

// HandleGetNodeDef retrieves a specific node definition by ID.
func HandleGetNodeDef(w http.ResponseWriter, r *http.Request) {
	path := "/api/v1/nodes/"

	if r.Method != http.MethodGet {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, path)
	if id == "" {
		metrics.StepCounter.WithLabelValues(path, "missing_id", "error").Inc()
		logger.Slog.Error("Missing node definition ID")
		http.Error(w, "Missing node definition ID", http.StatusBadRequest)
		return
	}

	nodeDef, err := dbClient.FindRecordByID("nodeDefs", "nodes", id)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve node definition from database", "error", err)
		http.Error(w, "Failed to retrieve node definition", http.StatusInternalServerError)
		return
	}

	metrics.StepCounter.WithLabelValues(path, "retrieval_success", "success").Inc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodeDef)
}

//-----------------------------------------------------------------------------
// 2. Agent (the "instance") structs and handler
//-----------------------------------------------------------------------------

// NodeInstance is a simplified node reference in the agent workflow.
type NodeInstance struct {
	ID            string                 `json:"id"`
	DefinitionRef string                 `json:"definition_ref"`
	Parameters    map[string]interface{} `json:"parameters,omitempty"`
}

// MultiString allows "from" or "to" to be either a single string or array of strings.
type MultiString []string

// UnmarshalJSON handles single strings and arrays for MultiString.
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

// Edge represents a link in the agent workflow graph.
type Edge struct {
	From MultiString `json:"from"`
	To   MultiString `json:"to"`
}

// Metadata holds creation/update timestamps.
type Metadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Agent is the main object that references node definitions via NodeInstance.
type Agent struct {
	Name        string         `json:"name"`
	User        string         `json:"user"`
	Description string         `json:"description"`
	Nodes       []NodeInstance `json:"nodes"`
	Edges       []Edge         `json:"edges"`
	Metadata    Metadata       `json:"metadata"`
}

// HandleCreateAgent handles the creation of an agent (which references node definitions).
func HandleCreateAgent(w http.ResponseWriter, r *http.Request) {
	var input Agent
	path := "/api/v1/agents"

	if r.Method != http.MethodPost {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		logger.Slog.Error("Invalid request method", "method", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "api_hit", "success").Inc()
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		metrics.StepCounter.WithLabelValues(path, "decode_error", "error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "decoding_request", "success").Inc()
	}

	// Populate metadata for Agent
	input.Metadata = Metadata{
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if input.Metadata.CreatedAt.IsZero() || input.Metadata.UpdatedAt.IsZero() {
		metrics.StepCounter.WithLabelValues(path, "populating_metadata", "error").Inc()
		logger.Slog.Error("Failed to populate metadata")
		http.Error(w, "Internal server error while populating metadata", http.StatusInternalServerError)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "populating_metadata", "success").Inc()
	}

	result, err := dbClient.InsertRecord("userAgents", "agents", input)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_insertion_error", "error").Inc()
		logger.Slog.Error("Failed to insert agent into database", "error", err)
		http.Error(w, "Failed to insert agent into database", http.StatusInternalServerError)
		return
	} else {
		metrics.StepCounter.WithLabelValues(path, "db_insertion", "success").Inc()
	}

	metrics.StepCounter.WithLabelValues(path, "create_success", "success").Inc()
	logger.Slog.Info("Agent inserted successfully", "ID", result.InsertedID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Agent created successfully",
		"agent_id": result.InsertedID,
	})
}

// HandleGetAllAgents retrieves all agents from the database with their id, _id, and name fields.
func HandleGetAllAgents(w http.ResponseWriter, r *http.Request) {
	path := "/api/v1/agents"

	if r.Method != http.MethodGet {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Define the projection to include id, _id, and name
	projection := bson.M{
		"user": 1, // Include the `user` field
		"_id":  1, // Include the MongoDB `_id` field
		"name": 1, // Include the `name` field
	}

	// Retrieve agents with the defined projection
	agents, err := dbClient.FindRecordsWithProjection("userAgents", "agents", bson.M{}, projection)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve agents from database", "error", err)
		http.Error(w, "Failed to retrieve agents", http.StatusInternalServerError)
		return
	}

	metrics.StepCounter.WithLabelValues(path, "retrieval_success", "success").Inc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// HandleGetAgent retrieves a specific agent by ID.
func HandleGetAgent(w http.ResponseWriter, r *http.Request) {
	path := "/api/v1/agents/"

	if r.Method != http.MethodGet {
		metrics.StepCounter.WithLabelValues(path, "invalid_method", "error").Inc()
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, path)
	if id == "" {
		metrics.StepCounter.WithLabelValues(path, "missing_id", "error").Inc()
		logger.Slog.Error("Missing agent ID")
		http.Error(w, "Missing agent ID", http.StatusBadRequest)
		return
	}

	agent, err := dbClient.FindRecordByID("userAgents", "agents", id)
	if err != nil {
		metrics.StepCounter.WithLabelValues(path, "db_retrieval_error", "error").Inc()
		logger.Slog.Error("Failed to retrieve agent from database", "error", err)
		http.Error(w, "Failed to retrieve agent", http.StatusInternalServerError)
		return
	}

	metrics.StepCounter.WithLabelValues(path, "retrieval_success", "success").Inc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}
