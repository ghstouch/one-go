package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ghstouch/one-go/internal/repository"
	"github.com/ghstouch/one-go/pkg/response"
)

// DashboardHandler handles dashboard HTTP requests
type DashboardHandler struct {
	providerRepo repository.ProviderRepository
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(providerRepo repository.ProviderRepository) *DashboardHandler {
	return &DashboardHandler{providerRepo: providerRepo}
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalProviders  int64   `json:"totalProviders"`
	ActiveProviders int64   `json:"activeProviders"`
	TotalCombos     int64   `json:"totalCombos"`
	TotalAPIKeys    int64   `json:"totalApiKeys"`
	TotalRequests   int64   `json:"totalRequests"`
	SuccessRate     float64 `json:"successRate"`
	TotalTokensUsed int64   `json:"totalTokensUsed"`
	TotalCost       float64 `json:"totalCost"`
	RequestsToday   int64   `json:"requestsToday"`
	TokensToday     int64   `json:"tokensToday"`
}

// GetDashboardStats returns dashboard statistics
func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	// TODO: Implement actual stats from database
	stats := DashboardStats{
		SuccessRate: 100.0,
	}
	response.OK(c, stats)
}

// GetRecentActivity returns recent activity
func (h *DashboardHandler) GetRecentActivity(c *gin.Context) {
	// TODO: Implement actual activity from database
	response.OK(c, []map[string]interface{}{})
}

// HealthCheck returns health status with provider info
func (h *DashboardHandler) HealthCheck(c *gin.Context) {
	result := gin.H{
		"status":  "healthy",
		"version": "1.0.0",
	}

	if h.providerRepo != nil {
		providers, err := h.providerRepo.ListProviders()
		if err == nil {
			var active int
			var providerStatus []gin.H
			for _, p := range providers {
				if p.Status == "active" {
					active++
				}
				providerStatus = append(providerStatus, gin.H{
					"name":   p.Name,
					"type":   p.Type,
					"status": p.Status,
				})
			}
			result["providers"] = gin.H{
				"total":  len(providers),
				"active": active,
				"list":   providerStatus,
			}
		}
	}

	response.OK(c, result)
}
