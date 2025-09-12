package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `yaml:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string `yaml:"secret"`
}

// Config 根配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
}

// LoadConfig 从指定路径加载配置
func LoadConfig() (*Config, error) {
	configPath := "config/config.yaml" // 默认配置文件路径
	if envPath := os.Getenv("APP_CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("无法读取配置文件 %s, 使用默认值. 错误: %v", configPath, err)
		return defaultConfigs(), nil
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// defaultConfigs 返回默认配置
func defaultConfigs() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
		},
		Database: DatabaseConfig{
			DSN: "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
		},
		JWT: JWTConfig{
			Secret: "your_super_secret_key",
		},
	}
}