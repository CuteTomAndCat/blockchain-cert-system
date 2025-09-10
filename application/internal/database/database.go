package database

import (
	"database/sql"
	"fmt"
	"cert-system/config"

	_ "github.com/go-sql-driver/mysql"
)

// DB 数据库连接结构
type DB struct {
	*sql.DB
}

// Init 初始化数据库连接
func Init(config config.DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库ping失败: %v", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)

	return &DB{db}, nil
}
