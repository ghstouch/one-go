package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/ghstouch/one-go/internal/config"
	"github.com/ghstouch/one-go/internal/handler"
	"github.com/ghstouch/one-go/internal/middleware"
	"github.com/ghstouch/one-go/internal/model"
	"github.com/ghstouch/one-go/internal/repository"
	"github.com/ghstouch/one-go/internal/service"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	db.AutoMigrate(
		&model.User{},
		&model.Settings{},
		&model.Provider{},
		&model.Combo{},
		&model.ComboTarget{},
		&model.ApiKey{},
		&model.UsageHistory{},
		&model.CallLog{},
		&model.Proxy{},
		&model.ProxyLog{},
	)
	return db
}

func setupTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	userRepo := repository.NewUserRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	providerRepo := repository.NewProviderRepository(db)
	comboRepo := repository.NewComboRepository(db)
	apiKeyRepo := repository.NewApiKeyRepository(db)
	usageRepo := repository.NewUsageRepository(db)
	proxyRepo := repository.NewProxyRepository(db)

	jwtCfg := &config.JWTConfig{Secret: "test-secret", ExpiryHours: 24, Issuer: "test"}
	authService := service.NewAuthService(userRepo, jwtCfg)
	settingsService := service.NewSettingsService(settingsRepo)
	providerService := service.NewProviderService(providerRepo)
	comboService := service.NewComboService(comboRepo)
	apiKeyService := service.NewApiKeyService(apiKeyRepo)
	usageService := service.NewUsageService(usageRepo)
	proxyService := service.NewProxyService(proxyRepo)
	routingService := service.NewRoutingService()
	quotaService := service.NewQuotaService(usageRepo)

	authHandler := handler.NewAuthHandler(authService)
	dashboardHandler := handler.NewDashboardHandler(providerRepo)
	settingsHandler := handler.NewSettingsHandler(settingsService)
	providerHandler := handler.NewProviderHandler(providerService)
	comboHandler := handler.NewComboHandler(comboService)
	apiKeyHandler := handler.NewApiKeyHandler(apiKeyService)
	usageHandler := handler.NewUsageHandler(usageService)
	proxyHandler := handler.NewProxyHandler(proxyService)
	v1Handler := handler.NewV1Handler(providerRepo, comboRepo, proxyRepo, usageService, routingService)
	quotaHandler := handler.NewQuotaHandler(quotaService)

	r := gin.New()
	r.Use(gin.Recovery())

	// Public
	r.POST("/api/auth/login", authHandler.Login)
	r.GET("/api/health", dashboardHandler.HealthCheck)

	// Protected
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(authService))
	{
		api.GET("/auth/me", authHandler.Me)
		api.GET("/dashboard/stats", dashboardHandler.GetDashboardStats)
		api.GET("/settings", settingsHandler.GetSettings)
		api.PUT("/settings", settingsHandler.UpdateSetting)
		api.GET("/providers", providerHandler.GetProviders)
		api.POST("/providers", providerHandler.CreateProvider)
		api.PUT("/providers/:id", providerHandler.UpdateProvider)
		api.DELETE("/providers/:id", providerHandler.DeleteProvider)
		api.POST("/providers/:id/test", providerHandler.TestProvider)
		api.POST("/providers/validate", providerHandler.ValidateAPIKey)
		api.GET("/combos", comboHandler.GetCombos)
		api.POST("/combos", comboHandler.CreateCombo)
		api.GET("/combos/:id", comboHandler.GetCombo)
		api.PUT("/combos/:id", comboHandler.UpdateCombo)
		api.DELETE("/combos/:id", comboHandler.DeleteCombo)
		api.GET("/keys", apiKeyHandler.GetKeys)
		api.POST("/keys", apiKeyHandler.CreateKey)
		api.PUT("/keys/:id", apiKeyHandler.UpdateKey)
		api.DELETE("/keys/:id", apiKeyHandler.DeleteKey)
		api.GET("/usage", usageHandler.GetUsage)
		api.GET("/usage/stats", usageHandler.GetUsageStats)
		api.GET("/usage/logs", usageHandler.GetCallLogs)
		api.GET("/quota", quotaHandler.GetQuotas)
		api.GET("/proxy", proxyHandler.GetProxies)
		api.POST("/proxy", proxyHandler.CreateProxy)
		api.DELETE("/proxy/:id", proxyHandler.DeleteProxy)
	}

	// v1 API (API key auth)
	v1 := r.Group("/v1")
	v1.Use(middleware.APIKeyMiddleware(apiKeyService))
	v1.Use(middleware.APIKeyRateLimitMiddleware(middleware.NewAPIKeyRateLimiter()))
	{
		v1.POST("/chat/completions", v1Handler.ChatCompletions)
		v1.POST("/completions", v1Handler.Completions)
		v1.POST("/embeddings", v1Handler.Embeddings)
		v1.GET("/models", v1Handler.ListModels)
	}

	return r
}

func seedTestUser(db *gorm.DB) {
	hash, _ := service.HashPasswordForTest("admin123")
	db.Create(&model.User{
		Username: "admin",
		Password: hash,
		Role:     "admin",
		IsActive: true,
	})
}

func loginAndGetToken(t *testing.T, r *gin.Engine) string {
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	return data["token"].(string)
}

func doRequest(r *gin.Engine, method, path string, token string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = bytes.NewReader([]byte{})
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	r.ServeHTTP(w, req)
	return w
}

// --- Tests ---

func TestIntegration_HealthCheck(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestRouter(db)

	w := doRequest(r, "GET", "/api/health", "", nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestIntegration_LoginFlow(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestRouter(db)
	seedTestUser(db)

	token := loginAndGetToken(t, r)
	if token == "" {
		t.Fatal("expected token")
	}

	// Access protected route
	w := doRequest(r, "GET", "/api/auth/me", token, nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestIntegration_FullProviderFlow(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestRouter(db)
	seedTestUser(db)
	token := loginAndGetToken(t, r)

	// Create provider
	w := doRequest(r, "POST", "/api/providers", token, map[string]interface{}{
		"name":    "Test OpenAI",
		"type":    "openai",
		"baseUrl": "https://api.openai.com/v1",
		"status":  "active",
	})
	if w.Code != 201 {
		t.Fatalf("create provider: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var providerResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &providerResp)
	providerData := providerResp["data"].(map[string]interface{})
	providerID := providerData["id"].(string)

	// List providers
	w = doRequest(r, "GET", "/api/providers", token, nil)
	if w.Code != 200 {
		t.Fatalf("list providers: expected 200, got %d", w.Code)
	}

	// Update provider
	w = doRequest(r, "PUT", "/api/providers/"+providerID, token, map[string]interface{}{
		"name":    "Updated OpenAI",
		"type":    "openai",
		"baseUrl": "https://api.openai.com/v1",
		"status":  "active",
	})
	if w.Code != 200 {
		t.Fatalf("update provider: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Delete provider
	w = doRequest(r, "DELETE", "/api/providers/"+providerID, token, nil)
	if w.Code != 200 {
		t.Fatalf("delete provider: expected 200, got %d", w.Code)
	}
}

func TestIntegration_ComboFlow(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestRouter(db)
	seedTestUser(db)
	token := loginAndGetToken(t, r)

	// Create combo
	w := doRequest(r, "POST", "/api/combos", token, map[string]interface{}{
		"name":     "Test Combo",
		"strategy": "priority",
	})
	if w.Code != 201 {
		t.Fatalf("create combo: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// List combos
	w = doRequest(r, "GET", "/api/combos", token, nil)
	if w.Code != 200 {
		t.Fatalf("list combos: expected 200, got %d", w.Code)
	}
}

func TestIntegration_APIKeyFlow(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestRouter(db)
	seedTestUser(db)
	token := loginAndGetToken(t, r)

	// Create API key
	w := doRequest(r, "POST", "/api/keys", token, map[string]interface{}{
		"name": "Test Key",
	})
	if w.Code != 201 {
		t.Fatalf("create key: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var keyResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &keyResp)
	keyData := keyResp["data"].(map[string]interface{})
	rawKey := keyData["key"].(string)

	if rawKey == "" {
		t.Fatal("expected raw key in response")
	}

	// List keys
	w = doRequest(r, "GET", "/api/keys", token, nil)
	if w.Code != 200 {
		t.Fatalf("list keys: expected 200, got %d", w.Code)
	}

	// Use API key to access v1/models
	w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+rawKey)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("v1/models with API key: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestIntegration_V1WithoutKey(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestRouter(db)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/models", nil)
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestIntegration_UnauthorizedAccess(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestRouter(db)

	w := doRequest(r, "GET", "/api/providers", "", nil)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestIntegration_ProxyFlow(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestRouter(db)
	seedTestUser(db)
	token := loginAndGetToken(t, r)

	// Create proxy
	w := doRequest(r, "POST", "/api/proxy", token, map[string]interface{}{
		"name": "Test Proxy",
		"type": "http",
		"host": "127.0.0.1",
		"port": 8080,
	})
	if w.Code != 201 {
		t.Fatalf("create proxy: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// List proxies
	w = doRequest(r, "GET", "/api/proxy", token, nil)
	if w.Code != 200 {
		t.Fatalf("list proxies: expected 200, got %d", w.Code)
	}
}

func TestIntegration_UsageStats(t *testing.T) {
	db := setupTestDB(t)
	r := setupTestRouter(db)
	seedTestUser(db)
	token := loginAndGetToken(t, r)

	// Get usage stats (empty)
	w := doRequest(r, "GET", "/api/usage/stats", token, nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Get quota
	w = doRequest(r, "GET", "/api/quota", token, nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
