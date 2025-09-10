-- 计量证书防伪溯源系统数据库初始化脚本

CREATE DATABASE IF NOT EXISTS cert_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE cert_system;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名',
    password_hash VARCHAR(128) NOT NULL COMMENT '密码哈希',
    role ENUM('admin', 'operator', 'viewer') DEFAULT 'viewer' COMMENT '用户角色',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 设备信息表
CREATE TABLE IF NOT EXISTS devices (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    device_addr VARCHAR(100) UNIQUE NOT NULL COMMENT '设备地址',
    device_name VARCHAR(200) NOT NULL COMMENT '设备名称',
    manufacturer VARCHAR(100) COMMENT '制造厂商',
    model VARCHAR(100) COMMENT '型号规格',
    accuracy_class VARCHAR(50) COMMENT '准确度等级',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 委托方信息表
CREATE TABLE IF NOT EXISTS customers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    customer_name VARCHAR(200) NOT NULL COMMENT '委托者名称',
    customer_address TEXT COMMENT '委托者地址',
    contact_person VARCHAR(100) COMMENT '联系人',
    contact_phone VARCHAR(20) COMMENT '联系电话',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 计量证书表
CREATE TABLE IF NOT EXISTS certificates (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    cert_number VARCHAR(100) UNIQUE NOT NULL COMMENT '证书编号',
    customer_id BIGINT NOT NULL COMMENT '委托方ID',
    instrument_name VARCHAR(200) NOT NULL COMMENT '器具名称',
    instrument_number VARCHAR(100) COMMENT '器具编号',
    manufacturer VARCHAR(100) COMMENT '制造厂',
    model_spec VARCHAR(100) COMMENT '型号规格',
    instrument_accuracy VARCHAR(50) COMMENT '器具准确度',
    test_date DATE NOT NULL COMMENT '检测日期',
    expire_date DATE COMMENT '有效期至',
    test_result ENUM('qualified', 'unqualified') DEFAULT 'qualified' COMMENT '检测结果',
    blockchain_tx_id VARCHAR(128) COMMENT '区块链交易ID',
    status ENUM('draft', 'testing', 'completed', 'issued') DEFAULT 'draft' COMMENT '证书状态',
    created_by BIGINT COMMENT '创建人ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES customers(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

-- 测试数据表（互感器测试数据）
CREATE TABLE IF NOT EXISTS test_data (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    cert_id BIGINT NOT NULL COMMENT '证书ID',
    device_addr VARCHAR(100) NOT NULL COMMENT '设备地址',
    data_type VARCHAR(50) NOT NULL COMMENT '数据类型',
    percentage_value DECIMAL(10,6) COMMENT '百分表值',
    ratio_error DECIMAL(10,6) COMMENT '比差值',
    angle_error DECIMAL(10,6) COMMENT '角差值',
    current_value DECIMAL(10,3) COMMENT '电流值(A)',
    voltage_value DECIMAL(10,3) COMMENT '电压值(V)',
    workstation_number VARCHAR(20) COMMENT '工位号',
    test_point VARCHAR(50) COMMENT '测试点',
    actual_percentage DECIMAL(10,6) COMMENT '实际值',
    test_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '测试时间',
    blockchain_hash VARCHAR(128) COMMENT '区块链哈希',
    encrypted_data TEXT COMMENT '国密加密后的敏感数据',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cert_id) REFERENCES certificates(id)
);

-- 区块链交易记录表
CREATE TABLE IF NOT EXISTS blockchain_transactions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    tx_id VARCHAR(128) UNIQUE NOT NULL COMMENT '交易ID',
    block_number BIGINT COMMENT '区块号',
    transaction_hash VARCHAR(128) COMMENT '交易哈希',
    cert_id BIGINT COMMENT '关联证书ID',
    operation_type ENUM('create', 'update', 'verify') COMMENT '操作类型',
    status ENUM('pending', 'confirmed', 'failed') DEFAULT 'pending' COMMENT '交易状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cert_id) REFERENCES certificates(id)
);

-- 创建索引
CREATE INDEX idx_cert_number ON certificates(cert_number);
CREATE INDEX idx_device_addr ON test_data(device_addr);
CREATE INDEX idx_test_timestamp ON test_data(test_timestamp);
CREATE INDEX idx_blockchain_tx ON blockchain_transactions(tx_id);

-- 插入初始管理员用户（密码：admin123，实际使用时应使用强密码）
INSERT INTO users (username, password_hash, role) VALUES 
('admin', SHA2('admin123', 256), 'admin');

-- 插入示例数据（用于测试，可根据需要修改）
INSERT INTO customers (customer_name, customer_address, contact_person, contact_phone) VALUES 
('XX电力公司', '北京市朝阳区XXX路123号', '张三', '13800138001'),
('YY制造企业', '上海市浦东新区YYY街456号', '李四', '13900139002');

INSERT INTO devices (device_addr, device_name, manufacturer, model, accuracy_class) VALUES 
('DEV001', '电流互感器测试装置', 'ABC仪器公司', 'CTT-2000', '0.1级'),
('DEV002', '电压互感器测试装置', 'XYZ测试设备', 'PTT-1000', '0.2级');
