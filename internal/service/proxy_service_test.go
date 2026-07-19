package service

import (
	"testing"

	"github.com/omniroute-go/internal/model"
)

// mockProxyRepo implements repository.ProxyRepository for testing
type mockProxyRepo struct {
	proxies map[string]*model.Proxy
	logs    []model.ProxyLog
}

func newMockProxyRepo() *mockProxyRepo {
	return &mockProxyRepo{proxies: make(map[string]*model.Proxy)}
}

func (m *mockProxyRepo) Create(proxy *model.Proxy) error {
	m.proxies[proxy.ID] = proxy
	return nil
}

func (m *mockProxyRepo) GetByID(id string) (*model.Proxy, error) {
	if p, ok := m.proxies[id]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockProxyRepo) Update(proxy *model.Proxy) error {
	m.proxies[proxy.ID] = proxy
	return nil
}

func (m *mockProxyRepo) Delete(id string) error {
	delete(m.proxies, id)
	return nil
}

func (m *mockProxyRepo) List() ([]model.Proxy, error) {
	var list []model.Proxy
	for _, p := range m.proxies {
		list = append(list, *p)
	}
	return list, nil
}

func (m *mockProxyRepo) CreateLog(log *model.ProxyLog) error {
	m.logs = append(m.logs, *log)
	return nil
}

func (m *mockProxyRepo) ListLogs(proxyID string, limit int) ([]model.ProxyLog, error) {
	return m.logs, nil
}

func TestCreateProxy_Success(t *testing.T) {
	repo := newMockProxyRepo()
	svc := NewProxyService(repo)

	p := &model.Proxy{Name: "Test Proxy", Type: "http", Host: "127.0.0.1", Port: 8080}
	err := svc.CreateProxy(p)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateProxy_MissingName(t *testing.T) {
	repo := newMockProxyRepo()
	svc := NewProxyService(repo)

	p := &model.Proxy{Type: "http", Host: "127.0.0.1", Port: 8080}
	err := svc.CreateProxy(p)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestCreateProxy_MissingHost(t *testing.T) {
	repo := newMockProxyRepo()
	svc := NewProxyService(repo)

	p := &model.Proxy{Name: "Test", Type: "http", Port: 8080}
	err := svc.CreateProxy(p)
	if err == nil {
		t.Fatal("expected error for missing host")
	}
}

func TestCreateProxy_InvalidPort(t *testing.T) {
	repo := newMockProxyRepo()
	svc := NewProxyService(repo)

	p := &model.Proxy{Name: "Test", Type: "http", Host: "127.0.0.1", Port: 0}
	err := svc.CreateProxy(p)
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestCreateProxy_InvalidType(t *testing.T) {
	repo := newMockProxyRepo()
	svc := NewProxyService(repo)

	p := &model.Proxy{Name: "Test", Type: "ftp", Host: "127.0.0.1", Port: 8080}
	err := svc.CreateProxy(p)
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestCreateProxy_ValidTypes(t *testing.T) {
	repo := newMockProxyRepo()
	svc := NewProxyService(repo)

	for _, typ := range []string{"http", "https", "socks5"} {
		p := &model.Proxy{Name: "Test-" + typ, Type: typ, Host: "127.0.0.1", Port: 8080}
		err := svc.CreateProxy(p)
		if err != nil {
			t.Fatalf("expected no error for type %s, got %v", typ, err)
		}
	}
}

func TestDeleteProxy_Success(t *testing.T) {
	repo := newMockProxyRepo()
	svc := NewProxyService(repo)

	repo.proxies["px1"] = &model.Proxy{ID: "px1", Name: "Test"}
	err := svc.DeleteProxy("px1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := repo.proxies["px1"]; exists {
		t.Fatal("expected proxy to be deleted")
	}
}

func TestDeleteProxy_EmptyID(t *testing.T) {
	repo := newMockProxyRepo()
	svc := NewProxyService(repo)

	err := svc.DeleteProxy("")
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestListProxies(t *testing.T) {
	repo := newMockProxyRepo()
	svc := NewProxyService(repo)

	repo.proxies["px1"] = &model.Proxy{ID: "px1", Name: "A"}
	repo.proxies["px2"] = &model.Proxy{ID: "px2", Name: "B"}

	list, err := svc.ListProxies()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 proxies, got %d", len(list))
	}
}
