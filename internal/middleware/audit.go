package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ghstouch/one-go/internal/service"
)

// AuditMiddleware logs admin actions automatically
func AuditMiddleware(auditSvc service.AuditLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only log mutating actions
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" {
			return
		}

		// Skip v1 endpoints (logged via usage tracking)
		if strings.HasPrefix(c.Request.URL.Path, "/v1/") {
			return
		}

		// Skip auth endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/api/auth/") {
			return
		}

		// Skip if request failed with 4xx/5xx
		status := c.Writer.Status()
		if status >= 400 {
			return
		}

		userID := GetUserIDFromContext(c)
		username := GetUsernameFromContext(c)
		ip := c.ClientIP()

		action := methodToAction(method)
		resource, resourceID := parseResource(c.Request.URL.Path)
		details := method + " " + c.Request.URL.Path

		auditSvc.Log(userID, username, action, resource, resourceID, details, ip)
	}
}

func methodToAction(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return method
	}
}

func parseResource(path string) (string, string) {
	parts := strings.Split(strings.TrimPrefix(path, "/api/"), "/")
	if len(parts) == 0 {
		return "", ""
	}

	resource := parts[0]
	// Map plural to singular
	switch resource {
	case "providers":
		resource = "provider"
	case "combos":
		resource = "combo"
	case "keys":
		resource = "api_key"
	case "proxy":
		resource = "proxy"
	case "settings":
		resource = "settings"
	}

	resourceID := ""
	if len(parts) > 1 {
		resourceID = parts[1]
	}

	return resource, resourceID
}
