const API_BASE = '/api';
let currentProjectId = null;

// 页面加载时初始化
document.addEventListener('DOMContentLoaded', () => {
    loadProjects();

    // 表单提交
    document.getElementById('project-form').addEventListener('submit', handleFormSubmit);
});

// 显示列表视图
function showList() {
    document.getElementById('view-list').classList.remove('hidden');
    document.getElementById('view-form').classList.add('hidden');
    document.getElementById('view-detail').classList.add('hidden');
    loadProjects();
}

// 显示创建表单
function showCreateForm() {
    document.getElementById('view-list').classList.add('hidden');
    document.getElementById('view-form').classList.remove('hidden');
    document.getElementById('view-detail').classList.add('hidden');
    document.getElementById('form-title').textContent = '新建项目';
    document.getElementById('project-form').reset();
    document.getElementById('project-id').value = '';

    // 设置AWS OIDC示例
    document.getElementById('refresh_headers').value = '{"Content-Type": "application/json"}';
    document.getElementById('refresh_body_template').value = '{"clientId":"{{.ClientId}}","clientSecret":"{{.ClientSecret}}","grantType":"refresh_token","refreshToken":"{{.RefreshToken}}"}';
}

// 显示编辑表单
function showEditForm(project) {
    document.getElementById('view-list').classList.add('hidden');
    document.getElementById('view-form').classList.remove('hidden');
    document.getElementById('view-detail').classList.add('hidden');
    document.getElementById('form-title').textContent = '编辑项目';

    // 填充表单
    document.getElementById('project-id').value = project.id;
    document.getElementById('name').value = project.name;
    document.getElementById('description').value = project.description || '';
    document.getElementById('refresh_url').value = project.refresh_url;
    document.getElementById('refresh_method').value = project.refresh_method;
    document.getElementById('refresh_headers').value = project.refresh_headers || '';
    document.getElementById('refresh_body_template').value = project.refresh_body_template || '';
    document.getElementById('access_token_path').value = project.access_token_path;
    document.getElementById('refresh_token_path').value = project.refresh_token_path;
    document.getElementById('expires_in_path').value = project.expires_in_path || '';
    document.getElementById('custom_variables').value = project.custom_variables || '';
    document.getElementById('current_refresh_token').value = project.current_refresh_token || '';
    document.getElementById('refresh_before_seconds').value = project.refresh_before_seconds;
}

// 显示项目详情
async function showDetail(projectId) {
    currentProjectId = projectId;
    document.getElementById('view-list').classList.add('hidden');
    document.getElementById('view-form').classList.add('hidden');
    document.getElementById('view-detail').classList.remove('hidden');

    try {
        const project = await fetchAPI(`/projects/${projectId}`);
        const token = await fetchAPI(`/projects/${projectId}/token`);
        const logs = await fetchAPI(`/projects/${projectId}/logs?limit=10`);

        document.getElementById('detail-title').textContent = project.name;
        document.getElementById('detail-access-token').value = token.access_token || '(未刷新)';
        document.getElementById('detail-refresh-token').value = token.refresh_token || '(未刷新)';

        if (token.expires_at && token.expires_at.Valid) {
            const expiresAt = new Date(token.expires_at.Time);
            const now = new Date();
            const diff = expiresAt - now;
            const hours = Math.floor(diff / 3600000);
            const minutes = Math.floor((diff % 3600000) / 60000);
            document.getElementById('detail-expires-at').textContent =
                `${expiresAt.toLocaleString()} (剩余 ${hours}小时 ${minutes}分钟)`;
        } else {
            document.getElementById('detail-expires-at').textContent = '未设置';
        }

        if (project.last_refresh_at && project.last_refresh_at.Valid) {
            const lastRefresh = new Date(project.last_refresh_at.Time);
            document.getElementById('detail-last-refresh').textContent =
                `${lastRefresh.toLocaleString()} - ${project.last_refresh_status}`;
        } else {
            document.getElementById('detail-last-refresh').textContent = '从未刷新';
        }

        // 显示日志（只显示最近10条）
        renderLogs(logs);
    } catch (error) {
        showToast('加载项目详情失败: ' + error.message, 'error');
    }
}

// 渲染日志列表
function renderLogs(logs) {
    const logsContainer = document.getElementById('logs-container');
    if (!logs || logs.length === 0) {
        logsContainer.innerHTML = '<p class="text-gray-500">暂无刷新日志</p>';
    } else {
        const logsHtml = logs.map(log => `
            <div class="border-l-4 ${log.status === 'success' ? 'border-green-500' : 'border-red-500'} pl-4 py-2 mb-3">
                <div class="flex justify-between">
                    <span class="font-medium ${log.status === 'success' ? 'text-green-700' : 'text-red-700'}">
                        ${log.status === 'success' ? '✓ 成功' : '✗ 失败'}
                    </span>
                    <span class="text-sm text-gray-500">${new Date(log.refresh_at).toLocaleString()}</span>
                </div>
                ${log.error_message ? `<p class="text-sm text-red-600 mt-1">错误: ${truncate(log.error_message, 100)}</p>` : ''}
                ${log.new_token_preview ? `<p class="text-sm text-gray-600 mt-1">新Access Token: ${log.new_token_preview}...</p>` : ''}
            </div>
        `).join('');

        logsContainer.innerHTML = logsHtml + `
            <div class="text-center mt-4">
                <button onclick="loadMoreLogs()" class="px-4 py-2 text-sm text-blue-600 hover:text-blue-800">
                    加载更多日志 (最多显示50条)
                </button>
            </div>
        `;
    }
}

// 加载更多日志
async function loadMoreLogs() {
    if (!currentProjectId) return;

    try {
        const logs = await fetchAPI(`/projects/${currentProjectId}/logs?limit=50`);
        const logsContainer = document.getElementById('logs-container');

        if (!logs || logs.length === 0) {
            logsContainer.innerHTML = '<p class="text-gray-500">暂无刷新日志</p>';
            return;
        }

        const logsHtml = logs.map(log => `
            <div class="border-l-4 ${log.status === 'success' ? 'border-green-500' : 'border-red-500'} pl-4 py-2 mb-3">
                <div class="flex justify-between">
                    <span class="font-medium ${log.status === 'success' ? 'text-green-700' : 'text-red-700'}">
                        ${log.status === 'success' ? '✓ 成功' : '✗ 失败'}
                    </span>
                    <span class="text-sm text-gray-500">${new Date(log.refresh_at).toLocaleString()}</span>
                </div>
                ${log.error_message ? `<p class="text-sm text-red-600 mt-1">错误: ${truncate(log.error_message, 100)}</p>` : ''}
                ${log.new_token_preview ? `<p class="text-sm text-gray-600 mt-1">新Access Token: ${log.new_token_preview}...</p>` : ''}
            </div>
        `).join('');

        logsContainer.innerHTML = logsHtml + `
            <div class="text-center mt-4">
                <p class="text-sm text-gray-500">已显示最近 ${logs.length} 条日志</p>
            </div>
        `;
    } catch (error) {
        showToast('加载日志失败: ' + error.message, 'error');
    }
}

// 加载项目列表
async function loadProjects() {
    try {
        const projects = await fetchAPI('/projects');
        const container = document.getElementById('projects-container');

        if (!projects || projects.length === 0) {
            container.innerHTML = '<p class="col-span-full text-center text-gray-500">暂无项目，点击"新建项目"开始</p>';
            return;
        }

        container.innerHTML = projects.map(project => {
            const status = getProjectStatus(project);
            return `
                <div class="bg-white shadow rounded-lg p-6 hover:shadow-lg transition-shadow">
                    <div class="flex justify-between items-start mb-4">
                        <div>
                            <h3 class="text-lg font-semibold text-gray-900">${project.name}</h3>
                            <p class="text-sm text-gray-500">${project.description || '无描述'}</p>
                        </div>
                        <span class="px-2 py-1 text-xs font-medium rounded ${status.class}">
                            ${status.text}
                        </span>
                    </div>

                    <div class="space-y-2 text-sm text-gray-600 mb-4">
                        <div>URL: ${truncate(project.refresh_url, 40)}</div>
                        ${project.last_refresh_at && project.last_refresh_at.Valid ?
                            `<div>最后刷新: ${new Date(project.last_refresh_at.Time).toLocaleString()}</div>` :
                            '<div>最后刷新: 从未</div>'}
                    </div>

                    <div class="flex space-x-2">
                        <button onclick="showDetail(${project.id})" class="flex-1 px-3 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700">
                            查看
                        </button>
                        <button onclick="refreshProject(${project.id})" class="px-3 py-2 bg-green-600 text-white text-sm rounded hover:bg-green-700">
                            刷新
                        </button>
                        <button onclick="toggleProject(${project.id})" class="px-3 py-2 ${project.enabled ? 'bg-yellow-600' : 'bg-gray-600'} text-white text-sm rounded hover:opacity-80">
                            ${project.enabled ? '禁用' : '启用'}
                        </button>
                        <button onclick="editProject(${project.id})" class="px-3 py-2 bg-gray-600 text-white text-sm rounded hover:bg-gray-700">
                            编辑
                        </button>
                        <button onclick="deleteProject(${project.id}, '${project.name}')" class="px-3 py-2 bg-red-600 text-white text-sm rounded hover:bg-red-700">
                            删除
                        </button>
                    </div>
                </div>
            `;
        }).join('');
    } catch (error) {
        showToast('加载项目列表失败: ' + error.message, 'error');
    }
}

// 获取项目状态
function getProjectStatus(project) {
    if (!project.enabled) {
        return { text: '已禁用', class: 'status-disabled' };
    }

    if (!project.token_expires_at || !project.token_expires_at.Valid) {
        return { text: '未刷新', class: 'status-warning' };
    }

    const expiresAt = new Date(project.token_expires_at.Time);
    const now = new Date();
    const diff = expiresAt - now;

    if (diff < 0) {
        return { text: '已过期', class: 'status-error' };
    } else if (diff < project.refresh_before_seconds * 1000) {
        return { text: '即将过期', class: 'status-warning' };
    } else {
        return { text: '正常', class: 'status-active' };
    }
}

// 表单提交处理
async function handleFormSubmit(e) {
    e.preventDefault();

    const projectId = document.getElementById('project-id').value;
    const data = {
        name: document.getElementById('name').value,
        description: document.getElementById('description').value,
        refresh_url: document.getElementById('refresh_url').value,
        refresh_method: document.getElementById('refresh_method').value,
        refresh_headers: document.getElementById('refresh_headers').value,
        refresh_body_template: document.getElementById('refresh_body_template').value,
        access_token_path: document.getElementById('access_token_path').value,
        refresh_token_path: document.getElementById('refresh_token_path').value,
        expires_in_path: document.getElementById('expires_in_path').value,
        custom_variables: document.getElementById('custom_variables').value,
        current_refresh_token: document.getElementById('current_refresh_token').value,
        refresh_before_seconds: parseInt(document.getElementById('refresh_before_seconds').value),
    };

    try {
        if (projectId) {
            await fetchAPI(`/projects/${projectId}`, 'PUT', data);
            showToast('项目更新成功');
        } else {
            await fetchAPI('/projects', 'POST', data);
            showToast('项目创建成功');
        }
        showList();
    } catch (error) {
        showToast('保存失败: ' + error.message, 'error');
    }
}

// 编辑项目
async function editProject(projectId) {
    try {
        const project = await fetchAPI(`/projects/${projectId}`);
        showEditForm(project);
    } catch (error) {
        showToast('加载项目失败: ' + error.message, 'error');
    }
}

// 删除项目
async function deleteProject(projectId, projectName) {
    if (!confirm(`确定要删除项目 "${projectName}" 吗？此操作不可恢复。`)) {
        return;
    }

    try {
        await fetchAPI(`/projects/${projectId}`, 'DELETE');
        showToast('项目删除成功');
        loadProjects();
    } catch (error) {
        showToast('删除失败: ' + error.message, 'error');
    }
}

// 切换项目启用状态
async function toggleProject(projectId) {
    try {
        await fetchAPI(`/projects/${projectId}/toggle`, 'POST');
        showToast('状态切换成功');
        loadProjects();
    } catch (error) {
        showToast('切换失败: ' + error.message, 'error');
    }
}

// 手动刷新项目
async function refreshProject(projectId) {
    try {
        showToast('正在刷新...');
        await fetchAPI(`/projects/${projectId}/refresh`, 'POST');
        showToast('刷新成功');
        loadProjects();
    } catch (error) {
        showToast('刷新失败: ' + error.message, 'error');
    }
}

// 复制Token
function copyToken(type) {
    const inputId = type === 'refresh' ? 'detail-refresh-token' : 'detail-access-token';
    const input = document.getElementById(inputId);
    input.select();
    document.execCommand('copy');
    const tokenType = type === 'refresh' ? 'Refresh Token' : 'Access Token';
    showToast(`${tokenType}已复制到剪贴板`);
}

// API请求封装
async function fetchAPI(endpoint, method = 'GET', data = null) {
    const options = {
        method,
        headers: {
            'Content-Type': 'application/json',
        },
    };

    if (data) {
        options.body = JSON.stringify(data);
    }

    const response = await fetch(API_BASE + endpoint, options);

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Request failed');
    }

    return response.json();
}

// 显示Toast通知
function showToast(message, type = 'success') {
    const toast = document.getElementById('toast');
    const toastMessage = document.getElementById('toast-message');

    toastMessage.textContent = message;
    toast.classList.remove('hidden');

    if (type === 'error') {
        toast.classList.add('bg-red-600');
        toast.classList.remove('bg-gray-900');
    } else {
        toast.classList.add('bg-gray-900');
        toast.classList.remove('bg-red-600');
    }

    setTimeout(() => {
        toast.classList.add('hidden');
    }, 3000);
}

// 工具函数
function truncate(str, len) {
    return str.length > len ? str.substring(0, len) + '...' : str;
}
