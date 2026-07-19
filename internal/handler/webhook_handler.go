package handler

import (
	"net/http"

	"github.com/ghstouch/one-go/internal/model"
	"github.com/ghstouch/one-go/internal/service"
	"github.com/ghstouch/one-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// WebhookHandler handles webhook HTTP requests
type WebhookHandler struct {
	svc service.WebhookService
}

func NewWebhookHandler(svc service.WebhookService) *WebhookHandler {
	return &WebhookHandler{svc: svc}
}

type CreateWebhookRequest struct {
	Name     string `json:"name" binding:"required"`
	URL      string `json:"url" binding:"required"`
	Secret   string `json:"secret"`
	Events   string `json:"events"`
	IsActive *bool  `json:"isActive"`
}

func (h *WebhookHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		response.InternalServerError(c, "Failed to fetch webhooks")
		return
	}
	response.OK(c, list)
}

func (h *WebhookHandler) Create(c *gin.Context) {
	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}

	webhook := &model.Webhook{
		Name:     req.Name,
		URL:      req.URL,
		Secret:   req.Secret,
		Events:   req.Events,
		IsActive: active,
	}
	if err := h.svc.Create(webhook); err != nil {
		response.InternalServerError(c, "Failed to create webhook: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, response.Response{Success: true, Data: webhook})
}

func (h *WebhookHandler) Update(c *gin.Context) {
	existing, err := h.svc.GetByID(c.Param("id"))
	if err != nil {
		response.NotFound(c, "Webhook not found")
		return
	}

	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	existing.Name = req.Name
	existing.URL = req.URL
	if req.Secret != "" {
		existing.Secret = req.Secret
	}
	existing.Events = req.Events
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := h.svc.Update(existing); err != nil {
		response.InternalServerError(c, "Failed to update webhook")
		return
	}
	response.OK(c, existing)
}

func (h *WebhookHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Param("id")); err != nil {
		response.InternalServerError(c, "Failed to delete webhook")
		return
	}
	response.OKWithMessage(c, "Webhook deleted", nil)
}

func (h *WebhookHandler) GetLogs(c *gin.Context) {
	webhookID := c.Query("webhookId")
	logs, err := h.svc.ListLogs(webhookID, 50)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch webhook logs")
		return
	}
	response.OK(c, logs)
}
