package routes

import (
	"net/http"

	"ea-agent-manager/handlers"
	"ea-agent-manager/metrics"
)

// RegisterRoutes sets up the routes and their corresponding handlers.
func RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/api/v1/metrics", metrics.MetricsHandler())
	mux.HandleFunc("/api/v1/agents", handlers.HandleCreateAgent)                      // POST for creating an agent
	mux.HandleFunc("/api/v1/agents/presets", handlers.HandleGetPresets)               // GET for retrieving presets
	mux.HandleFunc("/api/v1/agents/{agentId}/nodes", handlers.HandleCreateAgentNode)  // POST for creating a node
	mux.HandleFunc("/api/v1/agents/{agentId}/nodes/{nodeId}", handlers.HandleGetNode) // GET for retrieving a specific node

}
