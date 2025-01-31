package routes

import (
	"ea-job-engine/handlers"
	"ea-job-engine/metrics"

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
	users := router.Group("/api/v1/jobs")
	{
		users.GET("", handlers.HandleGetAllJobs)      // List all jobs
		users.POST("", handlers.HandleCreateJob)      // Create new job
		users.GET("/:user_id", handlers.HandleGetJob) // Get job by ID

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
