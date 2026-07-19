package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Proxy represents a proxy configuration
type Proxy struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Type      string         `gorm:"type:varchar(10);not null" json:"type"` // http, https, socks5
	Host      string         `gorm:"type:varchar(255);not null" json:"host"`
	Port      int            `gorm:"not null" json:"port"`
	Username  string         `gorm:"type:varchar(255)" json:"username,omitempty"`
	Password  string         `gorm:"type:text" json:"-"`
	IsActive  bool           `gorm:"default:true" json:"isActive"`
	IsGlobal  bool           `gorm:"default:false" json:"isGlobal"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (p *Proxy) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

func (Proxy) TableName() string {
	return "proxies"
}

// ProxyLog represents a proxy usage log entry
type ProxyLog struct {
	ID             string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	ProxyID        string    `gorm:"type:varchar(36);index" json:"proxyId"`
	ProxyName      string    `gorm:"type:varchar(100)" json:"proxyName"`
	ConnectionID   string    `gorm:"type:varchar(36)" json:"connectionId,omitempty"`
	Status         string    `gorm:"type:varchar(20)" json:"status"` // success, error, timeout
	ResponseTimeMs int       `gorm:"default:0" json:"responseTimeMs"`
	ErrorMessage   string    `gorm:"type:text" json:"errorMessage,omitempty"`
	ClientIP       string    `gorm:"type:varchar(45)" json:"clientIp,omitempty"`
	UserAgent      string    `gorm:"type:varchar(500)" json:"userAgent,omitempty"`
	Timestamp      time.Time `gorm:"index" json:"timestamp"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (pl *ProxyLog) BeforeCreate(tx *gorm.DB) error {
	if pl.ID == "" {
		pl.ID = uuid.New().String()
	}
	if pl.Timestamp.IsZero() {
		pl.Timestamp = time.Now()
	}
	return nil
}

func (ProxyLog) TableName() string {
	return "proxy_logs"
}

const (
	ProxyTypeHTTP   = "http"
	ProxyTypeHTTPS  = "https"
	ProxyTypeSOCKS5 = "socks5"
)
