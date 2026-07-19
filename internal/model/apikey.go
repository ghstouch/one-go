package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ApiKey represents an API key for authentication
type ApiKey struct {
	ID                   string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name                 string         `gorm:"type:varchar(100);not null" json:"name"`
	Key                  string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	AllowedModels        string         `gorm:"type:text" json:"allowedModels,omitempty"` // comma-separated or JSON
	BlockedModels        string         `gorm:"type:text" json:"blockedModels,omitempty"` // comma-separated or JSON
	AllowedCombos        string         `gorm:"type:text" json:"allowedCombos,omitempty"` // comma-separated or JSON
	Scopes               string         `gorm:"type:text" json:"scopes,omitempty"`        // comma-separated: chat,completions,embeddings,models
	IsActive             bool           `gorm:"default:true" json:"isActive"`
	IsBanned             bool           `gorm:"default:false" json:"isBanned"`
	ExpiresAt            *time.Time     `json:"expiresAt,omitempty"`
	MaxRequestsPerDay    int            `gorm:"default:0" json:"maxRequestsPerDay"`    // 0 = unlimited
	MaxRequestsPerMinute int            `gorm:"default:0" json:"maxRequestsPerMinute"` // 0 = unlimited
	RequestsToday        int            `gorm:"default:0" json:"requestsToday"`
	RequestsThisMinute   int            `gorm:"default:0" json:"requestsThisMinute"`
	LastUsedAt           *time.Time     `json:"lastUsedAt,omitempty"`
	TotalRequests        int64          `gorm:"default:0" json:"totalRequests"`
	CreatedAt            time.Time      `json:"createdAt"`
	UpdatedAt            time.Time      `json:"updatedAt"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (k *ApiKey) BeforeCreate(tx *gorm.DB) error {
	if k.ID == "" {
		k.ID = uuid.New().String()
	}
	return nil
}

func (ApiKey) TableName() string {
	return "api_keys"
}

// Scope constants
const (
	ScopeChat        = "chat"
	ScopeCompletions = "completions"
	ScopeEmbeddings  = "embeddings"
	ScopeModels      = "models"
)

// HasScope checks if the API key has a specific scope
// Empty scopes = all scopes allowed (unrestricted)
func (k *ApiKey) HasScope(scope string) bool {
	if k.Scopes == "" {
		return true // unrestricted
	}
	for _, s := range strings.Split(k.Scopes, ",") {
		if strings.TrimSpace(s) == scope {
			return true
		}
	}
	return false
}

// HasModelAccess checks if the API key can access a specific model
// Empty allowedModels = all models allowed
func (k *ApiKey) HasModelAccess(model string) bool {
	if k.AllowedModels != "" {
		found := false
		for _, m := range strings.Split(k.AllowedModels, ",") {
			if strings.TrimSpace(m) == model {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if k.BlockedModels != "" {
		for _, m := range strings.Split(k.BlockedModels, ",") {
			if strings.TrimSpace(m) == model {
				return false
			}
		}
	}
	return true
}
