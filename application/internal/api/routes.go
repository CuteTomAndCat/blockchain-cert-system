package api

import (
	"cert-system/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置API路由
func SetupRoutes(router *gin.Engine, 
	certService *service.CertificateService,
	testDataService *service.TestDataService,
	authService *service.AuthService) {

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "计量证书防伪溯源系统运行正常",
		})
	})

	// API版本组
	v1 := router.Group("/api/v1")
	{
		// 认证相关路由
		auth := v1.Group("/auth")
		{
			auth.POST("/login", NewAuthHandler(authService).Login)
			auth.POST("/logout", AuthMiddleware(), NewAuthHandler(authService).Logout)
			auth.GET("/profile", AuthMiddleware(), NewAuthHandler(authService).GetProfile)
		}

		// 证书相关路由
		certificates := v1.Group("/certificates")
		certificates.Use(AuthMiddleware()) // 所有证书API都需要认证
		{
			certHandler := NewCertificateHandler(certService)
			certificates.POST("", certHandler.CreateCertificate)
			certificates.GET("", certHandler.GetAllCertificates)
			certificates.GET("/:certNumber", certHandler.GetCertificate)
			certificates.PUT("/:certNumber", certHandler.UpdateCertificate)
			certificates.POST("/:certNumber/verify", certHandler.VerifyCertificate)
			certificates.GET("/:certNumber/history", certHandler.GetCertificateHistory)
		}

		// 测试数据相关路由
		testData := v1.Group("/test-data")
		testData.Use(AuthMiddleware())
		{
			testHandler := NewTestDataHandler(testDataService)
			testData.POST("", testHandler.AddTestData)
			testData.POST("/batch", testHandler.BatchAddTestData)
			testData.GET("/certificate/:certId", testHandler.GetTestDataByCert)
			testData.POST("/generate/:certId", testHandler.GenerateTestData)
		}

		// 公开验证接口（不需要认证）
		public := v1.Group("/public")
		{
			public.POST("/verify/:certNumber", NewCertificateHandler(certService).PublicVerifyCertificate)
		}

		// 系统管理相关路由（仅管理员）
		admin := v1.Group("/admin")
		admin.Use(AuthMiddleware(), AdminMiddleware())
		{
			adminHandler := NewAdminHandler(authService)
			admin.GET("/users", adminHandler.GetAllUsers)
			admin.POST("/users", adminHandler.CreateUser)
			admin.PUT("/users/:id", adminHandler.UpdateUser)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)
		}
	}
}
