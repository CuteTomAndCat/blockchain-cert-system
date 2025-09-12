package api

import (
    "cert-system/internal/service"
    "cert-system/internal/models"
    "github.com/gin-gonic/gin"
    "net/http"
    "time"
)

// TestDataHandler 测试数据处理器
type TestDataHandler struct {
    testDataService *service.TestDataService
    certService     *service.CertificateService
}

// NewTestDataHandler 创建 TestDataHandler 实例
func NewTestDataHandler(testDataService *service.TestDataService, certService *service.CertificateService) *TestDataHandler {
    return &TestDataHandler{
        testDataService: testDataService,
        certService:     certService,
    }
}

// AddTestData 批量添加测试点数据
func (h *TestDataHandler) AddTestData(c *gin.Context) {
    var req models.AddTestDataRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Code:    400,
            Message: "请求参数错误: " + err.Error(),
        })
        return
    }

    // 查证书
    cert, err := h.certService.GetCertificateByNumber(req.CertNumber)
    if err != nil || cert == nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Code:    400,
            Message: "证书不存在: " + req.CertNumber,
        })
        return
    }

    var dataList []*models.TestData
    for _, d := range req.Data {
        t, err := time.Parse(time.RFC3339, d.TestTimestamp)
        if err != nil {
            c.JSON(http.StatusBadRequest, models.APIResponse{
                Code:    400,
                Message: "testTimestamp 格式错误: " + err.Error(),
            })
            return
        }

        dataList = append(dataList, &models.TestData{
            CertID:           cert.ID,
            CertNumber:       req.CertNumber,
            DeviceAddr:       d.DeviceAddr,
            TestPoint:        d.TestPoint,
            ActualPercentage: d.ActualPercentage,
            RatioError:       d.RatioError,
            AngleError:       d.AngleError,
            TestTimestamp:    t,
        })
    }

    if err := h.testDataService.BatchAddTestData(dataList); err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Code:    500,
            Message: "添加测试数据失败: " + err.Error(),
        })
        return
    }

    c.JSON(http.StatusCreated, models.APIResponse{
        Code:    201,
        Message: "测试数据批量添加成功",
        Data:    dataList,
    })
}

// GetTestDataByCert 获取证书相关的测试数据
func (h *TestDataHandler) GetTestDataByCert(c *gin.Context) {
    certNumber := c.Param("certId")
    if certNumber == "" {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Code:    400,
            Message: "证书编号不能为空",
        })
        return
    }

    cert, err := h.certService.GetCertificateByNumber(certNumber)
    if err != nil || cert == nil {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Code:    404,
            Message: "证书不存在: " + certNumber,
        })
        return
    }

    data, err := h.testDataService.GetTestDataByCertId(cert.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Code:    500,
            Message: "获取测试数据失败: " + err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, models.APIResponse{
        Code:    200,
        Message: "获取测试数据成功",
        Data:    data,
    })
}

// GenerateTestData 生成测试数据（占位逻辑）
func (h *TestDataHandler) GenerateTestData(c *gin.Context) {
    c.JSON(http.StatusOK, models.APIResponse{
        Code:    200,
        Message: "生成测试数据成功（占位）",
    })
}
