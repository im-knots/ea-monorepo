package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"ea-agent-manager/logger"
	"ea-agent-manager/metrics"
	"ea-agent-manager/mongo"
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

// Agent represents the structure of an agent record.
type Agent struct {
	Name        string   `json:"name"`
	User        string   `json:"user"`
	Description string   `json:"description"`
	Nodes       []Node   `json:"nodes"`
	Edges       []Edge   `json:"edges"`
	Metadata    Metadata `json:"metadata"`
}

type Node struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Data     []NodeData     `json:"data,omitempty"`
	Metadata []NodeMetadata `json:"metadata,omitempty"`
}

type NodeData struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type NodeMetadata struct {
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Additional  map[string]interface{} `json:"additional,omitempty"`
}

type Edge struct {
	From MultiString `json:"from"`
	To   MultiString `json:"to"`
}

type MultiString []string

type Metadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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

// HandleCreateAgent handles the creation of an agent.
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

// HandleCreateNodeDef handles the creation of a node definition
func HandleCreateNodeDef(w http.ResponseWriter, r *http.Request) {
	var input Node
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
		"message":  "Node definition created successfully",
		"agent_id": result.InsertedID,
	})
}
