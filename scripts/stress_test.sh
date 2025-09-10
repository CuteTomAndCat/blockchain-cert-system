#!/bin/bash

# 压力测试脚本

API_BASE="http://localhost:8080/api/v1"
CONCURRENT_USERS=10
REQUESTS_PER_USER=50

echo "开始压力测试..."
echo "并发用户数: $CONCURRENT_USERS"
echo "每用户请求数: $REQUESTS_PER_USER"

# 登录获取token
get_token() {
    response=$(curl -s -X POST "$API_BASE/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "admin",
            "password": "admin123"
        }')
    
    echo "$response" | jq -r '.data.token' 2>/dev/null
}

# 压力测试函数
stress_test_worker() {
    local worker_id=$1
    local token=$2
    local success_count=0
    local error_count=0
    
    for i in $(seq 1 $REQUESTS_PER_USER); do
        cert_number="STRESS_${worker_id}_${i}_$(date +%s)"
        
        response=$(curl -s -w "%{http_code}" -X POST "$API_BASE/certificates" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $token" \
            -d "{
                \"certNumber\": \"$cert_number\",
                \"customerId\": 1,
                \"instrumentName\": \"压力测试互感器\",
                \"instrumentNumber\": \"STRESS-$worker_id-$i\",
                \"manufacturer\": \"测试公司\",
                \"modelSpec\": \"TEST-1000\",
                \"instrumentAccuracy\": \"0.2级\",
                \"testDate\": \"$(date +%Y-%m-%d)\",
                \"testResult\": \"qualified\"
            }")
        
        http_code="${response: -3}"
        if [ "$http_code" == "200" ]; then
            ((success_count++))
        else
            ((error_count++))
        fi
        
        # 简单的进度显示
        if [ $((i % 10)) -eq 0 ]; then
            echo "Worker $worker_id: $i/$REQUESTS_PER_USER requests completed"
        fi
    done
    
    echo "Worker $worker_id completed: $success_count success, $error_count errors"
}

# 获取认证token
echo "获取认证token..."
TOKEN=$(get_token)

if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
    echo "获取token失败"
    exit 1
fi

echo "开始并发测试..."
start_time=$(date +%s)

# 启动并发工作进程
for i in $(seq 1 $CONCURRENT_USERS); do
    stress_test_worker $i "$TOKEN" &
done

# 等待所有进程完成
wait

end_time=$(date +%s)
duration=$((end_time - start_time))
total_requests=$((CONCURRENT_USERS * REQUESTS_PER_USER))

echo ""
echo "压力测试完成！"
echo "总用时: ${duration}秒"
echo "总请求数: $total_requests"
echo "平均QPS: $((total_requests / duration))"
