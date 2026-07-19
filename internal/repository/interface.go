package repository

import (
	"github.com/ghstouch/one-go/internal/model"
)

// UserRepository interface defines user data operations
type UserRepository interface {
	Create(user *model.User) error
	GetByID(id string) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	Update(user *model.User) error
	Delete(id string) error
	List(page, pageSize int) ([]model.User, int64, error)
	UpdateLastLogin(id string) error
}

// SettingsRepository interface defines settings data operations
type SettingsRepository interface {
	Get(key string) (*model.Settings, error)
	Set(key, value, settingType string) error
	GetAll() ([]model.Settings, error)
	Delete(key string) error
	GetMultiple(keys []string) (map[string]string, error)
}
