package fabric

import (
	"encoding/json"
	"fmt"
	certconfig "cert-system/config"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

// Client Fabric客户端结构
type Client struct {
	SDK           *fabsdk.FabricSDK
	ChannelClient *channel.Client
	ChannelName   string
	ChaincodeName string
}

// NewClient 创建新的Fabric客户端
func NewClient(cfg certconfig.FabricConfig) (*Client, error) {
	// 加载SDK配置
	configProvider := config.FromFile(cfg.ConfigPath)
	
	// 创建SDK实例
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		return nil, fmt.Errorf("创建Fabric SDK失败: %v", err)
	}

	// 创建通道客户端
	clientChannelContext := sdk.ChannelContext(cfg.ChannelName, fabsdk.WithUser(cfg.UserName), fabsdk.WithOrg(cfg.OrgName))
	channelClient, err := channel.New(clientChannelContext)
	if err != nil {
		return nil, fmt.Errorf("创建通道客户端失败: %v", err)
	}

	return &Client{
		SDK:           sdk,
		ChannelClient: channelClient,
		ChannelName:   cfg.ChannelName,
		ChaincodeName: cfg.ChaincodeName,
	}, nil
}

// InvokeChaincode 调用链码
func (c *Client) InvokeChaincode(function string, args [][]byte) ([]byte, error) {
	request := channel.Request{
		ChaincodeID: c.ChaincodeName,
		Fcn:         function,
		Args:        args,
	}

	response, err := c.ChannelClient.Execute(request)
	if err != nil {
		return nil, fmt.Errorf("链码调用失败: %v", err)
	}

	return response.Payload, nil
}

// QueryChaincode 查询链码
func (c *Client) QueryChaincode(function string, args [][]byte) ([]byte, error) {
	request := channel.Request{
		ChaincodeID: c.ChaincodeName,
		Fcn:         function,
		Args:        args,
	}

	response, err := c.ChannelClient.Query(request)
	if err != nil {
		return nil, fmt.Errorf("链码查询失败: %v", err)
	}

	return response.Payload, nil
}

// CreateCertificate 在区块链上创建证书
func (c *Client) CreateCertificate(cert interface{}) error {
	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	_, err = c.InvokeChaincode("CreateCertificate", [][]byte{certJSON})
	return err
}

// GetCertificate 从区块链获取证书
func (c *Client) GetCertificate(certNumber string) ([]byte, error) {
	return c.QueryChaincode("GetCertificate", [][]byte{[]byte(certNumber)})
}

// AddTestData 在区块链上添加测试数据
func (c *Client) AddTestData(testData interface{}) error {
	testDataJSON, err := json.Marshal(testData)
	if err != nil {
		return err
	}

	_, err = c.InvokeChaincode("AddTestData", [][]byte{testDataJSON})
	return err
}

// GetTestDataByCert 获取证书的测试数据
func (c *Client) GetTestDataByCert(certNumber string) ([]byte, error) {
	return c.QueryChaincode("GetTestDataByCert", [][]byte{[]byte(certNumber)})
}

// VerifyCertificate 验证证书
func (c *Client) VerifyCertificate(certNumber string) ([]byte, error) {
	return c.QueryChaincode("VerifyCertificate", [][]byte{[]byte(certNumber)})
}

// Close 关闭客户端
func (c *Client) Close() {
	if c.SDK != nil {
		c.SDK.Close()
	}
}
