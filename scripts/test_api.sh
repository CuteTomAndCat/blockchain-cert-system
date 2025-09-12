#!/bin/bash

API_URL="http://localhost:8080/api/v1"
PUBLIC_URL="http://localhost:8080/api/v1/public"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "====== 开始执行 API 测试 ======"

# ---> 用户登录
echo -e "\n---> 测试用户登录"
LOGIN_RESP=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}')

TOKEN=$(echo $LOGIN_RESP | jq -r '.data.token')

if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
  echo -e "${GREEN}✔ 用户登录成功，Token 已获取。${NC}"
else
  echo -e "${RED}✖ 用户登录失败，响应: $LOGIN_RESP${NC}"
  exit 1
fi

AUTH_HEADER="Authorization: Bearer $TOKEN"

# ---> 创建证书
CERT_NUMBER="CERT-TEST-$(date +%s)"
echo -e "\n---> 测试创建证书"
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
    \"testDate\": \"2023-10-26\",
    \"expireDate\": \"2025-10-26\",
    \"testResult\": \"qualified\"
  }")

CERT_ID=$(echo $CREATE_RESP | jq -r '.data.id')

if [ "$CERT_ID" != "null" ] && [ -n "$CERT_ID" ]; then
  echo -e "${GREEN}✔ 证书创建成功，证书编号: $CERT_NUMBER, ID: $CERT_ID${NC}"
else
  echo -e "${RED}✖ 证书创建失败，响应: $CREATE_RESP${NC}"
  exit 1
fi

# ---> 更新证书
UPDATED_CERT_NUMBER="${CERT_NUMBER}-UPDATED"
echo -e "\n---> 测试更新证书"
UPDATE_RESP=$(curl -s -X PUT "$API_URL/certificates/$CERT_NUMBER" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "{
    \"certNumber\": \"$UPDATED_CERT_NUMBER\",
    \"customerId\": 1,
    \"instrumentName\": \"万用表(更新版)\",
    \"instrumentNumber\": \"MTR-12345-V2\",
    \"manufacturer\": \"XYZ Corp\",
    \"modelSpec\": \"Model B\",
    \"instrumentAccuracy\": \"0.05%\",
    \"testDate\": \"2023-10-26T10:00:00Z\",
    \"expireDate\": \"2026-10-26T10:00:00Z\",
    \"testResult\": \"qualified\",
    \"status\": \"draft\"
  }")

UPDATE_CODE=$(echo $UPDATE_RESP | jq -r '.code')

if [ "$UPDATE_CODE" == "200" ]; then
  echo -e "${GREEN}✔ 证书更新成功（新编号: $UPDATED_CERT_NUMBER）。${NC}"
else
  echo -e "${RED}✖ 证书更新失败，响应: $UPDATE_RESP${NC}"
  exit 1
fi

# ---> 验证证书（存在）
echo -e "\n---> 测试验证证书 (存在)"
VERIFY_RESP=$(curl -s -X GET "$PUBLIC_URL/verify/$UPDATED_CERT_NUMBER")
VERIFY_CODE=$(echo $VERIFY_RESP | jq -r '.code')

if [ "$VERIFY_CODE" == "200" ]; then
  echo -e "${GREEN}✔ 验证证书成功（存在）。${NC}"
else
  echo -e "${RED}✖ 验证证书失败 (存在)，响应: $VERIFY_RESP${NC}"
fi

# ---> 验证证书（不存在）
echo -e "\n---> 测试验证证书 (不存在)"
VERIFY_RESP2=$(curl -s -X GET "$PUBLIC_URL/verify/NON-EXISTENT-$CERT_NUMBER")
VERIFY_CODE2=$(echo $VERIFY_RESP2 | jq -r '.code')

if [ "$VERIFY_CODE2" == "200" ]; then
  echo -e "${GREEN}✔ 验证证书成功（不存在时返回正确提示）。${NC}"
else
  echo -e "${YELLOW}⚠ 验证证书失败 (不存在)，响应: $VERIFY_RESP2${NC}"
fi

# ---> 批量添加测试数据
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
        \"testTimestamp\": \"2023-10-26T10:00:00Z\"
      },
      {
        \"deviceAddr\": \"DEV-01\",
        \"testPoint\": \"P2\",
        \"actualPercentage\": 80.0,
        \"ratioError\": 0.3,
        \"angleError\": 0.15,
        \"testTimestamp\": \"2023-10-26T10:05:00Z\"
      }
    ]
  }")

ADD_TEST_CODE=$(echo $ADD_TEST_RESP | jq -r '.code')

if [ "$ADD_TEST_CODE" == "201" ]; then
  echo -e "${GREEN}✔ 批量测试数据添加成功。${NC}"
else
  echo -e "${RED}✖ 测试数据添加失败，响应: $ADD_TEST_RESP${NC}"
  exit 1
fi

# ---> 获取测试数据
echo -e "\n---> 测试获取测试数据"
GET_TEST_RESP=$(curl -s -X GET "$API_URL/test-data/certificate/$UPDATED_CERT_NUMBER" \
  -H "$AUTH_HEADER")

GET_TEST_CODE=$(echo $GET_TEST_RESP | jq -r '.code')

if [ "$GET_TEST_CODE" == "200" ]; then
  echo -e "${GREEN}✔ 获取测试数据成功。${NC}"
else
  echo -e "${RED}✖ 获取测试数据失败，响应: $GET_TEST_RESP${NC}"
fi

echo -e "\n====== API 测试执行完毕 ======"
