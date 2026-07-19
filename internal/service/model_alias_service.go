package service

import (
	"errors"

	"github.com/omniroute-go/internal/model"
	"github.com/omniroute-go/internal/repository"
)

// ModelAliasService defines model alias operations
type ModelAliasService interface {
	Create(alias *model.ModelAlias) error
	Resolve(alias string) (string, error)
	List() ([]model.ModelAlias, error)
	Delete(id string) error
}

type modelAliasService struct {
	repo repository.ModelAliasRepository
}

func NewModelAliasService(repo repository.ModelAliasRepository) ModelAliasService {
	return &modelAliasService{repo: repo}
}

func (s *modelAliasService) Create(alias *model.ModelAlias) error {
	if alias.Alias == "" {
		return errors.New("alias is required")
	}
	if alias.ActualModel == "" {
		return errors.New("actual model is required")
	}
	return s.repo.Create(alias)
}

func (s *modelAliasService) Resolve(alias string) (string, error) {
	m, err := s.repo.GetByAlias(alias)
	if err != nil {
		return alias, nil // not found, return as-is
	}
	return m.ActualModel, nil
}

func (s *modelAliasService) List() ([]model.ModelAlias, error) {
	return s.repo.List()
}

func (s *modelAliasService) Delete(id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	return s.repo.Delete(id)
}
