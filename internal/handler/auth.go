package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ghstouch/one-go/internal/middleware"
	"github.com/ghstouch/one-go/internal/service"
	"github.com/ghstouch/one-go/pkg/response"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest represents login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=6"`
}

// Login handles user login
// @Summary Login
// @Description Authenticate user and get JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=service.TokenResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	tokenResponse, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, tokenResponse)
}

// RefreshToken handles token refresh
// @Summary Refresh Token
// @Description Refresh JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=service.TokenResponse}
// @Failure 401 {object} response.Response
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) <= 7 {
		response.Unauthorized(c, "Invalid authorization header")
		return
	}

	tokenString := authHeader[7:] // Remove "Bearer "

	tokenResponse, err := h.authService.RefreshToken(tokenString)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, tokenResponse)
}

// Me returns current user info
// @Summary Get Current User
// @Description Get current authenticated user info
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=model.SafeUser}
// @Failure 401 {object} response.Response
// @Router /api/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	user, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	response.OK(c, user)
}

// Logout handles user logout
// @Summary Logout
// @Description Logout user (client should discard token)
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// In JWT-based auth, logout is typically handled client-side
	// by discarding the token. Server can optionally blacklist the token.
	response.OKWithMessage(c, "Logged out successfully", nil)
}
