package service

import (
	"cert-system/internal/database"
	"cert-system/internal/helper" // 假设你有一个用于哈希的 helper 包
	"cert-system/internal/models"
	"errors"
	"log"
	"time"
	"crypto/sha256" // 导入 sha256 包
	"encoding/hex"  // 导入 encoding/hex 包
)

// AuthService 认证服务
type AuthService struct {
	dbClient *database.Client
}

// NewAuthService 创建新的 AuthService
func NewAuthService(dbClient *database.Client) *AuthService {
	return &AuthService{
		dbClient: dbClient,
	}
}

// hashPassword 使用 SHA256 对密码进行哈希（与数据库中的存储方式一致）
// 在实际应用中，请考虑使用更安全的 bcrypt 或 Argon2
func hashPassword(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Login 登录功能
func (s *AuthService) Login(username, password string) (*models.LoginResponse, error) {
	// 1. 根据用户名查询数据库中的用户
	var user models.User
	// 注意：这里的查询现在会使用 models.User 结构体，并且 GORM 会自动忽略 PasswordHash 字段（如果它的 GORM tag 是 json:"-"）
	// 我们需要显式查询 PasswordHash
	result := s.dbClient.DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, database.ErrRecordNotFound) {
			log.Printf("用户 %s 未找到", username)
			return nil, errors.New("用户名或密码错误")
		}
		log.Printf("数据库查询失败: %v", result.Error)
		return nil, result.Error
	}

	// 2. 验证密码哈希
	// 将用户输入的明文密码进行哈希，然后与数据库中存储的哈希值进行比较
	inputPasswordHash := hashPassword(password)

	if user.PasswordHash != inputPasswordHash {
		log.Printf("用户 %s 密码验证失败", username)
		return nil, errors.New("用户名或密码错误")
	}

	// 3. 生成JWT令牌
	// 假设 helper.GenerateJWT 函数能正确生成 token
	tokenString, expirationTime, err := helper.GenerateJWT(user.ID, user.Username, user.Role)
	if err != nil {
		log.Printf("生成JWT令牌失败: %v", err)
		return nil, errors.New("登录失败，请稍后再试")
	}

	// 4. 返回登录响应
	resp := &models.LoginResponse{
		Code:    200,
		Message: "登录成功",
		Data: struct {
			Token     string `json:"token"`
			UserID    int64  `json:"userId"`
			Username  string `json:"username"`
			Role      string `json:"role"`
			ExpiresAt string `json:"expiresAt"`
		}{
			Token:     tokenString,
			UserID:    user.ID,
			Username:  user.Username,
			Role:      user.Role,
			ExpiresAt: expirationTime.Format(time.RFC3339),
		},
	}

	return resp, nil
}

// GetProfile 获取用户信息
func (s *AuthService) GetProfile(userID int64) (*models.User, error) {
	var user models.User
	// 注意：这里查询的是 User 结构体，GORM 会自动知道 PasswordHash 是要忽略的
	result := s.dbClient.DB.First(&user, userID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}