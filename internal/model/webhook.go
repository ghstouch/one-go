package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Webhook represents a webhook subscription
type Webhook struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	URL       string         `gorm:"type:text;not null" json:"url"`
	Secret    string         `gorm:"type:text" json:"secret,omitempty"`
	Events    string         `gorm:"type:text" json:"events"` // comma-separated: provider.created,provider.deleted,usage.error,etc.
	IsActive  bool           `gorm:"default:true" json:"isActive"`
	LastSentAt *time.Time    `json:"lastSentAt,omitempty"`
	LastStatus int           `json:"lastStatus,omitempty"`
	FailCount  int           `gorm:"default:0" json:"failCount"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (w *Webhook) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

func (Webhook) TableName() string {
	return "webhooks"
}

// WebhookLog represents a webhook delivery attempt
type WebhookLog struct {
	ID         string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	WebhookID  string    `gorm:"type:varchar(36);index" json:"webhookId"`
	Event      string    `gorm:"type:varchar(100)" json:"event"`
	StatusCode int       `json:"statusCode"`
	Success    bool      `json:"success"`
	Error      string    `gorm:"type:text" json:"error,omitempty"`
	Timestamp  time.Time `gorm:"index" json:"timestamp"`
}

func (wl *WebhookLog) BeforeCreate(tx *gorm.DB) error {
	if wl.ID == "" {
		wl.ID = uuid.New().String()
	}
	if wl.Timestamp.IsZero() {
		wl.Timestamp = time.Now()
	}
	return nil
}

func (WebhookLog) TableName() string {
	return "webhook_logs"
}

// Webhook event constants
const (
	EventProviderCreated = "provider.created"
	EventProviderUpdated = "provider.updated"
	EventProviderDeleted = "provider.deleted"
	EventComboCreated    = "combo.created"
	EventComboUpdated    = "combo.updated"
	EventComboDeleted    = "combo.deleted"
	EventAPIKeyCreated   = "api_key.created"
	EventAPIKeyDeleted   = "api_key.deleted"
	EventUsageError      = "usage.error"
	EventProxyCreated    = "proxy.created"
	EventProxyDeleted    = "proxy.deleted"
)
