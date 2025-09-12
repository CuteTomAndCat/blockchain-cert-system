package service

import (
	"cert-system/internal/database"
	"cert-system/internal/models"
	"gorm.io/gorm"
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
func (s *CertificateService) CreateCertificate(cert *models.Certificate) error {
	result := s.dbClient.DB.Create(cert)
	return result.Error
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
func (s *CertificateService) VerifyCertificate(certNumber string) (*models.Certificate, error) {
	var cert models.Certificate
	result := s.dbClient.DB.Where("cert_number = ?", certNumber).First(&cert)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Return nil for both if not found
		}
		return nil, result.Error
	}
	return &cert, nil
}

// GetCertificateHistory 获取证书历史（占位符，需要根据实际业务逻辑实现）
func (s *CertificateService) GetCertificateHistory(certNumber string) ([]*models.HistoryRecord, error) {
	// 这里是占位符，实际逻辑需要从区块链或数据库中查询历史记录
	// 暂时返回一个空切片，以保证编译通过
	return []*models.HistoryRecord{}, nil
}