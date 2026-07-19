package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Combo represents a routing combo that maps multiple providers
type Combo struct {
	ID              string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name            string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Description     string         `gorm:"type:text" json:"description,omitempty"`
	Strategy        string         `gorm:"type:varchar(20);default:'priority'" json:"strategy"` // priority, round-robin, weighted, fallback
	IsHidden        bool           `gorm:"default:false" json:"isHidden"`
	MaxRetries      int            `gorm:"default:3" json:"maxRetries"`
	RetryDelayMs    int            `gorm:"default:1000" json:"retryDelayMs"`
	FallbackDelayMs int            `gorm:"default:0" json:"fallbackDelayMs"`
	IsActive        bool           `gorm:"default:true" json:"isActive"`
	Targets         []ComboTarget  `gorm:"foreignKey:ComboID" json:"targets,omitempty"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// ComboTarget represents a target within a combo
type ComboTarget struct {
	ID         string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	ComboID    string    `gorm:"type:varchar(36);not null;index" json:"comboId"`
	ProviderID string    `gorm:"type:varchar(36);not null" json:"providerId"`
	ModelName  string    `gorm:"type:varchar(100)" json:"modelName"`
	Priority   int       `gorm:"default:0" json:"priority"`
	Weight     int       `gorm:"default:1" json:"weight"`
	IsActive   bool      `gorm:"default:true" json:"isActive"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (c *Combo) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

func (Combo) TableName() string {
	return "combos"
}

func (ct *ComboTarget) BeforeCreate(tx *gorm.DB) error {
	if ct.ID == "" {
		ct.ID = uuid.New().String()
	}
	return nil
}

func (ComboTarget) TableName() string {
	return "combo_targets"
}

const (
	StrategyPriority   = "priority"
	StrategyRoundRobin = "round-robin"
	StrategyWeighted   = "weighted"
	StrategyFallback   = "fallback"
)
