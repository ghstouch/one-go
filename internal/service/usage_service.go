package service

import (
	"time"

	"github.com/omniroute-go/internal/model"
	"github.com/omniroute-go/internal/repository"
)

// UsageService defines usage operations
type UsageService interface {
	RecordUsage(usage *model.UsageHistory) error
	ListUsage(page, pageSize int, provider string, from, to *time.Time) ([]model.UsageHistory, int64, error)
	GetUsageStats(from, to *time.Time) (*repository.UsageStats, error)
	RecordCallLog(log *model.CallLog) error
	ListCallLogs(page, pageSize int, provider string, isError *bool) ([]model.CallLog, int64, error)
	GetCallLogByID(id string) (*model.CallLog, error)
}

type usageService struct {
	repo repository.UsageRepository
}

func NewUsageService(repo repository.UsageRepository) UsageService {
	return &usageService{repo: repo}
}

func (s *usageService) RecordUsage(usage *model.UsageHistory) error {
	return s.repo.CreateUsage(usage)
}

func (s *usageService) ListUsage(page, pageSize int, provider string, from, to *time.Time) ([]model.UsageHistory, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListUsage(page, pageSize, provider, from, to)
}

func (s *usageService) GetUsageStats(from, to *time.Time) (*repository.UsageStats, error) {
	return s.repo.GetUsageStats(from, to)
}

func (s *usageService) RecordCallLog(log *model.CallLog) error {
	return s.repo.CreateCallLog(log)
}

func (s *usageService) ListCallLogs(page, pageSize int, provider string, isError *bool) ([]model.CallLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListCallLogs(page, pageSize, provider, isError)
}

func (s *usageService) GetCallLogByID(id string) (*model.CallLog, error) {
	if id == "" {
		return nil, nil
	}
	return s.repo.GetCallLogByID(id)
}
