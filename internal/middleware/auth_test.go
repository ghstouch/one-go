package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ghstouch/one-go/internal/model"
	"github.com/ghstouch/one-go/internal/service"
	"github.com/gin-gonic/gin"
)

// mockAuthService implements service.AuthService for testing
type mockAuthService struct {
	validateErr bool
	claims      *service.Claims
}

func (m *mockAuthService) Login(username, password string) (*service.TokenResponse, error) {
	return nil, nil
}
func (m *mockAuthService) ValidateToken(tokenString string) (*service.Claims, error) {
	if m.validateErr {
		return nil, nil
	}
	return m.claims, nil
}
func (m *mockAuthService) RefreshToken(tokenString string) (*service.TokenResponse, error) {
	return nil, nil
}
func (m *mockAuthService) HashPassword(password string) (string, error) { return "", nil }
func (m *mockAuthService) CheckPassword(password, hash string) bool     { return false }
func (m *mockAuthService) GetCurrentUser(userID string) (*model.SafeUser, error) {
	return nil, nil
}

func TestAuthMiddleware_NoHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)

	authSvc := &mockAuthService{}
	router.Use(AuthMiddleware(authSvc))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, c.Request)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)

	authSvc := &mockAuthService{}
	router.Use(AuthMiddleware(authSvc))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Basic abc123")
	router.ServeHTTP(w, c.Request)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)

	authSvc := &mockAuthService{
		claims: &service.Claims{UserID: "u1", Username: "admin", Role: "admin"},
	}
	router.Use(AuthMiddleware(authSvc))
	router.GET("/test", func(c *gin.Context) {
		uid, _ := c.Get(UserIDKey)
		c.JSON(200, gin.H{"userId": uid})
	})

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer valid-token")
	router.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)

	router.Use(func(c *gin.Context) { c.Set(RoleKey, "admin") })
	router.Use(RequireRole("admin"))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequireRole_Denied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)

	router.Use(func(c *gin.Context) { c.Set(RoleKey, "viewer") })
	router.Use(RequireRole("admin"))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, c.Request)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestRequireRole_NoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)

	router.Use(RequireRole("admin"))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, c.Request)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireRole_MultipleRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)

	router.Use(func(c *gin.Context) { c.Set(RoleKey, "operator") })
	router.Use(RequireRole("admin", "operator"))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestOptionalAuthMiddleware_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)

	authSvc := &mockAuthService{}
	router.Use(OptionalAuthMiddleware(authSvc))
	router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestOptionalAuthMiddleware_WithToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)

	authSvc := &mockAuthService{
		claims: &service.Claims{UserID: "u1", Username: "admin", Role: "admin"},
	}
	router.Use(OptionalAuthMiddleware(authSvc))
	router.GET("/test", func(c *gin.Context) {
		uid := GetUserIDFromContext(c)
		c.JSON(200, gin.H{"userId": uid})
	})

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer valid-token")
	router.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetContextHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Empty context
	if GetUserIDFromContext(c) != "" {
		t.Fatal("expected empty user ID")
	}
	if GetUsernameFromContext(c) != "" {
		t.Fatal("expected empty username")
	}
	if GetRoleFromContext(c) != "" {
		t.Fatal("expected empty role")
	}

	// Set values
	c.Set(UserIDKey, "u1")
	c.Set(UsernameKey, "admin")
	c.Set(RoleKey, "admin")

	if GetUserIDFromContext(c) != "u1" {
		t.Fatal("expected u1")
	}
	if GetUsernameFromContext(c) != "admin" {
		t.Fatal("expected admin")
	}
	if GetRoleFromContext(c) != "admin" {
		t.Fatal("expected admin")
	}
}
