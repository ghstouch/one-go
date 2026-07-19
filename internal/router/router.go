package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ghstouch/one-go/internal/config"
	"github.com/ghstouch/one-go/internal/database"
	"github.com/ghstouch/one-go/internal/handler"
	"github.com/ghstouch/one-go/internal/middleware"
	"github.com/ghstouch/one-go/internal/repository"
	"github.com/ghstouch/one-go/internal/service"
)

// Setup initializes and returns the router
func Setup(cfg *config.Config) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create router
	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.CORSMiddleware(
		cfg.Security.CORSAllowedOrigins,
		cfg.Security.CORSAllowedMethods,
		cfg.Security.CORSAllowedHeaders,
	))
	r.Use(middleware.BodyLimitMiddleware(10 << 20)) // 10MB max body

	// Rate limiter
	var rateLimiter *middleware.RateLimiter
	if cfg.Security.RateLimitEnabled {
		rateLimiter = middleware.NewRateLimiter(
			cfg.Security.RateLimitPerMinute,
			time.Minute,
		)
	}

	// Get database instance
	db := database.Get()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	providerRepo := repository.NewProviderRepository(db)
	comboRepo := repository.NewComboRepository(db)
	apiKeyRepo := repository.NewApiKeyRepository(db)
	usageRepo := repository.NewUsageRepository(db)
	proxyRepo := repository.NewProxyRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, &cfg.JWT)
	settingsService := service.NewSettingsService(settingsRepo)
	providerService := service.NewProviderService(providerRepo)
	comboService := service.NewComboService(comboRepo)
	apiKeyService := service.NewApiKeyService(apiKeyRepo)
	usageService := service.NewUsageService(usageRepo)
	proxyService := service.NewProxyService(proxyRepo)
	routingService := service.NewRoutingService()
	quotaService := service.NewQuotaService(usageRepo)
	modelAliasRepo := repository.NewModelAliasRepository(db)
	modelAliasService := service.NewModelAliasService(modelAliasRepo)
	auditLogRepo := repository.NewAuditLogRepository(db)
	auditLogService := service.NewAuditLogService(auditLogRepo)
	webhookRepo := repository.NewWebhookRepository(db)
	webhookService := service.NewWebhookService(webhookRepo)

	// Initialize handlers
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
	modelAliasHandler := handler.NewModelAliasHandler(modelAliasService)
	auditLogHandler := handler.NewAuditLogHandler(auditLogService)
	webhookHandler := handler.NewWebhookHandler(webhookService)

	// Static files (for frontend)
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/**/*")

	// Health check (public)
	r.GET("/api/health", dashboardHandler.HealthCheck)

	// Auth routes (public)
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
	}

	// Protected API routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(authService))
	if rateLimiter != nil {
		api.Use(middleware.RateLimitMiddleware(rateLimiter))
	}
	api.Use(middleware.AuditMiddleware(auditLogService))
	{
		// Auth
		api.POST("/auth/refresh", authHandler.RefreshToken)
		api.GET("/auth/me", authHandler.Me)
		api.POST("/auth/logout", authHandler.Logout)

		// Dashboard
		api.GET("/dashboard/stats", dashboardHandler.GetDashboardStats)
		api.GET("/dashboard/activity", dashboardHandler.GetRecentActivity)

		// Settings
		api.GET("/settings", settingsHandler.GetSettings)
		api.PUT("/settings", middleware.RequireRole("admin"), settingsHandler.UpdateSetting)

		// Providers
		api.GET("/providers", middleware.RequireRole("admin"), providerHandler.GetProviders)
		api.POST("/providers", middleware.RequireRole("admin"), providerHandler.CreateProvider)
		api.PUT("/providers/:id", middleware.RequireRole("admin"), providerHandler.UpdateProvider)
		api.DELETE("/providers/:id", middleware.RequireRole("admin"), providerHandler.DeleteProvider)
		api.POST("/providers/:id/test", middleware.RequireRole("admin"), providerHandler.TestProvider)
		api.POST("/providers/validate", middleware.RequireRole("admin"), providerHandler.ValidateAPIKey)
		api.POST("/providers/batch-delete", middleware.RequireRole("admin"), providerHandler.BatchDeleteProviders)

		// Combos
		api.GET("/combos", middleware.RequireRole("admin"), comboHandler.GetCombos)
		api.POST("/combos", middleware.RequireRole("admin"), comboHandler.CreateCombo)
		api.GET("/combos/:id", middleware.RequireRole("admin"), comboHandler.GetCombo)
		api.PUT("/combos/:id", middleware.RequireRole("admin"), comboHandler.UpdateCombo)
		api.DELETE("/combos/:id", middleware.RequireRole("admin"), comboHandler.DeleteCombo)

		// API Keys
		api.GET("/keys", middleware.RequireRole("admin"), apiKeyHandler.GetKeys)
		api.POST("/keys", middleware.RequireRole("admin"), apiKeyHandler.CreateKey)
		api.GET("/keys/:id", middleware.RequireRole("admin"), apiKeyHandler.GetKey)
		api.PUT("/keys/:id", middleware.RequireRole("admin"), apiKeyHandler.UpdateKey)
		api.DELETE("/keys/:id", middleware.RequireRole("admin"), apiKeyHandler.DeleteKey)

		// Usage
		api.GET("/usage", usageHandler.GetUsage)
		api.GET("/usage/stats", usageHandler.GetUsageStats)
		api.GET("/usage/logs", usageHandler.GetCallLogs)
		api.GET("/usage/logs/:id", usageHandler.GetCallLogDetail)
		api.GET("/usage/export", usageHandler.ExportUsageCSV)

		// Quota
		api.GET("/quota", quotaHandler.GetQuotas)
		api.GET("/quota/:provider", quotaHandler.GetProviderQuota)

		// Model Aliases
		api.GET("/models/aliases", modelAliasHandler.List)
		api.POST("/models/aliases", middleware.RequireRole("admin"), modelAliasHandler.Create)
		api.DELETE("/models/aliases/:id", middleware.RequireRole("admin"), modelAliasHandler.Delete)
		api.GET("/models/resolve", modelAliasHandler.Resolve)

		// Audit Logs
		api.GET("/audit-logs", middleware.RequireRole("admin"), auditLogHandler.List)
		api.GET("/audit-logs/:id", middleware.RequireRole("admin"), auditLogHandler.GetByID)

		// Webhooks
		api.GET("/webhooks", middleware.RequireRole("admin"), webhookHandler.List)
		api.POST("/webhooks", middleware.RequireRole("admin"), webhookHandler.Create)
		api.PUT("/webhooks/:id", middleware.RequireRole("admin"), webhookHandler.Update)
		api.DELETE("/webhooks/:id", middleware.RequireRole("admin"), webhookHandler.Delete)
		api.GET("/webhooks/logs", middleware.RequireRole("admin"), webhookHandler.GetLogs)

		// Proxy
		api.GET("/proxy", middleware.RequireRole("admin"), proxyHandler.GetProxies)
		api.POST("/proxy", middleware.RequireRole("admin"), proxyHandler.CreateProxy)
		api.PUT("/proxy/:id", middleware.RequireRole("admin"), proxyHandler.UpdateProxy)
		api.DELETE("/proxy/:id", middleware.RequireRole("admin"), proxyHandler.DeleteProxy)
		api.POST("/proxy/:id/test", middleware.RequireRole("admin"), proxyHandler.TestProxy)
		api.GET("/proxy/logs", middleware.RequireRole("admin"), proxyHandler.GetProxyLogs)
	}

	// OpenAI-compatible v1 API (uses API key auth)
	v1 := r.Group("/v1")
	apiKeyRateLimiter := middleware.NewAPIKeyRateLimiter()
	v1.Use(middleware.APIKeyMiddleware(apiKeyService))
	v1.Use(middleware.APIKeyRateLimitMiddleware(apiKeyRateLimiter))
	{
		v1.POST("/chat/completions", v1Handler.ChatCompletions)
		v1.POST("/completions", v1Handler.Completions)
		v1.POST("/embeddings", v1Handler.Embeddings)
		v1.GET("/models", v1Handler.ListModels)
	}

	// Web routes (HTML pages)
	web := r.Group("/")
	web.Use(middleware.OptionalAuthMiddleware(authService))
	{
		web.GET("/", func(c *gin.Context) {
			// Check if user is authenticated
			userID := middleware.GetUserIDFromContext(c)
			if userID == "" {
				c.Redirect(302, "/login")
				return
			}
			c.Redirect(302, "/dashboard")
		})

		web.GET("/login", func(c *gin.Context) {
			c.HTML(200, "login.html", gin.H{
				"title": "Login - One",
			})
		})

		// Protected web routes
		webProtected := web.Group("/")
		webProtected.Use(middleware.AuthMiddleware(authService))
		{
			webProtected.GET("/dashboard", func(c *gin.Context) {
				username := middleware.GetUsernameFromContext(c)
				c.HTML(200, "dashboard.html", gin.H{
					"title":    "Dashboard - One",
					"username": username,
					"active":   "dashboard",
				})
			})

			webProtected.GET("/providers", func(c *gin.Context) {
				username := middleware.GetUsernameFromContext(c)
				c.HTML(200, "providers.html", gin.H{
					"title":    "Providers - One",
					"username": username,
					"active":   "providers",
				})
			})

			webProtected.GET("/combos", func(c *gin.Context) {
				username := middleware.GetUsernameFromContext(c)
				c.HTML(200, "combos.html", gin.H{
					"title":    "Combos - One",
					"username": username,
					"active":   "combos",
				})
			})

			webProtected.GET("/usage", func(c *gin.Context) {
				username := middleware.GetUsernameFromContext(c)
				c.HTML(200, "usage.html", gin.H{
					"title":    "Usage - One",
					"username": username,
					"active":   "usage",
				})
			})

			webProtected.GET("/logs", func(c *gin.Context) {
				username := middleware.GetUsernameFromContext(c)
				c.HTML(200, "logs.html", gin.H{
					"title":    "Logs - One",
					"username": username,
					"active":   "logs",
				})
			})

			webProtected.GET("/api-keys", func(c *gin.Context) {
				username := middleware.GetUsernameFromContext(c)
				c.HTML(200, "apikeys.html", gin.H{
					"title":    "API Keys - One",
					"username": username,
					"active":   "api-keys",
				})
			})

			webProtected.GET("/proxy", func(c *gin.Context) {
				username := middleware.GetUsernameFromContext(c)
				c.HTML(200, "proxy.html", gin.H{
					"title":    "Proxy - One",
					"username": username,
					"active":   "proxy",
				})
			})

			webProtected.GET("/settings-page", func(c *gin.Context) {
				username := middleware.GetUsernameFromContext(c)
				c.HTML(200, "settings.html", gin.H{
					"title":    "Settings - One",
					"username": username,
					"active":   "settings",
				})
			})
		}
	}

	return r
}
