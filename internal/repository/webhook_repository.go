package repository

import (
	"github.com/omniroute-go/internal/model"
	"gorm.io/gorm"
)

// WebhookRepository defines webhook data operations
type WebhookRepository interface {
	Create(webhook *model.Webhook) error
	GetByID(id string) (*model.Webhook, error)
	List() ([]model.Webhook, error)
	Update(webhook *model.Webhook) error
	Delete(id string) error
	GetByEvent(event string) ([]model.Webhook, error)
	CreateLog(log *model.WebhookLog) error
	ListLogs(webhookID string, limit int) ([]model.WebhookLog, error)
}

type webhookRepo struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) WebhookRepository {
	return &webhookRepo{db: db}
}

func (r *webhookRepo) Create(webhook *model.Webhook) error {
	return r.db.Create(webhook).Error
}

func (r *webhookRepo) GetByID(id string) (*model.Webhook, error) {
	var w model.Webhook
	err := r.db.Where("id = ?", id).First(&w).Error
	return &w, err
}

func (r *webhookRepo) List() ([]model.Webhook, error) {
	var list []model.Webhook
	err := r.db.Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *webhookRepo) Update(webhook *model.Webhook) error {
	return r.db.Save(webhook).Error
}

func (r *webhookRepo) Delete(id string) error {
	return r.db.Delete(&model.Webhook{}, "id = ?", id).Error
}

func (r *webhookRepo) GetByEvent(event string) ([]model.Webhook, error) {
	var list []model.Webhook
	err := r.db.Where("is_active = 1 AND (events = '' OR events LIKE ?)", "%"+event+"%").Find(&list).Error
	return list, err
}

func (r *webhookRepo) CreateLog(log *model.WebhookLog) error {
	return r.db.Create(log).Error
}

func (r *webhookRepo) ListLogs(webhookID string, limit int) ([]model.WebhookLog, error) {
	var logs []model.WebhookLog
	query := r.db.Order("timestamp DESC")
	if webhookID != "" {
		query = query.Where("webhook_id = ?", webhookID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&logs).Error
	return logs, err
}
