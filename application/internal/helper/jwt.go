package helper

import (
	"cert-system/internal/models"
	"fmt"
	"time"
	"github.com/golang-jwt/jwt/v4"
)

// JWTSecretKey 用于签署JWT的密钥
// 在生产环境中，请将其放在环境变量或配置中心中
var JWTSecretKey = []byte("your_super_secret_key")

// GenerateJWT 生成JWT令牌
func GenerateJWT(userID int64, username string, role string) (tokenString string, expirationTime time.Time, err error) {
	expirationTime = time.Now().Add(24 * time.Hour) // 令牌有效期为24小时
	claims := &models.JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(JWTSecretKey)
	return
}

// ParseJWT 解析JWT令牌
func ParseJWT(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非法的签名方法: %v", token.Header["alg"])
		}
		return JWTSecretKey, nil
	})

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}