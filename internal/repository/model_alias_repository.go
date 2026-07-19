package repository

import (
	"github.com/omniroute-go/internal/model"
	"gorm.io/gorm"
)

// ModelAliasRepository defines model alias data operations
type ModelAliasRepository interface {
	Create(alias *model.ModelAlias) error
	GetByAlias(alias string) (*model.ModelAlias, error)
	List() ([]model.ModelAlias, error)
	Update(alias *model.ModelAlias) error
	Delete(id string) error
}

type modelAliasRepo struct {
	db *gorm.DB
}

func NewModelAliasRepository(db *gorm.DB) ModelAliasRepository {
	return &modelAliasRepo{db: db}
}

func (r *modelAliasRepo) Create(alias *model.ModelAlias) error {
	return r.db.Create(alias).Error
}

func (r *modelAliasRepo) GetByAlias(alias string) (*model.ModelAlias, error) {
	var m model.ModelAlias
	err := r.db.Where("alias = ?", alias).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *modelAliasRepo) List() ([]model.ModelAlias, error) {
	var list []model.ModelAlias
	err := r.db.Order("alias ASC").Find(&list).Error
	return list, err
}

func (r *modelAliasRepo) Update(alias *model.ModelAlias) error {
	return r.db.Save(alias).Error
}

func (r *modelAliasRepo) Delete(id string) error {
	return r.db.Delete(&model.ModelAlias{}, "id = ?", id).Error
}
