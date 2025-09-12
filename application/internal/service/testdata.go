package service

import (
	"cert-system/internal/database"
	"cert-system/internal/models"
)

// TestDataService 测试数据服务
type TestDataService struct {
	dbClient *database.Client
}

// NewTestDataService 创建新的 TestDataService
func NewTestDataService(dbClient *database.Client) *TestDataService {
	return &TestDataService{
		dbClient: dbClient,
	}
}

// AddTestData 添加单条测试数据
func (s *TestDataService) AddTestData(data *models.TestData) error {
	result := s.dbClient.DB.Create(data)
	return result.Error
}

// BatchAddTestData 批量添加测试数据
func (s *TestDataService) BatchAddTestData(data []*models.TestData) error {
	tx := s.dbClient.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, d := range data {
		if err := tx.Create(d).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetTestDataByCertId 根据证书ID获取所有测试数据
func (s *TestDataService) GetTestDataByCertId(certId int64) ([]*models.TestData, error) {
	var testData []*models.TestData
	result := s.dbClient.DB.Where("cert_id = ?", certId).Find(&testData)
	if result.Error != nil {
		return nil, result.Error
	}
	return testData, nil
}