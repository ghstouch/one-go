package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UsageHistory represents a request usage record
type UsageHistory struct {
	ID              string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Provider        string    `gorm:"type:varchar(100);index" json:"provider"`
	Model           string    `gorm:"type:varchar(100);index" json:"model"`
	ConnectionID    string    `gorm:"type:varchar(36)" json:"connectionId,omitempty"`
	ApiKeyName      string    `gorm:"type:varchar(100)" json:"apiKeyName,omitempty"`
	TokensInput     int       `gorm:"default:0" json:"tokensInput"`
	TokensOutput    int       `gorm:"default:0" json:"tokensOutput"`
	TokensCacheRead int       `gorm:"default:0" json:"tokensCacheRead"`
	TotalTokens     int       `gorm:"default:0" json:"totalTokens"`
	ServiceTier     string    `gorm:"type:varchar(50)" json:"serviceTier,omitempty"`
	Status          string    `gorm:"type:varchar(20);index" json:"status"` // success, error
	Success         bool      `gorm:"default:true" json:"success"`
	LatencyMs       int       `gorm:"default:0" json:"latencyMs"`
	TtftMs          int       `gorm:"default:0" json:"ttftMs"`
	ErrorCode       string    `gorm:"type:varchar(50)" json:"errorCode,omitempty"`
	ErrorMessage    string    `gorm:"type:text" json:"errorMessage,omitempty"`
	Cost            float64   `gorm:"default:0" json:"cost"`
	Timestamp       time.Time `gorm:"index" json:"timestamp"`
	CreatedAt       time.Time `json:"createdAt"`
}

func (u *UsageHistory) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	if u.Timestamp.IsZero() {
		u.Timestamp = time.Now()
	}
	return nil
}

func (UsageHistory) TableName() string {
	return "usage_history"
}

// CallLog represents a detailed request/response log
type CallLog struct {
	ID             string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Provider       string    `gorm:"type:varchar(100);index" json:"provider"`
	Model          string    `gorm:"type:varchar(100)" json:"model"`
	ApiKeyName     string    `gorm:"type:varchar(100)" json:"apiKeyName,omitempty"`
	RequestMethod  string    `gorm:"type:varchar(10)" json:"requestMethod"`
	RequestPath    string    `gorm:"type:varchar(500)" json:"requestPath"`
	RequestHeaders string    `gorm:"type:text" json:"requestHeaders,omitempty"`
	RequestBody    string    `gorm:"type:text" json:"requestBody,omitempty"`
	ResponseStatus int       `json:"responseStatus"`
	ResponseBody   string    `gorm:"type:text" json:"responseBody,omitempty"`
	StatusCode     int       `json:"statusCode"`
	LatencyMs      int       `gorm:"default:0" json:"latencyMs"`
	ErrorMessage   string    `gorm:"type:text" json:"errorMessage,omitempty"`
	IsError        bool      `gorm:"default:false" json:"isError"`
	Timestamp      time.Time `gorm:"index" json:"timestamp"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (cl *CallLog) BeforeCreate(tx *gorm.DB) error {
	if cl.ID == "" {
		cl.ID = uuid.New().String()
	}
	if cl.Timestamp.IsZero() {
		cl.Timestamp = time.Now()
	}
	return nil
}

func (CallLog) TableName() string {
	return "call_logs"
}
