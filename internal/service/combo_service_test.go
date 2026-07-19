package service

import (
	"testing"

	"github.com/omniroute-go/internal/model"
)

// mockComboRepo implements repository.ComboRepository for testing
type mockComboRepo struct {
	combos map[string]*model.Combo
}

func newMockComboRepo() *mockComboRepo {
	return &mockComboRepo{combos: make(map[string]*model.Combo)}
}

func (m *mockComboRepo) Create(combo *model.Combo) error {
	m.combos[combo.ID] = combo
	return nil
}

func (m *mockComboRepo) GetByID(id string) (*model.Combo, error) {
	if c, ok := m.combos[id]; ok {
		return c, nil
	}
	return nil, nil
}

func (m *mockComboRepo) Update(combo *model.Combo) error {
	m.combos[combo.ID] = combo
	return nil
}

func (m *mockComboRepo) Delete(id string) error {
	delete(m.combos, id)
	return nil
}

func (m *mockComboRepo) List() ([]model.Combo, error) {
	var list []model.Combo
	for _, c := range m.combos {
		list = append(list, *c)
	}
	return list, nil
}

func TestCreateCombo_Success(t *testing.T) {
	repo := newMockComboRepo()
	svc := NewComboService(repo)

	c := &model.Combo{Name: "Test Combo", Strategy: "priority"}
	err := svc.CreateCombo(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateCombo_MissingName(t *testing.T) {
	repo := newMockComboRepo()
	svc := NewComboService(repo)

	c := &model.Combo{Strategy: "priority"}
	err := svc.CreateCombo(c)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestCreateCombo_DefaultStrategy(t *testing.T) {
	repo := newMockComboRepo()
	svc := NewComboService(repo)

	c := &model.Combo{Name: "Test", Strategy: "invalid"}
	err := svc.CreateCombo(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c.Strategy != model.StrategyPriority {
		t.Fatalf("expected default strategy 'priority', got '%s'", c.Strategy)
	}
}

func TestGetComboByID_Success(t *testing.T) {
	repo := newMockComboRepo()
	svc := NewComboService(repo)

	repo.combos["c1"] = &model.Combo{ID: "c1", Name: "Test"}

	c, err := svc.GetComboByID("c1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c.Name != "Test" {
		t.Fatalf("expected name 'Test', got '%s'", c.Name)
	}
}

func TestGetComboByID_EmptyID(t *testing.T) {
	repo := newMockComboRepo()
	svc := NewComboService(repo)

	_, err := svc.GetComboByID("")
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestDeleteCombo_Success(t *testing.T) {
	repo := newMockComboRepo()
	svc := NewComboService(repo)

	repo.combos["c1"] = &model.Combo{ID: "c1", Name: "Test"}
	err := svc.DeleteCombo("c1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := repo.combos["c1"]; exists {
		t.Fatal("expected combo to be deleted")
	}
}

func TestListCombos(t *testing.T) {
	repo := newMockComboRepo()
	svc := NewComboService(repo)

	repo.combos["c1"] = &model.Combo{ID: "c1", Name: "A"}
	repo.combos["c2"] = &model.Combo{ID: "c2", Name: "B"}

	list, err := svc.ListCombos()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 combos, got %d", len(list))
	}
}
