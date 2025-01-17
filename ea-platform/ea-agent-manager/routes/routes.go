package routes

import (
	"net/http"

	"ea-agent-manager/handlers"
	"ea-agent-manager/metrics"
)

// RegisterRoutes sets up the routes and their corresponding handlers.
func RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/api/v1/metrics", metrics.MetricsHandler())
	mux.Handle("/api/v1/agents", metrics.MetricsMiddleware(http.HandlerFunc(handlers.HandleCreateAgent)))  // POST for creating an agent
	mux.Handle("/api/v1/nodes", metrics.MetricsMiddleware(http.HandlerFunc(handlers.HandleCreateNodeDef))) // POST for creating a node definition
	// mux.Handle("/api/v1/agents/presets", metrics.MetricsMiddleware(http.HandlerFunc(handlers.HandleGetPresets)))               // GET for retrieving presets
	// mux.Handle("/api/v1/agents/{agentId}/nodes", metrics.MetricsMiddleware(http.HandlerFunc(handlers.HandleCreateAgentNode)))  // POST for creating a node
	// mux.Handle("/api/v1/agents/{agentId}/nodes/{nodeId}", metrics.MetricsMiddleware(http.HandlerFunc(handlers.HandleGetNode))) // GET for retrieving a specific node

}
