package models

import (
	"time"
)

// Certificate 证书模型
type Certificate struct {
	ID                 int64     `json:"id" db:"id"`
	CertNumber         string    `json:"certNumber" db:"cert_number"`
	CustomerID         int64     `json:"customerId" db:"customer_id"`
	InstrumentName     string    `json:"instrumentName" db:"instrument_name"`
	InstrumentNumber   string    `json:"instrumentNumber" db:"instrument_number"`
	Manufacturer       string    `json:"manufacturer" db:"manufacturer"`
	ModelSpec          string    `json:"modelSpec" db:"model_spec"`
	InstrumentAccuracy string    `json:"instrumentAccuracy" db:"instrument_accuracy"`
	TestDate           time.Time `json:"testDate" db:"test_date"`
	ExpireDate         time.Time `json:"expireDate" db:"expire_date"`
	TestResult         string    `json:"testResult" db:"test_result"`
	BlockchainTxID     string    `json:"blockchainTxId" db:"blockchain_tx_id"`
	Status             string    `json:"status" db:"status"`
	CreatedBy          int64     `json:"createdBy" db:"created_by"`
	CreatedAt          time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time `json:"updatedAt" db:"updated_at"`

	// 关联数据
	Customer Customer `json:"customer"`
}

// Customer 客户模型
type Customer struct {
	ID              int64     `json:"id" db:"id"`
	CustomerName    string    `json:"customerName" db:"customer_name"`
	CustomerAddress string    `json:"customerAddress" db:"customer_address"`
	ContactPerson   string    `json:"contactPerson" db:"contact_person"`
	ContactPhone    string    `json:"contactPhone" db:"contact_phone"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time `json:"updatedAt" db:"updated_at"`
}

// TestData 测试数据模型
type TestData struct {
	ID                int64     `json:"id" db:"id"`
	CertID            int64     `json:"certId" db:"cert_id"`
	CertNumber        string    `json:"certNumber"`
	DeviceAddr        string    `json:"deviceAddr" db:"device_addr"`
	DataType          string    `json:"dataType" db:"data_type"`
	PercentageValue   float64   `json:"percentageValue" db:"percentage_value"`
	RatioError        float64   `json:"ratioError" db:"ratio_error"`
	AngleError        float64   `json:"angleError" db:"angle_error"`
	CurrentValue      float64   `json:"currentValue" db:"current_value"`
	VoltageValue      float64   `json:"voltageValue" db:"voltage_value"`
	WorkstationNumber string    `json:"workstationNumber" db:"workstation_number"`
	TestPoint         string    `json:"testPoint" db:"test_point"`
	ActualPercentage  float64   `json:"actualPercentage" db:"actual_percentage"`
	TestTimestamp     time.Time `json:"testTimestamp" db:"test_timestamp"`
	BlockchainHash    string    `json:"blockchainHash" db:"blockchain_hash"`
	EncryptedData     string    `json:"-" db:"encrypted_data"` // 不序列化到JSON
	DecryptedData     string    `json:"decryptedData"`         // 解密后的数据
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
}

// User 用户模型
type User struct {
	ID           int64     `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"` // 不序列化密码
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// Device 设备模型
type Device struct {
	ID           int64     `json:"id" db:"id"`
	DeviceAddr   string    `json:"deviceAddr" db:"device_addr"`
	DeviceName   string    `json:"deviceName" db:"device_name"`
	Manufacturer string    `json:"manufacturer" db:"manufacturer"`
	Model        string    `json:"model" db:"model"`
	AccuracyClass string   `json:"accuracyClass" db:"accuracy_class"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// BlockchainTransaction 区块链交易模型
type BlockchainTransaction struct {
	ID              int64     `json:"id" db:"id"`
	TxID            string    `json:"txId" db:"tx_id"`
	BlockNumber     int64     `json:"blockNumber" db:"block_number"`
	TransactionHash string    `json:"transactionHash" db:"transaction_hash"`
	CertID          int64     `json:"certId" db:"cert_id"`
	OperationType   string    `json:"operationType" db:"operation_type"`
	Status          string    `json:"status" db:"status"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
}

// 区块链相关模型

// BlockchainCertificate 区块链证书模型
type BlockchainCertificate struct {
	CertNumber         string `json:"certNumber"`
	CustomerName       string `json:"customerName"`
	CustomerAddress    string `json:"customerAddress"`
	InstrumentName     string `json:"instrumentName"`
	Manufacturer       string `json:"manufacturer"`
	ModelSpec          string `json:"modelSpec"`
	InstrumentNumber   string `json:"instrumentNumber"`
	InstrumentAccuracy string `json:"instrumentAccuracy"`
	TestDate           string `json:"testDate"`
	ExpireDate         string `json:"expireDate"`
	TestResult         string `json:"testResult"`
	Status             string `json:"status"`
}

// BlockchainTestData 区块链测试数据模型
type BlockchainTestData struct {
	CertNumber         string  `json:"certNumber"`
	DeviceAddr         string  `json:"deviceAddr"`
	DataType           string  `json:"dataType"`
	PercentageValue    float64 `json:"percentageValue"`
	RatioError         float64 `json:"ratioError"`
	AngleError         float64 `json:"angleError"`
	CurrentValue       float64 `json:"currentValue"`
	VoltageValue       float64 `json:"voltageValue"`
	WorkstationNumber  string  `json:"workstationNumber"`
	TestPoint          string  `json:"testPoint"`
	ActualPercentage   float64 `json:"actualPercentage"`
	TestTimestamp      string  `json:"testTimestamp"`
	EncryptedData      string  `json:"encryptedData"`
}

// API相关模型

// APIResponse 通用API响应
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
	CertNumber  string       `json:"certNumber"`
	IsValid     bool         `json:"isValid"`
	Message     string       `json:"message"`
	VerifiedAt  time.Time    `json:"verifiedAt"`
	Certificate *Certificate `json:"certificate,omitempty"`
}

// HistoryRecord 历史记录
type HistoryRecord struct {
	TxID      string      `json:"txId"`
	Value     interface{} `json:"value"`
	Timestamp string      `json:"timestamp"`
	IsDelete  bool        `json:"isDelete"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Role     string `json:"role"`
	ExpiresAt time.Time `json:"expiresAt"`
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
	TestDate           string `json:"testDate" binding:"required"`
	ExpireDate         string `json:"expireDate"`
	TestResult         string `json:"testResult"`
}

// AddTestDataRequest 添加测试数据请求
type AddTestDataRequest struct {
	CertNumber         string  `json:"certNumber" binding:"required"`
	DeviceAddr         string  `json:"deviceAddr" binding:"required"`
	DataType           string  `json:"dataType" binding:"required"`
	PercentageValue    float64 `json:"percentageValue"`
	RatioError         float64 `json:"ratioError"`
	AngleError         float64 `json:"angleError"`
	CurrentValue       float64 `json:"currentValue"`
	VoltageValue       float64 `json:"voltageValue"`
	WorkstationNumber  string  `json:"workstationNumber"`
	TestPoint          string  `json:"testPoint"`
	ActualPercentage   float64 `json:"actualPercentage"`
}

// BatchTestDataRequest 批量测试数据请求
type BatchTestDataRequest struct {
	CertNumber string              `json:"certNumber" binding:"required"`
	TestData   []AddTestDataRequest `json:"testData" binding:"required"`
}
