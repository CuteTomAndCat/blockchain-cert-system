#!/bin/bash

# API测试脚本 - 基于main.go实现的接口测试

API_BASE="http://localhost:8080/api/v1"
TOKEN=""  # 实际项目中需要替换为有效的认证token获取逻辑

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 工具函数
print_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 测试健康检查接口
test_health() {
    print_test "测试健康检查接口..."
    response=$(curl -s "http://localhost:8080/health")
    if echo "$response" | grep -q "ok"; then
        print_success "健康检查通过"
    else
        print_error "健康检查失败"
        echo "$response"
    fi
}

# 测试系统信息接口
test_system_info() {
    print_test "测试系统信息接口..."
    response=$(curl -s "$API_BASE/info")
    if echo "$response" | grep -q '"code":200'; then
        print_success "系统信息获取成功"
    else
        print_error "系统信息获取失败"
        echo "$response"
    fi
}

# 测试创建客户接口
test_create_customer() {
    print_test "测试创建客户接口..."
    customer_name="测试客户_$(date +%s)"
    response=$(curl -s -X POST "$API_BASE/customers" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d '{
            "customerName": "'"$customer_name"'",
            "customerAddress": "测试地址",
            "contactPerson": "测试联系人",
            "contactPhone": "13800138000"
        }')
    
    if echo "$response" | grep -q '"code":200'; then
        print_success "客户创建成功"
        # 提取客户ID供后续测试使用
        export TEST_CUSTOMER_ID=$(echo "$response" | jq -r '.data.id')
    else
        print_error "客户创建失败"
        echo "$response"
    fi
}

# 测试获取客户列表接口
test_get_customers() {
    print_test "测试获取客户列表接口..."
    response=$(curl -s -X GET "$API_BASE/customers" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$response" | grep -q '"code":200'; then
        print_success "客户列表获取成功"
        count=$(echo "$response" | jq '.data | length')
        print_success "共获取到 $count 个客户"
    else
        print_error "客户列表获取失败"
        echo "$response"
    fi
}

# 测试创建证书接口
test_create_certificate() {
    print_test "测试创建证书接口..."
    
    if [ -z "$TEST_CUSTOMER_ID" ]; then
        print_error "没有可用的测试客户ID，跳过测试"
        return
    fi
    
    cert_number="CERT$(date +%Y%m%d%H%M%S)"
    test_date=$(date +%Y-%m-%d)
    expire_date=$(date -d "+3 years" +%Y-%m-%d)
    
    response=$(curl -s -X POST "$API_BASE/certificates" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"certNumber\": \"$cert_number\",
            \"customerId\": $TEST_CUSTOMER_ID,
            \"instrumentName\": \"电流互感器\",
            \"instrumentNumber\": \"CT-$(date +%s)\",
            \"manufacturer\": \"测试厂商\",
            \"modelSpec\": \"CT-1000/5A\",
            \"instrumentAccuracy\": \"0.2级\",
            \"testDate\": \"$test_date\",
            \"expireDate\": \"$expire_date\",
            \"testResult\": \"qualified\"
        }")
    
    if echo "$response" | grep -q '"code":200'; then
        print_success "证书创建成功: $cert_number"
        export TEST_CERT_NUMBER="$cert_number"
    else
        print_error "证书创建失败"
        echo "$response"
    fi
}

# 测试获取证书列表接口
test_get_certificates() {
    print_test "测试获取证书列表接口..."
    response=$(curl -s -X GET "$API_BASE/certificates?page=1&pageSize=10" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$response" | grep -q '"code":200'; then
        print_success "证书列表获取成功"
        total=$(echo "$response" | jq '.total')
        print_success "总证书数量: $total"
    else
        print_error "证书列表获取失败"
        echo "$response"
    fi
}

# 测试获取证书详情接口
test_get_certificate_detail() {
    print_test "测试获取证书详情接口..."
    
    if [ -z "$TEST_CERT_NUMBER" ]; then
        print_error "没有可用的测试证书编号，跳过测试"
        return
    fi
    
    response=$(curl -s -X GET "$API_BASE/certificates/$TEST_CERT_NUMBER" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$response" | grep -q '"code":200'; then
        print_success "证书详情获取成功"
    else
        print_error "证书详情获取失败"
        echo "$response"
    fi
}

# 测试添加测试数据接口
test_add_test_data() {
    print_test "测试添加测试数据接口..."
    
    if [ -z "$TEST_CERT_NUMBER" ]; then
        print_error "没有可用的测试证书编号，跳过测试"
        return
    fi
    
    response=$(curl -s -X POST "$API_BASE/test-data" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"certNumber\": \"$TEST_CERT_NUMBER\",
            \"deviceAddr\": \"DEV-$(date +%s)\",
            \"dataType\": \"电流互感器测试\",
            \"percentageValue\": 100.0,
            \"ratioError\": 0.02,
            \"angleError\": 1.5,
            \"currentValue\": 5.0,
            \"voltageValue\": 220.0,
            \"workstationNumber\": \"WS001\",
            \"testPoint\": \"100%额定电流\",
            \"actualPercentage\": 99.98
        }")
    
    if echo "$response" | grep -q '"code":200'; then
        print_success "测试数据添加成功"
    else
        print_error "测试数据添加失败"
        echo "$response"
    fi
}

# 测试获取测试数据接口
test_get_test_data() {
    print_test "测试获取测试数据接口..."
    
    if [ -z "$TEST_CERT_NUMBER" ]; then
        print_error "没有可用的测试证书编号，跳过测试"
        return
    fi
    
    response=$(curl -s -X GET "$API_BASE/test-data/$TEST_CERT_NUMBER" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$response" | grep -q '"code":200'; then
        print_success "测试数据获取成功"
        count=$(echo "$response" | jq '.data | length')
        print_success "共获取到 $count 条测试数据"
    else
        print_error "测试数据获取失败"
        echo "$response"
    fi
}

# 测试证书验证接口(GET方式)
test_verify_certificate_get() {
    print_test "测试证书验证接口(GET)..."
    
    if [ -z "$TEST_CERT_NUMBER" ]; then
        print_error "没有可用的测试证书编号，跳过测试"
        return
    fi
    
    response=$(curl -s -X GET "$API_BASE/verify/$TEST_CERT_NUMBER")
    
    if echo "$response" | grep -q '"isValid":true'; then
        print_success "证书验证成功(GET)"
    else
        print_error "证书验证失败(GET)"
        echo "$response"
    fi
}

# 测试证书验证接口(POST方式)
test_verify_certificate_post() {
    print_test "测试证书验证接口(POST)..."
    
    if [ -z "$TEST_CERT_NUMBER" ]; then
        print_error "没有可用的测试证书编号，跳过测试"
        return
    fi
    
    response=$(curl -s -X POST "$API_BASE/verify/$TEST_CERT_NUMBER")
    
    if echo "$response" | grep -q '"isValid":true'; then
        print_success "证书验证成功(POST)"
    else
        print_error "证书验证失败(POST)"
        echo "$response"
    fi
}

# 运行所有测试
run_all_tests() {
    echo "开始运行API测试套件..."
    echo "======================"
    
    test_health
    echo ""
    
    test_system_info
    echo ""
    
    test_create_customer
    echo ""
    
    test_get_customers
    echo ""
    
    test_create_certificate
    echo ""
    
    test_get_certificates
    echo ""
    
    test_get_certificate_detail
    echo ""
    
    test_add_test_data
    echo ""
    
    test_get_test_data
    echo ""
    
    test_verify_certificate_get
    echo ""
    
    test_verify_certificate_post
    echo ""
    
    echo "======================"
    print_success "所有API测试执行完成！"
}

# 检查依赖工具
if ! command -v curl &> /dev/null; then
    print_error "未安装curl，请先安装curl工具"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    print_error "未安装jq，部分JSON解析功能将无法正常工作"
fi

# 执行测试
run_all_tests