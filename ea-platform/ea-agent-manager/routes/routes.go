package routes

import (
	"ea-agent-manager/handlers"
	"ea-agent-manager/metrics"

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

	// Agents routes
	agents := router.Group("/api/v1/agents")
	{
		agents.GET("", handlers.HandleGetAllAgents)       // List all Agents
		agents.POST("", handlers.HandleCreateAgent)       // Create new Agent
		agents.GET("/:agent_id", handlers.HandleGetAgent) // Get Agent by ID
	}

	// Nodes routes
	nodes := router.Group("/api/v1/nodes")
	{
		nodes.GET("", handlers.HandleGetAllNodeDefs)      // List all nodes
		nodes.POST("", handlers.HandleCreateNodeDef)      // Create new node
		nodes.GET("/:node_id", handlers.HandleGetNodeDef) // Get node by ID
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
