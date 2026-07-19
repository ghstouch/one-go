package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ghstouch/one-go/pkg/logger"
)

// LoggingMiddleware creates a request logging middleware
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status
		status := c.Writer.Status()

		// Build log fields
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		if raw != "" {
			path = path + "?" + raw
		}

		// Get user info if available
		userID := GetUserIDFromContext(c)

		// Log based on status (mask sensitive headers)
		log := logger.With(
			"status", status,
			"method", method,
			"path", path,
			"ip", clientIP,
			"latency", latency.String(),
			"user_agent", userAgent,
		)

		if userID != "" {
			log = log.With("user_id", userID)
		}

		if len(c.Errors) > 0 {
			log.Errorw("Request completed with errors", "errors", c.Errors.String())
		} else if status >= 500 {
			log.Error("Request completed")
		} else if status >= 400 {
			log.Warn("Request completed")
		} else {
			log.Info("Request completed")
		}
	}
}

// BodyLimitMiddleware limits request body size
// maxSize is in bytes (e.g., 10 << 20 for 10MB)
func BodyLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.AbortWithStatusJSON(413, gin.H{
				"success": false,
				"error":   "Request body too large",
			})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// MaskCredential masks a sensitive string, showing only first/last chars
func MaskCredential(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "..." + s[len(s)-4:]
}
