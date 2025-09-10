#!/bin/bash

# Hyperledger Fabric网络搭建脚本

FABRIC_CFG_PATH=$(pwd)/configs
CHANNEL_NAME="certchannel"
CHAINCODE_NAME="certchaincode"

echo "开始搭建Hyperledger Fabric网络..."

# 生成加密材料
echo "生成组织证书..."
cryptogen generate --config=configs/crypto-config.yaml --output="crypto-config"

# 生成创世区块
echo "生成创世区块..."
configtxgen -profile TwoOrgsOrdererGenesis -channelID system-channel -outputBlock ./channel-artifacts/genesis.block

# 生成通道配置
echo "生成通道配置..."
configtxgen -profile TwoOrgsChannel -outputCreateChannelTx ./channel-artifacts/${CHANNEL_NAME}.tx -channelID $CHANNEL_NAME

# 生成锚节点配置
configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org1MSP
configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org2MSP

echo "Fabric网络配置文件生成完成！"