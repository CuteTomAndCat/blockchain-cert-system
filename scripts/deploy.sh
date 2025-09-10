#!/bin/bash

# 计量证书防伪溯源系统一键部署脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "开始部署计量证书防伪溯源系统..."
echo "项目根目录: $PROJECT_ROOT"

# 检查是否以root权限运行
#if [[ $EUID -eq 0 ]]; then
#   echo "请不要以root权限运行此脚本"
#   exit 1
#fi

# 设置颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查系统依赖
check_dependencies() {
    print_status "检查系统依赖..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker未安装，请先运行install_prerequisites.sh"
        exit 1
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose未安装，请先运行install_prerequisites.sh"
        exit 1
    fi
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        print_error "Go未安装，请先运行install_prerequisites.sh"
        exit 1
    fi
    
    print_status "系统依赖检查完成"
}

# 创建目录结构
create_directories() {
    print_status "创建项目目录结构..."
    
    cd "$PROJECT_ROOT"
    mkdir -p {scripts,chaincode,application,database,configs,docker,channel-artifacts,crypto-config}
    
    print_status "目录结构创建完成"
}

# 生成Fabric网络配置
generate_fabric_config() {
    print_status "生成Fabric网络配置..."
    
    cd "$PROJECT_ROOT"
    
    # 设置Fabric环境变量
    export FABRIC_CFG_PATH="$PROJECT_ROOT/configs"
    
    # 生成加密材料
    if [ ! -d "crypto-config" ] || [ -z "$(ls -A crypto-config 2>/dev/null)" ]; then
        print_status "生成组织证书..."
        cryptogen generate --config=configs/crypto-config.yaml --output="crypto-config"
    fi
    
    # 创建channel-artifacts目录
    mkdir -p channel-artifacts
    
    # 生成创世区块
    if [ ! -f "channel-artifacts/genesis.block" ]; then
        print_status "生成创世区块..."
        configtxgen -profile TwoOrgsOrdererGenesis -channelID system-channel -outputBlock ./channel-artifacts/genesis.block
    fi
    
    # 生成通道配置
    if [ ! -f "channel-artifacts/certchannel.tx" ]; then
        print_status "生成通道配置..."
        configtxgen -profile TwoOrgsChannel -outputCreateChannelTx ./channel-artifacts/certchannel.tx -channelID certchannel
    fi
    
    # 生成锚节点配置
    if [ ! -f "channel-artifacts/Org1MSPanchors.tx" ]; then
        print_status "生成锚节点配置..."
        configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx -channelID certchannel -asOrg Org1MSP
        configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors.tx -channelID certchannel -asOrg Org2MSP
    fi
    
    print_status "Fabric网络配置生成完成"
}

# 构建链码
build_chaincode() {
    print_status "构建链码..."
    
    cd "$PROJECT_ROOT/chaincode/cert-chaincode"
    
    # 初始化Go模块
    if [ ! -f "go.mod" ]; then
        go mod init cert-chaincode
    fi
    
    # 下载依赖
    go mod tidy
    
    # 构建链码
    go build -o cert-chaincode
    
    print_status "链码构建完成"
}

# 构建应用程序
build_application() {
    print_status "构建后端应用程序..."
    
    cd "$PROJECT_ROOT/application"
    
    # 初始化Go模块
    if [ ! -f "go.mod" ]; then
        go mod init cert-system
    fi
    
    # 下载依赖
    go mod tidy
    
    # 构建应用程序
    go build -o cert-system ./main.go
    
    print_status "应用程序构建完成"
}

# 启动基础设施
start_infrastructure() {
    print_status "启动基础设施服务..."
    
    cd "$PROJECT_ROOT"
    
    # 停止可能存在的容器
    docker-compose -f docker/docker-compose.yaml down 2>/dev/null || true
    
    # 启动基础设施
    docker-compose -f docker/docker-compose.yaml up -d mysql
    
    # 等待MySQL启动
    print_status "等待MySQL启动..."
    sleep 5
    
    # 检查MySQL是否就绪
    for i in {1..30}; do
        if docker exec cert-mysql mysql -ucertuser -pcertpass123 -e "SELECT 1;" &>/dev/null; then
            print_status "MySQL已就绪"
            break
        fi
        if [ $i -eq 30 ]; then
            print_error "MySQL启动超时"
            exit 1
        fi
        sleep 2
    done
    
    print_status "基础设施启动完成"
}

# 启动Fabric网络
start_fabric_network() {
    print_status "启动Fabric网络..."
    
    cd "$PROJECT_ROOT"
    
    # 启动Fabric节点
    docker-compose -f docker/docker-compose.yaml up -d
    
    # 等待Fabric网络启动
    print_status "等待Fabric网络启动..."
    sleep 5
    
    print_status "Fabric网络启动完成"
}

# 部署链码
deploy_chaincode() {
    print_status "部署链码..."
    
    cd "$PROJECT_ROOT"
    
    # 设置环境变量
    export CORE_PEER_TLS_ENABLED=true
    export CORE_PEER_LOCALMSPID="Org1MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE="$PROJECT_ROOT/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
    export CORE_PEER_MSPCONFIGPATH="$PROJECT_ROOT/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"
    export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
    
    # 创建通道
    print_status "创建通道..."
    peer channel create -o orderer.example.com:7050 -c certchannel -f ./channel-artifacts/certchannel.tx --outputBlock ./channel-artifacts/certchannel.block --tls --cafile "$PROJECT_ROOT/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem"
    
    # 加入通道
    print_status "加入通道..."
    peer channel join -b ./channel-artifacts/certchannel.block 2>/dev/null || true
    
    # 打包链码
    print_status "打包链码..."
    cd "$PROJECT_ROOT/chaincode"
    peer lifecycle chaincode package certchaincode.tar.gz --path ./cert-chaincode --lang golang --label certchaincode_1.0 2>/dev/null || true
    
    # 安装链码
    print_status "安装链码..."
    peer lifecycle chaincode install certchaincode.tar.gz 2>/dev/null || true
    print_status "安装完了..."
    
    # 查询已安装的链码包ID
    PACKAGE_ID=$(peer lifecycle chaincode queryinstalled --output json | jq -r '.installed_chaincodes[0].package_id' 2>/dev/null)
    if [ "$PACKAGE_ID" != "null" ] && [ -n "$PACKAGE_ID" ]; then
        print_status "链码包ID: $PACKAGE_ID"
        
       # 批准链码
        print_status "批准链码..."
        peer lifecycle chaincode approveformyorg -o localhost:7050 --channelID certchannel --name certchaincode --version 1.0 --package-id $PACKAGE_ID --sequence 1 --tls --cafile "$PROJECT_ROOT/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" 2>/dev/null || true
        
        # 提交链码
        print_status "提交链码..."
        peer lifecycle chaincode commit -o localhost:7050 --channelID certchannel --name certchaincode --version 1.0 --sequence 1 --tls --cafile "$PROJECT_ROOT/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" --peerAddresses localhost:7051 --tlsRootCertFiles "$PROJECT_ROOT/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" 2>/dev/null || true

        print_status "链码部署完成"
    else
        print_warning "链码安装可能失败，请检查日志"
    fi
}

# 启动应用程序
start_application() {
    print_status "启动后端应用程序..."
    
    cd "$PROJECT_ROOT/application"
    
    # 设置环境变量
    export DB_HOST=localhost
    export DB_PORT=3306
    export DB_USERNAME=certuser
    export DB_PASSWORD=certpass123
    export DB_DATABASE=cert_system
    export SERVER_PORT=8080
    
    # 启动应用程序
    nohup ./cert-system > app.log 2>&1 &
    
    # 等待应用程序启动
    print_status "等待应用程序启动..."
    sleep 10
    
    # 检查应用程序是否启动成功
    if curl -s http://localhost:8080/health > /dev/null; then
        print_status "应用程序启动成功"
    else
        print_warning "应用程序可能启动失败，请检查app.log"
    fi
}

# 运行测试
run_tests() {
    print_status "运行系统测试..."
    
    # 测试数据库连接
    if docker exec cert-mysql mysql -ucertuser -pcertpass123 cert_system -e "SHOW TABLES;" &>/dev/null; then
        print_status "数据库连接测试成功"
    else
        print_error "数据库连接测试失败"
    fi
    
    # 测试API
    if curl -s http://localhost:8080/health | grep -q "ok"; then
        print_status "API健康检查成功"
    else
        print_error "API健康检查失败"
   fi
   
   print_status "系统测试完成"
}

# 显示部署信息
show_deployment_info() {
   print_status "部署完成！"
   echo ""
   echo "=========================================="
   echo "  计量证书防伪溯源系统部署信息"
   echo "=========================================="
   echo ""
   echo "🔗 服务地址:"
   echo "   - 后端API: http://localhost:8080"
   echo "   - 健康检查: http://localhost:8080/health"
   echo "   - MySQL: localhost:3306"
   echo "   - CouchDB: http://localhost:5984"
   echo ""
   echo "🔑 默认账户:"
   echo "   - 用户名: admin"
   echo "   - 密码: admin123"
   echo ""
   echo "📂 重要文件:"
   echo "   - 应用日志: $PROJECT_ROOT/application/app.log"
   echo "   - Docker日志: docker-compose -f $PROJECT_ROOT/docker/docker-compose.yaml logs"
   echo ""
   echo "🛠 管理命令:"
   echo "   - 停止服务: docker-compose -f $PROJECT_ROOT/docker/docker-compose.yaml down"
   echo "   - 查看日志: docker-compose -f $PROJECT_ROOT/docker/docker-compose.yaml logs -f"
   echo "   - 重启服务: $SCRIPT_DIR/restart.sh"
   echo ""
   echo "📖 API文档:"
   echo "   - 登录: POST /api/v1/auth/login"
   echo "   - 创建证书: POST /api/v1/certificates"
   echo "   - 验证证书: POST /api/v1/public/verify/{certNumber}"
   echo "   - 添加测试数据: POST /api/v1/test-data"
   echo ""
   print_status "系统已就绪，可以开始使用！"
}

# 主函数
main() {
   echo "计量证书防伪溯源系统自动部署脚本"
   echo "======================================="
   
   # 检查依赖
   check_dependencies
   
   # 创建目录
   create_directories
   
   # 生成Fabric配置
   generate_fabric_config
   
   # 构建代码
   build_chaincode
   build_application
   
   # 启动服务
   start_infrastructure
   start_fabric_network
   
   # 部署链码
   deploy_chaincode
   
   # 启动应用
   start_application
   
   # 运行测试
   run_tests
   
   # 显示部署信息
   show_deployment_info
}

# 错误处理
trap 'print_error "部署过程中发生错误，请检查日志"; exit 1' ERR

# 运行主函数
main "$@"
