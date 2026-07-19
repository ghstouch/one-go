package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ghstouch/one-go/internal/service"
	"github.com/ghstouch/one-go/pkg/response"
)

// ApiKeyHandler handles API key HTTP requests
type ApiKeyHandler struct {
	apiKeySvc service.ApiKeyService
}

func NewApiKeyHandler(apiKeySvc service.ApiKeyService) *ApiKeyHandler {
	return &ApiKeyHandler{apiKeySvc: apiKeySvc}
}

type CreateApiKeyRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type UpdateApiKeyRequest struct {
	Name                 string `json:"name"`
	IsActive             *bool  `json:"isActive"`
	IsBanned             *bool  `json:"isBanned"`
	MaxRequestsPerDay    *int   `json:"maxRequestsPerDay"`
	MaxRequestsPerMinute *int   `json:"maxRequestsPerMinute"`
	AllowedModels        string `json:"allowedModels"`
	BlockedModels        string `json:"blockedModels"`
	AllowedCombos        string `json:"allowedCombos"`
	Scopes               string `json:"scopes"`
}

func (h *ApiKeyHandler) GetKeys(c *gin.Context) {
	keys, err := h.apiKeySvc.ListKeys()
	if err != nil {
		response.InternalServerError(c, "Failed to fetch API keys")
		return
	}
	// Mask keys for security - only show last 8 chars
	type SafeKey struct {
		ID                   string  `json:"id"`
		Name                 string  `json:"name"`
		KeyPreview           string  `json:"keyPreview"`
		IsActive             bool    `json:"isActive"`
		IsBanned             bool    `json:"isBanned"`
		MaxRequestsPerDay    int     `json:"maxRequestsPerDay"`
		MaxRequestsPerMinute int     `json:"maxRequestsPerMinute"`
		TotalRequests        int64   `json:"totalRequests"`
		AllowedModels        string  `json:"allowedModels,omitempty"`
		CreatedAt            string  `json:"createdAt"`
		LastUsedAt           *string `json:"lastUsedAt,omitempty"`
	}
	var safeKeys []SafeKey
	for _, k := range keys {
		preview := "sk-..."
		if len(k.Key) > 8 {
			preview = "sk-..." + k.Key[len(k.Key)-8:]
		}
		sk := SafeKey{
			ID:                   k.ID,
			Name:                 k.Name,
			KeyPreview:           preview,
			IsActive:             k.IsActive,
			IsBanned:             k.IsBanned,
			MaxRequestsPerDay:    k.MaxRequestsPerDay,
			MaxRequestsPerMinute: k.MaxRequestsPerMinute,
			TotalRequests:        k.TotalRequests,
			AllowedModels:        k.AllowedModels,
			CreatedAt:            k.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if k.LastUsedAt != nil {
			s := k.LastUsedAt.Format("2006-01-02 15:04:05")
			sk.LastUsedAt = &s
		}
		safeKeys = append(safeKeys, sk)
	}
	response.OK(c, safeKeys)
}

func (h *ApiKeyHandler) CreateKey(c *gin.Context) {
	var req CreateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	apiKey, rawKey, err := h.apiKeySvc.CreateKey(req.Name)
	if err != nil {
		response.InternalServerError(c, "Failed to create API key: "+err.Error())
		return
	}

	// Return the raw key only once
	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Data: map[string]interface{}{
			"id":        apiKey.ID,
			"name":      apiKey.Name,
			"key":       rawKey,
			"createdAt": apiKey.CreatedAt,
		},
		Message: "Save this key - it will not be shown again",
	})
}

func (h *ApiKeyHandler) GetKey(c *gin.Context) {
	key, err := h.apiKeySvc.GetKeyByID(c.Param("id"))
	if err != nil {
		response.NotFound(c, "API key not found")
		return
	}
	response.OK(c, map[string]interface{}{
		"id":                   key.ID,
		"name":                 key.Name,
		"isActive":             key.IsActive,
		"isBanned":             key.IsBanned,
		"maxRequestsPerDay":    key.MaxRequestsPerDay,
		"maxRequestsPerMinute": key.MaxRequestsPerMinute,
		"totalRequests":        key.TotalRequests,
		"createdAt":            key.CreatedAt,
	})
}

func (h *ApiKeyHandler) UpdateKey(c *gin.Context) {
	key, err := h.apiKeySvc.GetKeyByID(c.Param("id"))
	if err != nil {
		response.NotFound(c, "API key not found")
		return
	}

	var req UpdateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if req.Name != "" {
		key.Name = req.Name
	}
	if req.IsActive != nil {
		key.IsActive = *req.IsActive
	}
	if req.IsBanned != nil {
		key.IsBanned = *req.IsBanned
	}
	if req.MaxRequestsPerDay != nil {
		key.MaxRequestsPerDay = *req.MaxRequestsPerDay
	}
	if req.MaxRequestsPerMinute != nil {
		key.MaxRequestsPerMinute = *req.MaxRequestsPerMinute
	}
	if req.AllowedModels != "" {
		key.AllowedModels = req.AllowedModels
	}
	if req.BlockedModels != "" {
		key.BlockedModels = req.BlockedModels
	}
	if req.AllowedCombos != "" {
		key.AllowedCombos = req.AllowedCombos
	}
	if req.Scopes != "" {
		key.Scopes = req.Scopes
	}

	if err := h.apiKeySvc.UpdateKey(key); err != nil {
		response.InternalServerError(c, "Failed to update API key")
		return
	}
	response.OK(c, key)
}

func (h *ApiKeyHandler) DeleteKey(c *gin.Context) {
	if err := h.apiKeySvc.DeleteKey(c.Param("id")); err != nil {
		response.InternalServerError(c, "Failed to delete API key")
		return
	}
	response.OKWithMessage(c, "API key deleted", nil)
}
