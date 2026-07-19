package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Provider represents an AI or external service provider
type Provider struct {
	ID            string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name          string         `gorm:"type:varchar(100);not null;index" json:"name"`
	Type          string         `gorm:"type:varchar(50);not null;index" json:"type"` // openai, anthropic, oauth, cookie
	Description   string         `gorm:"type:text" json:"description,omitempty"`
	APIKey        string         `gorm:"type:text" json:"-"` // for OpenAI/Anthropic
	ClientID      string         `gorm:"type:text" json:"-"` // for OAuth
	ClientSecret  string         `gorm:"type:text" json:"-"` // for OAuth
	AuthURL       string         `gorm:"type:text" json:"authUrl,omitempty"`
	TokenURL      string         `gorm:"type:text" json:"tokenUrl,omitempty"`
	RedirectURI   string         `gorm:"type:text" json:"redirectUri,omitempty"`
	BaseURL       string         `gorm:"type:text;not null" json:"baseUrl"`
	Status        string         `gorm:"type:varchar(20);default:'active'" json:"status"` // active/inactive/pending
	ProxyID       string         `gorm:"type:varchar(36)" json:"proxyId,omitempty"`
	EndpointsJSON []string       `gorm:"type:text;serializer:json" json:"endpoints,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate generates UUID
func (p *Provider) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// TableName returns the table name
func (Provider) TableName() string {
	return "providers"
}

// ProviderType constants
const (
	ProviderTypeOpenAI    = "openai"
	ProviderTypeAnthropic = "anthropic"
	ProviderTypeOAuth     = "oauth"
	ProviderTypeCookie    = "cookie"
)

// ProviderEndpointType constants
const (
	EndpointTypeChat      = "chat_completion"
	EndpointTypeImage     = "image"
	EndpointTypeAudio     = "audio"
	EndpointTypeEmbedding = "embedding"
)

// AutoDetectEndpoints detects endpoint types from Name and BaseURL
func (p *Provider) AutoDetectEndpoints() []string {
	var endpoints []string

	name := strings.ToLower(p.Name)
	baseURL := strings.ToLower(p.BaseURL)

	if strings.Contains(name, "embedding") || strings.Contains(baseURL, "embedding") {
		endpoints = append(endpoints, EndpointTypeEmbedding)
	}

	if strings.Contains(name, "dall-e") || strings.Contains(name, "image") || strings.Contains(name, "stable-diffusion") {
		endpoints = append(endpoints, EndpointTypeImage)
	}

	if strings.Contains(name, "whisper") || strings.Contains(name, "tts") || strings.Contains(name, "audio") {
		endpoints = append(endpoints, EndpointTypeAudio)
	}

	if strings.Contains(name, "claude") || strings.Contains(name, "gpt") {
		endpoints = append(endpoints, EndpointTypeChat)
	}

	return endpoints
}
