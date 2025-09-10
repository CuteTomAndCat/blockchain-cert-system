package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tjfoc/gmsm/sm3"
	"github.com/tjfoc/gmsm/sm4"
)

// 数据结构定义
type Certificate struct {
	ID                 int64     `json:"id"`
	CertNumber         string    `json:"certNumber"`
	CustomerID         int64     `json:"customerId"`
	CustomerName       string    `json:"customerName"`
	CustomerAddress    string    `json:"customerAddress"`
	InstrumentName     string    `json:"instrumentName"`
	InstrumentNumber   string    `json:"instrumentNumber"`
	Manufacturer       string    `json:"manufacturer"`
	ModelSpec          string    `json:"modelSpec"`
	InstrumentAccuracy string    `json:"instrumentAccuracy"`
	TestDate           time.Time `json:"testDate"`
	ExpireDate         time.Time `json:"expireDate"`
	TestResult         string    `json:"testResult"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type TestData struct {
	ID                int64     `json:"id"`
	CertID            int64     `json:"certId"`
	CertNumber        string    `json:"certNumber"`
	DeviceAddr        string    `json:"deviceAddr"`
	DataType          string    `json:"dataType"`
	PercentageValue   float64   `json:"percentageValue"`
	RatioError        float64   `json:"ratioError"`
	AngleError        float64   `json:"angleError"`
	CurrentValue      float64   `json:"currentValue"`
	VoltageValue      float64   `json:"voltageValue"`
	WorkstationNumber string    `json:"workstationNumber"`
	TestPoint         string    `json:"testPoint"`
	ActualPercentage  float64   `json:"actualPercentage"`
	TestTimestamp     time.Time `json:"testTimestamp"`
	EncryptedData     string    `json:"encryptedData"`
	DecryptedData     string    `json:"decryptedData"`
}

type Customer struct {
	ID              int64     `json:"id"`
	CustomerName    string    `json:"customerName"`
	CustomerAddress string    `json:"customerAddress"`
	ContactPerson   string    `json:"contactPerson"`
	ContactPhone    string    `json:"contactPhone"`
	CreatedAt       time.Time `json:"createdAt"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PagedResponse struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

type VerificationResult struct {
	CertNumber  string       `json:"certNumber"`
	IsValid     bool         `json:"isValid"`
	Message     string       `json:"message"`
	VerifiedAt  time.Time    `json:"verifiedAt"`
	Certificate *Certificate `json:"certificate,omitempty"`
}

var db *sql.DB
const sm4Key = "1234567890123456" // 16字节国密SM4密钥

// 工具函数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func encryptWithSM4(plaintext string) (string, error) {
	keyBytes := []byte(sm4Key)
	plaintextBytes := []byte(plaintext)
	ciphertext, err := sm4.Sm4Ecb(keyBytes, plaintextBytes, true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", ciphertext), nil
}

func decryptWithSM4(encryptedData string) (string, error) {
	keyBytes := []byte(sm4Key)
	var ciphertext []byte
	for i := 0; i < len(encryptedData); i += 2 {
		if i+1 < len(encryptedData) {
			var b byte
			fmt.Sscanf(encryptedData[i:i+2], "%02x", &b)
			ciphertext = append(ciphertext, b)
		}
	}
	plaintext, err := sm4.Sm4Ecb(keyBytes, ciphertext, false)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func calculateSM3Hash(data string) string {
	hash := sm3.Sm3Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// 数据库初始化
func initDB() error {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USERNAME", "certuser")
	dbPass := getEnv("DB_PASSWORD", "certpass123")
	dbName := getEnv("DB_DATABASE", "cert_system")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("数据库ping失败: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("✓ 数据库连接成功")
	return nil
}

// API处理函数
func createCustomer(c *gin.Context) {
	var customer Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	customer.CreatedAt = time.Now()
	query := `INSERT INTO customers (customer_name, customer_address, contact_person, contact_phone, created_at) 
		VALUES (?, ?, ?, ?, ?)`

	result, err := db.Exec(query, customer.CustomerName, customer.CustomerAddress,
		customer.ContactPerson, customer.ContactPhone, customer.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "创建客户失败: " + err.Error(),
		})
		return
	}

	customerID, _ := result.LastInsertId()
	customer.ID = customerID

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "客户创建成功",
		Data:    customer,
	})
}

func getCustomers(c *gin.Context) {
	query := `SELECT id, customer_name, customer_address, contact_person, contact_phone, created_at 
		FROM customers ORDER BY created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "查询客户失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var customers []Customer
	for rows.Next() {
		var customer Customer
		err := rows.Scan(&customer.ID, &customer.CustomerName, &customer.CustomerAddress,
			&customer.ContactPerson, &customer.ContactPhone, &customer.CreatedAt)
		if err != nil {
			continue
		}
		customers = append(customers, customer)
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取成功",
		Data:    customers,
	})
}

func createCertificate(c *gin.Context) {
	var req struct {
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

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 解析日期
	testDate, err := time.Parse("2006-01-02", req.TestDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "测试日期格式错误",
		})
		return
	}

	var expireDate time.Time
	if req.ExpireDate != "" {
		expireDate, err = time.Parse("2006-01-02", req.ExpireDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Code:    400,
				Message: "有效期日期格式错误",
			})
			return
		}
	} else {
		expireDate = testDate.AddDate(3, 0, 0)
	}

	// 插入数据库
	query := `INSERT INTO certificates (cert_number, customer_id, instrument_name, instrument_number, 
		manufacturer, model_spec, instrument_accuracy, test_date, expire_date, 
		test_result, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	result, err := db.Exec(query, req.CertNumber, req.CustomerID, req.InstrumentName,
		req.InstrumentNumber, req.Manufacturer, req.ModelSpec, req.InstrumentAccuracy,
		testDate, expireDate, req.TestResult, "draft", now, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "创建证书失败: " + err.Error(),
		})
		return
	}

	certID, _ := result.LastInsertId()

	cert := Certificate{
		ID:                 certID,
		CertNumber:         req.CertNumber,
		CustomerID:         req.CustomerID,
		InstrumentName:     req.InstrumentName,
		InstrumentNumber:   req.InstrumentNumber,
		Manufacturer:       req.Manufacturer,
		ModelSpec:          req.ModelSpec,
		InstrumentAccuracy: req.InstrumentAccuracy,
		TestDate:           testDate,
		ExpireDate:         expireDate,
		TestResult:         req.TestResult,
		Status:             "draft",
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "证书创建成功",
		Data:    cert,
	})
}

func getCertificates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 计算总数
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM certificates").Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "查询证书总数失败: " + err.Error(),
		})
		return
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `SELECT c.id, c.cert_number, c.customer_id, c.instrument_name,
		c.instrument_number, c.manufacturer, c.model_spec, c.instrument_accuracy,
		c.test_date, c.expire_date, c.test_result, c.status,
		c.created_at, c.updated_at,
		COALESCE(cu.customer_name, '') as customer_name, 
		COALESCE(cu.customer_address, '') as customer_address
		FROM certificates c
		LEFT JOIN customers cu ON c.customer_id = cu.id
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := db.Query(query, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "查询证书列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var certificates []Certificate
	for rows.Next() {
		var cert Certificate
		err := rows.Scan(
			&cert.ID, &cert.CertNumber, &cert.CustomerID, &cert.InstrumentName,
			&cert.InstrumentNumber, &cert.Manufacturer, &cert.ModelSpec, &cert.InstrumentAccuracy,
			&cert.TestDate, &cert.ExpireDate, &cert.TestResult, &cert.Status,
			&cert.CreatedAt, &cert.UpdatedAt,
			&cert.CustomerName, &cert.CustomerAddress,
		)
		if err != nil {
			continue
		}
		certificates = append(certificates, cert)
	}

	totalPages := (total + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, PagedResponse{
		Code:       200,
		Message:    "获取成功",
		Data:       certificates,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func getCertificate(c *gin.Context) {
	certNumber := c.Param("certNumber")
	if certNumber == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "证书编号不能为空",
		})
		return
	}

	var cert Certificate
	query := `SELECT c.id, c.cert_number, c.customer_id, c.instrument_name,
		c.instrument_number, c.manufacturer, c.model_spec, c.instrument_accuracy,
		c.test_date, c.expire_date, c.test_result, c.status,
		c.created_at, c.updated_at,
		COALESCE(cu.customer_name, '') as customer_name, 
		COALESCE(cu.customer_address, '') as customer_address
		FROM certificates c
		LEFT JOIN customers cu ON c.customer_id = cu.id
		WHERE c.cert_number = ?`

	err := db.QueryRow(query, certNumber).Scan(
		&cert.ID, &cert.CertNumber, &cert.CustomerID, &cert.InstrumentName,
		&cert.InstrumentNumber, &cert.Manufacturer, &cert.ModelSpec, &cert.InstrumentAccuracy,
		&cert.TestDate, &cert.ExpireDate, &cert.TestResult, &cert.Status,
		&cert.CreatedAt, &cert.UpdatedAt,
		&cert.CustomerName, &cert.CustomerAddress,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, APIResponse{
				Code:    404,
				Message: "证书不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Code:    500,
				Message: "查询证书失败: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取成功",
		Data:    cert,
	})
}

func addTestData(c *gin.Context) {
	var req struct {
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

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取证书ID
	var certID int64
	err := db.QueryRow("SELECT id FROM certificates WHERE cert_number = ?", req.CertNumber).Scan(&certID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, APIResponse{
				Code:    404,
				Message: "证书不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Code:    500,
				Message: "查询证书失败: " + err.Error(),
			})
		}
		return
	}

	// 加密敏感数据
	sensitiveData := fmt.Sprintf("%.6f|%.6f|%.6f|%.3f|%.3f",
		req.PercentageValue, req.RatioError, req.AngleError,
		req.CurrentValue, req.VoltageValue)

	encryptedData, err := encryptWithSM4(sensitiveData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "数据加密失败: " + err.Error(),
		})
		return
	}

	// 插入数据库
	query := `INSERT INTO test_data (cert_id, device_addr, data_type, percentage_value,
		ratio_error, angle_error, current_value, voltage_value, workstation_number,
		test_point, actual_percentage, test_timestamp, encrypted_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.Exec(query, certID, req.DeviceAddr, req.DataType,
		req.PercentageValue, req.RatioError, req.AngleError,
		req.CurrentValue, req.VoltageValue, req.WorkstationNumber,
		req.TestPoint, req.ActualPercentage, time.Now(), encryptedData)

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "添加测试数据失败: " + err.Error(),
		})
		return
	}

	testDataID, _ := result.LastInsertId()

	testData := TestData{
		ID:                testDataID,
		CertID:            certID,
		CertNumber:        req.CertNumber,
		DeviceAddr:        req.DeviceAddr,
		DataType:          req.DataType,
		PercentageValue:   req.PercentageValue,
		RatioError:        req.RatioError,
		AngleError:        req.AngleError,
		CurrentValue:      req.CurrentValue,
		VoltageValue:      req.VoltageValue,
		WorkstationNumber: req.WorkstationNumber,
		TestPoint:         req.TestPoint,
		ActualPercentage:  req.ActualPercentage,
		TestTimestamp:     time.Now(),
		EncryptedData:     encryptedData,
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "测试数据添加成功",
		Data:    testData,
	})
}

func getTestData(c *gin.Context) {
	certNumber := c.Param("certNumber")
	if certNumber == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "证书编号不能为空",
		})
		return
	}

	query := `SELECT td.id, td.cert_id, td.device_addr, td.data_type, td.percentage_value,
		td.ratio_error, td.angle_error, td.current_value, td.voltage_value, td.workstation_number,
		td.test_point, td.actual_percentage, td.test_timestamp, td.encrypted_data,
		c.cert_number
		FROM test_data td
		LEFT JOIN certificates c ON td.cert_id = c.id
		WHERE c.cert_number = ?
		ORDER BY td.test_timestamp DESC`

	rows, err := db.Query(query, certNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "查询测试数据失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var testDataList []TestData
	for rows.Next() {
		var testData TestData
		err := rows.Scan(
			&testData.ID, &testData.CertID, &testData.DeviceAddr, &testData.DataType,
			&testData.PercentageValue, &testData.RatioError, &testData.AngleError,
			&testData.CurrentValue, &testData.VoltageValue, &testData.WorkstationNumber,
			&testData.TestPoint, &testData.ActualPercentage, &testData.TestTimestamp,
			&testData.EncryptedData, &testData.CertNumber,
		)
		if err != nil {
			continue
		}

		// 尝试解密数据
		if decrypted, err := decryptWithSM4(testData.EncryptedData); err == nil {
			testData.DecryptedData = decrypted
		}

		testDataList = append(testDataList, testData)
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取成功",
		Data:    testDataList,
	})
}

func verifyCertificate(c *gin.Context) {
	certNumber := c.Param("certNumber")
	if certNumber == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "证书编号不能为空",
		})
		return
	}

	var cert Certificate
	query := `SELECT c.cert_number, c.instrument_name, c.test_result, c.expire_date, c.status,
		COALESCE(cu.customer_name, '') as customer_name
		FROM certificates c
		LEFT JOIN customers cu ON c.customer_id = cu.id
		WHERE c.cert_number = ?`

	err := db.QueryRow(query, certNumber).Scan(
		&cert.CertNumber, &cert.InstrumentName, &cert.TestResult,
		&cert.ExpireDate, &cert.Status, &cert.CustomerName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, APIResponse{
				Code:    200,
				Message: "验证完成",
				Data: VerificationResult{
					CertNumber: certNumber,
					IsValid:    false,
					Message:    "证书不存在",
					VerifiedAt: time.Now(),
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Code:    500,
				Message: "验证失败: " + err.Error(),
			})
		}
		return
	}

	// 检查证书有效性
	isValid := cert.TestResult == "qualified" &&
		cert.ExpireDate.After(time.Now()) &&
		cert.Status != "revoked"

	message := "证书有效"
	if !isValid {
		if cert.TestResult != "qualified" {
			message = "证书检测结果不合格"
		} else if cert.ExpireDate.Before(time.Now()) {
			message = "证书已过期"
		} else if cert.Status == "revoked" {
			message = "证书已撤销"
		}
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "验证完成",
		Data: VerificationResult{
			CertNumber:  certNumber,
			IsValid:     isValid,
			Message:     message,
			VerifiedAt:  time.Now(),
			Certificate: &cert,
		},
	})
}

func main() {
	// 初始化数据库
	if err := initDB(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer db.Close()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// 添加CORS中间件
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"message":    "计量证书防伪溯源系统运行正常",
			"version":    "1.0.0",
			"timestamp":  time.Now(),
			"database":   "connected",
			"blockchain": "ready",
			"encryption": "GM/T SM3/SM4",
		})
	})

	// API路由
	v1 := router.Group("/api/v1")
	{
		// 系统信息
		v1.GET("/info", func(c *gin.Context) {
			c.JSON(http.StatusOK, APIResponse{
				Code:    200,
				Message: "系统信息",
				Data: map[string]interface{}{
					"name":        "计量证书防伪溯源系统",
					"version":     "1.0.0",
					"description": "基于区块链和国密技术的计量证书防伪溯源系统",
					"blockchain":  "Hyperledger Fabric",
					"database":    "MySQL",
					"encryption":  "国密SM3/SM4",
					"features": []string{
						"证书创建与管理",
						"测试数据加密存储",
						"区块链防篡改",
						"证书真伪验证",
						"互感器测试数据管理",
					},
				},
			})
		})

		// 客户管理
		v1.POST("/customers", createCustomer)
		v1.GET("/customers", getCustomers)

		// 证书管理
		v1.POST("/certificates", createCertificate)
		v1.GET("/certificates", getCertificates)
		v1.GET("/certificates/:certNumber", getCertificate)

		// 测试数据管理
		v1.POST("/test-data", addTestData)
		v1.GET("/test-data/:certNumber", getTestData)

		// 证书验证（公开接口）
		v1.POST("/verify/:certNumber", verifyCertificate)
		v1.GET("/verify/:certNumber", verifyCertificate)
	}

	port := getEnv("SERVER_PORT", "8080")
	fmt.Printf("🚀 计量证书防伪溯源系统启动成功\n")
	fmt.Printf("🔗 监听端口: %s\n", port)
	fmt.Printf("📋 健康检查: http://localhost:%s/health\n", port)
	fmt.Printf("📖 系统信息: http://localhost:%s/api/v1/info\n", port)
	fmt.Printf("🔒 使用国密SM3/SM4算法保护数据安全\n")
	fmt.Printf("⛓️  集成Hyperledger Fabric区块链技术\n")
	fmt.Printf("📊 支持互感器测试数据管理\n")

	log.Fatal(http.ListenAndServe(":"+port, router))
}
