package handler

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omniroute-go/internal/service"
	"github.com/omniroute-go/pkg/response"
)

// UsageHandler handles usage HTTP requests
type UsageHandler struct {
	usageSvc service.UsageService
}

func NewUsageHandler(usageSvc service.UsageService) *UsageHandler {
	return &UsageHandler{usageSvc: usageSvc}
}

func (h *UsageHandler) GetUsage(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	provider := c.Query("provider")

	var from, to *time.Time
	if f := c.Query("from"); f != "" {
		t, err := time.Parse("2006-01-02", f)
		if err == nil {
			from = &t
		}
	}
	if t := c.Query("to"); t != "" {
		parsed, err := time.Parse("2006-01-02", t)
		if err == nil {
			endOfDay := parsed.Add(24*time.Hour - time.Second)
			to = &endOfDay
		}
	}

	items, total, err := h.usageSvc.ListUsage(page, pageSize, provider, from, to)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch usage")
		return
	}
	response.Paginated(c, items, page, pageSize, total)
}

func (h *UsageHandler) GetUsageStats(c *gin.Context) {
	var from, to *time.Time
	if f := c.Query("from"); f != "" {
		t, err := time.Parse("2006-01-02", f)
		if err == nil {
			from = &t
		}
	}
	if t := c.Query("to"); t != "" {
		parsed, err := time.Parse("2006-01-02", t)
		if err == nil {
			endOfDay := parsed.Add(24*time.Hour - time.Second)
			to = &endOfDay
		}
	}

	stats, err := h.usageSvc.GetUsageStats(from, to)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch stats")
		return
	}
	response.OK(c, stats)
}

func (h *UsageHandler) GetCallLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	provider := c.Query("provider")

	var isError *bool
	if e := c.Query("isError"); e == "true" {
		b := true
		isError = &b
	} else if e == "false" {
		b := false
		isError = &b
	}

	items, total, err := h.usageSvc.ListCallLogs(page, pageSize, provider, isError)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch logs")
		return
	}
	response.Paginated(c, items, page, pageSize, total)
}

func (h *UsageHandler) GetCallLogDetail(c *gin.Context) {
	log, err := h.usageSvc.GetCallLogByID(c.Param("id"))
	if err != nil {
		response.NotFound(c, "Log not found")
		return
	}
	response.OK(c, log)
}

// ExportUsageCSV exports usage data as CSV
func (h *UsageHandler) ExportUsageCSV(c *gin.Context) {
	provider := c.Query("provider")
	var from, to *time.Time
	if f := c.Query("from"); f != "" {
		t, err := time.Parse("2006-01-02", f)
		if err == nil {
			from = &t
		}
	}
	if t := c.Query("to"); t != "" {
		parsed, err := time.Parse("2006-01-02", t)
		if err == nil {
			endOfDay := parsed.Add(24*time.Hour - time.Second)
			to = &endOfDay
		}
	}

	// Fetch all (up to 10000)
	items, _, err := h.usageSvc.ListUsage(1, 10000, provider, from, to)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch usage data")
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=usage_%s.csv", time.Now().Format("20060102_150405")))

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"Time", "Provider", "Model", "Status", "Success", "Tokens In", "Tokens Out", "Total Tokens", "Latency (ms)", "Error"})

	for _, u := range items {
		w.Write([]string{
			u.Timestamp.Format(time.RFC3339),
			u.Provider,
			u.Model,
			u.Status,
			fmt.Sprintf("%t", u.Success),
			strconv.Itoa(u.TokensInput),
			strconv.Itoa(u.TokensOutput),
			strconv.Itoa(u.TotalTokens),
			strconv.Itoa(u.LatencyMs),
			u.ErrorMessage,
		})
	}
	w.Flush()
}
