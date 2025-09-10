package service

import (
	"cert-system/internal/database"
	"cert-system/internal/fabric"
	"cert-system/internal/models"
	"fmt"
	"time"

	"github.com/tjfoc/gmsm/sm4"
)

// TestDataService 测试数据服务
type TestDataService struct {
	db           *database.DB
	fabricClient *fabric.Client
	sm4Key       string
}

// NewTestDataService 创建测试数据服务实例
func NewTestDataService(db *database.DB, fabricClient *fabric.Client) *TestDataService {
	return &TestDataService{
		db:           db,
		fabricClient: fabricClient,
		sm4Key:       "1234567890123456", // 16字节密钥，实际应用中应从配置获取
	}
}

// AddTestData 添加测试数据
func (s *TestDataService) AddTestData(testData *models.TestData) error {
	// 开始数据库事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 对敏感数据进行国密SM4加密
	sensitiveData := fmt.Sprintf("%.6f|%.6f|%.6f|%.3f|%.3f",
		testData.PercentageValue, testData.RatioError, testData.AngleError,
		testData.CurrentValue, testData.VoltageValue)

	encryptedData, err := s.encryptWithSM4(sensitiveData)
	if err != nil {
		return fmt.Errorf("数据加密失败: %v", err)
	}

	// 在数据库中插入测试数据
	query := `INSERT INTO test_data (cert_id, device_addr, data_type, percentage_value,
		ratio_error, angle_error, current_value, voltage_value, workstation_number,
		test_point, actual_percentage, test_timestamp, encrypted_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := tx.Exec(query,
		testData.CertID, testData.DeviceAddr, testData.DataType, testData.PercentageValue,
		testData.RatioError, testData.AngleError, testData.CurrentValue, testData.VoltageValue,
		testData.WorkstationNumber, testData.TestPoint, testData.ActualPercentage,
		testData.TestTimestamp, encryptedData)

	if err != nil {
		return fmt.Errorf("数据库插入失败: %v", err)
	}

	testDataID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取插入ID失败: %v", err)
	}
	testData.ID = testDataID

	// 准备区块链测试数据
	blockchainTestData := models.BlockchainTestData{
		CertNumber:         testData.CertNumber,
		DeviceAddr:         testData.DeviceAddr,
		DataType:           testData.DataType,
		PercentageValue:    testData.PercentageValue,
		RatioError:         testData.RatioError,
		AngleError:         testData.AngleError,
		CurrentValue:       testData.CurrentValue,
		VoltageValue:       testData.VoltageValue,
		WorkstationNumber:  testData.WorkstationNumber,
		TestPoint:          testData.TestPoint,
		ActualPercentage:   testData.ActualPercentage,
		TestTimestamp:      testData.TestTimestamp.Format(time.RFC3339),
		EncryptedData:      encryptedData,
	}

	// 在区块链上添加测试数据
	err = s.fabricClient.AddTestData(blockchainTestData)
	if err != nil {
		return fmt.Errorf("区块链添加测试数据失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetTestDataByCert 根据证书获取测试数据
func (s *TestDataService) GetTestDataByCert(certID int64) ([]*models.TestData, error) {
	query := `SELECT id, cert_id, device_addr, data_type, percentage_value,
		ratio_error, angle_error, current_value, voltage_value, workstation_number,
		test_point, actual_percentage, test_timestamp, blockchain_hash, encrypted_data, created_at
		FROM test_data WHERE cert_id = ? ORDER BY test_timestamp DESC`

	rows, err := s.db.Query(query, certID)
	if err != nil {
		return nil, fmt.Errorf("查询测试数据失败: %v", err)
	}
	defer rows.Close()

	var testDataList []*models.TestData
	for rows.Next() {
		var testData models.TestData
		var encryptedData string
		err := rows.Scan(
			&testData.ID, &testData.CertID, &testData.DeviceAddr, &testData.DataType,
			&testData.PercentageValue, &testData.RatioError, &testData.AngleError,
			&testData.CurrentValue, &testData.VoltageValue, &testData.WorkstationNumber,
			&testData.TestPoint, &testData.ActualPercentage, &testData.TestTimestamp,
			&testData.BlockchainHash, &encryptedData, &testData.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描测试数据失败: %v", err)
		}

		// 解密敏感数据用于显示（在实际应用中可能需要权限控制）
		decryptedData, err := s.decryptWithSM4(encryptedData)
		if err == nil {
			testData.DecryptedData = decryptedData
		}

		testDataList = append(testDataList, &testData)
	}

	return testDataList, nil
}

// BatchAddTestData 批量添加测试数据
func (s *TestDataService) BatchAddTestData(testDataList []*models.TestData) error {
	// 开始数据库事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 准备批量插入语句
	query := `INSERT INTO test_data (cert_id, device_addr, data_type, percentage_value,
		ratio_error, angle_error, current_value, voltage_value, workstation_number,
		test_point, actual_percentage, test_timestamp, encrypted_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("准备语句失败: %v", err)
	}
	defer stmt.Close()

	// 批量插入并收集区块链数据
	var blockchainDataList []models.BlockchainTestData

	for _, testData := range testDataList {
		// 加密敏感数据
		sensitiveData := fmt.Sprintf("%.6f|%.6f|%.6f|%.3f|%.3f",
			testData.PercentageValue, testData.RatioError, testData.AngleError,
			testData.CurrentValue, testData.VoltageValue)

		encryptedData, err := s.encryptWithSM4(sensitiveData)
		if err != nil {
			return fmt.Errorf("数据加密失败: %v", err)
		}

		// 插入数据库
		_, err = stmt.Exec(
			testData.CertID, testData.DeviceAddr, testData.DataType, testData.PercentageValue,
			testData.RatioError, testData.AngleError, testData.CurrentValue, testData.VoltageValue,
			testData.WorkstationNumber, testData.TestPoint, testData.ActualPercentage,
			testData.TestTimestamp, encryptedData)

		if err != nil {
			return fmt.Errorf("批量插入失败: %v", err)
		}

		// 准备区块链数据
		blockchainData := models.BlockchainTestData{
			CertNumber:         testData.CertNumber,
			DeviceAddr:         testData.DeviceAddr,
			DataType:           testData.DataType,
			PercentageValue:    testData.PercentageValue,
			RatioError:         testData.RatioError,
			AngleError:         testData.AngleError,
			CurrentValue:       testData.CurrentValue,
			VoltageValue:       testData.VoltageValue,
			WorkstationNumber:  testData.WorkstationNumber,
			TestPoint:          testData.TestPoint,
			ActualPercentage:   testData.ActualPercentage,
			TestTimestamp:      testData.TestTimestamp.Format(time.RFC3339),
			EncryptedData:      encryptedData,
		}
		blockchainDataList = append(blockchainDataList, blockchainData)
	}

	// 批量提交到区块链
	for _, blockchainData := range blockchainDataList {
		err = s.fabricClient.AddTestData(blockchainData)
		if err != nil {
			return fmt.Errorf("区块链批量添加失败: %v", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GenerateTestData 生成模拟测试数据（用于测试和演示）
func (s *TestDataService) GenerateTestData(certID int64, certNumber string, count int) ([]*models.TestData, error) {
	var testDataList []*models.TestData

	// 模拟数据生成 - 互感器测试数据
	testPoints := []string{"5%额定电流", "20%额定电流", "100%额定电流", "120%额定电流"}
	deviceAddrs := []string{"DEV001", "DEV002", "DEV003"}
	workstations := []string{"WS001", "WS002", "WS003"}

	for i := 0; i < count; i++ {
		testData := &models.TestData{
			CertID:             certID,
			CertNumber:         certNumber,
			DeviceAddr:         deviceAddrs[i%len(deviceAddrs)],
			DataType:           "电流互感器测试",
			PercentageValue:    float64(100 + (i%5-2)*5), // 95-105范围
			RatioError:         float64((i%10-5)) * 0.01,  // -0.05到0.05范围
			AngleError:         float64((i%8-4)) * 0.5,    // -2到2范围
			CurrentValue:       float64(5 + i%10),         // 5-15A范围
			VoltageValue:       float64(220 + i%20),       // 220-240V范围
			WorkstationNumber:  workstations[i%len(workstations)],
			TestPoint:          testPoints[i%len(testPoints)],
			ActualPercentage:   float64(98 + i%5),         // 98-102范围
			TestTimestamp:      time.Now().Add(-time.Duration(i) * time.Minute),
		}

		testDataList = append(testDataList, testData)
	}

	return testDataList, nil
}

// encryptWithSM4 使用国密SM4算法加密数据
func (s *TestDataService) encryptWithSM4(plaintext string) (string, error) {
	keyBytes := []byte(s.sm4Key)
	plaintextBytes := []byte(plaintext)

	ciphertext, err := sm4.Sm4Ecb(keyBytes, plaintextBytes, true)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", ciphertext), nil
}

// decryptWithSM4 使用国密SM4算法解密数据
func (s *TestDataService) decryptWithSM4(encryptedData string) (string, error) {
	keyBytes := []byte(s.sm4Key)

	// 将十六进制字符串转换为字节数组
	var ciphertext []byte
	for i := 0; i < len(encryptedData); i += 2 {
		var b byte
		fmt.Sscanf(encryptedData[i:i+2], "%02x", &b)
		ciphertext = append(ciphertext, b)
	}

	plaintext, err := sm4.Sm4Ecb(keyBytes, ciphertext, false)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
