package service

import (
	"testing"

	"github.com/omniroute-go/internal/model"
)

// mockProviderRepo implements repository.ProviderRepository for testing
type mockProviderRepo struct {
	providers map[string]*model.Provider
	createErr error
	getErr    error
}

func newMockProviderRepo() *mockProviderRepo {
	return &mockProviderRepo{providers: make(map[string]*model.Provider)}
}

func (m *mockProviderRepo) Create(provider *model.Provider) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.providers[provider.ID] = provider
	return nil
}

func (m *mockProviderRepo) ListProviders() ([]model.Provider, error) {
	var list []model.Provider
	for _, p := range m.providers {
		list = append(list, *p)
	}
	return list, nil
}

func (m *mockProviderRepo) GetByID(id string) (*model.Provider, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	p, ok := m.providers[id]
	if !ok {
		return nil, nil
	}
	return p, nil
}

func (m *mockProviderRepo) Update(provider *model.Provider) error {
	m.providers[provider.ID] = provider
	return nil
}

func (m *mockProviderRepo) Delete(id string) error {
	delete(m.providers, id)
	return nil
}

func TestCreateProvider_Success(t *testing.T) {
	repo := newMockProviderRepo()
	svc := NewProviderService(repo)

	p := &model.Provider{
		Name:    "Test OpenAI",
		Type:    "openai",
		BaseURL: "https://api.openai.com/v1",
	}
	err := svc.CreateProvider(p)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateProvider_MissingName(t *testing.T) {
	repo := newMockProviderRepo()
	svc := NewProviderService(repo)

	p := &model.Provider{Type: "openai", BaseURL: "https://api.openai.com/v1"}
	err := svc.CreateProvider(p)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestCreateProvider_MissingBaseURL(t *testing.T) {
	repo := newMockProviderRepo()
	svc := NewProviderService(repo)

	p := &model.Provider{Name: "Test", Type: "openai"}
	err := svc.CreateProvider(p)
	if err == nil {
		t.Fatal("expected error for missing base URL")
	}
}

func TestCreateProvider_MissingType(t *testing.T) {
	repo := newMockProviderRepo()
	svc := NewProviderService(repo)

	p := &model.Provider{Name: "Test", BaseURL: "https://api.openai.com/v1"}
	err := svc.CreateProvider(p)
	if err == nil {
		t.Fatal("expected error for missing type")
	}
}

func TestGetProviderByID_Success(t *testing.T) {
	repo := newMockProviderRepo()
	svc := NewProviderService(repo)

	p := &model.Provider{ID: "p1", Name: "Test", Type: "openai", BaseURL: "https://api.openai.com/v1"}
	repo.providers["p1"] = p

	result, err := svc.GetProviderByID("p1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Name != "Test" {
		t.Fatalf("expected name 'Test', got '%s'", result.Name)
	}
}

func TestGetProviderByID_EmptyID(t *testing.T) {
	repo := newMockProviderRepo()
	svc := NewProviderService(repo)

	_, err := svc.GetProviderByID("")
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestUpdateProvider_Success(t *testing.T) {
	repo := newMockProviderRepo()
	svc := NewProviderService(repo)

	p := &model.Provider{ID: "p1", Name: "Old", Type: "openai", BaseURL: "https://api.openai.com/v1"}
	repo.providers["p1"] = p

	p.Name = "Updated"
	err := svc.UpdateProvider(p)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.providers["p1"].Name != "Updated" {
		t.Fatal("expected name to be updated")
	}
}

func TestDeleteProvider_Success(t *testing.T) {
	repo := newMockProviderRepo()
	svc := NewProviderService(repo)

	repo.providers["p1"] = &model.Provider{ID: "p1", Name: "Test"}
	err := svc.DeleteProvider("p1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := repo.providers["p1"]; exists {
		t.Fatal("expected provider to be deleted")
	}
}

func TestDeleteProvider_EmptyID(t *testing.T) {
	repo := newMockProviderRepo()
	svc := NewProviderService(repo)

	err := svc.DeleteProvider("")
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestListProviders(t *testing.T) {
	repo := newMockProviderRepo()
	svc := NewProviderService(repo)

	repo.providers["p1"] = &model.Provider{ID: "p1", Name: "A"}
	repo.providers["p2"] = &model.Provider{ID: "p2", Name: "B"}

	list, err := svc.ListProviders()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(list))
	}
}

func TestValidateAPIKey_MissingBaseURL(t *testing.T) {
	repo := newMockProviderRepo()
	svc := NewProviderService(repo)

	_, err := svc.ValidateAPIKey("", "sk-test", "openai")
	if err == nil {
		t.Fatal("expected error for missing base URL")
	}
}

func TestValidateAPIKey_MissingKey(t *testing.T) {
	repo := newMockProviderRepo()
	svc := NewProviderService(repo)

	_, err := svc.ValidateAPIKey("https://api.openai.com/v1", "", "openai")
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}
