package models

import (
	"time"
	"github.com/golang-jwt/jwt/v4"
)

// APIResponse 通用API响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PagedResponse 分页响应
type PagedResponse struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

// VerificationResult 验证结果
type VerificationResult struct {
	CertNumber string      `json:"certNumber"`
	IsValid    bool        `json:"isValid"`
	Message    string      `json:"message"`
	VerifiedAt time.Time   `json:"verifiedAt"`
	Certificate *Certificate `json:"certificate,omitempty"`
}

// HistoryRecord 历史记录
type HistoryRecord struct {
	TxID      string      `json:"txId"`
	Value     interface{} `json:"value"`
	Timestamp string      `json:"timestamp"`
	IsDelete  bool        `json:"isDelete"`
}

// --- 请求/响应模型 ---

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"` // 明文密码
}

// LoginResponse 用户登录响应
type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token     string `json:"token"`
		UserID    int64  `json:"userId"`
		Username  string `json:"username"`
		Role      string `json:"role"`
		ExpiresAt string `json:"expiresAt"` // JWT 实际的过期时间字符串
	} `json:"data"`
}

// CreateCertificateRequest 创建证书请求
type CreateCertificateRequest struct {
	CertNumber         string `json:"certNumber" binding:"required"`
	CustomerID         int64  `json:"customerId" binding:"required"`
	InstrumentName     string `json:"instrumentName" binding:"required"`
	InstrumentNumber   string `json:"instrumentNumber"`
	Manufacturer       string `json:"manufacturer"`
	ModelSpec          string `json:"modelSpec"`
	InstrumentAccuracy string `json:"instrumentAccuracy"`
	TestDate           string `json:"testDate" binding:"required"` // 接收 string，在 handlers 中解析
	ExpireDate         string `json:"expireDate"`                   // 接收 string，在 handlers 中解析
	TestResult         string `json:"testResult" binding:"required"` // 必须是 "qualified" 或 "unqualified"
}

// BatchTestDataRequest 批量测试数据请求
type BatchTestDataRequest struct {
	CertNumber string             `json:"certNumber" binding:"required"`
	TestData   []AddTestDataRequest `json:"testData" binding:"required"`
}

// --- 数据库/核心模型 ---

// User 用户模型
type User struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"column:username;uniqueIndex" json:"username"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"` // 存储密码哈希
	Role         string    `gorm:"column:role" json:"role"`      // ENUM ('admin', 'operator', 'viewer')
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
	// No DeletedAt field to match the provided schema without soft delete
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Certificate 证书模型
type Certificate struct {
	ID                 int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CertNumber         string    `json:"certNumber" gorm:"column:cert_number;uniqueIndex"`
	CustomerID         int64     `json:"customerId" gorm:"column:customer_id"`
	InstrumentName     string    `json:"instrumentName" gorm:"column:instrument_name"`
	InstrumentNumber   string    `json:"instrumentNumber" gorm:"column:instrument_number"`
	Manufacturer       string    `json:"manufacturer" gorm:"column:manufacturer"`
	ModelSpec          string    `json:"modelSpec" gorm:"column:model_spec"`
	InstrumentAccuracy string    `json:"instrumentAccuracy" gorm:"column:instrument_accuracy"`
	TestDate           time.Time `json:"testDate" gorm:"column:test_date"`
	ExpireDate         time.Time `json:"expireDate" gorm:"column:expire_date"`
	TestResult         string    `json:"testResult" gorm:"column:test_result"` // ENUM('qualified', 'unqualified')
	BlockchainTxID     string    `json:"blockchainTxId" gorm:"column:blockchain_tx_id"`
	Status             string    `json:"status" gorm:"column:status"`
	CreatedBy          int64     `json:"createdBy" gorm:"column:created_by"`
	CreatedAt          time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt          time.Time `gorm:"column:updated_at" json:"updatedAt"`

	// 关联数据
	Customer Customer `json:"customer" gorm:"foreignKey:CustomerID"`
}

// Customer 客户模型
type Customer struct {
	ID              int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CustomerName    string    `json:"customerName" gorm:"column:customer_name"`
	CustomerAddress string    `json:"customerAddress" gorm:"column:customer_address"`
	ContactPerson   string    `json:"contactPerson" gorm:"column:contact_person"`
	ContactPhone    string    `json:"contactPhone" gorm:"column:contact_phone"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TestData 测试数据模型
type TestData struct {
    ID                int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
    CertID            int64     `json:"certId" gorm:"column:cert_id"`
    CertNumber        string    `json:"certNumber" gorm:"-"` // 不映射数据库
    DeviceAddr        string    `json:"deviceAddr" gorm:"column:device_addr"`
    DataType          string    `json:"dataType" gorm:"column:data_type"`
    PercentageValue   float64   `json:"percentageValue" gorm:"column:percentage_value"`
    RatioError        float64   `json:"ratioError" gorm:"column:ratio_error"`
    AngleError        float64   `json:"angleError" gorm:"column:angle_error"`
    CurrentValue      float64   `json:"currentValue" gorm:"column:current_value"`
    VoltageValue      float64   `json:"voltageValue" gorm:"column:voltage_value"`
    WorkstationNumber string    `json:"workstationNumber" gorm:"column:workstation_number"`
    TestPoint         string    `json:"testPoint" gorm:"column:test_point"`
    ActualPercentage  float64   `json:"actualPercentage" gorm:"column:actual_percentage"`
    TestTimestamp     time.Time `json:"testTimestamp" gorm:"column:test_timestamp"`
    BlockchainHash    string    `json:"blockchainHash" gorm:"column:blockchain_hash"`
    EncryptedData     string    `json:"-" gorm:"column:encrypted_data"` // 不返回给前端
    DecryptedData     string    `json:"decryptedData" gorm:"-"`        // 不存数据库
    CreatedAt         time.Time `gorm:"column:created_at" json:"createdAt"`
}



// Device 设备模型
type Device struct {
	ID            int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	DeviceAddr    string    `json:"deviceAddr" gorm:"column:device_addr"`
	DeviceName    string    `json:"deviceName" gorm:"column:device_name"`
	Manufacturer  string    `json:"manufacturer" gorm:"column:manufacturer"`
	Model         string    `json:"model" gorm:"column:model"`
	AccuracyClass string    `json:"accuracyClass" gorm:"column:accuracy_class"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// BlockchainTransaction 区块链交易模型
type BlockchainTransaction struct {
	ID              int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	TxID            string    `json:"txId" gorm:"column:tx_id"`
	BlockNumber     int64     `json:"blockNumber" gorm:"column:block_number"`
	TransactionHash string    `json:"transactionHash" gorm:"column:transaction_hash"`
	CertID          int64     `json:"certId" gorm:"column:cert_id"`
	OperationType   string    `json:"operationType" gorm:"column:operation_type"`
	Status          string    `json:"status" gorm:"column:status"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"createdAt"`
}

// BlockchainCertificate 区块链证书模型
type BlockchainCertificate struct {
	CertNumber         string  `json:"certNumber"`
	CustomerName       string  `json:"customerName"`
	CustomerAddress    string  `json:"customerAddress"`
	InstrumentName     string  `json:"instrumentName"`
	Manufacturer       string  `json:"manufacturer"`
	ModelSpec          string  `json:"modelSpec"`
	InstrumentNumber   string  `json:"instrumentNumber"`
	InstrumentAccuracy string  `json:"instrumentAccuracy"`
	TestDate           string  `json:"testDate"`
	ExpireDate         string  `json:"expireDate"`
	TestResult         string  `json:"testResult"`
	Status             string  `json:"status"`
}

// BlockchainTestData 区块链测试数据模型
type BlockchainTestData struct {
	CertNumber        string  `json:"certNumber"`
	DeviceAddr        string  `json:"deviceAddr"`
	DataType          string  `json:"dataType"`
	PercentageValue   float64 `json:"percentageValue"`
	RatioError        float64 `json:"ratioError"`
	AngleError        float64 `json:"angleError"`
	CurrentValue      float64 `json:"currentValue"`
	VoltageValue      float64 `json:"voltageValue"`
	WorkstationNumber string  `json:"workstationNumber"`
	TestPoint         string  `json:"testPoint"`
	ActualPercentage  float64 `json:"actualPercentage"`
	TestTimestamp     string  `json:"testTimestamp"`
	EncryptedData     string  `json:"encryptedData"`
}