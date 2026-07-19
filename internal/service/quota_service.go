package service

import (
	"time"

	"github.com/omniroute-go/internal/repository"
)

// QuotaService provides quota tracking per provider
type QuotaService interface {
	GetProviderQuota(providerName string) (*QuotaInfo, error)
	GetAllQuotas() ([]QuotaInfo, error)
}

type QuotaInfo struct {
	Provider      string  `json:"provider"`
	DailyRequests int64   `json:"dailyRequests"`
	DailyTokens   int64   `json:"dailyTokens"`
	DailyCost     float64 `json:"dailyCost"`
	MonthlyTokens int64   `json:"monthlyTokens"`
	MonthlyCost   float64 `json:"monthlyCost"`
}

type quotaService struct {
	usageRepo repository.UsageRepository
}

func NewQuotaService(usageRepo repository.UsageRepository) QuotaService {
	return &quotaService{usageRepo: usageRepo}
}

func (s *quotaService) GetProviderQuota(providerName string) (*QuotaInfo, error) {
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Get daily stats
	dayFrom := dayStart
	dayTo := now
	dailyStats, err := s.usageRepo.GetUsageStats(&dayFrom, &dayTo)
	if err != nil {
		return nil, err
	}

	// Get monthly stats
	monthFrom := monthStart
	monthTo := now
	monthlyStats, err := s.usageRepo.GetUsageStats(&monthFrom, &monthTo)
	if err != nil {
		return nil, err
	}

	return &QuotaInfo{
		Provider:      providerName,
		DailyRequests: dailyStats.TotalRequests,
		DailyTokens:   dailyStats.TotalTokens,
		DailyCost:     dailyStats.TotalCost,
		MonthlyTokens: monthlyStats.TotalTokens,
		MonthlyCost:   monthlyStats.TotalCost,
	}, nil
}

func (s *quotaService) GetAllQuotas() ([]QuotaInfo, error) {
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	dayFrom := dayStart
	dayTo := now
	dailyStats, err := s.usageRepo.GetUsageStats(&dayFrom, &dayTo)
	if err != nil {
		return nil, err
	}

	monthFrom := monthStart
	monthTo := now
	monthlyStats, err := s.usageRepo.GetUsageStats(&monthFrom, &monthTo)
	if err != nil {
		return nil, err
	}

	return []QuotaInfo{{
		Provider:      "all",
		DailyRequests: dailyStats.TotalRequests,
		DailyTokens:   dailyStats.TotalTokens,
		DailyCost:     dailyStats.TotalCost,
		MonthlyTokens: monthlyStats.TotalTokens,
		MonthlyCost:   monthlyStats.TotalCost,
	}}, nil
}
