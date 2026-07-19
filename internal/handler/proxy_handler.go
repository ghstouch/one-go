package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ghstouch/one-go/internal/model"
	"github.com/ghstouch/one-go/internal/service"
	"github.com/ghstouch/one-go/pkg/response"
)

// ProxyHandler handles proxy HTTP requests
type ProxyHandler struct {
	proxySvc service.ProxyService
}

func NewProxyHandler(proxySvc service.ProxyService) *ProxyHandler {
	return &ProxyHandler{proxySvc: proxySvc}
}

type CreateProxyRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Type     string `json:"type" binding:"required,oneof=http https socks5"`
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required,min=1,max=65535"`
	Username string `json:"username"`
	Password string `json:"password"`
	IsActive *bool  `json:"isActive"`
	IsGlobal *bool  `json:"isGlobal"`
}

func (h *ProxyHandler) GetProxies(c *gin.Context) {
	proxies, err := h.proxySvc.ListProxies()
	if err != nil {
		response.InternalServerError(c, "Failed to fetch proxies")
		return
	}
	response.OK(c, proxies)
}

func (h *ProxyHandler) CreateProxy(c *gin.Context) {
	var req CreateProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	global := false
	if req.IsGlobal != nil {
		global = *req.IsGlobal
	}

	proxy := model.Proxy{
		Name:     req.Name,
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
		IsActive: active,
		IsGlobal: global,
	}

	if err := h.proxySvc.CreateProxy(&proxy); err != nil {
		response.InternalServerError(c, "Failed to create proxy: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, response.Response{Success: true, Data: proxy})
}

func (h *ProxyHandler) UpdateProxy(c *gin.Context) {
	existing, err := h.proxySvc.GetProxyByID(c.Param("id"))
	if err != nil {
		response.NotFound(c, "Proxy not found")
		return
	}

	var req CreateProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	existing.Name = req.Name
	existing.Type = req.Type
	existing.Host = req.Host
	existing.Port = req.Port
	if req.Username != "" {
		existing.Username = req.Username
	}
	if req.Password != "" {
		existing.Password = req.Password
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.IsGlobal != nil {
		existing.IsGlobal = *req.IsGlobal
	}

	if err := h.proxySvc.UpdateProxy(existing); err != nil {
		response.InternalServerError(c, "Failed to update proxy")
		return
	}
	response.OK(c, existing)
}

func (h *ProxyHandler) DeleteProxy(c *gin.Context) {
	if err := h.proxySvc.DeleteProxy(c.Param("id")); err != nil {
		response.InternalServerError(c, "Failed to delete proxy")
		return
	}
	response.OKWithMessage(c, "Proxy deleted", nil)
}

func (h *ProxyHandler) TestProxy(c *gin.Context) {
	result, err := h.proxySvc.TestProxy(c.Param("id"))
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OKWithMessage(c, result, nil)
}

func (h *ProxyHandler) GetProxyLogs(c *gin.Context) {
	proxyID := c.Query("proxyId")
	logs, err := h.proxySvc.ListProxyLogs(proxyID, 50)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch proxy logs")
		return
	}
	response.OK(c, logs)
}
