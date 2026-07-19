package service

import (
	"strconv"

	"github.com/omniroute-go/internal/model"
	"github.com/omniroute-go/internal/repository"
)

// SettingsService interface defines settings operations
type SettingsService interface {
	Get(key string) (string, error)
	GetInt(key string, defaultValue int) int
	GetBool(key string, defaultValue bool) bool
	Set(key, value string) error
	SetInt(key string, value int) error
	SetBool(key string, value bool) error
	GetAll() (map[string]string, error)
	Delete(key string) error
}

type settingsService struct {
	repo repository.SettingsRepository
}

// NewSettingsService creates a new settings service
func NewSettingsService(repo repository.SettingsRepository) SettingsService {
	return &settingsService{repo: repo}
}

// Get retrieves a setting value by key
func (s *settingsService) Get(key string) (string, error) {
	setting, err := s.repo.Get(key)
	if err != nil {
		return "", err
	}
	if setting == nil {
		return "", nil
	}
	return setting.Value, nil
}

// GetInt retrieves a setting as int
func (s *settingsService) GetInt(key string, defaultValue int) int {
	value, err := s.Get(key)
	if err != nil {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// GetBool retrieves a setting as bool
func (s *settingsService) GetBool(key string, defaultValue bool) bool {
	value, err := s.Get(key)
	if err != nil || value == "" {
		return defaultValue
	}
	return value == "true" || value == "1"
}

// Set saves a string setting
func (s *settingsService) Set(key, value string) error {
	return s.repo.Set(key, value, model.SettingTypeString)
}

// SetInt saves an int setting
func (s *settingsService) SetInt(key string, value int) error {
	return s.repo.Set(key, strconv.Itoa(value), model.SettingTypeInt)
}

// SetBool saves a bool setting
func (s *settingsService) SetBool(key string, value bool) error {
	strValue := "false"
	if value {
		strValue = "true"
	}
	return s.repo.Set(key, strValue, model.SettingTypeBool)
}

// GetAll retrieves all settings as a map
func (s *settingsService) GetAll() (map[string]string, error) {
	settings, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, setting := range settings {
		result[setting.Key] = setting.Value
	}
	return result, nil
}

// Delete removes a setting
func (s *settingsService) Delete(key string) error {
	return s.repo.Delete(key)
}
