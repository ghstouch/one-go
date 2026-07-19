package service

import (
	"strings"
	"testing"
	"time"

	"github.com/omniroute-go/internal/model"
)

// mockApiKeyRepo implements repository.ApiKeyRepository for testing
type mockApiKeyRepo struct {
	keys map[string]*model.ApiKey
}

func newMockApiKeyRepo() *mockApiKeyRepo {
	return &mockApiKeyRepo{keys: make(map[string]*model.ApiKey)}
}

func (m *mockApiKeyRepo) Create(key *model.ApiKey) error {
	m.keys[key.ID] = key
	m.keys[key.Key] = key
	return nil
}

func (m *mockApiKeyRepo) GetByID(id string) (*model.ApiKey, error) {
	if k, ok := m.keys[id]; ok {
		return k, nil
	}
	return nil, nil
}

func (m *mockApiKeyRepo) GetByKey(key string) (*model.ApiKey, error) {
	if k, ok := m.keys[key]; ok {
		return k, nil
	}
	return nil, nil
}

func (m *mockApiKeyRepo) Update(key *model.ApiKey) error {
	m.keys[key.ID] = key
	m.keys[key.Key] = key
	return nil
}

func (m *mockApiKeyRepo) Delete(id string) error {
	if k, ok := m.keys[id]; ok {
		delete(m.keys, k.Key)
	}
	delete(m.keys, id)
	return nil
}

func (m *mockApiKeyRepo) List() ([]model.ApiKey, error) {
	var list []model.ApiKey
	seen := make(map[string]bool)
	for _, k := range m.keys {
		if !seen[k.ID] {
			list = append(list, *k)
			seen[k.ID] = true
		}
	}
	return list, nil
}

func (m *mockApiKeyRepo) IncrementUsage(id string) error {
	return nil
}

func TestCreateKey_Success(t *testing.T) {
	repo := newMockApiKeyRepo()
	svc := NewApiKeyService(repo)

	apiKey, rawKey, err := svc.CreateKey("Test Key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.HasPrefix(rawKey, "sk-") {
		t.Fatalf("expected key to start with 'sk-', got '%s'", rawKey)
	}
	if apiKey.Name != "Test Key" {
		t.Fatalf("expected name 'Test Key', got '%s'", apiKey.Name)
	}
	if !apiKey.IsActive {
		t.Fatal("expected key to be active")
	}
}

func TestCreateKey_EmptyName(t *testing.T) {
	repo := newMockApiKeyRepo()
	svc := NewApiKeyService(repo)

	_, _, err := svc.CreateKey("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestValidateKey_Success(t *testing.T) {
	repo := newMockApiKeyRepo()
	svc := NewApiKeyService(repo)

	apiKey, rawKey, _ := svc.CreateKey("Test")
	// Store by raw key for lookup
	repo.keys[rawKey] = apiKey

	validated, err := svc.ValidateKey(rawKey)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if validated.ID != apiKey.ID {
		t.Fatal("expected same key ID")
	}
}

func TestValidateKey_InvalidKey(t *testing.T) {
	repo := newMockApiKeyRepo()
	svc := NewApiKeyService(repo)

	_, err := svc.ValidateKey("sk-invalid")
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestValidateKey_EmptyKey(t *testing.T) {
	repo := newMockApiKeyRepo()
	svc := NewApiKeyService(repo)

	_, err := svc.ValidateKey("")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestValidateKey_InactiveKey(t *testing.T) {
	repo := newMockApiKeyRepo()
	svc := NewApiKeyService(repo)

	apiKey, rawKey, _ := svc.CreateKey("Test")
	apiKey.IsActive = false
	repo.keys[rawKey] = apiKey

	_, err := svc.ValidateKey(rawKey)
	if err == nil {
		t.Fatal("expected error for inactive key")
	}
}

func TestValidateKey_BannedKey(t *testing.T) {
	repo := newMockApiKeyRepo()
	svc := NewApiKeyService(repo)

	apiKey, rawKey, _ := svc.CreateKey("Test")
	apiKey.IsBanned = true
	repo.keys[rawKey] = apiKey

	_, err := svc.ValidateKey(rawKey)
	if err == nil {
		t.Fatal("expected error for banned key")
	}
}

func TestValidateKey_ExpiredKey(t *testing.T) {
	repo := newMockApiKeyRepo()
	svc := NewApiKeyService(repo)

	apiKey, rawKey, _ := svc.CreateKey("Test")
	past := time.Now().Add(-1 * time.Hour)
	apiKey.ExpiresAt = &past
	repo.keys[rawKey] = apiKey

	_, err := svc.ValidateKey(rawKey)
	if err == nil {
		t.Fatal("expected error for expired key")
	}
}

func TestDeleteKey_Success(t *testing.T) {
	repo := newMockApiKeyRepo()
	svc := NewApiKeyService(repo)

	apiKey, rawKey, _ := svc.CreateKey("Test")
	apiKey.ID = "key-1"
	repo.keys["key-1"] = apiKey
	repo.keys[rawKey] = apiKey

	err := svc.DeleteKey("key-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestListKeys(t *testing.T) {
	repo := newMockApiKeyRepo()
	svc := NewApiKeyService(repo)

	k1, _, _ := svc.CreateKey("Key1")
	k1.ID = "k1"
	repo.keys["k1"] = k1
	k2, _, _ := svc.CreateKey("Key2")
	k2.ID = "k2"
	repo.keys["k2"] = k2

	list, err := svc.ListKeys()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(list))
	}
}
