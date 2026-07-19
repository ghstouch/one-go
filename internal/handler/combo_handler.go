package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ghstouch/one-go/internal/model"
	"github.com/ghstouch/one-go/internal/service"
	"github.com/ghstouch/one-go/pkg/response"
)

// ComboHandler handles combo HTTP requests
type ComboHandler struct {
	comboSvc service.ComboService
}

func NewComboHandler(comboSvc service.ComboService) *ComboHandler {
	return &ComboHandler{comboSvc: comboSvc}
}

type CreateComboRequest struct {
	Name            string           `json:"name" binding:"required,min=1,max=100"`
	Description     string           `json:"description"`
	Strategy        string           `json:"strategy" binding:"omitempty,oneof=priority round-robin weighted fallback"`
	IsHidden        bool             `json:"isHidden"`
	MaxRetries      int              `json:"maxRetries"`
	RetryDelayMs    int              `json:"retryDelayMs"`
	FallbackDelayMs int              `json:"fallbackDelayMs"`
	Targets         []ComboTargetReq `json:"targets"`
}

type ComboTargetReq struct {
	ProviderID string `json:"providerId" binding:"required"`
	ModelName  string `json:"modelName"`
	Priority   int    `json:"priority"`
	Weight     int    `json:"weight"`
}

func (h *ComboHandler) GetCombos(c *gin.Context) {
	combos, err := h.comboSvc.ListCombos()
	if err != nil {
		response.InternalServerError(c, "Failed to fetch combos")
		return
	}
	response.OK(c, combos)
}

func (h *ComboHandler) GetCombo(c *gin.Context) {
	combo, err := h.comboSvc.GetComboByID(c.Param("id"))
	if err != nil {
		response.NotFound(c, "Combo not found")
		return
	}
	response.OK(c, combo)
}

func (h *ComboHandler) CreateCombo(c *gin.Context) {
	var req CreateComboRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	combo := model.Combo{
		Name:            req.Name,
		Description:     req.Description,
		Strategy:        req.Strategy,
		IsHidden:        req.IsHidden,
		MaxRetries:      req.MaxRetries,
		RetryDelayMs:    req.RetryDelayMs,
		FallbackDelayMs: req.FallbackDelayMs,
	}
	if combo.Strategy == "" {
		combo.Strategy = model.StrategyPriority
	}
	if combo.MaxRetries == 0 {
		combo.MaxRetries = 3
	}

	for _, t := range req.Targets {
		combo.Targets = append(combo.Targets, model.ComboTarget{
			ProviderID: t.ProviderID,
			ModelName:  t.ModelName,
			Priority:   t.Priority,
			Weight:     t.Weight,
		})
	}

	if err := h.comboSvc.CreateCombo(&combo); err != nil {
		response.InternalServerError(c, "Failed to create combo: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, response.Response{Success: true, Data: combo})
}

func (h *ComboHandler) UpdateCombo(c *gin.Context) {
	id := c.Param("id")
	existing, err := h.comboSvc.GetComboByID(id)
	if err != nil {
		response.NotFound(c, "Combo not found")
		return
	}

	var req CreateComboRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.Strategy = req.Strategy
	existing.IsHidden = req.IsHidden
	existing.MaxRetries = req.MaxRetries
	existing.RetryDelayMs = req.RetryDelayMs
	existing.FallbackDelayMs = req.FallbackDelayMs

	// Replace targets
	existing.Targets = nil
	for _, t := range req.Targets {
		existing.Targets = append(existing.Targets, model.ComboTarget{
			ComboID:    id,
			ProviderID: t.ProviderID,
			ModelName:  t.ModelName,
			Priority:   t.Priority,
			Weight:     t.Weight,
		})
	}

	if err := h.comboSvc.UpdateCombo(existing); err != nil {
		response.InternalServerError(c, "Failed to update combo")
		return
	}
	response.OK(c, existing)
}

func (h *ComboHandler) DeleteCombo(c *gin.Context) {
	if err := h.comboSvc.DeleteCombo(c.Param("id")); err != nil {
		response.InternalServerError(c, "Failed to delete combo")
		return
	}
	response.OKWithMessage(c, "Combo deleted", nil)
}
