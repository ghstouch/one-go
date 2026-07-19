package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ghstouch/one-go/internal/service"
	"github.com/ghstouch/one-go/pkg/response"
)

// AuditLogHandler handles audit log HTTP requests
type AuditLogHandler struct {
	svc service.AuditLogService
}

func NewAuditLogHandler(svc service.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{svc: svc}
}

func (h *AuditLogHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	action := c.Query("action")
	resource := c.Query("resource")

	items, total, err := h.svc.List(page, pageSize, action, resource)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch audit logs")
		return
	}
	response.Paginated(c, items, page, pageSize, total)
}

func (h *AuditLogHandler) GetByID(c *gin.Context) {
	log, err := h.svc.GetByID(c.Param("id"))
	if err != nil {
		response.NotFound(c, "Audit log not found")
		return
	}
	response.OK(c, log)
}
