package api

import (
	"cert-system/internal/service"
	"cert-system/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建新的 AuthHandler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login 处理登录请求
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 调用 AuthService 的登录方法
	resp, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		// 根据错误类型返回不同的HTTP状态码
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Code:    401,
			Message: err.Error(),
		})
		return
	}
	
	// 登录成功，返回响应
	c.JSON(http.StatusOK, resp)
}

// Logout 处理登出请求
func (h *AuthHandler) Logout(c *gin.Context) {
	// 登出逻辑，例如将令牌加入黑名单或仅仅返回成功
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "登出成功",
	})
}

// GetProfile 获取用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Code: 401,
			Message: "未找到用户信息",
		})
		return
	}

	user, err := h.authService.GetProfile(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "获取用户信息失败: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "获取成功",
		Data:    user,
	})
}