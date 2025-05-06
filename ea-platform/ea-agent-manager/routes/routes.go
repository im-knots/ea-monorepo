package routes

import (
	"ea-agent-manager/handlers"
	"ea-agent-manager/logger"
	"ea-agent-manager/metrics"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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
	agents.Use(authMiddleware())
	{
		agents.POST("", handlers.HandleCreateAgent)             // Create new Agent
		agents.GET("", handlers.HandleGetAllAgents)             // List all Agents
		agents.GET("/:agent_id", handlers.HandleGetAgent)       // Get Agent by ID
		agents.PUT("/:agent_id", handlers.HandleUpdateAgent)    // Update Agent by ID
		agents.DELETE("/:agent_id", handlers.HandleDeleteAgent) // Delete Agent by ID
	}

	// Nodes routes
	nodes := router.Group("/api/v1/nodes")
	nodes.Use(authMiddleware())
	{
		nodes.POST("", handlers.HandleCreateNodeDef)            // Create new node
		nodes.GET("", handlers.HandleGetAllNodeDefs)            // List all nodes
		nodes.GET("/:node_id", handlers.HandleGetNodeDef)       // Get node by ID
		nodes.PUT("/:node_id", handlers.HandleUpdateNodeDef)    // Update Node Definition by ID
		nodes.DELETE("/:node_id", handlers.HandleDeleteNodeDef) // Delete node by ID
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

// authMiddleware extracts JWT claims and sets authenticated user ID in context
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log all incoming request headers for debugging
		logger.Slog.Info("Request Headers:")
		for key, values := range c.Request.Header {
			for _, value := range values {
				logger.Slog.Info("Header", "key", key, "value", value)
			}
		}

		internalHeader := c.GetHeader("X-EA-Internal")
		if internalHeader == "internal" {
			// Internal request bypasses JWT validation
			c.Set("AuthenticatedUserID", "internal")
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Slog.Error("missing auth header")
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		parser := new(jwt.Parser)
		token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			logger.Slog.Error("missing bearer in auth header")
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.Slog.Error("missing token claims")
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			logger.Slog.Error("missing token sub claim")
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		// Set authenticated user ID in context
		c.Set("AuthenticatedUserID", sub)

		c.Next()
	}
}
