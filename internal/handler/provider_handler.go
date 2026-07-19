package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omniroute-go/internal/model"
	"github.com/omniroute-go/internal/service"
	"github.com/omniroute-go/pkg/response"
)

// ProviderHandler handles provider HTTP requests
type ProviderHandler struct {
	providerSvc service.ProviderService
}

// NewProviderHandler creates a new provider handler
func NewProviderHandler(providerSvc service.ProviderService) *ProviderHandler {
	return &ProviderHandler{
		providerSvc: providerSvc,
	}
}

// CreateProviderRequest represents the request body for creating a provider
type CreateProviderRequest struct {
	Name         string   `json:"name" binding:"required,min=1,max=100"`
	Type         string   `json:"type" binding:"required,oneof=openai anthropic oauth cookie"`
	Description  string   `json:"description"`
	APIKey       string   `json:"apiKey"`
	ClientID     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	AuthURL      string   `json:"authUrl"`
	TokenURL     string   `json:"tokenUrl"`
	RedirectURI  string   `json:"redirectUri"`
	BaseURL      string   `json:"baseUrl" binding:"required"`
	Status       string   `json:"status" binding:"omitempty,oneof=active inactive pending"`
	Endpoints    []string `json:"endpoints"`
}

// GetProviders lists all providers
func (h *ProviderHandler) GetProviders(c *gin.Context) {
	providers, err := h.providerSvc.ListProviders()
	if err != nil {
		response.InternalServerError(c, "Failed to fetch providers")
		return
	}
	response.OK(c, providers)
}

// CreateProvider creates a new provider
func (h *ProviderHandler) CreateProvider(c *gin.Context) {
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	provider := model.Provider{
		Name:          req.Name,
		Type:          req.Type,
		Description:   req.Description,
		APIKey:        req.APIKey,
		ClientID:      req.ClientID,
		ClientSecret:  req.ClientSecret,
		AuthURL:       req.AuthURL,
		TokenURL:      req.TokenURL,
		RedirectURI:   req.RedirectURI,
		BaseURL:       req.BaseURL,
		Status:        req.Status,
		EndpointsJSON: req.Endpoints,
	}
	if provider.Status == "" {
		provider.Status = "active"
	}

	if err := h.providerSvc.CreateProvider(&provider); err != nil {
		response.InternalServerError(c, "Failed to create provider")
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Data:    provider,
	})
}

// UpdateProvider updates an existing provider
func (h *ProviderHandler) UpdateProvider(c *gin.Context) {
	id := c.Param("id")

	existing, err := h.providerSvc.GetProviderByID(id)
	if err != nil {
		response.NotFound(c, "Provider not found")
		return
	}

	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	existing.Name = req.Name
	existing.Type = req.Type
	existing.Description = req.Description
	existing.APIKey = req.APIKey
	existing.ClientID = req.ClientID
	existing.ClientSecret = req.ClientSecret
	existing.AuthURL = req.AuthURL
	existing.TokenURL = req.TokenURL
	existing.RedirectURI = req.RedirectURI
	existing.BaseURL = req.BaseURL
	if req.Status != "" {
		existing.Status = req.Status
	}
	existing.EndpointsJSON = req.Endpoints

	if err := h.providerSvc.UpdateProvider(existing); err != nil {
		response.InternalServerError(c, "Failed to update provider")
		return
	}

	response.OK(c, existing)
}

// DeleteProvider deletes a provider
func (h *ProviderHandler) DeleteProvider(c *gin.Context) {
	id := c.Param("id")

	_, err := h.providerSvc.GetProviderByID(id)
	if err != nil {
		response.NotFound(c, "Provider not found")
		return
	}

	if err := h.providerSvc.DeleteProvider(id); err != nil {
		response.InternalServerError(c, "Failed to delete provider")
		return
	}
	response.OKWithMessage(c, "Provider deleted", nil)
}

// TestProvider tests a provider connection
func (h *ProviderHandler) TestProvider(c *gin.Context) {
	id := c.Param("id")

	result, err := h.providerSvc.TestProvider(id)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OKWithMessage(c, result, nil)
}

// ValidateAPIKey validates a provider API key without saving
func (h *ProviderHandler) ValidateAPIKey(c *gin.Context) {
	var req struct {
		BaseURL string `json:"baseUrl" binding:"required"`
		APIKey  string `json:"apiKey" binding:"required"`
		Type    string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	result, err := h.providerSvc.ValidateAPIKey(req.BaseURL, req.APIKey, req.Type)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKWithMessage(c, result, nil)
}

// BatchDeleteProviders deletes multiple providers
func (h *ProviderHandler) BatchDeleteProviders(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	var deleted int
	for _, id := range req.IDs {
		if err := h.providerSvc.DeleteProvider(id); err == nil {
			deleted++
		}
	}
	response.OKWithMessage(c, fmt.Sprintf("Deleted %d providers", deleted), nil)
}
