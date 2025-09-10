#!/bin/bash

# 重启系统服务脚本

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "重启计量证书防伪溯源系统..."

# 停止应用程序
echo "停止应用程序..."
pkill -f cert-system || true

# 停止Docker服务
echo "停止Docker服务..."
cd "$PROJECT_ROOT"
docker-compose -f docker/docker-compose.yaml down

# 等待服务完全停止
sleep 5

# 启动Docker服务
echo "启动Docker服务..."
docker-compose -f docker/docker-compose.yaml up -d

# 等待服务启动
echo "等待服务启动..."
sleep 30

# 启动应用程序
echo "启动应用程序..."
cd "$PROJECT_ROOT/application"
export DB_HOST=localhost
export DB_PORT=3306
export DB_USERNAME=certuser
export DB_PASSWORD=certpass123
export DB_DATABASE=cert_system
export SERVER_PORT=8080

nohup ./cert-system > app.log 2>&1 &

echo "系统重启完成！"

