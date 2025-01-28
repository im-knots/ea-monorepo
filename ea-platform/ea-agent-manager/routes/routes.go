package routes

import (
	"net/http"

	"ea-agent-manager/handlers"
	"ea-agent-manager/metrics"
)

// RegisterRoutes sets up the routes and their corresponding handlers.
func RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/api/v1/metrics", metrics.MetricsHandler())

	// Agent routes
	mux.Handle("/api/v1/agents", metrics.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.HandleCreateAgent(w, r) // POST: Create an agent
		} else if r.Method == http.MethodGet {
			handlers.HandleGetAllAgents(w, r) // GET: Retrieve all agents
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/v1/agents/", metrics.MetricsMiddleware(http.HandlerFunc(handlers.HandleGetAgent))) // GET: Retrieve a specific agent by ID

	// Node definition routes
	mux.Handle("/api/v1/nodes", metrics.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.HandleCreateNodeDef(w, r) // POST: Create a node definition
		} else if r.Method == http.MethodGet {
			handlers.HandleGetAllNodeDefs(w, r) // GET: Retrieve all node definitions
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/v1/nodes/", metrics.MetricsMiddleware(http.HandlerFunc(handlers.HandleGetNodeDef))) // GET: Retrieve a specific node definition by ID
}
