"use strict";
let csrfToken = '';
let currentUser = null;
async function api(method, path, body) {
    const headers = { 'Content-Type': 'application/json' };
    if (csrfToken)
        headers['X-CSRF-Token'] = csrfToken;
    const opts = { method, headers, credentials: 'include' };
    if (body !== undefined)
        opts.body = JSON.stringify(body);
    const res = await fetch(path, opts);
    if (!res.ok) {
        const err = await res.json().catch(() => ({ error: res.statusText }));
        throw new Error(err.error || 'Request failed');
    }
    return res.json();
}
async function init() {
    try {
        currentUser = await api('GET', '/api/user/me');
        if (currentUser.role !== 'admin') {
            window.location.href = '/app';
            return;
        }
        const tokenRes = await api('GET', '/api/csrf-token');
        csrfToken = tokenRes.token;
        showAdminTab('users');
        loadUsers();
        loadBackupSettings();
    }
    catch {
        window.location.href = '/login';
    }
}
function showAdminTab(tab) {
    document.querySelectorAll('.tab').forEach(el => el.classList.remove('active'));
    document.getElementById('tab' + capitalize(tab))?.classList.add('active');
    ['users', 'compendium', 'backup'].forEach(s => {
        document.getElementById('admin' + capitalize(s)).style.display = s === tab ? 'block' : 'none';
    });
    if (tab === 'users')
        loadUsers();
    if (tab === 'backup') {
        loadBackupSettings();
        loadBackupList();
    }
}
window.showAdminTab = showAdminTab;
// ─── Users ───
async function loadUsers() {
    try {
        const users = await api('GET', '/api/admin/users');
        const tbody = document.querySelector('#userTable tbody');
        tbody.innerHTML = users.map((u) => `
      <tr>
        <td>${u.id}</td>
        <td>${esc(u.username)}</td>
        <td>${esc(u.display_name)}</td>
        <td><span class="badge ${u.role === 'admin' ? 'badge-blood' : 'badge-gold'}">${u.role}</span></td>
        <td>${u.created_at}</td>
        <td>
          <button class="btn btn-sm" onclick="editUser(${u.id},'${esc(u.username)}','${esc(u.display_name)}','${u.role}')">Edit</button>
          <button class="btn btn-sm btn-danger" onclick="deleteUser(${u.id})">Delete</button>
          <button class="btn btn-sm" onclick="resetPass(${u.id})">Reset PW</button>
        </td>
      </tr>
    `).join('');
    }
    catch (e) {
        toast(e.message, true);
    }
}
window.showAddUser = function () {
    showModal(`
    <h2>Add User</h2>
    <div class="form-group"><label>Username</label><input id="addUsername"></div>
    <div class="form-group"><label>Password</label><input type="password" id="addPassword"></div>
    <div class="form-group"><label>Display Name</label><input id="addDisplay"></div>
    <div class="form-group">
      <label>Role</label>
      <select id="addRole"><option value="user">User</option><option value="admin">Admin</option></select>
    </div>
    <button class="btn btn-primary" onclick="saveNewUser(this)">Create</button>
  `);
};
window.saveNewUser = async function (btn) {
    try {
        await api('POST', '/api/admin/users', {
            username: document.getElementById('addUsername').value,
            password: document.getElementById('addPassword').value,
            display_name: document.getElementById('addDisplay').value,
            role: document.getElementById('addRole').value,
        });
        btn.closest('.modal-overlay')?.remove();
        loadUsers();
        toast('User created');
    }
    catch (e) {
        toast(e.message, true);
    }
};
window.editUser = function (id, username, display, role) {
    showModal(`
    <h2>Edit User</h2>
    <div class="form-group"><label>Username</label><input id="editUsername" value="${esc(username)}"></div>
    <div class="form-group"><label>Display Name</label><input id="editDisplay" value="${esc(display)}"></div>
    <div class="form-group">
      <label>Role</label>
      <select id="editRole"><option value="user" ${role === 'user' ? 'selected' : ''}>User</option><option value="admin" ${role === 'admin' ? 'selected' : ''}>Admin</option></select>
    </div>
    <button class="btn btn-primary" onclick="saveEditUser(${id}, this)">Save</button>
  `);
};
window.saveEditUser = async function (id, btn) {
    try {
        await api('PUT', `/api/admin/users/${id}`, {
            username: document.getElementById('editUsername').value,
            display_name: document.getElementById('editDisplay').value,
            role: document.getElementById('editRole').value,
        });
        btn.closest('.modal-overlay')?.remove();
        loadUsers();
        toast('User updated');
    }
    catch (e) {
        toast(e.message, true);
    }
};
window.deleteUser = async function (id) {
    if (!confirm('Delete this user?'))
        return;
    try {
        await api('DELETE', `/api/admin/users/${id}`);
        loadUsers();
        toast('User deleted');
    }
    catch (e) {
        toast(e.message, true);
    }
};
window.resetPass = function (id) {
    showModal(`
    <h2>Reset Password</h2>
    <div class="form-group"><label>New Password</label><input type="password" id="resetPass"></div>
    <button class="btn btn-primary" onclick="doResetPass(${id}, this)">Reset</button>
  `);
};
window.doResetPass = async function (id, btn) {
    try {
        await api('PUT', `/api/admin/users/${id}/password`, {
            password: document.getElementById('resetPass').value,
        });
        btn.closest('.modal-overlay')?.remove();
        toast('Password reset');
    }
    catch (e) {
        toast(e.message, true);
    }
};
// ─── Compendium ───
async function loadCompEntries() {
    const type = document.getElementById('compType').value;
    const el = document.getElementById('compEntries');
    try {
        const entries = await api('GET', `/api/compendium/${type}`);
        el.innerHTML = `<table><thead><tr><th>Name</th><th>Actions</th></tr></thead><tbody>
      ${entries.map((e) => `<tr>
        <td>${esc(e.name)}</td>
        <td>
          <button class="btn btn-sm btn-danger" onclick="deleteCompEntry('${type}', ${e.id})">Delete</button>
        </td>
      </tr>`).join('')}
    </tbody></table>`;
    }
    catch {
        el.innerHTML = '<p style="color:var(--text-muted)">Failed to load entries</p>';
    }
}
window.loadCompEntries = loadCompEntries;
window.showAddCompEntry = function () {
    const type = document.getElementById('compType').value;
    const fields = getCompFields(type);
    showModal(`
    <h2>Add ${capitalize(type)}</h2>
    ${fields}
    <button class="btn btn-primary" onclick="saveCompEntry('${type}', this)">Create</button>
  `);
};
function getCompFields(type) {
    switch (type) {
        case 'spells':
            return `
        <div class="form-group"><label>Name</label><input id="compName"></div>
        <div class="form-row">
          <div class="form-group"><label>Level</label><input id="compLevel" type="number" value="0"></div>
          <div class="form-group"><label>School</label><input id="compSchool"></div>
        </div>
        <div class="form-group"><label>Description</label><textarea id="compDesc"></textarea></div>`;
        case 'races':
            return `
        <div class="form-group"><label>Name</label><input id="compName"></div>
        <div class="form-group"><label>Description</label><textarea id="compDesc"></textarea></div>
        <div class="form-row">
          <div class="form-group"><label>Speed</label><input id="compSpeed" type="number" value="30"></div>
          <div class="form-group"><label>Size</label><input id="compSize" value="Medium"></div>
        </div>`;
        default:
            return `
        <div class="form-group"><label>Name</label><input id="compName"></div>
        <div class="form-group"><label>Description</label><textarea id="compDesc"></textarea></div>`;
    }
}
window.saveCompEntry = async function (type, btn) {
    const entry = { name: document.getElementById('compName').value };
    if (type === 'spells') {
        entry.level = +document.getElementById('compLevel').value || 0;
        entry.school = document.getElementById('compSchool').value;
        entry.description = document.getElementById('compDesc').value;
    }
    else if (type === 'races') {
        entry.description = document.getElementById('compDesc').value;
        entry.speed = +document.getElementById('compSpeed').value || 30;
        entry.size = document.getElementById('compSize').value || 'Medium';
    }
    else {
        entry.description = document.getElementById('compDesc').value;
    }
    try {
        await api('POST', `/api/admin/compendium/${type}`, entry);
        btn.closest('.modal-overlay')?.remove();
        loadCompEntries();
        toast('Entry created');
    }
    catch (e) {
        toast(e.message, true);
    }
};
window.deleteCompEntry = async function (type, id) {
    if (!confirm('Delete this entry?'))
        return;
    try {
        await api('DELETE', `/api/admin/compendium/${type}/${id}`);
        loadCompEntries();
        toast('Deleted');
    }
    catch (e) {
        toast(e.message, true);
    }
};
// ─── Backup ───
async function loadBackupSettings() {
    try {
        const settings = await api('GET', '/api/backup/settings');
        document.getElementById('backupEnabled').checked = settings.enabled;
        document.getElementById('backupInterval').value = settings.interval_hours || 168;
    }
    catch { }
}
window.saveBackupSettings = async function () {
    try {
        await api('PUT', '/api/backup/settings', {
            enabled: document.getElementById('backupEnabled').checked,
            interval_hours: +document.getElementById('backupInterval').value || 168,
        });
        toast('Settings saved');
    }
    catch (e) {
        toast(e.message, true);
    }
};
async function loadBackupList() {
    try {
        const backups = await api('GET', '/api/backup/list');
        const el = document.getElementById('backupList');
        el.innerHTML = backups.length > 0
            ? `<table><thead><tr><th>Name</th><th>Size</th></tr></thead><tbody>
          ${backups.map((b) => `<tr><td>${esc(b.name)}</td><td>${formatSize(b.size)}</td></tr>`).join('')}
        </tbody></table>`
            : '<p style="color:var(--text-muted)">No backups yet</p>';
    }
    catch { }
}
window.triggerBackup = async function () {
    try {
        const result = await api('POST', '/api/backup/trigger');
        toast('Backup created: ' + result.path);
        loadBackupList();
    }
    catch (e) {
        toast(e.message, true);
    }
};
// ─── Utils ───
function showModal(html) {
    const overlay = document.createElement('div');
    overlay.className = 'modal-overlay';
    overlay.innerHTML = `<div class="modal">${html}</div>`;
    overlay.onclick = (e) => { if (e.target === overlay)
        overlay.remove(); };
    document.body.appendChild(overlay);
}
function esc(s) {
    const div = document.createElement('div');
    div.textContent = s;
    return div.innerHTML;
}
function capitalize(s) {
    return s.charAt(0).toUpperCase() + s.slice(1);
}
function formatSize(bytes) {
    if (bytes < 1024)
        return bytes + ' B';
    if (bytes < 1024 * 1024)
        return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}
function toast(msg, isError = false) {
    const el = document.createElement('div');
    el.className = 'toast' + (isError ? ' error' : '');
    el.textContent = msg;
    document.body.appendChild(el);
    setTimeout(() => el.remove(), 4000);
}
window.logout = async function () {
    await api('POST', '/api/logout');
    window.location.href = '/login';
};
init();
//# sourceMappingURL=admin.js.map