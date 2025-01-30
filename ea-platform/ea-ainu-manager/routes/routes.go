package routes

import (
	"net/http"
	"strings"

	"ea-ainu-manager/handlers"
	"ea-ainu-manager/metrics"
)

// RegisterRoutes sets up the routes and their corresponding handlers.
func RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/api/v1/metrics", metrics.MetricsHandler())

	mux.Handle("/api/v1/users/", metrics.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		segments := strings.Split(strings.TrimPrefix(path, "/api/v1/users/"), "/")

		if len(segments) > 2 && segments[1] == "devices" {
			// Compute Device routes under /api/v1/users/{user_id}/devices/{device_id}
			if r.Method == http.MethodDelete {
				handlers.HandleDeleteComputeDevice(w, r) // Delete a compute device
			} else if r.Method == http.MethodPost {
				handlers.HandleAddComputeDevice(w, r) // Add a compute device
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if len(segments) > 1 && segments[1] == "credits" {
			// Compute Credits update route /api/v1/users/{user_id}/credits
			if r.Method == http.MethodPut {
				handlers.HandleUpdateComputeCredits(w, r) // Update compute credits
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if len(segments) == 1 {
			if r.Method == http.MethodGet {
				handlers.HandleGetUser(w, r) // GET: Retrieve a specific user
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		http.Error(w, "Not Found", http.StatusNotFound)
	})))

	mux.Handle("/api/v1/users", metrics.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.HandleCreateUser(w, r) // Create a User
		} else if r.Method == http.MethodGet {
			handlers.HandleGetAllUsers(w, r) // Retrieve all Users
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
}
