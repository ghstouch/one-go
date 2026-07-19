package service

import (
	"github.com/ghstouch/one-go/internal/model"
	"github.com/ghstouch/one-go/internal/repository"
)

// AuditLogService defines audit log operations
type AuditLogService interface {
	Log(userID, username, action, resource, resourceID, details, ip string) error
	List(page, pageSize int, action, resource string) ([]model.AuditLog, int64, error)
	GetByID(id string) (*model.AuditLog, error)
}

type auditLogService struct {
	repo repository.AuditLogRepository
}

func NewAuditLogService(repo repository.AuditLogRepository) AuditLogService {
	return &auditLogService{repo: repo}
}

func (s *auditLogService) Log(userID, username, action, resource, resourceID, details, ip string) error {
	entry := &model.AuditLog{
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		IPAddress:  ip,
	}
	return s.repo.Create(entry)
}

func (s *auditLogService) List(page, pageSize int, action, resource string) ([]model.AuditLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(page, pageSize, action, resource)
}

func (s *auditLogService) GetByID(id string) (*model.AuditLog, error) {
	return s.repo.GetByID(id)
}
