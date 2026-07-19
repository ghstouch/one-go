package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Settings represents application settings
type Settings struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Key       string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"key"`
	Value     string         `gorm:"type:text" json:"value"`
	Type      string         `gorm:"type:varchar(50);default:'string'" json:"type"` // string, int, bool, json
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to generate UUID
func (s *Settings) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// TableName returns the table name
func (Settings) TableName() string {
	return "settings"
}

// Setting keys constants
const (
	SettingDefaultModel          = "default_model"
	SettingRateLimitEnabled      = "rate_limit_enabled"
	SettingRateLimitPerMinute    = "rate_limit_per_minute"
	SettingRequireLogin          = "require_login"
	SettingComboDefaultStrategy  = "combo_default_strategy"
	SettingComboMaxRetries       = "combo_max_retries"
	SettingComboRetryDelayMs     = "combo_retry_delay_ms"
	SettingComboFallbackDelayMs  = "combo_fallback_delay_ms"
	SettingTrackMetrics          = "track_metrics"
	SettingLogRetentionDays      = "log_retention_days"
	SettingCallLogRetentionDays  = "call_log_retention_days"
	SettingGlobalProxyEnabled    = "global_proxy_enabled"
	SettingGlobalProxyID         = "global_proxy_id"
)

// SettingType constants
const (
	SettingTypeString = "string"
	SettingTypeInt    = "int"
	SettingTypeBool   = "bool"
	SettingTypeJSON   = "json"
)

// DefaultSettings returns default application settings
func DefaultSettings() []Settings {
	return []Settings{
		{Key: SettingDefaultModel, Value: "", Type: SettingTypeString},
		{Key: SettingRateLimitEnabled, Value: "true", Type: SettingTypeBool},
		{Key: SettingRateLimitPerMinute, Value: "60", Type: SettingTypeInt},
		{Key: SettingRequireLogin, Value: "true", Type: SettingTypeBool},
		{Key: SettingComboDefaultStrategy, Value: "priority", Type: SettingTypeString},
		{Key: SettingComboMaxRetries, Value: "1", Type: SettingTypeInt},
		{Key: SettingComboRetryDelayMs, Value: "2000", Type: SettingTypeInt},
		{Key: SettingComboFallbackDelayMs, Value: "0", Type: SettingTypeInt},
		{Key: SettingTrackMetrics, Value: "true", Type: SettingTypeBool},
		{Key: SettingLogRetentionDays, Value: "30", Type: SettingTypeInt},
		{Key: SettingCallLogRetentionDays, Value: "7", Type: SettingTypeInt},
		{Key: SettingGlobalProxyEnabled, Value: "false", Type: SettingTypeBool},
		{Key: SettingGlobalProxyID, Value: "", Type: SettingTypeString},
	}
}
