package api

import (
	"cert-system/internal/helper" // 导入 helper 包
	"cert-system/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{Code: 401, Message: "未提供认证令牌"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{Code: 401, Message: "令牌格式错误"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := helper.ParseJWT(tokenString) // 调用 helper 包中的解析函数
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{Code: 401, Message: "无效或过期的令牌"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role.(string) != "admin" {
			c.JSON(http.StatusForbidden, models.APIResponse{Code: 403, Message: "权限不足"})
			c.Abort()
			return
		}
		c.Next()
	}
}