package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/tjfoc/gmsm/sm3"
	"github.com/tjfoc/gmsm/sm4"
)

// CertChaincode 计量证书链码结构
type CertChaincode struct {
	contractapi.Contract
}

// Certificate 证书结构体
type Certificate struct {
	CertNumber        string    `json:"certNumber"`        // 证书编号
	CustomerName      string    `json:"customerName"`      // 委托者
	CustomerAddress   string    `json:"customerAddress"`   // 委托者地址
	InstrumentName    string    `json:"instrumentName"`    // 器具名称
	Manufacturer      string    `json:"manufacturer"`      // 制造厂
	ModelSpec         string    `json:"modelSpec"`         // 型号规格
	InstrumentNumber  string    `json:"instrumentNumber"`  // 器具编号
	InstrumentAccuracy string   `json:"instrumentAccuracy"` // 器具准确度
	TestDate          string    `json:"testDate"`          // 检测日期
	ExpireDate        string    `json:"expireDate"`        // 有效期
	TestResult        string    `json:"testResult"`        // 检测结果
	Status            string    `json:"status"`            // 证书状态
	CreatedAt         string    `json:"createdAt"`         // 创建时间
	UpdatedAt         string    `json:"updatedAt"`         // 更新时间
	TestDataHash      string    `json:"testDataHash"`      // 测试数据哈希
	BlockchainTxID    string    `json:"blockchainTxId"`    // 区块链交易ID
	BlockchainHash    string    `json:"blockchainHash"`     // 区块链哈希
}

// TestData 测试数据结构体
type TestData struct {
	CertNumber       string  `json:"certNumber"`
	DeviceAddr       string  `json:"deviceAddr"`
	TestPoint        string  `json:"testPoint"`
	ActualPercentage float64 `json:"actualPercentage"`
	RatioError       float64 `json:"ratioError"`
	AngleError       float64 `json:"angleError"`
	TestTimestamp    string  `json:"testTimestamp"`
	EncryptedData    string  `json:"encryptedData"`
}

// QueryResult 查询结果结构体
type QueryResult struct {
	Key    string      `json:"Key"`
	Record interface{} `json:"Record"`
}

// HistoryQueryResult 历史查询结果
type HistoryQueryResult struct {
	TxId      string      `json:"TxId"`
	Value     interface{} `json:"Value"`
	Timestamp string      `json:"Timestamp"`
	IsDelete  bool        `json:"IsDelete"`
}

// InitLedger 初始化账本
func (c *CertChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	log.Println("计量证书防伪溯源系统链码初始化完成")
	return nil
}

// CreateCertificate 创建证书
func (c *CertChaincode) CreateCertificate(ctx contractapi.TransactionContextInterface, certData string) (string, error) {
	var cert Certificate
	err := json.Unmarshal([]byte(certData), &cert)
	if err != nil {
		return "", fmt.Errorf("证书数据解析失败: %v", err)
	}

	// 检查证书是否已存在
	exists, err := c.CertificateExists(ctx, cert.CertNumber)
	if err != nil {
		return "", err
	}
	if exists {
		return "", fmt.Errorf("证书 %s 已存在", cert.CertNumber)
	}

	// 获取交易ID
	txID := ctx.GetStub().GetTxID()
	
	// 设置创建时间和区块链交易ID
	cert.CreatedAt = time.Now().Format(time.RFC3339)
	cert.UpdatedAt = cert.CreatedAt
	cert.Status = "created"
	cert.BlockchainTxID = txID  // 添加这个字段

	certJSON, err := json.Marshal(cert)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(cert.CertNumber, certJSON)
	if err != nil {
		return "", err
	}
	
	return txID, nil  // 返回交易ID
}

// CertificateExists 检查证书是否存在
func (c *CertChaincode) CertificateExists(ctx contractapi.TransactionContextInterface, certNumber string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(certNumber)
	if err != nil {
		return false, fmt.Errorf("读取证书状态失败: %v", err)
	}

	return certJSON != nil, nil
}

// GetCertificate 获取证书信息
func (c *CertChaincode) GetCertificate(ctx contractapi.TransactionContextInterface, certNumber string) (*Certificate, error) {
	certJSON, err := ctx.GetStub().GetState(certNumber)
	if err != nil {
		return nil, fmt.Errorf("读取证书失败: %v", err)
	}
	if certJSON == nil {
		return nil, fmt.Errorf("证书 %s 不存在", certNumber)
	}

	var cert Certificate
	err = json.Unmarshal(certJSON, &cert)
	if err != nil {
		return nil, err
	}

	return &cert, nil
}

// UpdateCertificate 更新证书
func (c *CertChaincode) UpdateCertificate(ctx contractapi.TransactionContextInterface, certNumber string, certData string) error {
	exists, err := c.CertificateExists(ctx, certNumber)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("证书 %s 不存在", certNumber)
	}

	var cert Certificate
	err = json.Unmarshal([]byte(certData), &cert)
	if err != nil {
		return fmt.Errorf("证书数据解析失败: %v", err)
	}

	cert.CertNumber = certNumber
	cert.UpdatedAt = time.Now().Format(time.RFC3339)

	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(certNumber, certJSON)
}

// AddTestData 添加测试数据
func (c *CertChaincode) AddTestData(ctx contractapi.TransactionContextInterface, testDataStr string) error {
	var testData TestData
	err := json.Unmarshal([]byte(testDataStr), &testData)
	if err != nil {
		return fmt.Errorf("测试数据解析失败: %v", err)
	}

	// 验证证书是否存在
	exists, err := c.CertificateExists(ctx, testData.CertNumber)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("证书 %s 不存在", testData.CertNumber)
	}

	// 生成测试数据的唯一键
	testDataKey := fmt.Sprintf("TESTDATA_%s_%d", testData.CertNumber, time.Now().UnixNano())
	
	// 设置测试时间
	testData.TestTimestamp = time.Now().Format(time.RFC3339)

	// 对敏感数据进行国密SM4加密（这里使用示例密钥，实际应用中应使用安全的密钥管理）
	sensitiveData := fmt.Sprintf("%.6f|%.6f|%s", 
    	testData.ActualPercentage, testData.RatioError, testData.TestPoint)
	
	encryptedData, err := c.encryptWithSM4(sensitiveData, "1234567890123456") // 16字节密钥
	if err != nil {
		return fmt.Errorf("数据加密失败: %v", err)
	}
	testData.EncryptedData = encryptedData

	testDataJSON, err := json.Marshal(testData)
	if err != nil {
		return err
	}

	// 存储测试数据
	err = ctx.GetStub().PutState(testDataKey, testDataJSON)
	if err != nil {
		return err
	}

	// 更新证书的测试数据哈希
	return c.updateCertificateTestDataHash(ctx, testData.CertNumber, testDataKey)
}

// 国密SM4加密函数
func (c *CertChaincode) encryptWithSM4(plaintext, key string) (string, error) {
	keyBytes := []byte(key)
	plaintextBytes := []byte(plaintext)
	
	ciphertext, err := sm4.Sm4Ecb(keyBytes, plaintextBytes, true)
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%x", ciphertext), nil
}

// 更新证书的测试数据哈希
func (c *CertChaincode) updateCertificateTestDataHash(ctx contractapi.TransactionContextInterface, certNumber string, testDataKey string) error {
	cert, err := c.GetCertificate(ctx, certNumber)
	if err != nil {
		return err
	}

	// 使用国密SM3算法计算哈希
	hashData := fmt.Sprintf("%s|%s|%s", cert.TestDataHash, testDataKey, time.Now().Format(time.RFC3339))
	hash := sm3.Sm3Sum([]byte(hashData))
	cert.TestDataHash = fmt.Sprintf("%x", hash)
	cert.UpdatedAt = time.Now().Format(time.RFC3339)

	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(certNumber, certJSON)
}

// GetTestDataByCert 根据证书编号获取测试数据
func (c *CertChaincode) GetTestDataByCert(ctx contractapi.TransactionContextInterface, certNumber string) ([]*TestData, error) {
	queryString := fmt.Sprintf(`{"selector":{"certNumber":"%s"}}`, certNumber)
	return c.getTestDataByQuery(ctx, queryString)
}

// 通用查询测试数据方法
func (c *CertChaincode) getTestDataByQuery(ctx contractapi.TransactionContextInterface, queryString string) ([]*TestData, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var testDataList []*TestData
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var testData TestData
		err = json.Unmarshal(queryResponse.Value, &testData)
		if err != nil {
			return nil, err
		}
		testDataList = append(testDataList, &testData)
	}

	return testDataList, nil
}

// GetAllCertificates 获取所有证书
func (c *CertChaincode) GetAllCertificates(ctx contractapi.TransactionContextInterface) ([]*QueryResult, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var results []*QueryResult
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		// 只返回证书数据（不包含测试数据）
		if len(queryResponse.Key) > 0 && queryResponse.Key[0:9] != "TESTDATA_" {
			var cert Certificate
			err = json.Unmarshal(queryResponse.Value, &cert)
			if err != nil {
				return nil, err
			}

			queryResult := &QueryResult{
				Key:    queryResponse.Key,
				Record: cert,
			}
			results = append(results, queryResult)
		}
	}

	return results, nil
}

// VerifyCertificate 验证证书真伪
func (c *CertChaincode) VerifyCertificate(ctx contractapi.TransactionContextInterface, certNumber string) (bool, error) {
	cert, err := c.GetCertificate(ctx, certNumber)
	if err != nil {
		return false, fmt.Errorf("证书验证失败: %v", err)
	}

	// 这里可以添加更多验证逻辑，比如验证数字签名等
	return cert != nil && cert.Status != "revoked", nil
}

// GetCertificateHistory 获取证书变更历史
func (c *CertChaincode) GetCertificateHistory(ctx contractapi.TransactionContextInterface, certNumber string) ([]*HistoryQueryResult, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(certNumber)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var results []*HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var cert Certificate
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &cert)
			if err != nil {
				return nil, err
			}
		}

		timestamp := time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).Format(time.RFC3339)

		record := &HistoryQueryResult{
			TxId:      response.TxId,
			Value:     cert,
			Timestamp: timestamp,
			IsDelete:  response.IsDelete,
		}
		results = append(results, record)
	}

	return results, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&CertChaincode{})
	if err != nil {
		log.Panicf("创建链码时出错: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("启动链码时出错: %v", err)
	}
}