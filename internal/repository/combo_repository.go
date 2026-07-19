package repository

import (
	"github.com/omniroute-go/internal/model"
	"gorm.io/gorm"
)

// ComboRepository defines combo data operations
type ComboRepository interface {
	Create(combo *model.Combo) error
	GetByID(id string) (*model.Combo, error)
	Update(combo *model.Combo) error
	Delete(id string) error
	List() ([]model.Combo, error)
}

type comboRepo struct {
	db *gorm.DB
}

func NewComboRepository(db *gorm.DB) ComboRepository {
	return &comboRepo{db: db}
}

func (r *comboRepo) Create(combo *model.Combo) error {
	return r.db.Create(combo).Error
}

func (r *comboRepo) GetByID(id string) (*model.Combo, error) {
	var combo model.Combo
	err := r.db.Preload("Targets").Where("id = ?", id).First(&combo).Error
	return &combo, err
}

func (r *comboRepo) Update(combo *model.Combo) error {
	return r.db.Save(combo).Error
}

func (r *comboRepo) Delete(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("combo_id = ?", id).Delete(&model.ComboTarget{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Combo{}, "id = ?", id).Error
	})
}

func (r *comboRepo) List() ([]model.Combo, error) {
	var combos []model.Combo
	err := r.db.Preload("Targets").Find(&combos).Error
	return combos, err
}
