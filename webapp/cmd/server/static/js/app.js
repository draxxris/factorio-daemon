// State
let instances = [];
let currentInstance = null;
let logStreamSource = null;

// DOM Elements
const instanceList = document.getElementById('instance-list');
const instanceModal = document.getElementById('instance-modal');
const newInstanceModal = document.getElementById('new-instance-modal');
const newInstanceBtn = document.getElementById('new-instance-btn');

// API Functions with timeout
async function fetchWithTimeout(url, options = {}, timeout = 10000) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);
    
    try {
        const response = await fetch(url, { ...options, signal: controller.signal });
        clearTimeout(timeoutId);
        return response;
    } catch (error) {
        clearTimeout(timeoutId);
        if (error.name === 'AbortError') {
            throw new Error('Request timed out');
        }
        throw error;
    }
}

async function apiGet(url) {
    const response = await fetchWithTimeout(url);
    if (!response.ok) {
        let errorMessage = `API error: ${response.statusText}`;
        try {
            const errorData = await response.json();
            if (errorData.error) {
                errorMessage = errorData.error;
            }
        } catch (e) {
            // If we can't parse JSON, use the status text
        }
        throw new Error(errorMessage);
    }
    return response.json();
}

async function apiPost(url, data = null) {
    const options = {
        method: 'POST',
    };
    if (data) {
        options.headers = { 'Content-Type': 'application/json' };
        options.body = JSON.stringify(data);
    }
    const response = await fetchWithTimeout(url, options);
    if (!response.ok) {
        let errorMessage = response.statusText;
        try {
            const error = await response.json();
            if (error.error) {
                errorMessage = error.error;
            }
        } catch (e) {
            // If we can't parse JSON, use the status text
        }
        throw new Error(errorMessage);
    }
    return response.json();
}

async function apiDelete(url) {
    const response = await fetchWithTimeout(url, { method: 'DELETE' });
    if (!response.ok) {
        let errorMessage = `API error: ${response.statusText}`;
        try {
            const error = await response.json();
            if (error.error) {
                errorMessage = error.error;
            }
        } catch (e) {
            // If we can't parse JSON, use the status text
        }
        throw new Error(errorMessage);
    }
    return response.json();
}

// Load instances
async function loadInstances() {
    try {
        instances = await apiGet('/api/instances');
        renderInstanceList();
    } catch (error) {
        console.error('Failed to load instances:', error);
    }
}

// Render instance list
async function renderInstanceList() {
    instanceList.innerHTML = '';
    
    if (instances.length === 0) {
        instanceList.innerHTML = '<p class="empty-hint">No instances found. Create one to get started.</p>';
        return;
    }
    
    instances.forEach(inst => {
        const card = document.createElement('div');
        card.className = 'instance-card';
        card.innerHTML = `
            <h3>${escapeHtml(inst.name)}</h3>
            <div class="meta">
                <div>Port: ${inst.port}</div>
                <div>Version: ${escapeHtml(inst.version)}</div>
                ${inst.title ? `<div>${escapeHtml(inst.title)}</div>` : ''}
            </div>
            <div class="status">
                <span class="status-dot ${inst.running ? 'running' : 'stopped'}"></span>
                <span>${inst.running ? 'Running' : 'Stopped'}</span>
                ${inst.enabled ? '<span class="badge">Autostart</span>' : '<span class="badge disabled">No Autostart</span>'}
            </div>
            <div class="server-info-mini" data-instance="${inst.name}">
                <div class="info-item">
                    <span class="info-label">Age:</span>
                    <span class="info-value server-age">-</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Players:</span>
                    <span class="info-value player-count">-</span>
                </div>
            </div>
        `;
        card.addEventListener('click', () => openInstanceModal(inst.name));
        instanceList.appendChild(card);
    });
    
    // Fetch server info for running instances
    instances.forEach(inst => {
        if (inst.running) {
            fetchServerInfoForCard(inst.name);
        }
    });
}

async function fetchServerInfoForCard(name) {
    const ageEl = document.querySelector(`[data-instance="${name}"] .server-age`);
    const countEl = document.querySelector(`[data-instance="${name}"] .player-count`);
    
    if (!ageEl || !countEl) return;
    
    try {
        const [timeData, playersData] = await Promise.all([
            apiGet(`/api/instances/${name}/rcon/time`).catch(() => null),
            apiGet(`/api/instances/${name}/rcon/players`).catch(() => null)
        ]);
        
        if (timeData && timeData.time) {
            ageEl.textContent = timeData.time;
        } else {
            ageEl.textContent = 'N/A';
        }
        
        if (playersData && playersData.players) {
            countEl.textContent = playersData.players.length;
        } else {
            countEl.textContent = '0';
        }
    } catch (error) {
        ageEl.textContent = 'N/A';
        countEl.textContent = '-';
    }
}

// Open instance modal
async function openInstanceModal(name) {
    currentInstance = instances.find(i => i.name === name);
    if (!currentInstance) {
        // Fetch from API
        try {
            currentInstance = await apiGet(`/api/instances/${name}`);
        } catch (error) {
            console.error('Failed to load instance:', error);
            return;
        }
    }
    
    // Update modal content
    document.getElementById('modal-title').textContent = currentInstance.name;
    updateStatusDisplay();
    
    // Reset to files tab
    switchTab('files');
    
    // Load staged files
    await loadStagedFiles();
    
    // Load backups
    await loadBackups();
    
    // Load RCON data (server time and players)
    await loadRconData();
    
    // Show modal
    instanceModal.classList.remove('hidden');
}

// Update status display
function updateStatusDisplay() {
    const statusDot = document.getElementById('status-indicator');
    const statusText = document.getElementById('status-text');
    const autostartBadge = document.getElementById('autostart-badge');
    const deployWarning = document.getElementById('deploy-warning');
    const deployBtn = document.getElementById('btn-deploy');
    
    statusDot.className = `status-dot ${currentInstance.running ? 'running' : 'stopped'}`;
    statusText.textContent = currentInstance.running ? 'Running' : 'Stopped';
    
    autostartBadge.textContent = currentInstance.enabled ? 'Autostart' : 'No Autostart';
    autostartBadge.className = `badge ${currentInstance.enabled ? '' : 'disabled'}`;
    
    // Update buttons
    document.getElementById('btn-start').disabled = currentInstance.running;
    document.getElementById('btn-stop').disabled = !currentInstance.running;
    document.getElementById('btn-enable').disabled = currentInstance.enabled;
    document.getElementById('btn-disable').disabled = !currentInstance.enabled;
    
    // Deploy warning
    if (currentInstance.running) {
        deployWarning.classList.remove('hidden');
        deployBtn.disabled = true;
    } else {
        deployWarning.classList.add('hidden');
        deployBtn.disabled = false;
    }
}

// Tab switching
function switchTab(tabName) {
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.tab === tabName);
    });
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.toggle('active', content.id === `tab-${tabName}`);
    });
    
    if (tabName === 'logs') {
        loadLogs();
    }
}

// Load RCON data (server time and player list)
async function loadRconData() {
    const serverTimeEl = document.getElementById('server-time');
    const playerListEl = document.getElementById('player-list');
    
    // Show loading state
    serverTimeEl.innerHTML = '<p class="loading-hint">Loading...</p>';
    playerListEl.innerHTML = '<p class="loading-hint">Loading...</p>';
    
    try {
        // Fetch server time
        const timeData = await apiGet(`/api/instances/${currentInstance.name}/rcon/time`);
        serverTimeEl.innerHTML = `<p class="time-display">${escapeHtml(timeData.time)}</p>`;
    } catch (error) {
        serverTimeEl.innerHTML = `<p class="error-hint">Failed to load server time: ${escapeHtml(error.message)}</p>`;
    }
    
    try {
        // Fetch player list
        const playersData = await apiGet(`/api/instances/${currentInstance.name}/rcon/players`);
        
        if (!playersData.players || playersData.players.length === 0) {
            playerListEl.innerHTML = '<p class="empty-hint">No players online</p>';
        } else {
            playerListEl.innerHTML = `
                <table class="player-table">
                    <thead>
                        <tr>
                            <th>Player Name</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${playersData.players.map(player => `
                            <tr>
                                <td>${escapeHtml(player)}</td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            `;
        }
    } catch (error) {
        playerListEl.innerHTML = `<p class="error-hint">Failed to load players: ${escapeHtml(error.message)}</p>`;
    }
}

// Add admin
async function addAdmin(playerName) {
    const messageEl = document.getElementById('admin-message');
    
    try {
        await apiPost(`/api/instances/${currentInstance.name}/rcon/admin`, { player: playerName });
        messageEl.textContent = `Successfully added ${playerName} as admin!`;
        messageEl.className = 'message success';
        messageEl.classList.remove('hidden');
        
        // Refresh player list
        await loadRconData();
        
        // Clear input
        document.getElementById('admin-player-name').value = '';
        
        // Hide message after 3 seconds
        setTimeout(() => {
            messageEl.classList.add('hidden');
        }, 3000);
    } catch (error) {
        messageEl.textContent = `Failed to add admin: ${escapeHtml(error.message)}`;
        messageEl.className = 'message error';
        messageEl.classList.remove('hidden');
    }
}

// Load staged files
async function loadStagedFiles() {
    const container = document.getElementById('staged-files');
    
    try {
        const files = await apiGet(`/api/instances/${currentInstance.name}/staged`);
        
        if (files.length === 0) {
            container.innerHTML = '<p class="empty-hint">No files staged</p>';
            return;
        }
        
        container.innerHTML = files.map(f => `
            <div class="file-item">
                <span>${escapeHtml(f)}</span>
            </div>
        `).join('');
    } catch (error) {
        container.innerHTML = '<p class="empty-hint">Failed to load staged files</p>';
    }
}

// Load backups
async function loadBackups() {
    const container = document.getElementById('backup-list');
    
    try {
        const backups = await apiGet(`/api/instances/${currentInstance.name}/backups`);
        
        if (backups.length === 0) {
            container.innerHTML = '<p class="empty-hint">No backups</p>';
            return;
        }
        
        container.innerHTML = backups.map(b => `
            <div class="backup-item">
                <div class="backup-info">
                    <span>${escapeHtml(b.filename)}</span>
                    <span class="backup-time">${new Date(b.timestamp).toLocaleString()} (${formatBytes(b.size)})</span>
                </div>
                <button class="btn btn-secondary btn-restore" data-filename="${escapeHtml(b.filename)}">Restore</button>
            </div>
        `).join('');
        
        // Add restore handlers
        container.querySelectorAll('.btn-restore').forEach(btn => {
            btn.addEventListener('click', () => restoreBackup(btn.dataset.filename));
        });
    } catch (error) {
        container.innerHTML = '<p class="empty-hint">Failed to load backups</p>';
    }
}

// Load logs
async function loadLogs() {
    const viewer = document.getElementById('log-viewer');
    
    try {
        const logs = await apiGet(`/api/instances/${currentInstance.name}/logs?lines=200`);
        
        viewer.innerHTML = logs.map(entry =>
            `<div class="log-entry"><span class="timestamp">[${new Date(entry.timestamp).toLocaleTimeString()}]</span> <span class="message">${escapeHtml(entry.message.trim())}</span></div>`
        ).join('');
        
        if (document.getElementById('auto-scroll').checked) {
            viewer.scrollTop = viewer.scrollHeight;
        }
    } catch (error) {
        viewer.innerHTML = '<p style="color: var(--text-muted)">Failed to load logs</p>';
    }
}

// Start log streaming
function startLogStream() {
    if (logStreamSource) {
        logStreamSource.close();
        logStreamSource = null;
        document.getElementById('btn-stream-logs').textContent = 'Stream Logs';
        return;
    }
    
    const viewer = document.getElementById('log-viewer');
    logStreamSource = new EventSource(`/api/instances/${currentInstance.name}/logs/stream`);
    
    logStreamSource.onmessage = (event) => {
        const data = JSON.parse(event.data);
        const entry = document.createElement('div');
        entry.className = 'log-entry';
        entry.innerHTML = `<span class="timestamp">[${new Date(data.timestamp).toLocaleTimeString()}]</span> <span class="message">${escapeHtml(data.message.trim())}</span>`;
        viewer.appendChild(entry);
        
        if (document.getElementById('auto-scroll').checked) {
            viewer.scrollTop = viewer.scrollHeight;
        }
    };
    
    logStreamSource.onerror = () => {
        logStreamSource.close();
        logStreamSource = null;
        document.getElementById('btn-stream-logs').textContent = 'Stream Logs';
    };
    
    document.getElementById('btn-stream-logs').textContent = 'Stop Streaming';
}

// File upload
function handleFileUpload(files) {
    Array.from(files).forEach(file => {
        const reader = new FileReader();
        reader.onload = async (e) => {
            try {
                const formData = new FormData();
                formData.append('file', file);
                
                const response = await fetch(`/api/instances/${currentInstance.name}/upload`, {
                    method: 'POST',
                    body: formData,
                });
                
                if (!response.ok) {
                    const error = await response.json();
                    throw new Error(error.error);
                }
                
                await loadStagedFiles();
            } catch (error) {
                alert(`Failed to upload ${file.name}: ${error.message}`);
            }
        };
        reader.readAsArrayBuffer(file);
    });
}

// Restore backup
async function restoreBackup(filename) {
    if (currentInstance.running) {
        alert('Instance must be stopped before restoring a backup!');
        return;
    }
    
    if (!confirm(`Restore backup ${filename}? This will overwrite the current save.`)) {
        return;
    }
    
    try {
        await apiPost(`/api/instances/${currentInstance.name}/backups/${filename}/restore`);
        alert('Backup restored successfully!');
    } catch (error) {
        alert(`Failed to restore backup: ${error.message}`);
    }
}

// Service control
async function startInstance() {
    try {
        await apiPost(`/api/instances/${currentInstance.name}/start`);
        await refreshCurrentInstance();
    } catch (error) {
        alert(`Failed to start: ${error.message}`);
    }
}

async function stopInstance() {
    try {
        await apiPost(`/api/instances/${currentInstance.name}/stop`);
        await refreshCurrentInstance();
    } catch (error) {
        alert(`Failed to stop: ${error.message}`);
    }
}

async function restartInstance() {
    try {
        await apiPost(`/api/instances/${currentInstance.name}/restart`);
        await refreshCurrentInstance();
    } catch (error) {
        alert(`Failed to restart: ${error.message}`);
    }
}

async function enableInstance() {
    try {
        await apiPost(`/api/instances/${currentInstance.name}/enable`);
        await refreshCurrentInstance();
    } catch (error) {
        alert(`Failed to enable: ${error.message}`);
    }
}

async function disableInstance() {
    try {
        await apiPost(`/api/instances/${currentInstance.name}/disable`);
        await refreshCurrentInstance();
    } catch (error) {
        alert(`Failed to disable: ${error.message}`);
    }
}

async function refreshCurrentInstance() {
    try {
        currentInstance = await apiGet(`/api/instances/${currentInstance.name}`);
        updateStatusDisplay();
        await loadInstances();
        
        // Refresh server info (server time and players)
        await loadRconData();
    } catch (error) {
        console.error('Failed to refresh:', error);
    }
}

// Deploy files
async function deployFiles() {
    if (currentInstance.running) {
        alert('Instance must be stopped before deploying!');
        return;
    }
    
    if (!confirm('Deploy staged files? This will overwrite existing files in the instance directory.')) {
        return;
    }
    
    try {
        await apiPost(`/api/instances/${currentInstance.name}/deploy`);
        await loadStagedFiles();
        alert('Files deployed successfully!');
    } catch (error) {
        alert(`Failed to deploy: ${error.message}`);
    }
}

// Clear staged files
async function clearStagedFiles() {
    try {
        await apiDelete(`/api/instances/${currentInstance.name}/staged`);
        await loadStagedFiles();
    } catch (error) {
        alert(`Failed to clear: ${error.message}`);
    }
}

// Backup save
async function backupSave() {
    try {
        await apiPost(`/api/instances/${currentInstance.name}/backup`);
        await loadBackups();
        alert('Save backed up successfully!');
    } catch (error) {
        alert(`Failed to backup: ${error.message}`);
    }
}

// Create new instance
async function createInstance(event) {
    event.preventDefault();
    
    const form = event.target;
    const formData = new FormData(form);
    
    const data = {
        name: formData.get('name'),
        version: formData.get('version') || 'latest',
        title: formData.get('title'),
        description: formData.get('description'),
        port: parseInt(formData.get('port')) || 34197,
        non_blocking_save: document.getElementById('new-non-blocking').checked,
    };
    
    try {
        await apiPost('/api/instances', data);
        
        // Enable and start if checked
        if (document.getElementById('new-enable-now').checked) {
            await apiPost(`/api/instances/${data.name}/enable`);
            await apiPost(`/api/instances/${data.name}/start`);
        }
        
        form.reset();
        document.getElementById('new-version').value = 'latest';
        document.getElementById('new-port').value = '34197';
        document.getElementById('new-non-blocking').checked = true;
        document.getElementById('new-enable-now').checked = true;
        
        newInstanceModal.classList.add('hidden');
        await loadInstances();
    } catch (error) {
        alert(`Failed to create instance: ${error.message}`);
    }
}

// Utility functions
function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function formatBytes(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

// Event Listeners
document.addEventListener('DOMContentLoaded', () => {
    loadInstances();
    
    // New instance button
    newInstanceBtn.addEventListener('click', () => {
        newInstanceModal.classList.remove('hidden');
    });
    
    // Close modals
    document.querySelectorAll('.close-btn, .close-modal').forEach(btn => {
        btn.addEventListener('click', () => {
            instanceModal.classList.add('hidden');
            newInstanceModal.classList.add('hidden');
            if (logStreamSource) {
                logStreamSource.close();
                logStreamSource = null;
            }
        });
    });
    
    // Close modals with ESC key
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            if (!instanceModal.classList.contains('hidden')) {
                instanceModal.classList.add('hidden');
                if (logStreamSource) {
                    logStreamSource.close();
                    logStreamSource = null;
                }
            }
            if (!newInstanceModal.classList.contains('hidden')) {
                newInstanceModal.classList.add('hidden');
            }
        }
    });

    // Note: Click outside modal does NOT close it (per user request)
    
    // Tab switching
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', () => switchTab(btn.dataset.tab));
    });
    
    // Service control buttons
    document.getElementById('btn-start').addEventListener('click', startInstance);
    document.getElementById('btn-stop').addEventListener('click', stopInstance);
    document.getElementById('btn-restart').addEventListener('click', restartInstance);
    document.getElementById('btn-enable').addEventListener('click', enableInstance);
    document.getElementById('btn-disable').addEventListener('click', disableInstance);
    
    // File upload
    const dropZone = document.getElementById('drop-zone');
    const fileInput = document.getElementById('file-input');
    
    dropZone.addEventListener('click', () => fileInput.click());
    
    dropZone.addEventListener('dragover', (e) => {
        e.preventDefault();
        dropZone.classList.add('dragover');
    });
    
    dropZone.addEventListener('dragleave', () => {
        dropZone.classList.remove('dragover');
    });
    
    dropZone.addEventListener('drop', (e) => {
        e.preventDefault();
        dropZone.classList.remove('dragover');
        handleFileUpload(e.dataTransfer.files);
    });
    
    fileInput.addEventListener('change', () => {
        handleFileUpload(fileInput.files);
        fileInput.value = '';
    });
    
    // Deploy and clear
    document.getElementById('btn-deploy').addEventListener('click', deployFiles);
    document.getElementById('btn-clear-staged').addEventListener('click', clearStagedFiles);
    
    // Logs
    document.getElementById('btn-refresh-logs').addEventListener('click', loadLogs);
    document.getElementById('btn-stream-logs').addEventListener('click', startLogStream);
    
    // Backup
    document.getElementById('btn-backup').addEventListener('click', backupSave);
    
    // Admin form
    document.getElementById('add-admin-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const playerName = document.getElementById('admin-player-name').value.trim();
        if (playerName) {
            await addAdmin(playerName);
        }
    });
    
    // New instance form
    document.getElementById('new-instance-form').addEventListener('submit', createInstance);
    
    // Auto-refresh instance list
    setInterval(loadInstances, 30000);
});
