package routes

import (
	"ea-ainu-manager/handlers"
	"ea-ainu-manager/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterRoutes sets up all API routes
func RegisterRoutes() *gin.Engine {
	router := gin.Default()

	// Enable CORS middleware
	router.Use(corsMiddleware())
	// Enable metrics middleware
	router.Use(metrics.MetricsMiddleware())

	router.GET("/api/v1/metrics", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/plain")
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})

	// User routes
	users := router.Group("/api/v1/users")
	{
		users.GET("", handlers.HandleGetAllUsers)      // List all users
		users.GET("/:user_id", handlers.HandleGetUser) // Get user by ID
		users.PUT("/:user_id/credits", handlers.HandleUpdateComputeCredits)

		// User Compute Devices
		users.POST("/:user_id/devices", handlers.HandleAddComputeDevice)
		users.DELETE("/:user_id/devices/:device_id", handlers.HandleDeleteComputeDevice)

		// User Jobs
		users.POST("/:user_id/jobs", handlers.HandleAddJob)
		users.DELETE("/:user_id/jobs/:job_id", handlers.HandleDeleteJob)
	}

	return router
}

// CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}
