// API配置
const API_BASE_URL = 'http://192.168.85.129:8080/api/v1';
let authToken = localStorage.getItem('authToken');
let currentUser = null;
let testDataRowCount = 1;

// 页面初始化
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
    setupEventListeners();
});

// 初始化应用
function initializeApp() {
    if (authToken) {
        validateToken();
    } else {
        showLoginModal();
    }
    setupNavigation();
}

// 设置事件监听器
function setupEventListeners() {
    document.getElementById('loginForm').addEventListener('submit', handleLogin);
    document.getElementById('logoutBtn').addEventListener('click', handleLogout);
    document.getElementById('createCertForm').addEventListener('submit', handleCreateCert);
    document.getElementById('editCertForm').addEventListener('submit', handleEditCert);
}

// 导航设置
function setupNavigation() {
    const navLinks = document.querySelectorAll('.nav-link');
    navLinks.forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            const target = this.getAttribute('href').substring(1);
            showSection(target);
            
            navLinks.forEach(l => l.classList.remove('active'));
            this.classList.add('active');
        });
    });
}

// 显示页面部分
function showSection(sectionId) {
    const sections = document.querySelectorAll('.content-section');
    sections.forEach(section => {
        section.classList.remove('active');
    });
    
    const targetSection = document.getElementById(sectionId);
    if (targetSection) {
        targetSection.classList.add('active');
        
        switch(sectionId) {
            case 'dashboard':
                loadDashboard();
                break;
            case 'certificates':
                loadCertificates();
                break;
            case 'reports':
                loadReports();
                break;
        }
    }
}

// 用户认证
async function handleLogin(e) {
    e.preventDefault();
    const username = document.getElementById('loginUsername').value;
    const password = document.getElementById('loginPassword').value;
    
    try {
        const response = await fetch(`${API_BASE_URL}/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, password })
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            authToken = data.data.token;
            localStorage.setItem('authToken', authToken);
            currentUser = {
                id: data.data.userId,
                username: data.data.username,
                role: data.data.role
            };
            
            document.getElementById('username').textContent = currentUser.username;
            document.getElementById('logoutBtn').style.display = 'block';
            closeModal('loginModal');
            loadDashboard();
            showNotification('登录成功', 'success');
        } else {
            showError('loginError', data.message || '登录失败');
        }
    } catch (error) {
        showError('loginError', '网络错误，请稍后重试');
    }
}

// 验证Token
async function validateToken() {
    try {
        const response = await fetch(`${API_BASE_URL}/auth/profile`, {
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            currentUser = data.data;
            document.getElementById('username').textContent = currentUser.username;
            document.getElementById('logoutBtn').style.display = 'block';
            loadDashboard();
        } else {
            showLoginModal();
        }
    } catch (error) {
        showLoginModal();
    }
}

// 登出
async function handleLogout() {
    try {
        await fetch(`${API_BASE_URL}/auth/logout`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
    } catch (error) {
        console.error('Logout error:', error);
    }
    
    localStorage.removeItem('authToken');
    authToken = null;
    currentUser = null;
    document.getElementById('username').textContent = '未登录';
    document.getElementById('logoutBtn').style.display = 'none';
    showLoginModal();
}

// 创建证书（包含测试数据）
async function handleCreateCert(e) {
    e.preventDefault();
    
    const formData = new FormData(e.target);
    
    // 收集证书基本信息
    const certData = {
        certNumber: formData.get('certNumber'),
        customerId: parseInt(formData.get('customerId')),
        instrumentName: formData.get('instrumentName'),
        instrumentNumber: formData.get('instrumentNumber'),
        manufacturer: formData.get('manufacturer'),
        modelSpec: formData.get('modelSpec'),
        instrumentAccuracy: formData.get('instrumentAccuracy'),
        testDate: formData.get('testDate'),
        expireDate: formData.get('expireDate'),
        testResult: formData.get('testResult')
    };
    
    // 收集测试数据
    const testDataArray = [];
    const testDataRows = document.querySelectorAll('.test-data-row');
    testDataRows.forEach((row, index) => {
        testDataArray.push({
            deviceAddr: formData.get(`deviceAddr_${index}`),
            testPoint: formData.get(`testPoint_${index}`),
            actualPercentage: parseFloat(formData.get(`actualPercentage_${index}`)),
            ratioError: parseFloat(formData.get(`ratioError_${index}`)),
            angleError: parseFloat(formData.get(`angleError_${index}`)),
            testTimestamp: new Date().toISOString()
        });
    });
    
    try {
        // 1. 创建证书
        const certResponse = await fetch(`${API_BASE_URL}/certificates`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify(certData)
        });
        
        const certResult = await response.json();
        
        if (certResult.code === 201 || certResult.code === 200) {
            // 2. 添加测试数据
            const testDataPayload = {
                certNumber: certData.certNumber,
                data: testDataArray
            };
            
            const testResponse = await fetch(`${API_BASE_URL}/test-data`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${authToken}`
                },
                body: JSON.stringify(testDataPayload)
            });
            
            const testResult = await testResponse.json();
            
            if (testResult.code === 201 || testResult.code === 200) {
                showNotification('证书创建成功，测试数据已添加', 'success');
                closeModal('createCertModal');
                loadCertificates();
                e.target.reset();
                resetTestDataRows();
            } else {
                showNotification('证书创建成功，但测试数据添加失败', 'warning');
            }
        } else {
            showNotification(certResult.message || '创建失败', 'error');
        }
    } catch (error) {
        showNotification('网络错误', 'error');
    }
}

// 添加测试数据行
function addTestDataRow() {
    const container = document.getElementById('testDataContainer');
    const newRow = document.createElement('div');
    newRow.className = 'test-data-row';
    newRow.dataset.index = testDataRowCount;
    
    newRow.innerHTML = `
        <div class="test-data-grid">
            <div class="form-group">
                <label>设备地址</label>
                <input type="text" name="deviceAddr_${testDataRowCount}" value="DEV-01" required>
            </div>
            <div class="form-group">
                <label>测试点</label>
                <input type="text" name="testPoint_${testDataRowCount}" placeholder="如: P${testDataRowCount + 1}" required>
            </div>
            <div class="form-group">
                <label>实际百分比</label>
                <input type="number" name="actualPercentage_${testDataRowCount}" step="0.01" required>
            </div>
            <div class="form-group">
                <label>比差值</label>
                <input type="number" name="ratioError_${testDataRowCount}" step="0.01" required>
            </div>
            <div class="form-group">
                <label>角差值</label>
                <input type="number" name="angleError_${testDataRowCount}" step="0.01" required>
            </div>
            <button type="button" class="btn btn-danger btn-small" onclick="removeTestDataRow(${testDataRowCount})">删除</button>
        </div>
    `;
    
    container.appendChild(newRow);
    testDataRowCount++;
}

// 删除测试数据行
function removeTestDataRow(index) {
    const row = document.querySelector(`.test-data-row[data-index="${index}"]`);
    if (row) {
        row.remove();
    }
}

// 重置测试数据行
function resetTestDataRows() {
    const container = document.getElementById('testDataContainer');
    container.innerHTML = `
        <div class="test-data-row" data-index="0">
            <div class="test-data-grid">
                <div class="form-group">
                    <label>设备地址</label>
                    <input type="text" name="deviceAddr_0" value="DEV-01" required>
                </div>
                <div class="form-group">
                    <label>测试点</label>
                    <input type="text" name="testPoint_0" placeholder="如: P1" required>
                </div>
                <div class="form-group">
                    <label>实际百分比</label>
                    <input type="number" name="actualPercentage_0" step="0.01" required>
                </div>
                <div class="form-group">
                    <label>比差值</label>
                    <input type="number" name="ratioError_0" step="0.01" required>
                </div>
                <div class="form-group">
                    <label>角差值</label>
                    <input type="number" name="angleError_0" step="0.01" required>
                </div>
            </div>
        </div>
    `;
    testDataRowCount = 1;
}

// 查看证书详情（包含测试数据和流程）
async function viewCertificate(certNumber) {
    try {
        // 获取证书信息
        const certResponse = await fetch(`${API_BASE_URL}/certificates/${certNumber}`, {
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        const certData = await certResponse.json();
        
        // 获取测试数据
        const testResponse = await fetch(`${API_BASE_URL}/test-data/certificate/${certNumber}`, {
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        const testData = await testResponse.json();
        
        if (certData.code === 200) {
            const cert = certData.data.certificate || certData.data;
            const blockchain = certData.data.blockchain;
            
            // 生成证书流程HTML
            const workflowHtml = generateWorkflowHtml(cert.status);
            
            // 生成测试数据表格
            let testDataHtml = '<h3>测试数据</h3>';
            if (testData.code === 200 && testData.data && testData.data.length > 0) {
                testDataHtml += `
                    <table class="data-table">
                        <thead>
                            <tr>
                                <th>设备地址</th>
                                <th>测试点</th>
                                <th>实际百分比</th>
                                <th>比差值</th>
                                <th>角差值</th>
                                <th>测试时间</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${testData.data.map(td => `
                                <tr>
                                    <td>${td.deviceAddr}</td>
                                    <td>${td.testPoint}</td>
                                    <td>${td.actualPercentage}%</td>
                                    <td>${td.ratioError}</td>
                                    <td>${td.angleError}</td>
                                    <td>${formatDateTime(td.testTimestamp)}</td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                `;
            } else {
                testDataHtml += '<p>暂无测试数据</p>';
            }
            
            const detailHtml = `
                <div class="cert-detail">
                    <h3>证书流程</h3>
                    ${workflowHtml}
                    
                    <h3>基本信息</h3>
                    <div class="detail-grid">
                        <div class="detail-item">
                            <label>证书编号:</label>
                            <span>${cert.certNumber}</span>
                        </div>
                        <div class="detail-item">
                            <label>器具名称:</label>
                            <span>${cert.instrumentName}</span>
                        </div>
                        <div class="detail-item">
                            <label>器具编号:</label>
                            <span>${cert.instrumentNumber || '-'}</span>
                        </div>
                        <div class="detail-item">
                            <label>制造厂:</label>
                            <span>${cert.manufacturer || '-'}</span>
                        </div>
                        <div class="detail-item">
                            <label>型号规格:</label>
                            <span>${cert.modelSpec || '-'}</span>
                        </div>
                        <div class="detail-item">
                            <label>准确度:</label>
                            <span>${cert.instrumentAccuracy || '-'}</span>
                        </div>
                        <div class="detail-item">
                            <label>测试日期:</label>
                            <span>${formatDate(cert.testDate)}</span>
                        </div>
                        <div class="detail-item">
                            <label>有效期至:</label>
                            <span>${formatDate(cert.expireDate)}</span>
                        </div>
                        <div class="detail-item">
                            <label>测试结果:</label>
                            <span>${cert.testResult === 'qualified' ? '✅ 合格' : '❌ 不合格'}</span>
                        </div>
                        <div class="detail-item">
                            <label>状态:</label>
                            <span class="status-badge status-${cert.status}">${getStatusText(cert.status)}</span>
                        </div>
                    </div>
                    
                    ${testDataHtml}
                    
                    ${cert.blockchainTxId ? `
                        <h3>区块链信息</h3>
                        <div class="blockchain-info">
                            <div class="blockchain-item">
                                <label>交易ID:</label>
                                <span class="hash-value">${cert.blockchainTxId}</span>
                            </div>
                            <div class="blockchain-item">
                                <label>区块链哈希:</label>
                                <span class="hash-value">${cert.blockchainHash || '-'}</span>
                            </div>
                        </div>
                    ` : '<p>该证书尚未上链</p>'}
                    
                    <div class="modal-footer">
                        <button class="btn btn-primary" onclick="editCertificate('${cert.certNumber}')">编辑证书</button>
                        <button class="btn btn-secondary" onclick="closeModal('viewCertModal')">关闭</button>
                    </div>
                </div>
            `;
            
            document.getElementById('certDetailContent').innerHTML = detailHtml;
            showModal('viewCertModal');
        }
    } catch (error) {
        showNotification('加载证书详情失败', 'error');
    }
}

// 编辑证书
async function editCertificate(certNumber) {
    try {
        const response = await fetch(`${API_BASE_URL}/certificates/${certNumber}`, {
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            const cert = data.data.certificate || data.data;
            
            // 填充编辑表单
            document.getElementById('editCertNumber').value = cert.certNumber;
            document.getElementById('editCertNumberDisplay').value = cert.certNumber;
            document.getElementById('editCustomerId').value = cert.customerId;
            document.getElementById('editInstrumentName').value = cert.instrumentName;
            document.getElementById('editInstrumentNumber').value = cert.instrumentNumber || '';
            document.getElementById('editManufacturer').value = cert.manufacturer || '';
            document.getElementById('editModelSpec').value = cert.modelSpec || '';
            document.getElementById('editInstrumentAccuracy').value = cert.instrumentAccuracy || '';
            document.getElementById('editTestDate').value = cert.testDate.split('T')[0];
            document.getElementById('editExpireDate').value = cert.expireDate.split('T')[0];
            document.getElementById('editTestResult').value = cert.testResult;
            document.getElementById('editStatus').value = cert.status;
            
            closeModal('viewCertModal');
            showModal('editCertModal');
        }
    } catch (error) {
        showNotification('加载证书信息失败', 'error');
    }
}

// 处理编辑证书提交
async function handleEditCert(e) {
    e.preventDefault();
    
    const certNumber = document.getElementById('editCertNumber').value;
    
    const certData = {
        certNumber: certNumber,
        customerId: parseInt(document.getElementById('editCustomerId').value),
        instrumentName: document.getElementById('editInstrumentName').value,
        instrumentNumber: document.getElementById('editInstrumentNumber').value,
        manufacturer: document.getElementById('editManufacturer').value,
        modelSpec: document.getElementById('editModelSpec').value,
        instrumentAccuracy: document.getElementById('editInstrumentAccuracy').value,
        testDate: document.getElementById('editTestDate').value,
        expireDate: document.getElementById('editExpireDate').value,
        testResult: document.getElementById('editTestResult').value,
        status: document.getElementById('editStatus').value
    };
    
    try {
        const response = await fetch(`${API_BASE_URL}/certificates/${certNumber}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify(certData)
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            showNotification('证书更新成功', 'success');
            closeModal('editCertModal');
            loadCertificates();
        } else {
            showNotification(data.message || '更新失败', 'error');
        }
    } catch (error) {
        showNotification('网络错误', 'error');
    }
}

// 验证证书（包含流程和测试数据）
async function verifyCertificate() {
    const certNumber = document.getElementById('verifyCertNumber').value;
    
    if (!certNumber) {
        showNotification('请输入证书编号', 'warning');
        return;
    }
    
    try {
        // 验证证书
        const verifyResponse = await fetch(`${API_BASE_URL}/public/verify/${certNumber}`);
        const verifyData = await verifyResponse.json();
        
        // 获取测试数据
        let testData = null;
        if (verifyData.data.isValid) {
            try {
                const testResponse = await fetch(`${API_BASE_URL}/test-data/certificate/${certNumber}`, {
                    headers: {
                        'Authorization': `Bearer ${authToken}`
                    }
                });
                testData = await testResponse.json();
            } catch (error) {
                console.log('获取测试数据失败');
            }
        }
        
        if (verifyData.code === 200) {
            const result = verifyData.data;
            const resultDiv = document.getElementById('verifyResult');
            const statusDiv = document.getElementById('verifyStatus');
            const detailsDiv = document.getElementById('certDetails');
            const workflowDiv = document.getElementById('certWorkflow');
            const testDataDiv = document.getElementById('certTestData');
            const blockchainDiv = document.getElementById('blockchainInfo');
            
            resultDiv.style.display = 'block';
            
            if (result.isValid) {
                statusDiv.className = 'verify-status valid';
                statusDiv.innerHTML = `
                    <h2>✅ 证书有效</h2>
                    <p>${result.message}</p>
                `;
                
                if (result.certificate) {
                    // 显示证书流程
                    workflowDiv.innerHTML = `
                        <h3>证书流程状态</h3>
                        ${generateWorkflowHtml(result.certificate.status)}
                    `;
                    
                    // 显示证书详情
                    detailsDiv.innerHTML = `
                        <h3>证书信息</h3>
                        <div class="detail-grid">
                            <div class="detail-item">
                                <label>证书编号:</label>
                                <span>${result.certificate.certNumber}</span>
                            </div>
                            <div class="detail-item">
                                <label>器具名称:</label>
                                <span>${result.certificate.instrumentName}</span>
                            </div>
                            <div class="detail-item">
                                <label>制造厂:</label>
                                <span>${result.certificate.manufacturer || '-'}</span>
                            </div>
                            <div class="detail-item">
                                <label>测试日期:</label>
                                <span>${formatDate(result.certificate.testDate)}</span>
                            </div>
                            <div class="detail-item">
                                <label>有效期至:</label>
                                <span>${formatDate(result.certificate.expireDate)}</span>
                            </div>
                            <div class="detail-item">
                                <label>状态:</label>
                                <span class="status-badge status-${result.certificate.status}">${getStatusText(result.certificate.status)}</span>
                            </div>
                        </div>
                    `;
                    
                    // 显示测试数据
                    if (testData && testData.data && testData.data.length > 0) {
                        testDataDiv.innerHTML = `
                            <h3>测试数据</h3>
                            <table class="data-table">
                                <thead>
                                    <tr>
                                        <th>测试点</th>
                                        <th>实际百分比</th>
                                        <th>比差值</th>
                                        <th>角差值</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${testData.data.map(td => `
                                        <tr>
                                            <td>${td.testPoint}</td>
                                            <td>${td.actualPercentage}%</td>
                                            <td>${td.ratioError}</td>
                                            <td>${td.angleError}</td>
                                        </tr>
                                    `).join('')}
                                </tbody>
                            </table>
                        `;
                    } else {
                        testDataDiv.innerHTML = '';
                    }
                }
                
                if (result.blockchainTxId) {
                    blockchainDiv.innerHTML = `
                        <h4>🔗 区块链验证信息</h4>
                        <div class="blockchain-item">
                            <span>交易ID:</span>
                            <span class="hash-value">${result.blockchainTxId}</span>
                        </div>
                        <div class="blockchain-item">
                            <span>区块链哈希:</span>
                            <span class="hash-value">${result.blockchainHash}</span>
                        </div>
                    `;
                }
            } else {
                statusDiv.className = 'verify-status invalid';
                statusDiv.innerHTML = `
                    <h2>❌ 证书无效</h2>
                    <p>${result.message}</p>
                `;
                detailsDiv.innerHTML = '';
                workflowDiv.innerHTML = '';
                testDataDiv.innerHTML = '';
                blockchainDiv.innerHTML = '';
            }
        }
    } catch (error) {
        showNotification('验证失败，请检查网络连接', 'error');
    }
}

// 生成证书流程HTML
function generateWorkflowHtml(currentStatus) {
    const workflow = [
        { status: 'draft', label: '草稿', icon: '📝' },
        { status: 'testing', label: '测试中', icon: '🔬' },
        { status: 'completed', label: '已完成', icon: '✔️' },
        { status: 'issued', label: '已签发', icon: '📜' }
    ];
    
    let currentIndex = workflow.findIndex(w => w.status === currentStatus);
    if (currentStatus === 'revoked') {
        currentIndex = -1; // 撤销状态特殊处理
    }
    
    let html = '<div class="workflow-container">';
    
    workflow.forEach((step, index) => {
        const isActive = index <= currentIndex;
        const isCurrent = step.status === currentStatus;
        
        html += `
            <div class="workflow-step ${isActive ? 'active' : ''} ${isCurrent ? 'current' : ''}">
                <div class="workflow-icon">${step.icon}</div>
                <div class="workflow-label">${step.label}</div>
            </div>
        `;
        
        if (index < workflow.length - 1) {
            html += `<div class="workflow-line ${isActive ? 'active' : ''}"></div>`;
        }
    });
    
    if (currentStatus === 'revoked') {
        html += `
            <div class="workflow-line"></div>
            <div class="workflow-step revoked current">
                <div class="workflow-icon">❌</div>
                <div class="workflow-label">已撤销</div>
            </div>
        `;
    }
    
    html += '</div>';
    return html;
}


// 加载仪表盘
async function loadDashboard() {
    try {
        const response = await fetch(`${API_BASE_URL}/certificates`, {
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            const certs = data.data || [];
            
            // 更新统计
            document.getElementById('totalCerts').textContent = data.total || certs.length;
            
            const validCount = certs.filter(c => c.status === 'issued').length;
            document.getElementById('validCerts').textContent = validCount;
            
            const expiringCount = certs.filter(c => {
                const expireDate = new Date(c.expireDate);
                const daysUntilExpire = (expireDate - new Date()) / (1000 * 60 * 60 * 24);
                return daysUntilExpire > 0 && daysUntilExpire < 30;
            }).length;
            document.getElementById('expiringSoon').textContent = expiringCount;
            
            const blockchainCount = certs.filter(c => c.blockchainTxId).length;
            document.getElementById('blockchainCerts').textContent = blockchainCount;
            
            // 更新活动列表
            updateActivityList(certs.slice(0, 5));
        }
    } catch (error) {
        console.error('Dashboard load error:', error);
    }
}

// 加载证书列表
async function loadCertificates(page = 1) {
    try {
        const searchTerm = document.getElementById('searchCert').value;
        const statusFilter = document.getElementById('statusFilter').value;
        
        let url = `${API_BASE_URL}/certificates?page=${page}&pageSize=10`;
        
        const response = await fetch(url, {
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            displayCertificates(data.data || []);
            setupPagination(data.total || 0, page, data.totalPages || 1);
        }
    } catch (error) {
        console.error('Load certificates error:', error);
        showNotification('加载证书列表失败', 'error');
    }
}

// 显示证书列表
function displayCertificates(certificates) {
    const tbody = document.getElementById('certTableBody');
    
    if (certificates.length === 0) {
        tbody.innerHTML = '<tr><td colspan="8" class="text-center">暂无数据</td></tr>';
        return;
    }
    
    tbody.innerHTML = certificates.map(cert => `
        <tr>
            <td>${cert.certNumber}</td>
            <td>${cert.instrumentName}</td>
            <td>客户${cert.customerId}</td>
            <td>${formatDate(cert.testDate)}</td>
            <td>${formatDate(cert.expireDate)}</td>
            <td><span class="status-badge status-${getStatusText(cert.status)}">${getStatusText(cert.status)}</span></td>
            <td>${cert.blockchainTxId ? '✅ 已上链' : '❌ 未上链'}</td>
            <td>
                <button class="btn btn-small btn-primary" onclick="viewCertificate('${cert.certNumber}')">查看</button>
                <button class="btn btn-small btn-warning" onclick="editCertificate('${cert.certNumber}')">编辑</button>
                ${cert.status !== 'revoked' ? 
                    `<button class="btn btn-small btn-danger" onclick="revokeCertificate('${cert.certNumber}')">撤销</button>` : 
                    ''}
            </td>
        </tr>
    `).join('');
}

// 创建证书
async function handleCreateCert(e) {
    e.preventDefault();
    
    const formData = new FormData(e.target);
    const certData = {
        certNumber: formData.get('certNumber'),
        customerId: parseInt(formData.get('customerId')),
        instrumentName: formData.get('instrumentName'),
        instrumentNumber: formData.get('instrumentNumber'),
        manufacturer: formData.get('manufacturer'),
        modelSpec: formData.get('modelSpec'),
        instrumentAccuracy: formData.get('instrumentAccuracy'),
        testDate: formData.get('testDate'),
        expireDate: formData.get('expireDate'),
        testResult: formData.get('testResult')
    };
    
    try {
        const response = await fetch(`${API_BASE_URL}/certificates`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify(certData)
        });
        
        const data = await response.json();
        
        if (data.code === 201 || data.code === 200) {
            showNotification('证书创建成功', 'success');
            closeModal('createCertModal');
            loadCertificates();
            e.target.reset();
        } else {
            showNotification(data.message || '创建失败', 'error');
        }
    } catch (error) {
        showNotification('网络错误', 'error');
    }
}

// 查看证书详情
async function viewCertificate(certNumber) {
    try {
        const response = await fetch(`${API_BASE_URL}/certificates/${certNumber}`, {
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            const cert = data.data.certificate || data.data;
            const blockchain = data.data.blockchain;
            
            const detailHtml = `
                <div class="cert-detail">
                    <h3>基本信息</h3>
                    <div class="detail-grid">
                        <div class="detail-item">
                            <label>证书编号:</label>
                            <span>${cert.certNumber}</span>
                        </div>
                        <div class="detail-item">
                            <label>器具名称:</label>
                            <span>${cert.instrumentName}</span>
                        </div>
                        <div class="detail-item">
                            <label>器具编号:</label>
                            <span>${cert.instrumentNumber || '-'}</span>
                        </div>
                        <div class="detail-item">
                            <label>制造厂:</label>
                            <span>${cert.manufacturer || '-'}</span>
                        </div>
                        <div class="detail-item">
                            <label>型号规格:</label>
                            <span>${cert.modelSpec || '-'}</span>
                        </div>
                        <div class="detail-item">
                            <label>准确度:</label>
                            <span>${cert.instrumentAccuracy || '-'}</span>
                        </div>
                        <div class="detail-item">
                            <label>测试日期:</label>
                            <span>${formatDate(cert.testDate)}</span>
                        </div>
                        <div class="detail-item">
                            <label>有效期至:</label>
                            <span>${formatDate(cert.expireDate)}</span>
                        </div>
                        <div class="detail-item">
                            <label>测试结果:</label>
                            <span>${cert.testResult === 'qualified' ? '✅ 合格' : '❌ 不合格'}</span>
                        </div>
                        <div class="detail-item">
                            <label>状态:</label>
                            <span class="status-badge status-${getStatusText(cert.status)}">${getStatusText(cert.status)}</span>
                        </div>
                    </div>
                    
                    ${cert.blockchainTxId ? `
                        <h3>区块链信息</h3>
                        <div class="blockchain-info">
                            <div class="blockchain-item">
                                <label>交易ID:</label>
                                <span class="hash-value">${cert.blockchainTxId}</span>
                            </div>
                            <div class="blockchain-item">
                                <label>区块链哈希:</label>
                                <span class="hash-value">${cert.blockchainHash || '-'}</span>
                            </div>
                        </div>
                    ` : '<p>该证书尚未上链</p>'}
                </div>
            `;
            
            document.getElementById('certDetailContent').innerHTML = detailHtml;
            showModal('viewCertModal');
        }
    } catch (error) {
        showNotification('加载证书详情失败', 'error');
    }
}

// 验证证书
async function verifyCertificate() {
    const certNumber = document.getElementById('verifyCertNumber').value;
    
    if (!certNumber) {
        showNotification('请输入证书编号', 'warning');
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE_URL}/public/verify/${certNumber}`);
        const data = await response.json();
        
        if (data.code === 200) {
            const result = data.data;
            const resultDiv = document.getElementById('verifyResult');
            const statusDiv = document.getElementById('verifyStatus');
            const detailsDiv = document.getElementById('certDetails');
            const blockchainDiv = document.getElementById('blockchainInfo');
            
            resultDiv.style.display = 'block';
            
            if (result.isValid) {
                statusDiv.className = 'verify-status valid';
                statusDiv.innerHTML = `
                    <h2>✅ 证书有效</h2>
                    <p>${result.message}</p>
                `;
                
                if (result.certificate) {
                    detailsDiv.innerHTML = `
                        <h3>证书信息</h3>
                        <p><strong>证书编号:</strong> ${result.certificate.certNumber}</p>
                        <p><strong>器具名称:</strong> ${result.certificate.instrumentName}</p>
                        <p><strong>有效期至:</strong> ${formatDate(result.certificate.expireDate)}</p>
                        <p><strong>状态:</strong> ${getStatusText(result.certificate.status)}</p>
                    `;
                }
                
                if (result.blockchainTxId) {
                    blockchainDiv.innerHTML = `
                        <h4>🔗 区块链验证信息</h4>
                        <div class="blockchain-item">
                            <span>交易ID:</span>
                            <span class="hash-value">${result.blockchainTxId}</span>
                        </div>
                        <div class="blockchain-item">
                            <span>区块链哈希:</span>
                            <span class="hash-value">${result.blockchainHash}</span>
                        </div>
                    `;
                }
            } else {
                statusDiv.className = 'verify-status invalid';
                statusDiv.innerHTML = `
                    <h2>❌ 证书无效</h2>
                    <p>${result.message}</p>
                `;
                detailsDiv.innerHTML = '';
                blockchainDiv.innerHTML = '';
            }
        }
    } catch (error) {
        showNotification('验证失败，请检查网络连接', 'error');
    }
}

// 添加测试数据
async function handleAddTestData(e) {
    e.preventDefault();
    
    const testData = {
        certNumber: document.getElementById('testCertNumber').value,
        data: [{
            deviceAddr: document.getElementById('deviceAddr').value,
            testPoint: document.getElementById('testPoint').value,
            actualPercentage: parseFloat(document.getElementById('actualPercentage').value),
            ratioError: parseFloat(document.getElementById('ratioError').value),
            angleError: parseFloat(document.getElementById('angleError').value),
            testTimestamp: new Date().toISOString()
        }]
    };
    
    try {
        const response = await fetch(`${API_BASE_URL}/test-data`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify(testData)
        });
        
        const data = await response.json();
        
        if (data.code === 201 || data.code === 200) {
            showNotification('测试数据添加成功', 'success');
            e.target.reset();
            loadTestDataForCert(testData.certNumber);
        } else {
            showNotification(data.message || '添加失败', 'error');
        }
    } catch (error) {
        showNotification('网络错误', 'error');
    }
}

// 工具函数
// 工具函数
function formatDate(dateString) {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleDateString('zh-CN');
}

function formatDateTime(dateString) {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN');
}

function getStatusText(status) {
    const statusMap = {
        'draft': '草稿',
        'testing': '测试中',
        'completed': '已完成',
        'issued': '已签发',
        'revoked': '已撤销'
    };
    return statusMap[status] || status;
}

function showModal(modalId) {
    document.getElementById(modalId).classList.add('show');
}

function closeModal(modalId) {
    document.getElementById(modalId).classList.remove('show');
}

function showCreateCertModal() {
    document.getElementById('newCertNumber').value = 'CERT-' + Date.now();
    showModal('createCertModal');
}

function showLoginModal() {
    showModal('loginModal');
}

function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.textContent = message;
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 1rem 1.5rem;
        background: ${type === 'success' ? '#22c55e' : type === 'error' ? '#ef4444' : type === 'warning' ? '#f59e0b' : '#3b82f6'};
        color: white;
        border-radius: 0.5rem;
        z-index: 10000;
        animation: slideIn 0.3s;
        box-shadow: 0 4px 6px rgba(0,0,0,0.1);
    `;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s';
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

function showError(elementId, message) {
    const errorElement = document.getElementById(elementId);
    errorElement.textContent = message;
    errorElement.classList.add('show');
    
    setTimeout(() => {
        errorElement.classList.remove('show');
    }, 5000);
}


function formatDate(dateString) {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleDateString('zh-CN');
}

function getStatusText(status) {
    const statusMap = {
        'draft': '草稿',
        'testing': '测试中',
        'completed': '已完成',
        'issued': '已签发',
        'revoked': '已撤销'
    };
    return statusMap[status] || status;
}

function setupPagination(total, currentPage, totalPages) {
    const pagination = document.getElementById('certPagination');
    
    if (totalPages <= 1) {
        pagination.innerHTML = '';
        return;
    }
    
    let html = '';
    
    // 上一页
    if (currentPage > 1) {
        html += `<button onclick="loadCertificates(${currentPage - 1})">上一页</button>`;
    }
    
    // 页码
    for (let i = 1; i <= Math.min(totalPages, 5); i++) {
        html += `<button class="${i === currentPage ? 'active' : ''}" onclick="loadCertificates(${i})">${i}</button>`;
    }
    
    // 下一页
    if (currentPage < totalPages) {
        html += `<button onclick="loadCertificates(${currentPage + 1})">下一页</button>`;
    }
    
    pagination.innerHTML = html;
}

// 更新活动列表
function updateActivityList(recentCerts) {
    const activityList = document.getElementById('activityList');
    
    if (recentCerts.length === 0) {
        activityList.innerHTML = '<li>暂无最近活动</li>';
        return;
    }
    
    activityList.innerHTML = recentCerts.map(cert => `
        <li>
            📋 证书 ${cert.certNumber} - ${cert.instrumentName} 
            <span style="color: var(--secondary-color); font-size: 0.9rem;">
                ${formatDate(cert.createdAt)}
            </span>
        </li>
    `).join('');
}

// 编辑证书
async function editCertificate(certNumber) {
    // 这里可以实现编辑功能
    showNotification('编辑功能开发中', 'info');
}

// 撤销证书
async function revokeCertificate(certNumber) {
    if (!confirm(`确定要撤销证书 ${certNumber} 吗？`)) {
        return;
    }
    
    try {
        // 先获取证书信息
        const getResponse = await fetch(`${API_BASE_URL}/certificates/${certNumber}`, {
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        const getData = await getResponse.json();
        const cert = getData.data.certificate || getData.data;
        
        // 更新状态为revoked
        cert.status = 'revoked';
        
        const response = await fetch(`${API_BASE_URL}/certificates/${certNumber}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify(cert)
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            showNotification('证书已撤销', 'success');
            loadCertificates();
        } else {
            showNotification('撤销失败', 'error');
        }
    } catch (error) {
        showNotification('操作失败', 'error');
    }
}

// 加载测试数据部分
function loadTestDataSection() {
    // 可以在这里加载测试数据列表
}

// 加载指定证书的测试数据
async function loadTestDataForCert(certNumber) {
    try {
        const response = await fetch(`${API_BASE_URL}/test-data/certificate/${certNumber}`, {
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        const data = await response.json();
        
        if (data.code === 200) {
            const testDataList = document.getElementById('testDataList');
            
            if (data.data && data.data.length > 0) {
                testDataList.innerHTML = `
                    <table class="data-table">
                        <thead>
                            <tr>
                                <th>设备地址</th>
                                <th>测试点</th>
                                <th>实际百分比</th>
                                <th>比差值</th>
                                <th>角差值</th>
                                <th>测试时间</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${data.data.map(td => `
                                <tr>
                                    <td>${td.deviceAddr}</td>
                                    <td>${td.testPoint}</td>
                                    <td>${td.actualPercentage}%</td>
                                    <td>${td.ratioError}</td>
                                    <td>${td.angleError}</td>
                                    <td>${formatDate(td.testTimestamp)}</td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                `;
            } else {
                testDataList.innerHTML = '<p>暂无测试数据</p>';
            }
        }
    } catch (error) {
        console.error('Load test data error:', error);
    }
}

// 加载报表
function loadReports() {
    // 这里可以集成图表库如Chart.js来显示统计图表
    showNotification('报表功能开发中', 'info');
}