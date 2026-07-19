package service

import (
	"errors"

	"github.com/ghstouch/one-go/internal/model"
	"github.com/ghstouch/one-go/internal/repository"
)

// ComboService defines combo operations
type ComboService interface {
	CreateCombo(combo *model.Combo) error
	GetComboByID(id string) (*model.Combo, error)
	UpdateCombo(combo *model.Combo) error
	DeleteCombo(id string) error
	ListCombos() ([]model.Combo, error)
}

type comboService struct {
	repo repository.ComboRepository
}

func NewComboService(repo repository.ComboRepository) ComboService {
	return &comboService{repo: repo}
}

func (s *comboService) CreateCombo(combo *model.Combo) error {
	if combo.Name == "" {
		return errors.New("combo name is required")
	}
	validStrategies := map[string]bool{
		model.StrategyPriority: true, model.StrategyRoundRobin: true,
		model.StrategyWeighted: true, model.StrategyFallback: true,
	}
	if !validStrategies[combo.Strategy] {
		combo.Strategy = model.StrategyPriority
	}
	return s.repo.Create(combo)
}

func (s *comboService) GetComboByID(id string) (*model.Combo, error) {
	if id == "" {
		return nil, errors.New("combo ID is required")
	}
	return s.repo.GetByID(id)
}

func (s *comboService) UpdateCombo(combo *model.Combo) error {
	if combo.ID == "" {
		return errors.New("combo ID is required")
	}
	return s.repo.Update(combo)
}

func (s *comboService) DeleteCombo(id string) error {
	if id == "" {
		return errors.New("combo ID is required")
	}
	return s.repo.Delete(id)
}

func (s *comboService) ListCombos() ([]model.Combo, error) {
	return s.repo.List()
}
