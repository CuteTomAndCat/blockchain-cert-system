#!/bin/bash

# 安装系统依赖环境脚本
echo "开始安装系统依赖..."

# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装基础工具
sudo apt install -y curl wget git vim build-essential

# 安装Docker
echo "安装Docker..."
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
sudo systemctl enable docker
sudo systemctl start docker

# 安装Docker Compose
echo "安装Docker Compose..."
sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 安装Go
echo "安装Go 1.19..."
wget https://go.dev/dl/go1.19.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.19.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# 安装Hyperledger Fabric工具
echo "安装Hyperledger Fabric工具..."
mkdir -p $HOME/go/src/github.com/hyperledger
cd $HOME/go/src/github.com/hyperledger
curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.4.0 1.5.0

# 设置Fabric环境变量
echo 'export PATH=$HOME/go/src/github.com/hyperledger/fabric-samples/bin:$PATH' >> ~/.bashrc
source ~/.bashrc

echo "环境安装完成！请重新登录以使Docker组权限生效。"