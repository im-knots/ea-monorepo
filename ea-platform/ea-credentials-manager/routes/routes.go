package routes

import (
	"ea-credentials-manager/handlers"
	"ea-credentials-manager/logger"
	"ea-credentials-manager/metrics"
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

	// Credentials routes
	credentials := router.Group("/api/v1/credentials")
	credentials.Use(authMiddleware())
	{
		credentials.PATCH("", handlers.HandleAddCredential) // Add a credential to a user secret
	}

	return router
}

// CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "PATCH, OPTIONS")
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
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		parser := new(jwt.Parser)
		token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		// Set authenticated user ID in context
		c.Set("AuthenticatedUserID", sub)

		c.Next()
	}
}
