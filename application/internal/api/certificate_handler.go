package api

import (
	"cert-system/internal/service"
	"cert-system/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
	"errors"
)

// CertificateHandler 证书处理器
type CertificateHandler struct {
	certService *service.CertificateService
}

// NewCertificateHandler 创建新的 CertificateHandler
func NewCertificateHandler(certService *service.CertificateService) *CertificateHandler {
	return &CertificateHandler{
		certService: certService,
	}
}

// parseDate 辅助函数，用于解析日期字符串
func parseDate(dateStr string) (time.Time, error) {
	layouts := []string{"2006-01-02", "2006/01/02"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("无法解析日期: " + dateStr + ". 支持格式如 YYYY-MM-DD 或 YYYY-MM-DDTHH:MM:SSZ")
}

// CreateCertificate 创建证书
func (h *CertificateHandler) CreateCertificate(c *gin.Context) {
	var req models.CreateCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "请求参数错误: " + err.Error()})
		return
	}

	// 转换日期格式
	testDate, err := parseDate(req.TestDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "testDate格式错误: " + err.Error()})
		return
	}
	expireDate, err := parseDate(req.ExpireDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "expireDate格式错误: " + err.Error()})
		return
	}

	// 验证 testResult 是否为有效枚举值
	if req.TestResult != "qualified" && req.TestResult != "unqualified" {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "testResult 必须是 'qualified' 或 'unqualified'"})
		return
	}

	// 从 JWT 中获取当前用户 ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{Code: 401, Message: "未找到用户信息"})
		return
	}

	cert := &models.Certificate{
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
		CreatedBy:          userID.(int64), 
	}

	if err := h.certService.CreateCertificate(cert); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Code: 500, Message: "创建证书失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{Code: 201, Message: "证书创建成功", Data: cert})
}

// GetAllCertificates 获取所有证书（支持分页）
func (h *CertificateHandler) GetAllCertificates(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	certs, total, err := h.certService.GetAllCertificates(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Code: 500, Message: "获取证书列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.PagedResponse{
		Code:       200,
		Message:    "获取证书列表成功",
		Data:       certs,
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (int(total) + pageSize - 1) / pageSize,
	})
}

// GetCertificate 根据证书编号获取证书
func (h *CertificateHandler) GetCertificate(c *gin.Context) {
	certNumber := c.Param("certNumber")
	if certNumber == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "证书编号不能为空"})
		return
	}

	cert, err := h.certService.GetCertificateByNumber(certNumber)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.APIResponse{Code: 404, Message: "证书未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{Code: 500, Message: "获取证书失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Code: 200, Message: "获取证书成功", Data: cert})
}

// UpdateCertificate 更新证书
func (h *CertificateHandler) UpdateCertificate(c *gin.Context) {
	certNumber := c.Param("certNumber")
	if certNumber == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "证书编号不能为空"})
		return
	}

	var updatedCertData struct {
		CertNumber         string `json:"certNumber"`
		CustomerID         int64  `json:"customerId"`
		InstrumentName     string `json:"instrumentName"`
		InstrumentNumber   string `json:"instrumentNumber"`
		Manufacturer       string `json:"manufacturer"`
		ModelSpec          string `json:"modelSpec"`
		InstrumentAccuracy string `json:"instrumentAccuracy"`
		TestDate           string `json:"testDate" binding:"required"`
		ExpireDate         string `json:"expireDate"`
		TestResult         string `json:"testResult" binding:"required"`
		Status             string `json:"status"`
	}

	if err := c.ShouldBindJSON(&updatedCertData); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "请求参数错误: " + err.Error()})
		return
	}

	// 转换日期格式
	testDate, err := parseDate(updatedCertData.TestDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "testDate格式错误: " + err.Error()})
		return
	}
	expireDate, err := parseDate(updatedCertData.ExpireDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "expireDate格式错误: " + err.Error()})
		return
	}

	// 验证 testResult 是否为有效枚举值
	if updatedCertData.TestResult != "qualified" && updatedCertData.TestResult != "unqualified" {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "testResult 必须是 'qualified' 或 'unqualified'"})
		return
	}

	existingCert, err := h.certService.GetCertificateByNumber(certNumber)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.APIResponse{Code: 404, Message: "证书未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{Code: 500, Message: "获取证书失败: " + err.Error()})
		return
	}

	// 更新字段
	existingCert.CertNumber = updatedCertData.CertNumber
	existingCert.CustomerID = updatedCertData.CustomerID
	existingCert.InstrumentName = updatedCertData.InstrumentName
	existingCert.InstrumentNumber = updatedCertData.InstrumentNumber
	existingCert.Manufacturer = updatedCertData.Manufacturer
	existingCert.ModelSpec = updatedCertData.ModelSpec
	existingCert.InstrumentAccuracy = updatedCertData.InstrumentAccuracy
	existingCert.TestDate = testDate
	existingCert.ExpireDate = expireDate
	existingCert.TestResult = updatedCertData.TestResult
	existingCert.Status = updatedCertData.Status
	existingCert.UpdatedAt = time.Now() // ✅ 自动更新时间

	if err := h.certService.UpdateCertificate(existingCert); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Code: 500, Message: "更新证书失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Code: 200, Message: "证书更新成功", Data: existingCert})
}

// DeleteCertificate 删除证书
func (h *CertificateHandler) DeleteCertificate(c *gin.Context) {
	certNumber := c.Param("certNumber")
	if certNumber == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "证书编号不能为空"})
		return
	}

	if err := h.certService.DeleteCertificateByNumber(certNumber); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Code: 500, Message: "删除证书失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Code: 200, Message: "证书删除成功"})
}

// VerifyCertificate 验证证书
func (h *CertificateHandler) VerifyCertificate(c *gin.Context) {
	certNumber := c.Param("certNumber")
	if certNumber == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "证书编号不能为空"})
		return
	}

	cert, err := h.certService.VerifyCertificate(certNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Code: 500, Message: "验证失败: " + err.Error()})
		return
	}

	var isValid bool
	var message string
	if cert != nil {
		isValid = true
		message = "证书存在且有效"
	} else {
		isValid = false
		message = "证书不存在"
	}

	result := models.VerificationResult{
		CertNumber:  certNumber,
		IsValid:     isValid,
		Message:     message,
		VerifiedAt:  time.Now(),
		Certificate: cert,
	}

	c.JSON(http.StatusOK, models.APIResponse{Code: 200, Message: "验证结果", Data: result})
}

// PublicVerifyCertificate 公开验证接口
func (h *CertificateHandler) PublicVerifyCertificate(c *gin.Context) {
	h.VerifyCertificate(c)
}

// GetCertificateHistory 获取证书历史
func (h *CertificateHandler) GetCertificateHistory(c *gin.Context) {
	certNumber := c.Param("certNumber")
	if certNumber == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "证书编号不能为空"})
		return
	}

	history, err := h.certService.GetCertificateHistory(certNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Code: 500, Message: "获取历史记录失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Code: 200, Message: "获取历史记录成功", Data: history})
}
