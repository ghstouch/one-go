package repository

import (
	"github.com/omniroute-go/internal/model"
	"gorm.io/gorm"
)

type settingsRepository struct {
	db *gorm.DB
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(db *gorm.DB) SettingsRepository {
	return &settingsRepository{db: db}
}

// Get retrieves a setting by key
func (r *settingsRepository) Get(key string) (*model.Settings, error) {
	var setting model.Settings
	if err := r.db.Where("key = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

// Set creates or updates a setting
func (r *settingsRepository) Set(key, value, settingType string) error {
	var setting model.Settings
	result := r.db.Where("key = ?", key).First(&setting)

	if result.Error == gorm.ErrRecordNotFound {
		// Create new setting
		setting = model.Settings{
			Key:   key,
			Value: value,
			Type:  settingType,
		}
		return r.db.Create(&setting).Error
	}

	if result.Error != nil {
		return result.Error
	}

	// Update existing setting
	setting.Value = value
	if settingType != "" {
		setting.Type = settingType
	}
	return r.db.Save(&setting).Error
}

// GetAll retrieves all settings
func (r *settingsRepository) GetAll() ([]model.Settings, error) {
	var settings []model.Settings
	if err := r.db.Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

// Delete removes a setting
func (r *settingsRepository) Delete(key string) error {
	return r.db.Delete(&model.Settings{}, "key = ?", key).Error
}

// GetMultiple retrieves multiple settings by keys
func (r *settingsRepository) GetMultiple(keys []string) (map[string]string, error) {
	var settings []model.Settings
	if err := r.db.Where("key IN ?", keys).Find(&settings).Error; err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}
