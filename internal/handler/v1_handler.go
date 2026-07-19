package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omniroute-go/internal/middleware"
	"github.com/omniroute-go/internal/model"
	"github.com/omniroute-go/internal/repository"
	"github.com/omniroute-go/internal/service"
	"github.com/omniroute-go/pkg/response"
)

// V1Handler handles OpenAI-compatible v1 API requests
type V1Handler struct {
	providerRepo repository.ProviderRepository
	comboRepo    repository.ComboRepository
	proxyRepo    repository.ProxyRepository
	usageSvc     service.UsageService
	routingSvc   service.RoutingService
}

func NewV1Handler(providerRepo repository.ProviderRepository, comboRepo repository.ComboRepository, proxyRepo repository.ProxyRepository, usageSvc service.UsageService, routingSvc service.RoutingService) *V1Handler {
	return &V1Handler{providerRepo: providerRepo, comboRepo: comboRepo, proxyRepo: proxyRepo, usageSvc: usageSvc, routingSvc: routingSvc}
}

// ChatCompletionRequest represents an OpenAI-compatible chat completion request
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Temperature *float64  `json:"temperature,omitempty"`
	MaxTokens   *int      `json:"max_tokens,omitempty"`
	TopP        *float64  `json:"top_p,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletions handles POST /v1/chat/completions
func (h *V1Handler) ChatCompletions(c *gin.Context) {
	apiKey := h.getAPIKeyFromContext(c)
	if apiKey == nil {
		response.Unauthorized(c, "API key not found in context")
		return
	}

	var req ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if req.Model == "" {
		response.BadRequest(c, "model is required")
		return
	}

	if !apiKey.HasScope(model.ScopeChat) {
		response.Forbidden(c, "API key does not have 'chat' scope")
		return
	}
	if !apiKey.HasModelAccess(req.Model) {
		response.Forbidden(c, "API key does not have access to model: "+req.Model)
		return
	}

	if len(req.Messages) == 0 {
		response.BadRequest(c, "messages is required")
		return
	}

	// Find an active provider
	providers, err := h.providerRepo.ListProviders()
	if err != nil || len(providers) == 0 {
		response.InternalServerError(c, "No providers configured")
		return
	}

	var provider *model.Provider
	for i, p := range providers {
		if p.Status == "active" && p.Type == "openai" {
			provider = &providers[i]
			break
		}
	}
	if provider == nil {
		// Fallback to any active provider
		for i, p := range providers {
			if p.Status == "active" {
				provider = &providers[i]
				break
			}
		}
	}
	if provider == nil {
		response.InternalServerError(c, "No active providers available")
		return
	}

	// Forward to provider
	start := time.Now()

	if req.Stream {
		// SSE streaming mode
		h.streamResponse(c, provider, &req)
		latency := int(time.Since(start).Milliseconds())
		h.usageSvc.RecordUsage(&model.UsageHistory{
			Provider: provider.Name, Model: req.Model, Status: "success", Success: true, LatencyMs: latency,
		})
		return
	}

	respBody, statusCode, err := h.forwardToProvider(provider, &req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		// Log error usage
		h.usageSvc.RecordUsage(&model.UsageHistory{
			Provider:     provider.Name,
			Model:        req.Model,
			Status:       "error",
			Success:      false,
			LatencyMs:    latency,
			ErrorCode:    fmt.Sprintf("%d", statusCode),
			ErrorMessage: err.Error(),
		})
		response.InternalServerError(c, "Provider error: "+err.Error())
		return
	}

	// Log successful usage
	h.usageSvc.RecordUsage(&model.UsageHistory{
		Provider:  provider.Name,
		Model:     req.Model,
		Status:    "success",
		Success:   true,
		LatencyMs: latency,
	})

	// Forward response
	c.Data(statusCode, "application/json", respBody)
}

func (h *V1Handler) forwardToProvider(provider *model.Provider, req *ChatCompletionRequest) ([]byte, int, error) {
	return h.forwardGeneric(provider, "/chat/completions", req)
}

// ListModels handles GET /v1/models
func (h *V1Handler) ListModels(c *gin.Context) {
	apiKey := h.getAPIKeyFromContext(c)
	if apiKey != nil && !apiKey.HasScope(model.ScopeModels) {
		response.Forbidden(c, "API key does not have 'models' scope")
		return
	}

	providers, err := h.providerRepo.ListProviders()
	if err != nil {
		response.InternalServerError(c, "Failed to fetch providers")
		return
	}

	type ModelInfo struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		OwnedBy string `json:"owned_by"`
	}

	var models []ModelInfo
	for _, p := range providers {
		if p.Status == "active" {
			for _, ep := range p.EndpointsJSON {
				models = append(models, ModelInfo{
					ID:      fmt.Sprintf("%s/%s", strings.ToLower(p.Name), ep),
					Object:  "model",
					OwnedBy: strings.ToLower(p.Name),
				})
			}
			// Always add a default model entry
			models = append(models, ModelInfo{
				ID:      fmt.Sprintf("%s/default", strings.ToLower(p.Name)),
				Object:  "model",
				OwnedBy: strings.ToLower(p.Name),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   models,
	})
}

// CompletionsRequest represents an OpenAI-compatible completions request
type CompletionsRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	Stream      bool     `json:"stream"`
	Temperature *float64 `json:"temperature,omitempty"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`
}

// Completions handles POST /v1/completions
func (h *V1Handler) Completions(c *gin.Context) {
	apiKey := h.getAPIKeyFromContext(c)
	if apiKey == nil {
		response.Unauthorized(c, "API key not found in context")
		return
	}

	var req CompletionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.Model == "" {
		response.BadRequest(c, "model is required")
		return
	}
	if !apiKey.HasScope(model.ScopeCompletions) {
		response.Forbidden(c, "API key does not have 'completions' scope")
		return
	}
	if !apiKey.HasModelAccess(req.Model) {
		response.Forbidden(c, "API key does not have access to model: "+req.Model)
		return
	}

	provider, err := h.findActiveProvider()
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	start := time.Now()
	respBody, statusCode, fwdErr := h.forwardGeneric(provider, "/completions", req)
	latency := int(time.Since(start).Milliseconds())

	if fwdErr != nil {
		h.usageSvc.RecordUsage(&model.UsageHistory{Provider: provider.Name, Model: req.Model, Status: "error", Success: false, LatencyMs: latency, ErrorMessage: fwdErr.Error()})
		response.InternalServerError(c, "Provider error: "+fwdErr.Error())
		return
	}
	h.usageSvc.RecordUsage(&model.UsageHistory{Provider: provider.Name, Model: req.Model, Status: "success", Success: true, LatencyMs: latency})
	c.Data(statusCode, "application/json", respBody)
}

// EmbeddingsRequest represents an OpenAI-compatible embeddings request
type EmbeddingsRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// Embeddings handles POST /v1/embeddings
func (h *V1Handler) Embeddings(c *gin.Context) {
	apiKey := h.getAPIKeyFromContext(c)
	if apiKey == nil {
		response.Unauthorized(c, "API key not found in context")
		return
	}

	var req EmbeddingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.Model == "" {
		response.BadRequest(c, "model is required")
		return
	}
	if !apiKey.HasScope(model.ScopeEmbeddings) {
		response.Forbidden(c, "API key does not have 'embeddings' scope")
		return
	}
	if !apiKey.HasModelAccess(req.Model) {
		response.Forbidden(c, "API key does not have access to model: "+req.Model)
		return
	}

	provider, err := h.findActiveProvider()
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	start := time.Now()
	respBody, statusCode, fwdErr := h.forwardGeneric(provider, "/embeddings", req)
	latency := int(time.Since(start).Milliseconds())

	if fwdErr != nil {
		h.usageSvc.RecordUsage(&model.UsageHistory{Provider: provider.Name, Model: req.Model, Status: "error", Success: false, LatencyMs: latency, ErrorMessage: fwdErr.Error()})
		response.InternalServerError(c, "Provider error: "+fwdErr.Error())
		return
	}
	h.usageSvc.RecordUsage(&model.UsageHistory{Provider: provider.Name, Model: req.Model, Status: "success", Success: true, LatencyMs: latency})
	c.Data(statusCode, "application/json", respBody)
}

// findActiveProvider returns the first active provider
func (h *V1Handler) findActiveProvider() (*model.Provider, error) {
	providers, err := h.providerRepo.ListProviders()
	if err != nil || len(providers) == 0 {
		return nil, fmt.Errorf("no providers configured")
	}
	for i, p := range providers {
		if p.Status == "active" {
			return &providers[i], nil
		}
	}
	return nil, fmt.Errorf("no active providers available")
}

// forwardGeneric forwards a request to a provider endpoint
func (h *V1Handler) forwardGeneric(provider *model.Provider, path string, reqBody interface{}) ([]byte, int, error) {
	url := strings.TrimRight(provider.BaseURL, "/") + path

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if provider.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}

	client := h.createHTTPClient(provider.ProxyID)
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}
	return respBody, resp.StatusCode, nil
}

// streamResponse forwards a streaming chat completion request to the provider
func (h *V1Handler) streamResponse(c *gin.Context, provider *model.Provider, req *ChatCompletionRequest) {
	url := strings.TrimRight(provider.BaseURL, "/") + "/chat/completions"

	body, err := json.Marshal(req)
	if err != nil {
		response.InternalServerError(c, "Failed to marshal request")
		return
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		response.InternalServerError(c, "Failed to create request")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if provider.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}

	client := h.createHTTPClient(provider.ProxyID)
	resp, err := client.Do(httpReq)
	if err != nil {
		response.InternalServerError(c, "Provider error: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, "application/json", respBody)
		return
	}

	// Stream SSE to client
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	buf := make([]byte, 4096)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			c.Writer.Write(buf[:n])
			c.Writer.Flush()
		}
		if readErr != nil {
			break
		}
	}
}

// createHTTPClient creates an HTTP client, optionally using a proxy
func (h *V1Handler) createHTTPClient(proxyID string) *http.Client {
	if proxyID == "" || h.proxyRepo == nil {
		return &http.Client{Timeout: 120 * time.Second}
	}

	proxy, err := h.proxyRepo.GetByID(proxyID)
	if err != nil || !proxy.IsActive {
		return &http.Client{Timeout: 120 * time.Second}
	}

	proxyURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", proxy.Host, proxy.Port),
	}
	if proxy.Type == "socks5" {
		proxyURL.Scheme = "socks5"
	}
	if proxy.Username != "" {
		proxyURL.User = url.UserPassword(proxy.Username, proxy.Password)
	}

	return &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			DialContext: (&net.Dialer{
				Timeout: 30 * time.Second,
			}).DialContext,
		},
	}
}

// getAPIKeyFromContext extracts the API key from the request context
func (h *V1Handler) getAPIKeyFromContext(c *gin.Context) *model.ApiKey {
	val, exists := c.Get(middleware.APIKeyContextKey)
	if !exists {
		return nil
	}
	apiKey, ok := val.(*model.ApiKey)
	if !ok {
		return nil
	}
	return apiKey
}
