package service

import (
	"cert-system/internal/database"
	"cert-system/internal/fabric"
	"cert-system/internal/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// CertificateService 证书服务
type CertificateService struct {
	db           *database.DB
	fabricClient *fabric.Client
}

// NewCertificateService 创建证书服务实例
func NewCertificateService(db *database.DB, fabricClient *fabric.Client) *CertificateService {
	return &CertificateService{
		db:           db,
		fabricClient: fabricClient,
	}
}

// CreateCertificate 创建证书
func (s *CertificateService) CreateCertificate(cert *models.Certificate) error {
	// 开始数据库事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 在数据库中创建证书记录
	query := `INSERT INTO certificates (cert_number, customer_id, instrument_name, 
		instrument_number, manufacturer, model_spec, instrument_accuracy, 
		test_date, expire_date, test_result, status, created_by) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := tx.Exec(query, cert.CertNumber, cert.CustomerID, cert.InstrumentName,
		cert.InstrumentNumber, cert.Manufacturer, cert.ModelSpec, cert.InstrumentAccuracy,
		cert.TestDate, cert.ExpireDate, cert.TestResult, cert.Status, cert.CreatedBy)
	
	if err != nil {
		return fmt.Errorf("数据库插入失败: %v", err)
	}

	certID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取插入ID失败: %v", err)
	}
	cert.ID = certID

	// 在区块链上创建证书记录
	blockchainCert := models.BlockchainCertificate{
		CertNumber:         cert.CertNumber,
		CustomerName:       cert.Customer.CustomerName,
		CustomerAddress:    cert.Customer.CustomerAddress,
		InstrumentName:     cert.InstrumentName,
		Manufacturer:       cert.Manufacturer,
		ModelSpec:          cert.ModelSpec,
		InstrumentNumber:   cert.InstrumentNumber,
		InstrumentAccuracy: cert.InstrumentAccuracy,
		TestDate:           cert.TestDate.Format("2006-01-02"),
		ExpireDate:         cert.ExpireDate.Format("2006-01-02"),
		TestResult:         cert.TestResult,
		Status:             cert.Status,
	}

	err = s.fabricClient.CreateCertificate(blockchainCert)
	if err != nil {
		return fmt.Errorf("区块链创建证书失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetCertificate 获取证书信息
func (s *CertificateService) GetCertificate(certNumber string) (*models.Certificate, error) {
	query := `SELECT c.id, c.cert_number, c.customer_id, c.instrument_name,
		c.instrument_number, c.manufacturer, c.model_spec, c.instrument_accuracy,
		c.test_date, c.expire_date, c.test_result, c.blockchain_tx_id, c.status,
		c.created_by, c.created_at, c.updated_at,
		cu.customer_name, cu.customer_address, cu.contact_person, cu.contact_phone
		FROM certificates c
		LEFT JOIN customers cu ON c.customer_id = cu.id
		WHERE c.cert_number = ?`

	var cert models.Certificate
	var customer models.Customer
	err := s.db.QueryRow(query, certNumber).Scan(
		&cert.ID, &cert.CertNumber, &cert.CustomerID, &cert.InstrumentName,
		&cert.InstrumentNumber, &cert.Manufacturer, &cert.ModelSpec, &cert.InstrumentAccuracy,
		&cert.TestDate, &cert.ExpireDate, &cert.TestResult, &cert.BlockchainTxID, &cert.Status,
		&cert.CreatedBy, &cert.CreatedAt, &cert.UpdatedAt,
		&customer.CustomerName, &customer.CustomerAddress, &customer.ContactPerson, &customer.ContactPhone,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("证书不存在")
		}
		return nil, fmt.Errorf("查询证书失败: %v", err)
	}

	cert.Customer = customer
	return &cert, nil
}

// GetAllCertificates 获取所有证书列表
func (s *CertificateService) GetAllCertificates(page, pageSize int) ([]*models.Certificate, int, error) {
	// 计算总数
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM certificates").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询证书总数失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `SELECT c.id, c.cert_number, c.customer_id, c.instrument_name,
		c.instrument_number, c.manufacturer, c.model_spec, c.instrument_accuracy,
		c.test_date, c.expire_date, c.test_result, c.blockchain_tx_id, c.status,
		c.created_by, c.created_at, c.updated_at,
		cu.customer_name, cu.customer_address
		FROM certificates c
		LEFT JOIN customers cu ON c.customer_id = cu.id
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询证书列表失败: %v", err)
	}
	defer rows.Close()

	var certificates []*models.Certificate
	for rows.Next() {
		var cert models.Certificate
		var customer models.Customer
		err := rows.Scan(
			&cert.ID, &cert.CertNumber, &cert.CustomerID, &cert.InstrumentName,
			&cert.InstrumentNumber, &cert.Manufacturer, &cert.ModelSpec, &cert.InstrumentAccuracy,
			&cert.TestDate, &cert.ExpireDate, &cert.TestResult, &cert.BlockchainTxID, &cert.Status,
			&cert.CreatedBy, &cert.CreatedAt, &cert.UpdatedAt,
			&customer.CustomerName, &customer.CustomerAddress,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描证书数据失败: %v", err)
		}

		cert.Customer = customer
		certificates = append(certificates, &cert)
	}

	return certificates, total, nil
}

// UpdateCertificate 更新证书
func (s *CertificateService) UpdateCertificate(cert *models.Certificate) error {
	// 开始数据库事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 更新数据库记录
	query := `UPDATE certificates SET customer_id=?, instrument_name=?,
		instrument_number=?, manufacturer=?, model_spec=?, instrument_accuracy=?,
		test_date=?, expire_date=?, test_result=?, status=?, updated_at=?
		WHERE cert_number=?`
	
	_, err = tx.Exec(query, cert.CustomerID, cert.InstrumentName,
		cert.InstrumentNumber, cert.Manufacturer, cert.ModelSpec, cert.InstrumentAccuracy,
		cert.TestDate, cert.ExpireDate, cert.TestResult, cert.Status, time.Now(),
		cert.CertNumber)
	
	if err != nil {
		return fmt.Errorf("数据库更新失败: %v", err)
	}

	// 在区块链上更新证书记录
	blockchainCert := models.BlockchainCertificate{
   	CertNumber:         cert.CertNumber,
   	CustomerName:       cert.Customer.CustomerName,
   	CustomerAddress:    cert.Customer.CustomerAddress,
   	InstrumentName:     cert.InstrumentName,
   	Manufacturer:       cert.Manufacturer,
   	ModelSpec:          cert.ModelSpec,
   	InstrumentNumber:   cert.InstrumentNumber,
   	InstrumentAccuracy: cert.InstrumentAccuracy,
   	TestDate:           cert.TestDate.Format("2006-01-02"),
   	ExpireDate:         cert.ExpireDate.Format("2006-01-02"),
   	TestResult:         cert.TestResult,
   	Status:             cert.Status,
   }

   certJSON, err := json.Marshal(blockchainCert)
   if err != nil {
   	return fmt.Errorf("证书序列化失败: %v", err)
   }

   _, err = s.fabricClient.InvokeChaincode("UpdateCertificate", [][]byte{[]byte(cert.CertNumber), certJSON})
   if err != nil {
   	return fmt.Errorf("区块链更新证书失败: %v", err)
   }

   // 提交事务
   if err := tx.Commit(); err != nil {
   	return fmt.Errorf("提交事务失败: %v", err)
   }

   return nil
}

// VerifyCertificate 验证证书真伪
func (s *CertificateService) VerifyCertificate(certNumber string) (*models.VerificationResult, error) {
   // 从区块链验证证书
   result, err := s.fabricClient.VerifyCertificate(certNumber)
   if err != nil {
   	return nil, fmt.Errorf("区块链验证失败: %v", err)
   }

   var isValid bool
   err = json.Unmarshal(result, &isValid)
   if err != nil {
   	return nil, fmt.Errorf("验证结果解析失败: %v", err)
   }

   // 从数据库获取证书信息进行交叉验证
   cert, err := s.GetCertificate(certNumber)
   if err != nil {
   	return &models.VerificationResult{
   		CertNumber: certNumber,
   		IsValid:    false,
   		Message:    "证书在数据库中不存在",
   		VerifiedAt: time.Now(),
   	}, nil
   }

   // 检查证书是否过期
   if cert.ExpireDate.Before(time.Now()) {
   	isValid = false
   }

   message := "证书有效"
   if !isValid {
   	message = "证书无效或已过期"
   }

   return &models.VerificationResult{
   	CertNumber: certNumber,
   	IsValid:    isValid,
   	Message:    message,
   	VerifiedAt: time.Now(),
   	Certificate: cert,
   }, nil
}

// GetCertificateHistory 获取证书变更历史
func (s *CertificateService) GetCertificateHistory(certNumber string) ([]models.HistoryRecord, error) {
   result, err := s.fabricClient.QueryChaincode("GetCertificateHistory", [][]byte{[]byte(certNumber)})
   if err != nil {
   	return nil, fmt.Errorf("获取证书历史失败: %v", err)
   }

   var history []models.HistoryRecord
   err = json.Unmarshal(result, &history)
   if err != nil {
   	return nil, fmt.Errorf("历史记录解析失败: %v", err)
   }

   return history, nil
}
