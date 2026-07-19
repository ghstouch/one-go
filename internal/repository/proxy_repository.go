package repository

import (
	"github.com/ghstouch/one-go/internal/model"
	"gorm.io/gorm"
)

// ProxyRepository defines proxy data operations
type ProxyRepository interface {
	Create(proxy *model.Proxy) error
	GetByID(id string) (*model.Proxy, error)
	Update(proxy *model.Proxy) error
	Delete(id string) error
	List() ([]model.Proxy, error)
	CreateLog(log *model.ProxyLog) error
	ListLogs(proxyID string, limit int) ([]model.ProxyLog, error)
}

type proxyRepo struct {
	db *gorm.DB
}

func NewProxyRepository(db *gorm.DB) ProxyRepository {
	return &proxyRepo{db: db}
}

func (r *proxyRepo) Create(proxy *model.Proxy) error {
	return r.db.Create(proxy).Error
}

func (r *proxyRepo) GetByID(id string) (*model.Proxy, error) {
	var proxy model.Proxy
	err := r.db.Where("id = ?", id).First(&proxy).Error
	return &proxy, err
}

func (r *proxyRepo) Update(proxy *model.Proxy) error {
	return r.db.Save(proxy).Error
}

func (r *proxyRepo) Delete(id string) error {
	return r.db.Delete(&model.Proxy{}, "id = ?", id).Error
}

func (r *proxyRepo) List() ([]model.Proxy, error) {
	var proxies []model.Proxy
	err := r.db.Order("created_at DESC").Find(&proxies).Error
	return proxies, err
}

func (r *proxyRepo) CreateLog(log *model.ProxyLog) error {
	return r.db.Create(log).Error
}

func (r *proxyRepo) ListLogs(proxyID string, limit int) ([]model.ProxyLog, error) {
	var logs []model.ProxyLog
	query := r.db.Order("timestamp DESC")
	if proxyID != "" {
		query = query.Where("proxy_id = ?", proxyID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&logs).Error
	return logs, err
}
