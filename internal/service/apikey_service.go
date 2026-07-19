package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/omniroute-go/internal/model"
	"github.com/omniroute-go/internal/repository"
)

// ApiKeyService defines API key operations
type ApiKeyService interface {
	CreateKey(name string) (*model.ApiKey, string, error)
	GetKeyByID(id string) (*model.ApiKey, error)
	UpdateKey(key *model.ApiKey) error
	DeleteKey(id string) error
	ListKeys() ([]model.ApiKey, error)
	ValidateKey(key string) (*model.ApiKey, error)
}

type apiKeyService struct {
	repo repository.ApiKeyRepository
}

func NewApiKeyService(repo repository.ApiKeyRepository) ApiKeyService {
	return &apiKeyService{repo: repo}
}

func generateKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sk-" + hex.EncodeToString(bytes), nil
}

func (s *apiKeyService) CreateKey(name string) (*model.ApiKey, string, error) {
	if name == "" {
		return nil, "", errors.New("API key name is required")
	}

	rawKey, err := generateKey()
	if err != nil {
		return nil, "", err
	}

	apiKey := &model.ApiKey{
		Name:     name,
		Key:      rawKey,
		IsActive: true,
		IsBanned: false,
	}

	if err := s.repo.Create(apiKey); err != nil {
		return nil, "", err
	}

	return apiKey, rawKey, nil
}

func (s *apiKeyService) GetKeyByID(id string) (*model.ApiKey, error) {
	if id == "" {
		return nil, errors.New("API key ID is required")
	}
	return s.repo.GetByID(id)
}

func (s *apiKeyService) UpdateKey(key *model.ApiKey) error {
	if key.ID == "" {
		return errors.New("API key ID is required")
	}
	return s.repo.Update(key)
}

func (s *apiKeyService) DeleteKey(id string) error {
	if id == "" {
		return errors.New("API key ID is required")
	}
	return s.repo.Delete(id)
}

func (s *apiKeyService) ListKeys() ([]model.ApiKey, error) {
	return s.repo.List()
}

func (s *apiKeyService) ValidateKey(key string) (*model.ApiKey, error) {
	if key == "" {
		return nil, errors.New("API key is required")
	}

	apiKey, err := s.repo.GetByKey(key)
	if err != nil || apiKey == nil {
		return nil, errors.New("invalid API key")
	}

	if !apiKey.IsActive {
		return nil, errors.New("API key is inactive")
	}
	if apiKey.IsBanned {
		return nil, errors.New("API key is banned")
	}
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("API key has expired")
	}

	return apiKey, nil
}
