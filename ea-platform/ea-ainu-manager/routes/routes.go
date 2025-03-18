package routes

import (
	"ea-ainu-manager/handlers"
	"ea-ainu-manager/logger"
	"ea-ainu-manager/metrics"
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

	// User routes
	users := router.Group("/api/v1/users")
	users.Use(jwtAuthMiddleware())
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

// jwtAuthMiddleware extracts JWT claims and sets authenticated user ID in context
func jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Slog.Error("Missing Authorization header")
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		parser := new(jwt.Parser)
		token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			logger.Slog.Error("Failed to parse JWT token", "error", err)
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.Slog.Error("JWT token claims invalid")
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			logger.Slog.Error("JWT missing 'sub' claim")
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		// Set authenticated user ID in context
		c.Set("AuthenticatedUserID", sub)

		c.Next()
	}
}
