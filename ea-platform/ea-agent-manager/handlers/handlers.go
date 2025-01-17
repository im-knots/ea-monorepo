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
	ID       string `json:"id"`
	Type     string `json:"type"`
	Data     string `json:"data,omitempty"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Metadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HandleCreateAgent handles the creation of an agent.
func HandleCreateAgent(w http.ResponseWriter, r *http.Request) {
	metrics.StepCounter.WithLabelValues("create_agent_api_hit").Inc()

	if r.Method != http.MethodPost {
		metrics.StepCounter.WithLabelValues("create_agent_invalid_method").Inc()
		logger.Slog.Error("Invalid request method", "method", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	metrics.StepCounter.WithLabelValues("create_agent_decoding_request").Inc()
	var input Agent
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		metrics.StepCounter.WithLabelValues("create_agent_decode_error").Inc()
		logger.Slog.Error("Failed to parse request body", "error", err)
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	metrics.StepCounter.WithLabelValues("create_agent_populating_metadata").Inc()
	input.Metadata = Metadata{
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	metrics.StepCounter.WithLabelValues("create_agent_db_insertion").Inc()
	result, err := dbClient.InsertRecord("userAgents", "agents", input)
	if err != nil {
		metrics.StepCounter.WithLabelValues("create_agent_db_insertion_error").Inc()
		logger.Slog.Error("Failed to insert agent into database", "error", err)
		http.Error(w, "Failed to insert agent into database", http.StatusInternalServerError)
		return
	}

	metrics.StepCounter.WithLabelValues("create_agent_success").Inc()
	logger.Slog.Info("Agent inserted successfully", "ID", result.InsertedID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Agent created successfully",
		"agent_id": result.InsertedID,
	})
}

// HandleGetNodes will pull a given Agents component nodes
// TODO: Define agent builder node type schema or declarative language?
func HandleGetNode(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("heres your list of agent builder node types! :)"))
}

// HandleGetPresets will populate a list of presets or a specific preset Agent and its associated component nodes
func HandleGetPresets(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Heres your list of Agent Presets and their component nodes!"))
}

// HandleCreateNode will create a Node component in a given Agent
func HandleCreateAgentNode(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("an Agent Node!"))
}

// HandleCreateJob will create a Job in a given Agent
func HandleCreateAgentJob(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("an Agent Job!"))
}
