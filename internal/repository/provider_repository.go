package repository

import (
	"github.com/omniroute-go/internal/model"
	"gorm.io/gorm"
)

// ProviderRepository defines provider data operations
type ProviderRepository interface {
	Create(provider *model.Provider) error
	ListProviders() ([]model.Provider, error)
	GetByID(id string) (*model.Provider, error)
	Update(provider *model.Provider) error
	Delete(id string) error
}

type providerRepo struct {
	db *gorm.DB
}

// NewProviderRepository creates a new provider repository
func NewProviderRepository(db *gorm.DB) ProviderRepository {
	return &providerRepo{db: db}
}

func (r *providerRepo) Create(provider *model.Provider) error {
	return r.db.Create(provider).Error
}

func (r *providerRepo) ListProviders() ([]model.Provider, error) {
	var providers []model.Provider
	err := r.db.Find(&providers).Error
	return providers, err
}

func (r *providerRepo) GetByID(id string) (*model.Provider, error) {
	var provider model.Provider
	err := r.db.Where("id = ?", id).First(&provider).Error
	return &provider, err
}

func (r *providerRepo) Update(provider *model.Provider) error {
	return r.db.Save(provider).Error
}

func (r *providerRepo) Delete(id string) error {
	return r.db.Delete(&model.Provider{}, "id = ?", id).Error
}
