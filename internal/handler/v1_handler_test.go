package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omniroute-go/internal/middleware"
	"github.com/omniroute-go/internal/model"
	"github.com/omniroute-go/internal/repository"
)

// --- Mock implementations ---

type mockProviderRepoV1 struct {
	providers []model.Provider
}

func (m *mockProviderRepoV1) Create(p *model.Provider) error             { return nil }
func (m *mockProviderRepoV1) ListProviders() ([]model.Provider, error)   { return m.providers, nil }
func (m *mockProviderRepoV1) GetByID(id string) (*model.Provider, error) { return nil, nil }
func (m *mockProviderRepoV1) Update(p *model.Provider) error             { return nil }
func (m *mockProviderRepoV1) Delete(id string) error                     { return nil }

type mockComboRepoV1 struct{}

func (m *mockComboRepoV1) Create(c *model.Combo) error             { return nil }
func (m *mockComboRepoV1) GetByID(id string) (*model.Combo, error) { return nil, nil }
func (m *mockComboRepoV1) Update(c *model.Combo) error             { return nil }
func (m *mockComboRepoV1) Delete(id string) error                  { return nil }
func (m *mockComboRepoV1) List() ([]model.Combo, error)            { return nil, nil }

type mockProxyRepoV1 struct{}

func (m *mockProxyRepoV1) Create(p *model.Proxy) error             { return nil }
func (m *mockProxyRepoV1) GetByID(id string) (*model.Proxy, error) { return nil, nil }
func (m *mockProxyRepoV1) Update(p *model.Proxy) error             { return nil }
func (m *mockProxyRepoV1) Delete(id string) error                  { return nil }
func (m *mockProxyRepoV1) List() ([]model.Proxy, error)            { return nil, nil }
func (m *mockProxyRepoV1) CreateLog(l *model.ProxyLog) error       { return nil }
func (m *mockProxyRepoV1) ListLogs(proxyID string, limit int) ([]model.ProxyLog, error) {
	return nil, nil
}

type mockUsageSvcV1 struct{}

func (m *mockUsageSvcV1) RecordUsage(u *model.UsageHistory) error { return nil }
func (m *mockUsageSvcV1) ListUsage(page, pageSize int, provider string, from, to *time.Time) ([]model.UsageHistory, int64, error) {
	return nil, 0, nil
}
func (m *mockUsageSvcV1) GetUsageStats(from, to *time.Time) (*repository.UsageStats, error) {
	return nil, nil
}
func (m *mockUsageSvcV1) RecordCallLog(l *model.CallLog) error { return nil }
func (m *mockUsageSvcV1) ListCallLogs(page, pageSize int, provider string, isError *bool) ([]model.CallLog, int64, error) {
	return nil, 0, nil
}
func (m *mockUsageSvcV1) GetCallLogByID(id string) (*model.CallLog, error) { return nil, nil }

type mockRoutingSvcV1 struct{}

func (m *mockRoutingSvcV1) SelectTarget(c *model.Combo) (*model.ComboTarget, error) {
	return nil, nil
}

// --- Helper ---

func setupV1Handler() *V1Handler {
	return NewV1Handler(
		&mockProviderRepoV1{},
		&mockComboRepoV1{},
		&mockProxyRepoV1{},
		&mockUsageSvcV1{},
		&mockRoutingSvcV1{},
	)
}

func setupGinContextWithAPIKey(apiKey *model.ApiKey) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.APIKeyContextKey, apiKey)
	return c, w
}

func makeAPIKey(scopes string) *model.ApiKey {
	return &model.ApiKey{
		ID:       "test-key-1",
		Name:     "Test Key",
		IsActive: true,
		Scopes:   scopes,
	}
}

// --- Tests ---

func TestChatCompletions_MissingAPIKey(t *testing.T) {
	h := setupV1Handler()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	c.Request, _ = http.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body))

	h.ChatCompletions(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestChatCompletions_MissingModel(t *testing.T) {
	h := setupV1Handler()
	c, w := setupGinContextWithAPIKey(makeAPIKey(""))

	body, _ := json.Marshal(ChatCompletionRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	c.Request, _ = http.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body))

	h.ChatCompletions(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestChatCompletions_MissingMessages(t *testing.T) {
	h := setupV1Handler()
	c, w := setupGinContextWithAPIKey(makeAPIKey(""))

	body, _ := json.Marshal(ChatCompletionRequest{Model: "gpt-4"})
	c.Request, _ = http.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body))

	h.ChatCompletions(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestChatCompletions_NoChatScope(t *testing.T) {
	h := setupV1Handler()
	c, w := setupGinContextWithAPIKey(makeAPIKey("embeddings"))

	body, _ := json.Marshal(ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	c.Request, _ = http.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body))

	h.ChatCompletions(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestCompletions_NoCompletionsScope(t *testing.T) {
	h := setupV1Handler()
	c, w := setupGinContextWithAPIKey(makeAPIKey("chat"))

	body, _ := json.Marshal(CompletionsRequest{Model: "gpt-3.5-turbo-instruct", Prompt: "Hello"})
	c.Request, _ = http.NewRequest("POST", "/v1/completions", bytes.NewReader(body))

	h.Completions(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestEmbeddings_NoEmbeddingsScope(t *testing.T) {
	h := setupV1Handler()
	c, w := setupGinContextWithAPIKey(makeAPIKey("chat"))

	body, _ := json.Marshal(EmbeddingsRequest{Model: "text-embedding-ada-002", Input: "hello"})
	c.Request, _ = http.NewRequest("POST", "/v1/embeddings", bytes.NewReader(body))

	h.Embeddings(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestEmbeddings_MissingModel(t *testing.T) {
	h := setupV1Handler()
	c, w := setupGinContextWithAPIKey(makeAPIKey(""))

	body, _ := json.Marshal(EmbeddingsRequest{Input: "hello"})
	c.Request, _ = http.NewRequest("POST", "/v1/embeddings", bytes.NewReader(body))

	h.Embeddings(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListModels_UnrestrictedScope(t *testing.T) {
	h := setupV1Handler()
	c, w := setupGinContextWithAPIKey(makeAPIKey("")) // empty = all scopes

	c.Request, _ = http.NewRequest("GET", "/v1/models", nil)

	h.ListModels(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestListModels_NoModelsScope(t *testing.T) {
	h := setupV1Handler()
	c, w := setupGinContextWithAPIKey(makeAPIKey("chat"))

	c.Request, _ = http.NewRequest("GET", "/v1/models", nil)

	h.ListModels(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestHasScope_Unrestricted(t *testing.T) {
	key := &model.ApiKey{Scopes: ""}
	if !key.HasScope("chat") {
		t.Fatal("empty scopes should allow all")
	}
}

func TestHasScope_SpecificScope(t *testing.T) {
	key := &model.ApiKey{Scopes: "chat,embeddings"}
	if !key.HasScope("chat") {
		t.Fatal("should have chat scope")
	}
	if !key.HasScope("embeddings") {
		t.Fatal("should have embeddings scope")
	}
	if key.HasScope("completions") {
		t.Fatal("should not have completions scope")
	}
}

func TestHasModelAccess_AllAllowed(t *testing.T) {
	key := &model.ApiKey{AllowedModels: "", BlockedModels: ""}
	if !key.HasModelAccess("gpt-4") {
		t.Fatal("empty allowed should allow all")
	}
}

func TestHasModelAccess_Whitelist(t *testing.T) {
	key := &model.ApiKey{AllowedModels: "gpt-4,gpt-3.5-turbo"}
	if !key.HasModelAccess("gpt-4") {
		t.Fatal("should allow gpt-4")
	}
	if key.HasModelAccess("claude-3") {
		t.Fatal("should not allow claude-3")
	}
}

func TestHasModelAccess_Blacklist(t *testing.T) {
	key := &model.ApiKey{BlockedModels: "gpt-4"}
	if key.HasModelAccess("gpt-4") {
		t.Fatal("should block gpt-4")
	}
	if !key.HasModelAccess("gpt-3.5-turbo") {
		t.Fatal("should allow gpt-3.5-turbo")
	}
}
