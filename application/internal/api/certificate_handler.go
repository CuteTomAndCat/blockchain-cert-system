package api

import (
	"cert-system/internal/models"
	"cert-system/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CertificateHandler 证书处理器
type CertificateHandler struct {
	certService *service.CertificateService
}

// NewCertificateHandler 创建证书处理器
func NewCertificateHandler(certService *service.CertificateService) *CertificateHandler {
	return &CertificateHandler{
		certService: certService,
	}
}

// CreateCertificate 创建证书
func (h *CertificateHandler) CreateCertificate(c *gin.Context) {
	var req models.CreateCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 解析日期
	testDate, err := time.Parse("2006-01-02", req.TestDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "测试日期格式错误",
		})
		return
	}

	var expireDate time.Time
	if req.ExpireDate != "" {
		expireDate, err = time.Parse("2006-01-02", req.ExpireDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Code:    400,
				Message: "有效期日期格式错误",
			})
			return
		}
	} else {
		// 默认有效期为测试日期后3年
		expireDate = testDate.AddDate(3, 0, 0)
	}

	// 获取当前用户ID
	userID, _ := c.Get("userID")

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

	err = h.certService.CreateCertificate(cert)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "创建证书失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "证书创建成功",
		Data:    cert,
	})
}

// GetAllCertificates 获取所有证书
func (h *CertificateHandler) GetAllCertificates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	certificates, total, err := h.certService.GetAllCertificates(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "获取证书列表失败: " + err.Error(),
		})
		return
	}

	totalPages := (total + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, models.PagedResponse{
		Code:       200,
		Message:    "获取成功",
		Data:       certificates,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetCertificate 获取单个证书
func (h *CertificateHandler) GetCertificate(c *gin.Context) {
	certNumber := c.Param("certNumber")
	if certNumber == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "证书编号不能为空",
		})
		return
	}

	cert, err := h.certService.GetCertificate(certNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Code:    404,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "获取成功",
		Data:    cert,
	})
}

// UpdateCertificate 更新证书
func (h *CertificateHandler) UpdateCertificate(c *gin.Context) {
	certNumber := c.Param("certNumber")
	if certNumber == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "证书编号不能为空",
		})
		return
	}

	var req models.CreateCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取现有证书
	cert, err := h.certService.GetCertificate(certNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Code:    404,
			Message: err.Error(),
		})
		return
	}

	// 解析日期
	testDate, err := time.Parse("2006-01-02", req.TestDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "测试日期格式错误",
		})
		return
	}

	var expireDate time.Time
	if req.ExpireDate != "" {
		expireDate, err = time.Parse("2006-01-02", req.ExpireDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Code:    400,
				Message: "有效期日期格式错误",
			})
			return
		}
	} else {
		expireDate = testDate.AddDate(3, 0, 0)
	}

	// 更新证书信息
	cert.CustomerID = req.CustomerID
	cert.InstrumentName = req.InstrumentName
	cert.InstrumentNumber = req.InstrumentNumber
	cert.Manufacturer = req.Manufacturer
	cert.ModelSpec = req.ModelSpec
	cert.InstrumentAccuracy = req.InstrumentAccuracy
	cert.TestDate = testDate
	cert.ExpireDate = expireDate
	cert.TestResult = req.TestResult

	err = h.certService.UpdateCertificate(cert)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "更新证书失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "证书更新成功",
		Data:    cert,
	})
}

// VerifyCertificate 验证证书（需要认证）
func (h *CertificateHandler) VerifyCertificate(c *gin.Context) {
	certNumber := c.Param("certNumber")
	if certNumber == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "证书编号不能为空",
		})
		return
	}

	result, err := h.certService.VerifyCertificate(certNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "验证失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "验证完成",
		Data:    result,
	})
}

// PublicVerifyCertificate 公开验证证书（无需认证）
func (h *CertificateHandler) PublicVerifyCertificate(c *gin.Context) {
	certNumber := c.Param("certNumber")
	if certNumber == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "证书编号不能为空",
		})
		return
	}

	result, err := h.certService.VerifyCertificate(certNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "验证失败: " + err.Error(),
		})
		return
	}

	// 公开验证只返回基本验证信息，不返回详细证书内容
	publicResult := map[string]interface{}{
		"certNumber": result.CertNumber,
		"isValid":    result.IsValid,
		"message":    result.Message,
		"verifiedAt": result.VerifiedAt,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "验证完成",
		Data:    publicResult,
	})
}

// GetCertificateHistory 获取证书变更历史
func (h *CertificateHandler) GetCertificateHistory(c *gin.Context) {
	certNumber := c.Param("certNumber")
	if certNumber == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "证书编号不能为空",
		})
		return
	}

	history, err := h.certService.GetCertificateHistory(certNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "获取历史记录失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "获取成功",
		Data:    history,
	})
}
