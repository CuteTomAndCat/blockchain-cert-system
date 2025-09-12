package main

import (
	"cert-system/internal/api"
	"cert-system/internal/database"
	"cert-system/internal/service"
	"cert-system/config" // 导入 config 包
	"log"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}

	// 初始化数据库客户端
	dbClient, err := database.NewClient(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}
	log.Println("数据库连接成功")
	
	// 这里可以添加数据库迁移逻辑，例如:
	// database.AutoMigrate(dbClient.DB)

	// 初始化服务层
	authService := service.NewAuthService(dbClient)
	certService := service.NewCertificateService(dbClient)
	testDataService := service.NewTestDataService(dbClient)
	
	// 初始化 Gin 路由器
	router := gin.Default()
	
	// 设置路由
	api.SetupRoutes(router, certService, testDataService, authService)

	// 启动服务器
	log.Printf("服务器在端口 %s 上运行", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}