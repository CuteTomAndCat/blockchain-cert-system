#!/bin/bash

# 停止系统服务脚本

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "停止计量证书防伪溯源系统..."

# 停止应用程序
echo "停止应用程序..."
pkill -f cert-system || true

# 停止Docker服务
echo "停止Docker服务..."
cd "$PROJECT_ROOT"
docker-compose -f docker/docker-compose.yaml down

# 清理Docker资源（可选）
read -p "是否清理Docker数据卷？(y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "清理Docker数据卷..."
    docker-compose -f docker/docker-compose.yaml down -v
    docker system prune -f
fi

echo "系统已停止！"

