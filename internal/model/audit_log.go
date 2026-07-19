package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLog represents an administrative action log
type AuditLog struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID    string    `gorm:"type:varchar(36);index" json:"userId"`
	Username  string    `gorm:"type:varchar(100)" json:"username"`
	Action    string    `gorm:"type:varchar(50);not null;index" json:"action"` // create, update, delete, login, etc.
	Resource  string    `gorm:"type:varchar(50);not null" json:"resource"`     // provider, combo, api_key, proxy, etc.
	ResourceID string   `gorm:"type:varchar(36)" json:"resourceId,omitempty"`
	Details   string    `gorm:"type:text" json:"details,omitempty"`
	IPAddress string    `gorm:"type:varchar(45)" json:"ipAddress,omitempty"`
	Timestamp time.Time `gorm:"index" json:"timestamp"`
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.Timestamp.IsZero() {
		a.Timestamp = time.Now()
	}
	return nil
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// Audit action constants
const (
	AuditActionCreate = "create"
	AuditActionUpdate = "update"
	AuditActionDelete = "delete"
	AuditActionLogin  = "login"
	AuditActionLogout = "logout"
	AuditActionTest   = "test"
	AuditActionBatch  = "batch_delete"
)

// Audit resource constants
const (
	AuditResourceProvider = "provider"
	AuditResourceCombo    = "combo"
	AuditResourceAPIKey   = "api_key"
	AuditResourceProxy    = "proxy"
	AuditResourceSettings = "settings"
	AuditResourceUser     = "user"
)
