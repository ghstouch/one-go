package service

import (
	"testing"
	"time"

	"github.com/ghstouch/one-go/internal/model"
	"github.com/ghstouch/one-go/internal/repository"
)

// mockUsageRepo implements repository.UsageRepository for testing
type mockUsageRepo struct {
	usages []model.UsageHistory
	logs   []model.CallLog
}

func newMockUsageRepo() *mockUsageRepo {
	return &mockUsageRepo{}
}

func (m *mockUsageRepo) CreateUsage(usage *model.UsageHistory) error {
	m.usages = append(m.usages, *usage)
	return nil
}

func (m *mockUsageRepo) ListUsage(page, pageSize int, provider string, from, to *time.Time) ([]model.UsageHistory, int64, error) {
	var filtered []model.UsageHistory
	for _, u := range m.usages {
		if provider != "" && u.Provider != provider {
			continue
		}
		filtered = append(filtered, u)
	}
	total := int64(len(filtered))
	start := (page - 1) * pageSize
	if start >= len(filtered) {
		return nil, total, nil
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total, nil
}

func (m *mockUsageRepo) GetUsageStats(from, to *time.Time) (*repository.UsageStats, error) {
	var stats repository.UsageStats
	stats.TotalRequests = int64(len(m.usages))
	for _, u := range m.usages {
		if u.Success {
			stats.SuccessCount++
		} else {
			stats.ErrorCount++
		}
		stats.TotalTokens += int64(u.TotalTokens)
		stats.TotalCost += u.Cost
	}
	return &stats, nil
}

func (m *mockUsageRepo) CreateCallLog(log *model.CallLog) error {
	m.logs = append(m.logs, *log)
	return nil
}

func (m *mockUsageRepo) ListCallLogs(page, pageSize int, provider string, isError *bool) ([]model.CallLog, int64, error) {
	var filtered []model.CallLog
	for _, l := range m.logs {
		if provider != "" && l.Provider != provider {
			continue
		}
		if isError != nil && l.IsError != *isError {
			continue
		}
		filtered = append(filtered, l)
	}
	return filtered, int64(len(filtered)), nil
}

func (m *mockUsageRepo) GetCallLogByID(id string) (*model.CallLog, error) {
	for i, l := range m.logs {
		if l.ID == id {
			return &m.logs[i], nil
		}
	}
	return nil, nil
}

func TestRecordUsage(t *testing.T) {
	repo := newMockUsageRepo()
	svc := NewUsageService(repo)

	usage := &model.UsageHistory{Provider: "openai", Model: "gpt-4", Status: "success", Success: true}
	err := svc.RecordUsage(usage)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.usages) != 1 {
		t.Fatalf("expected 1 usage record, got %d", len(repo.usages))
	}
}

func TestListUsage_Pagination(t *testing.T) {
	repo := newMockUsageRepo()
	svc := NewUsageService(repo)

	for i := 0; i < 5; i++ {
		repo.usages = append(repo.usages, model.UsageHistory{Provider: "openai", Model: "gpt-4"})
	}

	items, total, err := svc.ListUsage(1, 3, "", nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 5 {
		t.Fatalf("expected total 5, got %d", total)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	items2, _, _ := svc.ListUsage(2, 3, "", nil, nil)
	if len(items2) != 2 {
		t.Fatalf("expected 2 items on page 2, got %d", len(items2))
	}
}

func TestListUsage_FilterProvider(t *testing.T) {
	repo := newMockUsageRepo()
	svc := NewUsageService(repo)

	repo.usages = append(repo.usages, model.UsageHistory{Provider: "openai"})
	repo.usages = append(repo.usages, model.UsageHistory{Provider: "anthropic"})
	repo.usages = append(repo.usages, model.UsageHistory{Provider: "openai"})

	items, total, err := svc.ListUsage(1, 10, "openai", nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 items, got %d", total)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestGetUsageStats(t *testing.T) {
	repo := newMockUsageRepo()
	svc := NewUsageService(repo)

	repo.usages = append(repo.usages, model.UsageHistory{Success: true, TotalTokens: 100})
	repo.usages = append(repo.usages, model.UsageHistory{Success: false, TotalTokens: 50})
	repo.usages = append(repo.usages, model.UsageHistory{Success: true, TotalTokens: 200})

	stats, err := svc.GetUsageStats(nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if stats.TotalRequests != 3 {
		t.Fatalf("expected 3 requests, got %d", stats.TotalRequests)
	}
	if stats.SuccessCount != 2 {
		t.Fatalf("expected 2 success, got %d", stats.SuccessCount)
	}
	if stats.ErrorCount != 1 {
		t.Fatalf("expected 1 error, got %d", stats.ErrorCount)
	}
	if stats.TotalTokens != 350 {
		t.Fatalf("expected 350 tokens, got %d", stats.TotalTokens)
	}
}

func TestRecordCallLog(t *testing.T) {
	repo := newMockUsageRepo()
	svc := NewUsageService(repo)

	log := &model.CallLog{Provider: "openai", RequestMethod: "POST", StatusCode: 200}
	err := svc.RecordCallLog(log)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(repo.logs))
	}
}

func TestListCallLogs_FilterError(t *testing.T) {
	repo := newMockUsageRepo()
	svc := NewUsageService(repo)

	repo.logs = append(repo.logs, model.CallLog{Provider: "openai", IsError: false})
	repo.logs = append(repo.logs, model.CallLog{Provider: "openai", IsError: true})
	repo.logs = append(repo.logs, model.CallLog{Provider: "openai", IsError: false})

	isError := true
	items, total, err := svc.ListCallLogs(1, 10, "", &isError)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 {
		t.Fatalf("expected 1 error log, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestGetCallLogByID(t *testing.T) {
	repo := newMockUsageRepo()
	svc := NewUsageService(repo)

	repo.logs = append(repo.logs, model.CallLog{ID: "log1", Provider: "openai"})

	log, err := svc.GetCallLogByID("log1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if log.Provider != "openai" {
		t.Fatalf("expected provider 'openai', got '%s'", log.Provider)
	}
}

func TestGetCallLogByID_NotFound(t *testing.T) {
	repo := newMockUsageRepo()
	svc := NewUsageService(repo)

	log, err := svc.GetCallLogByID("nonexistent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if log != nil {
		t.Fatal("expected nil for not found")
	}
}
