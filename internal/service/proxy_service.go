package service

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/omniroute-go/internal/model"
	"github.com/omniroute-go/internal/repository"
)

// ProxyService defines proxy operations
type ProxyService interface {
	CreateProxy(proxy *model.Proxy) error
	GetProxyByID(id string) (*model.Proxy, error)
	UpdateProxy(proxy *model.Proxy) error
	DeleteProxy(id string) error
	ListProxies() ([]model.Proxy, error)
	TestProxy(id string) (string, error)
	ListProxyLogs(proxyID string, limit int) ([]model.ProxyLog, error)
}

type proxyService struct {
	repo repository.ProxyRepository
}

func NewProxyService(repo repository.ProxyRepository) ProxyService {
	return &proxyService{repo: repo}
}

func (s *proxyService) CreateProxy(proxy *model.Proxy) error {
	if proxy.Name == "" {
		return errors.New("proxy name is required")
	}
	if proxy.Host == "" {
		return errors.New("proxy host is required")
	}
	if proxy.Port <= 0 || proxy.Port > 65535 {
		return errors.New("invalid proxy port")
	}
	validTypes := map[string]bool{
		model.ProxyTypeHTTP: true, model.ProxyTypeHTTPS: true, model.ProxyTypeSOCKS5: true,
	}
	if !validTypes[proxy.Type] {
		return errors.New("invalid proxy type (http, https, socks5)")
	}
	return s.repo.Create(proxy)
}

func (s *proxyService) GetProxyByID(id string) (*model.Proxy, error) {
	if id == "" {
		return nil, errors.New("proxy ID is required")
	}
	return s.repo.GetByID(id)
}

func (s *proxyService) UpdateProxy(proxy *model.Proxy) error {
	if proxy.ID == "" {
		return errors.New("proxy ID is required")
	}
	return s.repo.Update(proxy)
}

func (s *proxyService) DeleteProxy(id string) error {
	if id == "" {
		return errors.New("proxy ID is required")
	}
	return s.repo.Delete(id)
}

func (s *proxyService) ListProxies() ([]model.Proxy, error) {
	return s.repo.List()
}

func (s *proxyService) TestProxy(id string) (string, error) {
	if id == "" {
		return "", errors.New("proxy ID is required")
	}

	proxy, err := s.repo.GetByID(id)
	if err != nil {
		return "", errors.New("proxy not found")
	}

	addr := fmt.Sprintf("%s:%d", proxy.Host, proxy.Port)
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	latency := time.Since(start).Milliseconds()

	log := &model.ProxyLog{
		ProxyID:        proxy.ID,
		ProxyName:      proxy.Name,
		ResponseTimeMs: int(latency),
	}

	if err != nil {
		log.Status = "error"
		log.ErrorMessage = err.Error()
		s.repo.CreateLog(log)
		return "", fmt.Errorf("connection failed: %w", err)
	}
	conn.Close()

	log.Status = "success"
	s.repo.CreateLog(log)
	return fmt.Sprintf("Connected in %dms", latency), nil
}

func (s *proxyService) ListProxyLogs(proxyID string, limit int) ([]model.ProxyLog, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListLogs(proxyID, limit)
}

// proxyURL builds a proxy URL for use with http.Transport
func ProxyURL(p *model.Proxy) *url.URL {
	scheme := "http"
	if p.Type == model.ProxyTypeSOCKS5 {
		scheme = "socks5"
	}
	u := &url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", p.Host, p.Port),
	}
	if p.Username != "" {
		u.User = url.UserPassword(p.Username, p.Password)
	}
	return u
}
