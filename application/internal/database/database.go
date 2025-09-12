package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ErrRecordNotFound 是一个公共的错误变量，用于表示未找到记录
var ErrRecordNotFound = gorm.ErrRecordNotFound

// Client 数据库客户端结构
type Client struct {
	DB *gorm.DB
}

// NewClient 创建并返回一个新的数据库客户端
func NewClient(dsn string) (*Client, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("无法连接到数据库: %w", err)
	}

	return &Client{DB: db}, nil
}