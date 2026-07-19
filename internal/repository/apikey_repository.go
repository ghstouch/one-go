package repository

import (
	"time"

	"github.com/omniroute-go/internal/model"
	"gorm.io/gorm"
)

// ApiKeyRepository defines API key data operations
type ApiKeyRepository interface {
	Create(key *model.ApiKey) error
	GetByID(id string) (*model.ApiKey, error)
	GetByKey(key string) (*model.ApiKey, error)
	Update(key *model.ApiKey) error
	Delete(id string) error
	List() ([]model.ApiKey, error)
	IncrementUsage(id string) error
}

type apiKeyRepo struct {
	db *gorm.DB
}

func NewApiKeyRepository(db *gorm.DB) ApiKeyRepository {
	return &apiKeyRepo{db: db}
}

func (r *apiKeyRepo) Create(key *model.ApiKey) error {
	return r.db.Create(key).Error
}

func (r *apiKeyRepo) GetByID(id string) (*model.ApiKey, error) {
	var key model.ApiKey
	err := r.db.Where("id = ?", id).First(&key).Error
	return &key, err
}

func (r *apiKeyRepo) GetByKey(key string) (*model.ApiKey, error) {
	var apiKey model.ApiKey
	err := r.db.Where("key = ?", key).First(&apiKey).Error
	return &apiKey, err
}

func (r *apiKeyRepo) Update(key *model.ApiKey) error {
	return r.db.Save(key).Error
}

func (r *apiKeyRepo) Delete(id string) error {
	return r.db.Delete(&model.ApiKey{}, "id = ?", id).Error
}

func (r *apiKeyRepo) List() ([]model.ApiKey, error) {
	var keys []model.ApiKey
	err := r.db.Order("created_at DESC").Find(&keys).Error
	return keys, err
}

func (r *apiKeyRepo) IncrementUsage(id string) error {
	now := time.Now()
	return r.db.Model(&model.ApiKey{}).Where("id = ?", id).Updates(map[string]interface{}{
		"total_requests": gorm.Expr("total_requests + 1"),
		"last_used_at":   now,
	}).Error
}
