export {};

let csrfToken = '';
let currentUser: any = null;

async function api(method: string, path: string, body?: any): Promise<any> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (csrfToken) headers['X-CSRF-Token'] = csrfToken;
  const opts: RequestInit = { method, headers, credentials: 'include' };
  if (body !== undefined) opts.body = JSON.stringify(body);
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
      window.location.href = '/';
      return;
    }
    const tokenRes = await api('GET', '/api/csrf-token');
    csrfToken = tokenRes.token;
    showAdminTab('users');
    loadUsers();
    loadBackupSettings();
  } catch {
    window.location.href = '/login';
  }
}

function showAdminTab(tab: string) {
  document.querySelectorAll('#adminTabs .nav-link').forEach(el => el.classList.remove('active'));
  document.getElementById('tab' + capitalize(tab) + 'Btn')?.classList.add('active');
  ['users', 'compendium', 'backup'].forEach(s => {
    document.getElementById('admin' + capitalize(s))!.style.display = s === tab ? 'block' : 'none';
  });
  if (tab === 'users') loadUsers();
  if (tab === 'backup') { loadBackupSettings(); loadBackupList(); }
}
(window as any).showAdminTab = showAdminTab;

// ─── Users ───

async function loadUsers() {
  try {
    const users = await api('GET', '/api/admin/users');
    const tbody = document.querySelector('#userTable tbody')!;
    tbody.innerHTML = users.map((u: any) => `
      <tr>
        <td>${u.id}</td>
        <td>${esc(u.username)}</td>
        <td>${esc(u.display_name)}</td>
        <td><span class="badge ${u.role === 'admin' ? 'badge-blood' : 'badge-gold'}">${u.role}</span></td>
        <td>${u.created_at}</td>
        <td>
          <button class="btn btn-outline-primary btn-sm" onclick="editUser(${u.id},'${esc(u.username)}','${esc(u.display_name)}','${u.role}')"><i class="fa-solid fa-pen"></i></button>
          <button class="btn btn-outline-danger btn-sm" onclick="deleteUser(${u.id})"><i class="fa-solid fa-trash"></i></button>
          <button class="btn btn-outline-secondary btn-sm" onclick="resetPass(${u.id})"><i class="fa-solid fa-key"></i></button>
        </td>
      </tr>
    `).join('');
  } catch (e: any) {
    toast(e.message, true);
  }
}

(window as any).showAddUser = function () {
  showModal('Add User', `
    <div class="mb-3"><label class="form-label">Username</label><input class="form-control" id="addUsername"></div>
    <div class="mb-3"><label class="form-label">Password</label><input class="form-control" type="password" id="addPassword"></div>
    <div class="mb-3"><label class="form-label">Display Name</label><input class="form-control" id="addDisplay"></div>
    <div class="mb-3">
      <label class="form-label">Role</label>
      <select class="form-select" id="addRole"><option value="user">User</option><option value="admin">Admin</option></select>
    </div>
    <button class="btn btn-primary w-100" onclick="saveNewUser()">Create</button>
  `);
};

(window as any).saveNewUser = async function () {
  try {
    await api('POST', '/api/admin/users', {
      username: (document.getElementById('addUsername') as HTMLInputElement).value,
      password: (document.getElementById('addPassword') as HTMLInputElement).value,
      display_name: (document.getElementById('addDisplay') as HTMLInputElement).value,
      role: (document.getElementById('addRole') as HTMLSelectElement).value,
    });
    hideModal();
    loadUsers();
    toast('User created');
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).editUser = function (id: number, username: string, display: string, role: string) {
  showModal('Edit User', `
    <div class="mb-3"><label class="form-label">Username</label><input class="form-control" id="editUsername" value="${esc(username)}"></div>
    <div class="mb-3"><label class="form-label">Display Name</label><input class="form-control" id="editDisplay" value="${esc(display)}"></div>
    <div class="mb-3">
      <label class="form-label">Role</label>
      <select class="form-select" id="editRole"><option value="user" ${role === 'user' ? 'selected' : ''}>User</option><option value="admin" ${role === 'admin' ? 'selected' : ''}>Admin</option></select>
    </div>
    <button class="btn btn-primary w-100" onclick="saveEditUser(${id})">Save</button>
  `);
};

(window as any).saveEditUser = async function (id: number) {
  try {
    await api('PUT', `/api/admin/users/${id}`, {
      username: (document.getElementById('editUsername') as HTMLInputElement).value,
      display_name: (document.getElementById('editDisplay') as HTMLInputElement).value,
      role: (document.getElementById('editRole') as HTMLSelectElement).value,
    });
    hideModal();
    loadUsers();
    toast('User updated');
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).deleteUser = async function (id: number) {
  if (!confirm('Delete this user?')) return;
  try {
    await api('DELETE', `/api/admin/users/${id}`);
    loadUsers();
    toast('User deleted');
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).resetPass = function (id: number) {
  showModal('Reset Password', `
    <div class="mb-3"><label class="form-label">New Password</label><input class="form-control" type="password" id="resetPass"></div>
    <button class="btn btn-primary w-100" onclick="doResetPass(${id})">Reset</button>
  `);
};

(window as any).doResetPass = async function (id: number) {
  try {
    await api('PUT', `/api/admin/users/${id}/password`, {
      password: (document.getElementById('resetPass') as HTMLInputElement).value,
    });
    hideModal();
    toast('Password reset');
  } catch (e: any) {
    toast(e.message, true);
  }
};

// ─── Compendium ───

async function loadCompEntries() {
  const type = (document.getElementById('compType') as HTMLSelectElement).value;
  const el = document.getElementById('compEntries')!;
  try {
    const entries = await api('GET', `/api/compendium/${type}`);
    el.innerHTML = `<table class="table table-hover mb-0"><thead><tr><th>Name</th><th style="width:100px">Actions</th></tr></thead><tbody>
      ${entries.map((e: any) => `<tr>
        <td>${esc(e.name)}</td>
        <td>
          <button class="btn btn-outline-danger btn-sm" onclick="deleteCompEntry('${type}', ${e.id})"><i class="fa-solid fa-trash"></i></button>
        </td>
      </tr>`).join('')}
    </tbody></table>`;
  } catch {
    el.innerHTML = '<p style="color:var(--text-muted)">Failed to load entries</p>';
  }
}
(window as any).loadCompEntries = loadCompEntries;

(window as any).showAddCompEntry = function () {
  const type = (document.getElementById('compType') as HTMLSelectElement).value;
  const fields = getCompFields(type);
  showModal(`Add ${capitalize(type)}`, fields + `<button class="btn btn-primary w-100 mt-3" onclick="saveCompEntry('${type}')">Create</button>`);
};

function getCompFields(type: string): string {
  switch (type) {
    case 'spells':
      return `
        <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="compName"></div>
        <div class="row g-3 mb-3">
          <div class="col-6"><label class="form-label">Level</label><input class="form-control" id="compLevel" type="number" value="0"></div>
          <div class="col-6"><label class="form-label">School</label><input class="form-control" id="compSchool"></div>
        </div>
        <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="compDesc" rows="3"></textarea></div>`;
    case 'races':
      return `
        <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="compName"></div>
        <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="compDesc" rows="3"></textarea></div>
        <div class="row g-3 mb-3">
          <div class="col-6"><label class="form-label">Speed</label><input class="form-control" id="compSpeed" type="number" value="30"></div>
          <div class="col-6"><label class="form-label">Size</label><input class="form-control" id="compSize" value="Medium"></div>
        </div>`;
    default:
      return `
        <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="compName"></div>
        <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="compDesc" rows="3"></textarea></div>`;
  }
}

(window as any).saveCompEntry = async function (type: string) {
  const entry: any = { name: (document.getElementById('compName') as HTMLInputElement).value };
  if (type === 'spells') {
    entry.level = +(document.getElementById('compLevel') as HTMLInputElement).value || 0;
    entry.school = (document.getElementById('compSchool') as HTMLInputElement).value;
    entry.description = (document.getElementById('compDesc') as HTMLTextAreaElement).value;
  } else if (type === 'races') {
    entry.description = (document.getElementById('compDesc') as HTMLTextAreaElement).value;
    entry.speed = +(document.getElementById('compSpeed') as HTMLInputElement).value || 30;
    entry.size = (document.getElementById('compSize') as HTMLInputElement).value || 'Medium';
  } else {
    entry.description = (document.getElementById('compDesc') as HTMLTextAreaElement).value;
  }
  try {
    await api('POST', `/api/admin/compendium/${type}`, entry);
    hideModal();
    loadCompEntries();
    toast('Entry created');
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).deleteCompEntry = async function (type: string, id: number) {
  if (!confirm('Delete this entry?')) return;
  try {
    await api('DELETE', `/api/admin/compendium/${type}/${id}`);
    loadCompEntries();
    toast('Deleted');
  } catch (e: any) {
    toast(e.message, true);
  }
};

// ─── Backup ───

async function loadBackupSettings() {
  try {
    const settings = await api('GET', '/api/backup/settings');
    (document.getElementById('backupEnabled') as HTMLInputElement).checked = settings.enabled;
    (document.getElementById('backupInterval') as HTMLInputElement).value = settings.interval_hours || 168;
  } catch {}
}

(window as any).saveBackupSettings = async function () {
  try {
    await api('PUT', '/api/backup/settings', {
      enabled: (document.getElementById('backupEnabled') as HTMLInputElement).checked,
      interval_hours: +(document.getElementById('backupInterval') as HTMLInputElement).value || 168,
    });
    toast('Settings saved');
  } catch (e: any) {
    toast(e.message, true);
  }
};

async function loadBackupList() {
  try {
    const backups = await api('GET', '/api/backup/list');
    const el = document.getElementById('backupList')!;
    el.innerHTML = backups.length > 0
      ? `<table class="table table-hover mb-0"><thead><tr><th>Name</th><th>Size</th></tr></thead><tbody>
          ${backups.map((b: any) => `<tr><td>${esc(b.name)}</td><td>${formatSize(b.size)}</td></tr>`).join('')}
        </tbody></table>`
      : '<p class="text-muted p-3">No backups yet</p>';
  } catch {}
}

(window as any).triggerBackup = async function () {
  try {
    const result = await api('POST', '/api/backup/trigger');
    toast('Backup created: ' + result.path);
    loadBackupList();
  } catch (e: any) {
    toast(e.message, true);
  }
};

// ─── Utils ───

let adminModal: any = null;
function getModal(): any {
  if (!adminModal) adminModal = new (window as any).bootstrap.Modal(document.getElementById('genericModal')!);
  return adminModal;
}
function showModal(title: string, bodyHtml: string) {
  document.getElementById('genericModalTitle')!.textContent = title;
  document.getElementById('genericModalBody')!.innerHTML = bodyHtml;
  getModal().show();
}
function hideModal() {
  getModal().hide();
}

function esc(s: string): string {
  const div = document.createElement('div');
  div.textContent = s;
  return div.innerHTML;
}

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

function toast(msg: string, isError = false) {
  const container = document.getElementById('toastContainer')!;
  const id = 'toast-' + Date.now();
  const bg = isError ? 'bg-danger' : 'bg-success';
  container.innerHTML += `
    <div class="toast align-items-center text-white ${bg} border-0 mb-2" id="${id}" role="alert">
      <div class="d-flex">
        <div class="toast-body">${esc(msg)}</div>
        <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
      </div>
    </div>`;
  const el = document.getElementById(id)!;
  new (window as any).bootstrap.Toast(el, { autohide: true, delay: 5000 }).show();
  setTimeout(() => el.remove(), 6000);
}

(window as any).logout = async function () {
  await api('POST', '/api/logout');
  window.location.href = '/login';
};

init();
