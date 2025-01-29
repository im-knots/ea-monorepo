package routes

import (
	"net/http"

	"brand-backend/handlers"
)

// RegisterRoutes sets up the routes and their corresponding handlers.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", handlers.HandleRoot)
	mux.HandleFunc("/contact", handlers.HandleContact)
	mux.HandleFunc("/subscribe", handlers.HandleSubscribe)
	mux.HandleFunc("/waitlist", handlers.HandleWaitlist)
}
