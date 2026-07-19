package repository

import (
	"github.com/omniroute-go/internal/model"
	"gorm.io/gorm"
)

// AuditLogRepository defines audit log data operations
type AuditLogRepository interface {
	Create(log *model.AuditLog) error
	List(page, pageSize int, action, resource string) ([]model.AuditLog, int64, error)
	GetByID(id string) (*model.AuditLog, error)
}

type auditLogRepo struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepo{db: db}
}

func (r *auditLogRepo) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditLogRepo) List(page, pageSize int, action, resource string) ([]model.AuditLog, int64, error) {
	var items []model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{})
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *auditLogRepo) GetByID(id string) (*model.AuditLog, error) {
	var log model.AuditLog
	err := r.db.Where("id = ?", id).First(&log).Error
	return &log, err
}
