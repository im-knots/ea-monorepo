package routes

import (
	"net/http"

	"ea-backend/handlers"
)

// RegisterRoutes sets up the routes and their corresponding handlers.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", handlers.HandleRoot)
}
