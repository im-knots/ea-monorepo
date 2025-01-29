package routes

import (
	"net/http"

	"ea-ainu-manager/handlers"
	"ea-ainu-manager/metrics"
)

// RegisterRoutes sets up the routes and their corresponding handlers.
func RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/api/v1/metrics", metrics.MetricsHandler())

	// User routes
	mux.Handle("/api/v1/users", metrics.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.HandleCreateUser(w, r) // POST: Create an User
		} else if r.Method == http.MethodGet {
			handlers.HandleGetAllUsers(w, r) // GET: Retrieve all Users
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/v1/users/", metrics.MetricsMiddleware(http.HandlerFunc(handlers.HandleGetUser))) // GET: Retrieve a specific User by ID

}
