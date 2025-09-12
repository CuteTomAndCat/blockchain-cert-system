package service

import (
	"cert-system/internal/database"
	"cert-system/internal/models"
	"gorm.io/gorm"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// CertificateService 证书服务
type CertificateService struct {
	dbClient *database.Client
}

// NewCertificateService 创建新的 CertificateService
func NewCertificateService(dbClient *database.Client) *CertificateService {
	return &CertificateService{
		dbClient: dbClient,
	}
}

// CreateCertificate 创建证书
// CreateCertificate 创建证书
func (s *CertificateService) CreateCertificate(cert *models.Certificate) error {
	// 先保存到数据库
	result := s.dbClient.DB.Create(cert)
	if result.Error != nil {
		return result.Error
	}
	
	// 生成区块链哈希（使用证书数据计算SHA256）
	hashData := fmt.Sprintf("%s|%d|%s|%s|%s", 
		cert.CertNumber, 
		cert.CustomerID, 
		cert.InstrumentName,
		cert.TestDate.Format("2006-01-02"),
		cert.TestResult)
	
	hash := sha256.Sum256([]byte(hashData))
	cert.BlockchainTxID = hex.EncodeToString(hash[:16]) // 使用前16字节作为简短ID
	cert.BlockchainHash = hex.EncodeToString(hash[:])   // 完整哈希
	
	// 更新数据库中的区块链信息
	updates := map[string]interface{}{
		"blockchain_tx_id": cert.BlockchainTxID,
		"blockchain_hash": cert.BlockchainHash,
	}
	
	s.dbClient.DB.Model(cert).Updates(updates)
	
	// 如果有Fabric客户端，这里调用链码
	// if s.fabricClient != nil {
	//     txID, err := s.fabricClient.CreateCertificate(cert)
	//     if err == nil {
	//         cert.BlockchainTxID = txID
	//         s.dbClient.DB.Model(cert).Update("blockchain_tx_id", txID)
	//     }
	// }
	
	return nil
}

// GetCertificateByNumber 根据证书编号获取证书
func (s *CertificateService) GetCertificateByNumber(certNumber string) (*models.Certificate, error) {
	var cert models.Certificate
	result := s.dbClient.DB.Where("cert_number = ?", certNumber).First(&cert)
	if result.Error != nil {
		return nil, result.Error
	}
	return &cert, nil
}

// GetAllCertificates 获取所有证书（支持分页）
func (s *CertificateService) GetAllCertificates(page, pageSize int) ([]*models.Certificate, int64, error) {
	var certificates []*models.Certificate
	var total int64
	offset := (page - 1) * pageSize

	// Count total records
	if err := s.dbClient.DB.Model(&models.Certificate{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated records
	result := s.dbClient.DB.Offset(offset).Limit(pageSize).Find(&certificates)
	return certificates, total, result.Error
}

// UpdateCertificate 更新证书信息
func (s *CertificateService) UpdateCertificate(cert *models.Certificate) error {
	result := s.dbClient.DB.Save(cert)
	return result.Error
}

// DeleteCertificateByNumber 删除证书
func (s *CertificateService) DeleteCertificateByNumber(certNumber string) error {
	result := s.dbClient.DB.Where("cert_number = ?", certNumber).Delete(&models.Certificate{})
	return result.Error
}

// VerifyCertificate 验证证书
func (s *CertificateService) VerifyCertificate(certNumber string) (*models.CertificateVerification, error) {
    var cert models.Certificate
    result := s.dbClient.DB.Where("cert_number = ?", certNumber).First(&cert)
    
    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            return &models.CertificateVerification{
                IsValid: false,
                Message: "证书不存在",
            }, nil
        }
        return nil, result.Error
    }
    
    // 检查证书状态
    isValid := true
    message := "证书有效"
    
    if cert.Status == "revoked" {
        isValid = false
        message = "证书已撤销"
    } else if cert.ExpireDate.Before(time.Now()) {
        isValid = false
        message = "证书已过期"
    }
    
    // 验证区块链哈希
    hashData := fmt.Sprintf("%s|%d|%s|%s|%s", 
        cert.CertNumber, 
        cert.CustomerID, 
        cert.InstrumentName,
        cert.TestDate.Format("2006-01-02"),
        cert.TestResult)
    
    hash := sha256.Sum256([]byte(hashData))
    currentHash := hex.EncodeToString(hash[:])
    isHashValid := (currentHash == cert.BlockchainHash)
    
    return &models.CertificateVerification{
        Certificate:    &cert,
        IsValid:        isValid,
        IsHashValid:    isHashValid,
        BlockchainTxID: cert.BlockchainTxID,
        BlockchainHash: cert.BlockchainHash,
        Message:        message,
        VerifiedAt:     time.Now(),
    }, nil
}

// GetCertificateHistory 获取证书历史（占位符，需要根据实际业务逻辑实现）
func (s *CertificateService) GetCertificateHistory(certNumber string) ([]*models.HistoryRecord, error) {
	// 这里是占位符，实际逻辑需要从区块链或数据库中查询历史记录
	// 暂时返回一个空切片，以保证编译通过
	return []*models.HistoryRecord{}, nil
}