package routes

import (
	"net/http"

	"ea-job-engine/handlers"
	"ea-job-engine/metrics"
)

// RegisterRoutes sets up the routes and their corresponding handlers.
func RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/api/v1/metrics", metrics.MetricsHandler())
	mux.Handle("/api/v1/jobs", metrics.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.HandleCreateJob(w, r) // POST: Create a Job
		} else if r.Method == http.MethodGet {
			handlers.HandleGetAllJobs(w, r) // GET: Retrieve all Jobs
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/v1/jobs/", metrics.MetricsMiddleware(http.HandlerFunc(handlers.HandleGetJob))) // GET: Retrieve a specific Job by ID
}
