package service

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ghstouch/one-go/internal/model"
	"github.com/ghstouch/one-go/internal/repository"
)

// ProviderService defines provider operations
type ProviderService interface {
	CreateProvider(provider *model.Provider) error
	ListProviders() ([]model.Provider, error)
	GetProviderByID(id string) (*model.Provider, error)
	UpdateProvider(provider *model.Provider) error
	DeleteProvider(id string) error
	TestProvider(id string) (string, error)
	ValidateAPIKey(baseURL, apiKey, providerType string) (string, error)
}

type providerService struct {
	repo repository.ProviderRepository
}

// NewProviderService creates a new provider service
func NewProviderService(repo repository.ProviderRepository) ProviderService {
	return &providerService{repo: repo}
}

func (s *providerService) CreateProvider(provider *model.Provider) error {
	if provider.Name == "" {
		return errors.New("provider name is required")
	}
	if provider.BaseURL == "" {
		return errors.New("provider base URL is required")
	}
	if provider.Type == "" {
		return errors.New("provider type is required")
	}

	// Auto-detect endpoints if none specified
	if len(provider.EndpointsJSON) == 0 {
		provider.EndpointsJSON = provider.AutoDetectEndpoints()
	}

	return s.repo.Create(provider)
}

func (s *providerService) ListProviders() ([]model.Provider, error) {
	return s.repo.ListProviders()
}

func (s *providerService) GetProviderByID(id string) (*model.Provider, error) {
	if id == "" {
		return nil, errors.New("provider ID is required")
	}
	return s.repo.GetByID(id)
}

func (s *providerService) UpdateProvider(provider *model.Provider) error {
	if provider.ID == "" {
		return errors.New("provider ID is required")
	}

	// Re-detect endpoints if none specified
	if len(provider.EndpointsJSON) == 0 {
		provider.EndpointsJSON = provider.AutoDetectEndpoints()
	}

	return s.repo.Update(provider)
}

func (s *providerService) DeleteProvider(id string) error {
	if id == "" {
		return errors.New("provider ID is required")
	}
	return s.repo.Delete(id)
}

func (s *providerService) TestProvider(id string) (string, error) {
	if id == "" {
		return "", errors.New("provider ID is required")
	}

	provider, err := s.repo.GetByID(id)
	if err != nil {
		return "", errors.New("provider not found")
	}

	// Simple connectivity test: GET /models endpoint
	if provider.BaseURL == "" {
		return "", errors.New("provider base URL is empty")
	}

	testURL := provider.BaseURL
	if !strings.HasSuffix(testURL, "/") {
		testURL += "/"
	}
	testURL += "models"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	if provider.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		// Update provider test status
		provider.Status = "inactive"
		s.repo.Update(provider)
		return "", fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		provider.Status = "active"
		s.repo.Update(provider)
		return fmt.Sprintf("Connected successfully (HTTP %d)", resp.StatusCode), nil
	}

	provider.Status = "inactive"
	s.repo.Update(provider)
	return "", fmt.Errorf("server returned HTTP %d", resp.StatusCode)
}

func (s *providerService) ValidateAPIKey(baseURL, apiKey, providerType string) (string, error) {
	if baseURL == "" {
		return "", errors.New("base URL is required")
	}
	if apiKey == "" {
		return "", errors.New("API key is required")
	}

	testURL := strings.TrimRight(baseURL, "/") + "/models"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return fmt.Sprintf("API key valid (HTTP %d)", resp.StatusCode), nil
	}
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return "", fmt.Errorf("API key rejected (HTTP %d)", resp.StatusCode)
	}
	return "", fmt.Errorf("server returned HTTP %d", resp.StatusCode)
}
