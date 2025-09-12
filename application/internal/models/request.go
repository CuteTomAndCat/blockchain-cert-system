package models

// 单个测试点数据请求
type TestDataPointRequest struct {
    DeviceAddr       string  `json:"deviceAddr" binding:"required"`
    TestPoint        string  `json:"testPoint" binding:"required"`
    ActualPercentage float64 `json:"actualPercentage" binding:"required"`
    RatioError       float64 `json:"ratioError"`
    AngleError       float64 `json:"angleError"`
    TestTimestamp    string  `json:"testTimestamp" binding:"required"`
}

// 添加多条测试数据请求
type AddTestDataRequest struct {
    CertNumber string                 `json:"certNumber" binding:"required"`
    Data       []TestDataPointRequest `json:"data" binding:"required"` // ✅ 改为数组
}
