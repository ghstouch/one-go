package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ghstouch/one-go/internal/service"
	"github.com/ghstouch/one-go/pkg/response"
)

// SettingsHandler handles settings HTTP requests
type SettingsHandler struct {
	settingsService service.SettingsService
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(settingsService service.SettingsService) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
	}
}

// UpdateSettingsRequest represents settings update request
type UpdateSettingsRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// GetSettings returns all application settings
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingsService.GetAll()
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve settings")
		return
	}

	response.OK(c, settings)
}

// UpdateSetting updates a single setting
func (h *SettingsHandler) UpdateSetting(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.settingsService.Set(req.Key, req.Value); err != nil {
		response.InternalServerError(c, "Failed to update setting")
		return
	}

	response.OKWithMessage(c, "Setting updated successfully", nil)
}
