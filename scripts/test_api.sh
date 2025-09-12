#!/bin/bash

API_URL="http://localhost:8080/api/v1"
PUBLIC_URL="http://localhost:8080/api/v1/public"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "====== 开始执行 API 测试 ======"
echo -e "${BLUE}测试时间: $(date '+%Y-%m-%d %H:%M:%S')${NC}"
echo ""

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试函数
run_test() {
    local test_name=$1
    local test_result=$2
    local expected=$3
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$test_result" == "$expected" ]; then
        echo -e "${GREEN}✔ $test_name${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}✖ $test_name${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# ========== 1. 健康检查 ==========
echo -e "\n${BLUE}[1] 健康检查${NC}"
# 先检查原始响应
RAW_HEALTH_RESP=$(curl -s http://localhost:8080/health)
echo "Debug - Raw health response: $RAW_HEALTH_RESP"

# 尝试不同的URL
if [ -z "$RAW_HEALTH_RESP" ]; then
    echo "尝试其他健康检查端点..."
    RAW_HEALTH_RESP=$(curl -s http://localhost:8080/api/v1/health)
    echo "Debug - API v1 health response: $RAW_HEALTH_RESP"
fi

# 解析响应
HEALTH_STATUS=$(echo $RAW_HEALTH_RESP | jq -r '.status' 2>/dev/null || echo "parse_error")
if [ "$HEALTH_STATUS" == "parse_error" ]; then
    # 可能是纯文本响应
    if [[ "$RAW_HEALTH_RESP" == *"ok"* ]] || [[ "$RAW_HEALTH_RESP" == *"运行正常"* ]]; then
        HEALTH_STATUS="ok"
    fi
fi
run_test "服务健康检查" "$HEALTH_STATUS" "ok"

# ========== 2. 用户认证测试 ==========
echo -e "\n${BLUE}[2] 用户认证测试${NC}"

# 2.1 错误密码登录测试
echo -e "\n---> 测试错误密码登录"
WRONG_LOGIN_RESP=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "wrongpass"}')
WRONG_LOGIN_CODE=$(echo $WRONG_LOGIN_RESP | jq -r '.code')
run_test "错误密码应被拒绝" "$WRONG_LOGIN_CODE" "401"

# 2.2 正确登录
echo -e "\n---> 测试用户登录"
LOGIN_RESP=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}')

TOKEN=$(echo $LOGIN_RESP | jq -r '.data.token')
USER_ID=$(echo $LOGIN_RESP | jq -r '.data.userId')
USERNAME=$(echo $LOGIN_RESP | jq -r '.data.username')

if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
    run_test "管理员登录成功" "success" "success"
    echo -e "${BLUE}  用户ID: $USER_ID, 用户名: $USERNAME${NC}"
else
    run_test "管理员登录成功" "failed" "success"
    echo -e "${RED}登录失败，响应: $LOGIN_RESP${NC}"
    exit 1
fi

AUTH_HEADER="Authorization: Bearer $TOKEN"

# 2.3 获取用户资料
echo -e "\n---> 测试获取用户资料"
PROFILE_RESP=$(curl -s -X GET "$API_URL/auth/profile" -H "$AUTH_HEADER")
PROFILE_CODE=$(echo $PROFILE_RESP | jq -r '.code')
run_test "获取用户资料" "$PROFILE_CODE" "200"

# ========== 3. 证书CRUD操作测试 ==========
echo -e "\n${BLUE}[3] 证书CRUD操作测试${NC}"

# 3.1 创建证书
CERT_NUMBER="CERT-TEST-$(date +%s)"
echo -e "\n---> 测试创建证书 (编号: $CERT_NUMBER)"
CREATE_RESP=$(curl -s -X POST "$API_URL/certificates" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "{
    \"certNumber\": \"$CERT_NUMBER\",
    \"customerId\": 1,
    \"instrumentName\": \"万用表\",
    \"instrumentNumber\": \"MTR-12345\",
    \"manufacturer\": \"XYZ Corp\",
    \"modelSpec\": \"Model A\",
    \"instrumentAccuracy\": \"0.1%\",
    \"testDate\": \"2024-10-26\",
    \"expireDate\": \"2025-10-26\",
    \"testResult\": \"qualified\"
  }")

CERT_ID=$(echo $CREATE_RESP | jq -r '.data.id')
BLOCKCHAIN_TX_ID=$(echo $CREATE_RESP | jq -r '.data.blockchainTxId')
BLOCKCHAIN_HASH=$(echo $CREATE_RESP | jq -r '.data.blockchainHash')

if [ "$CERT_ID" != "null" ] && [ -n "$CERT_ID" ]; then
    run_test "证书创建成功" "success" "success"
    echo -e "${BLUE}  证书ID: $CERT_ID${NC}"
    if [ "$BLOCKCHAIN_TX_ID" != "null" ]; then
        echo -e "${BLUE}  区块链交易ID: ${BLOCKCHAIN_TX_ID:0:20}...${NC}"
        echo -e "${BLUE}  区块链哈希: ${BLOCKCHAIN_HASH:0:20}...${NC}"
    fi
else
    run_test "证书创建成功" "failed" "success"
    echo -e "${RED}响应: $CREATE_RESP${NC}"
fi

# 3.2 查询单个证书
echo -e "\n---> 测试查询单个证书"
GET_CERT_RESP=$(curl -s -X GET "$API_URL/certificates/$CERT_NUMBER" -H "$AUTH_HEADER")
GET_CERT_CODE=$(echo $GET_CERT_RESP | jq -r '.code')
run_test "查询单个证书" "$GET_CERT_CODE" "200"

# 3.3 查询所有证书（分页）
echo -e "\n---> 测试查询证书列表（分页）"
LIST_RESP=$(curl -s -X GET "$API_URL/certificates?page=1&pageSize=10" -H "$AUTH_HEADER")
LIST_CODE=$(echo $LIST_RESP | jq -r '.code')
TOTAL_COUNT=$(echo $LIST_RESP | jq -r '.total')
run_test "查询证书列表" "$LIST_CODE" "200"
echo -e "${BLUE}  总证书数: $TOTAL_COUNT${NC}"

# 3.4 更新证书状态为testing
echo -e "\n---> 测试更新证书状态为testing"
UPDATE_STATUS_RESP=$(curl -s -X PUT "$API_URL/certificates/$CERT_NUMBER" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "{
    \"certNumber\": \"$CERT_NUMBER\",
    \"customerId\": 1,
    \"instrumentName\": \"万用表\",
    \"instrumentNumber\": \"MTR-12345\",
    \"manufacturer\": \"XYZ Corp\",
    \"modelSpec\": \"Model A\",
    \"instrumentAccuracy\": \"0.1%\",
    \"testDate\": \"2024-10-26\",
    \"expireDate\": \"2025-10-26\",
    \"testResult\": \"qualified\",
    \"status\": \"testing\"
  }")

UPDATE_STATUS_CODE=$(echo $UPDATE_STATUS_RESP | jq -r '.code')
run_test "更新证书状态为testing" "$UPDATE_STATUS_CODE" "200"

# 3.5 更新证书信息
UPDATED_CERT_NUMBER="${CERT_NUMBER}-UPDATED"
echo -e "\n---> 测试更新证书信息"
UPDATE_RESP=$(curl -s -X PUT "$API_URL/certificates/$CERT_NUMBER" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "{
    \"certNumber\": \"$UPDATED_CERT_NUMBER\",
    \"customerId\": 1,
    \"instrumentName\": \"万用表(更新版)\",
    \"instrumentNumber\": \"MTR-12345-V2\",
    \"manufacturer\": \"XYZ Corp Updated\",
    \"modelSpec\": \"Model B\",
    \"instrumentAccuracy\": \"0.05%\",
    \"testDate\": \"2024-10-26\",
    \"expireDate\": \"2026-10-26\",
    \"testResult\": \"qualified\",
    \"status\": \"completed\"
  }")

UPDATE_CODE=$(echo $UPDATE_RESP | jq -r '.code')
run_test "更新证书信息" "$UPDATE_CODE" "200"

# ========== 4. 测试数据管理 ==========
echo -e "\n${BLUE}[4] 测试数据管理${NC}"

# 4.1 批量添加测试数据
echo -e "\n---> 测试批量添加测试数据"
ADD_TEST_RESP=$(curl -s -X POST "$API_URL/test-data" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "{
    \"certNumber\": \"$UPDATED_CERT_NUMBER\",
    \"data\": [
      {
        \"deviceAddr\": \"DEV-01\",
        \"testPoint\": \"P1\",
        \"actualPercentage\": 100.0,
        \"ratioError\": 0.2,
        \"angleError\": 0.1,
        \"testTimestamp\": \"2024-10-26T10:00:00Z\"
      },
      {
        \"deviceAddr\": \"DEV-01\",
        \"testPoint\": \"P2\",
        \"actualPercentage\": 80.0,
        \"ratioError\": 0.3,
        \"angleError\": 0.15,
        \"testTimestamp\": \"2024-10-26T10:05:00Z\"
      },
      {
        \"deviceAddr\": \"DEV-01\",
        \"testPoint\": \"P3\",
        \"actualPercentage\": 60.0,
        \"ratioError\": 0.25,
        \"angleError\": 0.12,
        \"testTimestamp\": \"2024-10-26T10:10:00Z\"
      }
    ]
  }")

ADD_TEST_CODE=$(echo $ADD_TEST_RESP | jq -r '.code')
run_test "批量添加测试数据" "$ADD_TEST_CODE" "201"

# 4.2 获取测试数据
echo -e "\n---> 测试获取测试数据"
GET_TEST_RESP=$(curl -s -X GET "$API_URL/test-data/certificate/$UPDATED_CERT_NUMBER" \
  -H "$AUTH_HEADER")

GET_TEST_CODE=$(echo $GET_TEST_RESP | jq -r '.code')
TEST_DATA_COUNT=$(echo $GET_TEST_RESP | jq '.data | length')
run_test "获取测试数据" "$GET_TEST_CODE" "200"
echo -e "${BLUE}  测试数据点数: $TEST_DATA_COUNT${NC}"

# ========== 5. 证书签发流程 ==========
echo -e "\n${BLUE}[5] 证书签发流程${NC}"

# 5.1 更新证书状态为issued（签发）
echo -e "\n---> 测试签发证书"
ISSUE_RESP=$(curl -s -X PUT "$API_URL/certificates/$UPDATED_CERT_NUMBER" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "{
    \"certNumber\": \"$UPDATED_CERT_NUMBER\",
    \"customerId\": 1,
    \"instrumentName\": \"万用表(更新版)\",
    \"instrumentNumber\": \"MTR-12345-V2\",
    \"manufacturer\": \"XYZ Corp Updated\",
    \"modelSpec\": \"Model B\",
    \"instrumentAccuracy\": \"0.05%\",
    \"testDate\": \"2024-10-26\",
    \"expireDate\": \"2026-10-26\",
    \"testResult\": \"qualified\",
    \"status\": \"issued\"
  }")

ISSUE_CODE=$(echo $ISSUE_RESP | jq -r '.code')
run_test "签发证书" "$ISSUE_CODE" "200"

# ========== 6. 证书验证测试 ==========
echo -e "\n${BLUE}[6] 证书验证测试${NC}"

# 6.1 验证存在的证书（公开接口）
echo -e "\n---> 测试公开验证证书（存在）"
VERIFY_RESP=$(curl -s -X GET "$PUBLIC_URL/verify/$UPDATED_CERT_NUMBER")
VERIFY_CODE=$(echo $VERIFY_RESP | jq -r '.code')
IS_VALID=$(echo $VERIFY_RESP | jq -r '.data.isValid')
BLOCKCHAIN_INFO=$(echo $VERIFY_RESP | jq -r '.data.blockchainTxId')

run_test "验证存在的证书" "$VERIFY_CODE" "200"
if [ "$IS_VALID" == "true" ]; then
    echo -e "${GREEN}  证书验证: 有效${NC}"
    if [ "$BLOCKCHAIN_INFO" != "null" ]; then
        echo -e "${BLUE}  区块链验证: 已上链${NC}"
    fi
fi

# 6.2 验证不存在的证书
echo -e "\n---> 测试公开验证证书（不存在）"
NON_EXIST_CERT="CERT-NON-EXIST-999999"
VERIFY_INVALID_RESP=$(curl -s -X GET "$PUBLIC_URL/verify/$NON_EXIST_CERT")
VERIFY_INVALID_CODE=$(echo $VERIFY_INVALID_RESP | jq -r '.code')
IS_INVALID=$(echo $VERIFY_INVALID_RESP | jq -r '.data.isValid')

run_test "验证不存在的证书返回200" "$VERIFY_INVALID_CODE" "200"
run_test "不存在的证书应无效" "$IS_INVALID" "false"

# 6.3 内部验证接口（需要认证）
echo -e "\n---> 测试内部验证证书接口"
INTERNAL_VERIFY_RESP=$(curl -s -X POST "$API_URL/certificates/$UPDATED_CERT_NUMBER/verify" \
  -H "$AUTH_HEADER")
INTERNAL_VERIFY_CODE=$(echo $INTERNAL_VERIFY_RESP | jq -r '.code')
run_test "内部证书验证" "$INTERNAL_VERIFY_CODE" "200"

# ========== 7. 证书历史记录 ==========
echo -e "\n${BLUE}[7] 证书历史记录${NC}"

echo -e "\n---> 测试获取证书历史"
HISTORY_RESP=$(curl -s -X GET "$API_URL/certificates/$UPDATED_CERT_NUMBER/history" \
  -H "$AUTH_HEADER")
HISTORY_CODE=$(echo $HISTORY_RESP | jq -r '.code')
run_test "获取证书历史记录" "$HISTORY_CODE" "200"

# ========== 8. 证书撤销测试 ==========
echo -e "\n${BLUE}[8] 证书撤销测试${NC}"

# 8.1 创建一个用于撤销的证书
REVOKE_CERT_NUMBER="CERT-REVOKE-$(date +%s)"
echo -e "\n---> 创建用于撤销的证书"
CREATE_REVOKE_RESP=$(curl -s -X POST "$API_URL/certificates" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "{
    \"certNumber\": \"$REVOKE_CERT_NUMBER\",
    \"customerId\": 2,
    \"instrumentName\": \"测试仪器\",
    \"instrumentNumber\": \"TEST-99999\",
    \"manufacturer\": \"Test Corp\",
    \"modelSpec\": \"Model X\",
    \"instrumentAccuracy\": \"0.2%\",
    \"testDate\": \"2024-10-01\",
    \"expireDate\": \"2025-10-01\",
    \"testResult\": \"qualified\"
  }")

REVOKE_CERT_ID=$(echo $CREATE_REVOKE_RESP | jq -r '.data.id')
run_test "创建待撤销证书" "$([ "$REVOKE_CERT_ID" != "null" ] && echo 'success' || echo 'failed')" "success"

# 8.2 撤销证书（更新状态为revoked）
echo -e "\n---> 测试撤销证书"
REVOKE_RESP=$(curl -s -X PUT "$API_URL/certificates/$REVOKE_CERT_NUMBER" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "{
    \"certNumber\": \"$REVOKE_CERT_NUMBER\",
    \"customerId\": 2,
    \"instrumentName\": \"测试仪器\",
    \"instrumentNumber\": \"TEST-99999\",
    \"manufacturer\": \"Test Corp\",
    \"modelSpec\": \"Model X\",
    \"instrumentAccuracy\": \"0.2%\",
    \"testDate\": \"2024-10-01\",
    \"expireDate\": \"2025-10-01\",
    \"testResult\": \"qualified\",
    \"status\": \"revoked\"
  }")

REVOKE_CODE=$(echo $REVOKE_RESP | jq -r '.code')
REVOKE_STATUS=$(echo $REVOKE_RESP | jq -r '.data.status')
run_test "撤销证书操作" "$REVOKE_CODE" "200"
run_test "证书状态为revoked" "$REVOKE_STATUS" "revoked"

# 8.3 验证已撤销的证书
echo -e "\n---> 测试验证已撤销的证书"
VERIFY_REVOKED_RESP=$(curl -s -X GET "$PUBLIC_URL/verify/$REVOKE_CERT_NUMBER")
REVOKED_VALID=$(echo $VERIFY_REVOKED_RESP | jq -r '.data.isValid')
REVOKED_MSG=$(echo $VERIFY_REVOKED_RESP | jq -r '.data.message')
run_test "已撤销证书应无效" "$REVOKED_VALID" "false"
echo -e "${BLUE}  撤销原因: $REVOKED_MSG${NC}"

# ========== 9. 权限测试 ==========
echo -e "\n${BLUE}[9] 权限测试${NC}"

# 9.1 未认证访问测试
echo -e "\n---> 测试未认证访问"
UNAUTH_RESP=$(curl -s -X GET "$API_URL/certificates")
UNAUTH_CODE=$(echo $UNAUTH_RESP | jq -r '.code')
run_test "未认证访问应被拒绝" "$UNAUTH_CODE" "401"

# 9.2 无效Token访问测试
echo -e "\n---> 测试无效Token访问"
INVALID_TOKEN_RESP=$(curl -s -X GET "$API_URL/certificates" \
  -H "Authorization: Bearer invalid_token_12345")
INVALID_TOKEN_CODE=$(echo $INVALID_TOKEN_RESP | jq -r '.code')
run_test "无效Token应被拒绝" "$INVALID_TOKEN_CODE" "401"

# ========== 10. 批量操作测试 ==========
echo -e "\n${BLUE}[10] 批量操作测试${NC}"

# 10.1 创建多个证书进行测试
echo -e "\n---> 批量创建证书测试"
BATCH_SUCCESS=0
for i in {1..3}; do
    BATCH_CERT_NUMBER="CERT-BATCH-$(date +%s)-$i"
    BATCH_CREATE_RESP=$(curl -s -X POST "$API_URL/certificates" \
      -H "Content-Type: application/json" \
      -H "$AUTH_HEADER" \
      -d "{
        \"certNumber\": \"$BATCH_CERT_NUMBER\",
        \"customerId\": 1,
        \"instrumentName\": \"批量测试仪器$i\",
        \"instrumentNumber\": \"BATCH-$i\",
        \"manufacturer\": \"Batch Corp\",
        \"modelSpec\": \"Model $i\",
        \"instrumentAccuracy\": \"0.${i}%\",
        \"testDate\": \"2024-10-0$i\",
        \"expireDate\": \"2025-10-0$i\",
        \"testResult\": \"qualified\"
      }")
    
    BATCH_CODE=$(echo $BATCH_CREATE_RESP | jq -r '.code')
    if [ "$BATCH_CODE" == "201" ]; then
        BATCH_SUCCESS=$((BATCH_SUCCESS + 1))
    fi
    sleep 0.5  # 避免请求过快
done

run_test "批量创建3个证书" "$BATCH_SUCCESS" "3"

# ========== 11. 清理测试 ==========
echo -e "\n${BLUE}[11] 清理测试数据${NC}"

# 删除测试部分
echo -e "\n---> 测试删除证书"
DELETE_RESP=$(curl -s -X DELETE "$API_URL/certificates/$UPDATED_CERT_NUMBER" \
  -H "$AUTH_HEADER")
echo "Delete Response: $DELETE_RESP" # 添加调试输出
DELETE_CODE=$(echo $DELETE_RESP | jq -r '.code' 2>/dev/null || echo "parse_error")
if [ "$DELETE_CODE" == "parse_error" ]; then
    echo -e "${RED}响应格式错误: $DELETE_RESP${NC}"
fi
run_test "删除证书" "$DELETE_CODE" "200"

# 11.2 验证删除后的证书
echo -e "\n---> 验证已删除的证书"
VERIFY_DELETED_RESP=$(curl -s -X GET "$API_URL/certificates/$UPDATED_CERT_NUMBER" \
  -H "$AUTH_HEADER")
VERIFY_DELETED_CODE=$(echo $VERIFY_DELETED_RESP | jq -r '.code')
run_test "已删除证书应返回404" "$VERIFY_DELETED_CODE" "404"

# ========== 12. 登出测试 ==========
echo -e "\n${BLUE}[12] 用户登出${NC}"

echo -e "\n---> 测试用户登出"
LOGOUT_RESP=$(curl -s -X POST "$API_URL/auth/logout" -H "$AUTH_HEADER")
LOGOUT_CODE=$(echo $LOGOUT_RESP | jq -r '.code')
run_test "用户登出" "$LOGOUT_CODE" "200"

# ========== 测试报告 ==========
echo -e "\n${BLUE}====== API 测试报告 ======${NC}"
echo -e "测试时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo -e "总测试数: $TOTAL_TESTS"
echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
echo -e "${RED}失败: $FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}✅ 所有测试通过！${NC}"
    exit 0
else
    echo -e "\n${RED}❌ 有 $FAILED_TESTS 个测试失败${NC}"
    exit 1
fi