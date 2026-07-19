package repository

import (
	"time"

	"github.com/omniroute-go/internal/model"
	"gorm.io/gorm"
)

// UsageRepository defines usage data operations
type UsageRepository interface {
	CreateUsage(usage *model.UsageHistory) error
	ListUsage(page, pageSize int, provider string, from, to *time.Time) ([]model.UsageHistory, int64, error)
	GetUsageStats(from, to *time.Time) (*UsageStats, error)
	CreateCallLog(log *model.CallLog) error
	ListCallLogs(page, pageSize int, provider string, isError *bool) ([]model.CallLog, int64, error)
	GetCallLogByID(id string) (*model.CallLog, error)
}

type UsageStats struct {
	TotalRequests int64   `json:"totalRequests"`
	SuccessCount  int64   `json:"successCount"`
	ErrorCount    int64   `json:"errorCount"`
	TotalTokens   int64   `json:"totalTokens"`
	TotalCost     float64 `json:"totalCost"`
	AvgLatencyMs  float64 `json:"avgLatencyMs"`
}

type usageRepo struct {
	db *gorm.DB
}

func NewUsageRepository(db *gorm.DB) UsageRepository {
	return &usageRepo{db: db}
}

func (r *usageRepo) CreateUsage(usage *model.UsageHistory) error {
	return r.db.Create(usage).Error
}

func (r *usageRepo) ListUsage(page, pageSize int, provider string, from, to *time.Time) ([]model.UsageHistory, int64, error) {
	var items []model.UsageHistory
	var total int64

	query := r.db.Model(&model.UsageHistory{})
	if provider != "" {
		query = query.Where("provider = ?", provider)
	}
	if from != nil {
		query = query.Where("timestamp >= ?", from)
	}
	if to != nil {
		query = query.Where("timestamp <= ?", to)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *usageRepo) GetUsageStats(from, to *time.Time) (*UsageStats, error) {
	var stats UsageStats
	query := r.db.Model(&model.UsageHistory{})
	if from != nil {
		query = query.Where("timestamp >= ?", from)
	}
	if to != nil {
		query = query.Where("timestamp <= ?", to)
	}

	err := query.Select(
		"COUNT(*) as total_requests",
		"SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as success_count",
		"SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) as error_count",
		"COALESCE(SUM(total_tokens), 0) as total_tokens",
		"COALESCE(SUM(cost), 0) as total_cost",
		"COALESCE(AVG(latency_ms), 0) as avg_latency_ms",
	).Scan(&stats).Error
	return &stats, err
}

func (r *usageRepo) CreateCallLog(log *model.CallLog) error {
	return r.db.Create(log).Error
}

func (r *usageRepo) ListCallLogs(page, pageSize int, provider string, isError *bool) ([]model.CallLog, int64, error) {
	var items []model.CallLog
	var total int64

	query := r.db.Model(&model.CallLog{})
	if provider != "" {
		query = query.Where("provider = ?", provider)
	}
	if isError != nil {
		query = query.Where("is_error = ?", *isError)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *usageRepo) GetCallLogByID(id string) (*model.CallLog, error) {
	var log model.CallLog
	err := r.db.Where("id = ?", id).First(&log).Error
	return &log, err
}
