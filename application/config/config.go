package config

import (
	"encoding/json"
	"os"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Fabric   FabricConfig   `json:"fabric"`
	Security SecurityConfig `json:"security"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `json:"port"`
	Host string `json:"host"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// FabricConfig Fabric配置
type FabricConfig struct {
	ConfigPath    string `json:"configPath"`
	ChannelName   string `json:"channelName"`
	ChaincodeName string `json:"chaincodeName"`
	OrgName       string `json:"orgName"`
	UserName      string `json:"userName"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret  string `json:"jwtSecret"`
	SM4Key     string `json:"sm4Key"` // 国密SM4密钥
	TokenExpiry int   `json:"tokenExpiry"` // JWT过期时间(小时)
}

// Load 加载配置
func Load() (*Config, error) {
	// 优先从环境变量读取配置文件路径
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config/app-config.json"
	}

	file, err := os.Open(configFile)
	if err != nil {
		// 如果配置文件不存在，使用默认配置
		return getDefaultConfig(), nil
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	// 从环境变量覆盖配置
	overrideFromEnv(&config)

	return &config, nil
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "3306",
			Username: "certuser",
			Password: "certpass123",
			Database: "cert_system",
		},
		Fabric: FabricConfig{
			ConfigPath:    "configs/fabric-config.yaml",
			ChannelName:   "certchannel",
			ChaincodeName: "certchaincode",
			OrgName:       "Org1MSP",
			UserName:      "User1",
		},
		Security: SecurityConfig{
			JWTSecret:   "your-jwt-secret-key-here",
			SM4Key:      "1234567890123456", // 16字节密钥
			TokenExpiry: 24, // 24小时
		},
	}
}

// overrideFromEnv 从环境变量覆盖配置
func overrideFromEnv(config *Config) {
	if port := os.Getenv("SERVER_PORT"); port != "" {
		config.Server.Port = port
	}
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		config.Database.Port = dbPort
	}
	if dbUser := os.Getenv("DB_USERNAME"); dbUser != "" {
		config.Database.Username = dbUser
	}
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		config.Database.Password = dbPass
	}
	if dbName := os.Getenv("DB_DATABASE"); dbName != "" {
		config.Database.Database = dbName
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.Security.JWTSecret = jwtSecret
	}
	if sm4Key := os.Getenv("SM4_KEY"); sm4Key != "" {
		config.Security.SM4Key = sm4Key
	}
}
