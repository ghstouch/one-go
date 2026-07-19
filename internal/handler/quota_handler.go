package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ghstouch/one-go/internal/service"
	"github.com/ghstouch/one-go/pkg/response"
)

// QuotaHandler handles quota HTTP requests
type QuotaHandler struct {
	quotaSvc service.QuotaService
}

func NewQuotaHandler(quotaSvc service.QuotaService) *QuotaHandler {
	return &QuotaHandler{quotaSvc: quotaSvc}
}

// GetQuotas returns all provider quotas
func (h *QuotaHandler) GetQuotas(c *gin.Context) {
	quotas, err := h.quotaSvc.GetAllQuotas()
	if err != nil {
		response.InternalServerError(c, "Failed to fetch quotas")
		return
	}
	response.OK(c, quotas)
}

// GetProviderQuota returns quota for a specific provider
func (h *QuotaHandler) GetProviderQuota(c *gin.Context) {
	providerName := c.Param("provider")
	quota, err := h.quotaSvc.GetProviderQuota(providerName)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch quota")
		return
	}
	response.OK(c, quota)
}
