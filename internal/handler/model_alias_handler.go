package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omniroute-go/internal/model"
	"github.com/omniroute-go/internal/service"
	"github.com/omniroute-go/pkg/response"
)

// ModelAliasHandler handles model alias HTTP requests
type ModelAliasHandler struct {
	svc service.ModelAliasService
}

func NewModelAliasHandler(svc service.ModelAliasService) *ModelAliasHandler {
	return &ModelAliasHandler{svc: svc}
}

type CreateAliasRequest struct {
	Alias        string `json:"alias" binding:"required"`
	ProviderType string `json:"providerType"`
	ActualModel  string `json:"actualModel" binding:"required"`
	Description  string `json:"description"`
}

func (h *ModelAliasHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		response.InternalServerError(c, "Failed to fetch aliases")
		return
	}
	response.OK(c, list)
}

func (h *ModelAliasHandler) Create(c *gin.Context) {
	var req CreateAliasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	alias := &model.ModelAlias{
		Alias:        req.Alias,
		ProviderType: req.ProviderType,
		ActualModel:  req.ActualModel,
		Description:  req.Description,
	}
	if err := h.svc.Create(alias); err != nil {
		response.InternalServerError(c, "Failed to create alias: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, response.Response{Success: true, Data: alias})
}

func (h *ModelAliasHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Param("id")); err != nil {
		response.InternalServerError(c, "Failed to delete alias")
		return
	}
	response.OKWithMessage(c, "Alias deleted", nil)
}

func (h *ModelAliasHandler) Resolve(c *gin.Context) {
	alias := c.Query("alias")
	if alias == "" {
		response.BadRequest(c, "alias query parameter required")
		return
	}
	actual, err := h.svc.Resolve(alias)
	if err != nil {
		response.InternalServerError(c, "Failed to resolve alias")
		return
	}
	response.OK(c, gin.H{"alias": alias, "model": actual})
}
