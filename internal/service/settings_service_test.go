package service

import (
	"testing"

	"github.com/ghstouch/one-go/internal/model"
)

// mockSettingsRepo implements repository.SettingsRepository for testing
type mockSettingsRepo struct {
	settings map[string]*model.Settings
}

func newMockSettingsRepo() *mockSettingsRepo {
	return &mockSettingsRepo{settings: make(map[string]*model.Settings)}
}

func (m *mockSettingsRepo) Get(key string) (*model.Settings, error) {
	if s, ok := m.settings[key]; ok {
		return s, nil
	}
	return nil, nil
}

func (m *mockSettingsRepo) Set(key, value, settingType string) error {
	m.settings[key] = &model.Settings{Key: key, Value: value, Type: settingType}
	return nil
}

func (m *mockSettingsRepo) GetAll() ([]model.Settings, error) {
	var list []model.Settings
	for _, s := range m.settings {
		list = append(list, *s)
	}
	return list, nil
}

func (m *mockSettingsRepo) Delete(key string) error {
	delete(m.settings, key)
	return nil
}

func (m *mockSettingsRepo) GetMultiple(keys []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, key := range keys {
		if s, ok := m.settings[key]; ok {
			result[key] = s.Value
		}
	}
	return result, nil
}

func TestSettingsGet(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := NewSettingsService(repo)

	repo.settings["test_key"] = &model.Settings{Key: "test_key", Value: "test_value"}

	val, err := svc.Get("test_key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "test_value" {
		t.Fatalf("expected 'test_value', got '%s'", val)
	}
}

func TestSettingsGet_NotFound(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := NewSettingsService(repo)

	val, err := svc.Get("nonexistent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got '%s'", val)
	}
}

func TestSettingsSet(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := NewSettingsService(repo)

	err := svc.Set("key1", "value1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.settings["key1"].Value != "value1" {
		t.Fatal("expected value to be set")
	}
}

func TestSettingsGetInt(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := NewSettingsService(repo)

	repo.settings["port"] = &model.Settings{Key: "port", Value: "8080"}

	val := svc.GetInt("port", 3000)
	if val != 8080 {
		t.Fatalf("expected 8080, got %d", val)
	}
}

func TestSettingsGetInt_Default(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := NewSettingsService(repo)

	val := svc.GetInt("nonexistent", 3000)
	if val != 3000 {
		t.Fatalf("expected default 3000, got %d", val)
	}
}

func TestSettingsGetInt_InvalidValue(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := NewSettingsService(repo)

	repo.settings["port"] = &model.Settings{Key: "port", Value: "not_a_number"}

	val := svc.GetInt("port", 3000)
	if val != 3000 {
		t.Fatalf("expected default 3000, got %d", val)
	}
}

func TestSettingsGetBool(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := NewSettingsService(repo)

	repo.settings["enabled"] = &model.Settings{Key: "enabled", Value: "true"}
	repo.settings["disabled"] = &model.Settings{Key: "disabled", Value: "false"}

	if !svc.GetBool("enabled", false) {
		t.Fatal("expected true")
	}
	if svc.GetBool("disabled", true) {
		t.Fatal("expected false")
	}
}

func TestSettingsGetBool_Default(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := NewSettingsService(repo)

	if !svc.GetBool("nonexistent", true) {
		t.Fatal("expected default true")
	}
	if svc.GetBool("nonexistent", false) {
		t.Fatal("expected default false")
	}
}

func TestSettingsSetInt(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := NewSettingsService(repo)

	err := svc.SetInt("port", 9090)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.settings["port"].Value != "9090" {
		t.Fatalf("expected '9090', got '%s'", repo.settings["port"].Value)
	}
}

func TestSettingsSetBool(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := NewSettingsService(repo)

	svc.SetBool("enabled", true)
	if repo.settings["enabled"].Value != "true" {
		t.Fatalf("expected 'true', got '%s'", repo.settings["enabled"].Value)
	}

	svc.SetBool("disabled", false)
	if repo.settings["disabled"].Value != "false" {
		t.Fatalf("expected 'false', got '%s'", repo.settings["disabled"].Value)
	}
}

func TestSettingsGetAll(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := NewSettingsService(repo)

	repo.settings["a"] = &model.Settings{Key: "a", Value: "1"}
	repo.settings["b"] = &model.Settings{Key: "b", Value: "2"}

	result, err := svc.GetAll()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 settings, got %d", len(result))
	}
	if result["a"] != "1" {
		t.Fatalf("expected '1', got '%s'", result["a"])
	}
}

func TestSettingsDelete(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := NewSettingsService(repo)

	repo.settings["key"] = &model.Settings{Key: "key", Value: "val"}
	err := svc.Delete("key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := repo.settings["key"]; exists {
		t.Fatal("expected setting to be deleted")
	}
}
