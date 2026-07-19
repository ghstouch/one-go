package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ghstouch/one-go/internal/service"
	"github.com/ghstouch/one-go/pkg/response"
)

const (
	// AuthorizationHeader is the header key for authorization
	AuthorizationHeader = "Authorization"
	// BearerPrefix is the prefix for bearer tokens
	BearerPrefix = "Bearer "
	// UserIDKey is the context key for user ID
	UserIDKey = "userId"
	// UsernameKey is the context key for username
	UsernameKey = "username"
	// RoleKey is the context key for user role
	RoleKey = "role"
)

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)

		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			response.Unauthorized(c, "Invalid authorization format")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user info in context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)
		c.Set(RoleKey, claims.Role)

		c.Next()
	}
}

// OptionalAuthMiddleware creates optional JWT authentication middleware
// It doesn't abort if no token is present, but validates if one is provided
func OptionalAuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)

		if authHeader == "" {
			c.Next()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// Set user info in context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)
		c.Set(RoleKey, claims.Role)

		c.Next()
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(RoleKey)
		if !exists {
			response.Unauthorized(c, "Unauthorized")
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			response.Unauthorized(c, "Invalid role")
			c.Abort()
			return
		}

		// Check if user role is in allowed roles
		for _, role := range roles {
			if roleStr == role {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "Insufficient permissions")
		c.Abort()
	}
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get(UserIDKey); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// GetUsernameFromContext retrieves username from context
func GetUsernameFromContext(c *gin.Context) string {
	if username, exists := c.Get(UsernameKey); exists {
		if name, ok := username.(string); ok {
			return name
		}
	}
	return ""
}

// GetRoleFromContext retrieves user role from context
func GetRoleFromContext(c *gin.Context) string {
	if role, exists := c.Get(RoleKey); exists {
		if r, ok := role.(string); ok {
			return r
		}
	}
	return ""
}

// APIKeyContextKey is the context key for validated API key
const APIKeyContextKey = "apiKey"

// APIKeyMiddleware validates API keys for v1 endpoints
// Accepts: Authorization: Bearer sk-...
func APIKeyMiddleware(apiKeySvc service.ApiKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" || !strings.HasPrefix(authHeader, BearerPrefix) {
			response.Unauthorized(c, "API key required. Use Authorization: Bearer sk-...")
			c.Abort()
			return
		}

		key := strings.TrimPrefix(authHeader, BearerPrefix)
		apiKey, err := apiKeySvc.ValidateKey(key)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// Store API key info in context
		c.Set(APIKeyContextKey, apiKey)
		c.Next()
	}
}
