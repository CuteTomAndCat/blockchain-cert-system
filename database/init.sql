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

-- 计量证书表（增加区块链哈希字段）
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
    blockchain_hash VARCHAR(256) COMMENT '区块链哈希值',
    status ENUM('draft', 'testing', 'completed', 'issued', 'revoked') DEFAULT 'draft' COMMENT '证书状态',
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
    FOREIGN KEY (cert_id) REFERENCES certificates(id) ON DELETE CASCADE
);

-- 区块链交易记录表
CREATE TABLE IF NOT EXISTS blockchain_transactions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    tx_id VARCHAR(128) UNIQUE NOT NULL COMMENT '交易ID',
    block_number BIGINT COMMENT '区块号',
    block_hash VARCHAR(256) COMMENT '区块哈希',
    transaction_hash VARCHAR(256) COMMENT '交易哈希',
    cert_id BIGINT COMMENT '关联证书ID',
    operation_type ENUM('create', 'update', 'verify', 'revoke') COMMENT '操作类型',
    operator_id BIGINT COMMENT '操作人ID',
    status ENUM('pending', 'confirmed', 'failed') DEFAULT 'pending' COMMENT '交易状态',
    gas_used INT COMMENT '消耗的Gas',
    confirmation_time TIMESTAMP NULL COMMENT '确认时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cert_id) REFERENCES certificates(id),
    FOREIGN KEY (operator_id) REFERENCES users(id)
);

-- 证书历史记录表（用于追踪证书变更）
CREATE TABLE IF NOT EXISTS certificate_history (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    cert_id BIGINT NOT NULL COMMENT '证书ID',
    cert_number VARCHAR(100) NOT NULL COMMENT '证书编号',
    operation_type ENUM('create', 'update', 'verify', 'issue', 'revoke') COMMENT '操作类型',
    changed_fields JSON COMMENT '变更的字段（JSON格式）',
    old_values JSON COMMENT '旧值（JSON格式）',
    new_values JSON COMMENT '新值（JSON格式）',
    blockchain_tx_id VARCHAR(128) COMMENT '区块链交易ID',
    operator_id BIGINT COMMENT '操作人ID',
    operation_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
    FOREIGN KEY (cert_id) REFERENCES certificates(id),
    FOREIGN KEY (operator_id) REFERENCES users(id)
);

-- 创建索引以提高查询性能
CREATE INDEX idx_cert_number ON certificates(cert_number);
CREATE INDEX idx_cert_status ON certificates(status);
CREATE INDEX idx_cert_test_date ON certificates(test_date);
CREATE INDEX idx_cert_expire_date ON certificates(expire_date);
CREATE INDEX idx_blockchain_tx_id ON certificates(blockchain_tx_id);
CREATE INDEX idx_blockchain_hash ON certificates(blockchain_hash);
CREATE INDEX idx_device_addr ON test_data(device_addr);
CREATE INDEX idx_test_timestamp ON test_data(test_timestamp);
CREATE INDEX idx_blockchain_tx ON blockchain_transactions(tx_id);
CREATE INDEX idx_block_number ON blockchain_transactions(block_number);
CREATE INDEX idx_cert_history ON certificate_history(cert_id, operation_time);

-- 插入初始管理员用户（密码：admin123，实际使用时应使用强密码）
INSERT INTO users (username, password_hash, role) VALUES 
('admin', SHA2('admin123', 256), 'admin'),
('operator1', SHA2('oper123', 256), 'operator'),
('viewer1', SHA2('view123', 256), 'viewer');

-- 插入示例数据（用于测试，可根据需要修改）
INSERT INTO customers (customer_name, customer_address, contact_person, contact_phone) VALUES 
('XX电力公司', '北京市朝阳区XXX路123号', '张三', '13800138001'),
('YY制造企业', '上海市浦东新区YYY街456号', '李四', '13900139002'),
('ZZ科技有限公司', '深圳市南山区ZZZ大道789号', '王五', '13700137003');

INSERT INTO devices (device_addr, device_name, manufacturer, model, accuracy_class) VALUES 
('DEV001', '电流互感器测试装置', 'ABC仪器公司', 'CTT-2000', '0.1级'),
('DEV002', '电压互感器测试装置', 'XYZ测试设备', 'PTT-1000', '0.2级'),
('DEV003', '多功能校准仪', 'DEF精密仪器', 'MFC-3000', '0.05级');

-- 插入示例证书数据（包含区块链信息）
INSERT INTO certificates (
    cert_number, customer_id, instrument_name, instrument_number, 
    manufacturer, model_spec, instrument_accuracy, test_date, expire_date, 
    test_result, blockchain_tx_id, blockchain_hash, status, created_by
) VALUES 
(
    'CERT-2024-001', 1, '电流互感器', 'CT-001', 
    'ABC电气', 'LMZ-10', '0.2级', '2024-01-15', '2025-01-15', 
    'qualified', 
    '0x1a2b3c4d5e6f7890abcdef1234567890abcdef12',
    'a5f3b8c9d2e1f4a7b6c5d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9',
    'issued', 1
),
(
    'CERT-2024-002', 2, '电压互感器', 'PT-002', 
    'XYZ电气', 'JDZ-10', '0.5级', '2024-02-20', '2025-02-20', 
    'qualified',
    '0x2b3c4d5e6f7890ab1234567890abcdef12345678',
    'b6f4c9d0e2f5a8b7c6d9f0a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2',
    'issued', 1
);

-- 插入示例测试数据
INSERT INTO test_data (
    cert_id, device_addr, data_type, percentage_value, ratio_error, angle_error,
    current_value, voltage_value, workstation_number, test_point, actual_percentage,
    blockchain_hash, encrypted_data
) VALUES 
(1, 'DEV001', 'current', 100.00, 0.15, 0.08, 5.00, 0, 'WS01', 'P1', 99.85,
 'c7f5d0a1e3f6b9c8d7e0f1a2b3c4d5e6', 'encrypted_data_1'),
(1, 'DEV001', 'current', 80.00, 0.12, 0.06, 4.00, 0, 'WS01', 'P2', 79.88,
 'd8f6e1b2f4a7c9d8e1f2a3b4c5d6e7f8', 'encrypted_data_2'),
(2, 'DEV002', 'voltage', 100.00, 0.10, 0.05, 0, 100.00, 'WS02', 'P1', 99.90,
 'e9f7f2c3a5b8d0e9f2a3b4c5d6e7f8a9', 'encrypted_data_3');

-- 插入示例区块链交易记录
INSERT INTO blockchain_transactions (
    tx_id, block_number, block_hash, transaction_hash, cert_id, 
    operation_type, operator_id, status, gas_used, confirmation_time
) VALUES 
(
    '0x1a2b3c4d5e6f7890abcdef1234567890abcdef12', 1001, 
    '0xblock1hash234567890abcdef', '0xtx1hash234567890abcdef',
    1, 'create', 1, 'confirmed', 21000, '2024-01-15 10:30:00'
),
(
    '0x2b3c4d5e6f7890ab1234567890abcdef12345678', 1002,
    '0xblock2hash345678901bcdef', '0xtx2hash345678901bcdef',
    2, 'create', 1, 'confirmed', 21500, '2024-02-20 14:45:00'
);

-- 创建视图：证书完整信息视图
CREATE OR REPLACE VIEW v_certificate_details AS
SELECT 
    c.id,
    c.cert_number,
    c.instrument_name,
    c.instrument_number,
    c.manufacturer,
    c.model_spec,
    c.instrument_accuracy,
    c.test_date,
    c.expire_date,
    c.test_result,
    c.blockchain_tx_id,
    c.blockchain_hash,
    c.status,
    c.created_at,
    c.updated_at,
    cust.customer_name,
    cust.customer_address,
    cust.contact_person,
    cust.contact_phone,
    u.username as created_by_username,
    CASE 
        WHEN c.expire_date < CURDATE() THEN '已过期'
        WHEN c.expire_date < DATE_ADD(CURDATE(), INTERVAL 30 DAY) THEN '即将过期'
        ELSE '有效'
    END as validity_status,
    CASE 
        WHEN c.blockchain_hash IS NOT NULL THEN '已上链'
        ELSE '未上链'
    END as blockchain_status
FROM certificates c
LEFT JOIN customers cust ON c.customer_id = cust.id
LEFT JOIN users u ON c.created_by = u.id;

-- 创建存储过程：验证证书区块链哈希
DELIMITER //
CREATE PROCEDURE sp_verify_certificate_hash(
    IN p_cert_number VARCHAR(100),
    OUT p_is_valid BOOLEAN,
    OUT p_message VARCHAR(200)
)
BEGIN
    DECLARE v_blockchain_hash VARCHAR(256);
    DECLARE v_exists INT;
    
    SELECT COUNT(*), blockchain_hash INTO v_exists, v_blockchain_hash
    FROM certificates
    WHERE cert_number = p_cert_number
    GROUP BY blockchain_hash;
    
    IF v_exists = 0 THEN
        SET p_is_valid = FALSE;
        SET p_message = '证书不存在';
    ELSEIF v_blockchain_hash IS NULL THEN
        SET p_is_valid = FALSE;
        SET p_message = '证书未上链';
    ELSE
        SET p_is_valid = TRUE;
        SET p_message = CONCAT('证书已验证，区块链哈希: ', LEFT(v_blockchain_hash, 32), '...');
    END IF;
END//
DELIMITER ;

-- 授权（如果需要的话）
-- GRANT ALL PRIVILEGES ON cert_system.* TO 'certuser'@'%' IDENTIFIED BY 'certpass123';
-- FLUSH PRIVILEGES;