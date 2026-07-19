package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ModelAlias maps user-friendly model names to actual provider model names
type ModelAlias struct {
	ID           string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Alias        string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"alias"`
	ProviderType string         `gorm:"type:varchar(50);not null" json:"providerType"`
	ActualModel  string         `gorm:"type:varchar(100);not null" json:"actualModel"`
	Description  string         `gorm:"type:text" json:"description,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (m *ModelAlias) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

func (ModelAlias) TableName() string {
	return "model_aliases"
}

// DefaultModelAliases returns commonly used model aliases
func DefaultModelAliases() []ModelAlias {
	return []ModelAlias{
		{Alias: "gpt-4", ProviderType: "openai", ActualModel: "gpt-4", Description: "GPT-4"},
		{Alias: "gpt-4o", ProviderType: "openai", ActualModel: "gpt-4o", Description: "GPT-4o"},
		{Alias: "gpt-4o-mini", ProviderType: "openai", ActualModel: "gpt-4o-mini", Description: "GPT-4o Mini"},
		{Alias: "gpt-3.5-turbo", ProviderType: "openai", ActualModel: "gpt-3.5-turbo", Description: "GPT-3.5 Turbo"},
		{Alias: "claude-3-opus", ProviderType: "anthropic", ActualModel: "claude-3-opus-20240229", Description: "Claude 3 Opus"},
		{Alias: "claude-3-sonnet", ProviderType: "anthropic", ActualModel: "claude-3-sonnet-20240229", Description: "Claude 3 Sonnet"},
		{Alias: "claude-3-haiku", ProviderType: "anthropic", ActualModel: "claude-3-haiku-20240307", Description: "Claude 3 Haiku"},
	}
}
