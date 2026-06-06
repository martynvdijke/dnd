(() => {

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

function toggleTheme() {
  const html = document.documentElement;
  const isDark = html.getAttribute('data-theme') === 'dark';
  const newTheme = isDark ? 'light' : 'dark';
  html.setAttribute('data-theme', newTheme);
  localStorage.setItem('villum-theme', newTheme);
  const icon = document.getElementById('themeIcon');
  if (icon) icon.className = isDark ? 'fa-solid fa-moon' : 'fa-solid fa-sun';
}
(window as any).toggleTheme = toggleTheme;

function initTheme() {
  const saved = localStorage.getItem('villum-theme') || 'light';
  document.documentElement.setAttribute('data-theme', saved);
  const icon = document.getElementById('themeIcon');
  if (icon) icon.className = saved === 'dark' ? 'fa-solid fa-sun' : 'fa-solid fa-moon';
}

async function init() {
  initTheme();
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
  } catch {
    window.location.href = '/login';
  }
}

let logRefreshInterval: any = null;

function showAdminTab(tab: string) {
  document.querySelectorAll('#adminTabs .nav-link').forEach(el => el.classList.remove('active'));
  document.getElementById('tab' + capitalize(tab) + 'Btn')?.classList.add('active');
  const tabs = ['users', 'schemas', 'compendium', 'backup', 'email', 'ai-endpoints', 'analytics', 'telemetry', 'import'];
  tabs.forEach(s => {
    const id = 'admin' + s.split('-').map((p, i) => i === 0 ? capitalize(p) : capitalize(p)).join('');
    document.getElementById(id)!.style.display = s === tab ? 'block' : 'none';
  });
  if (tab === 'users') loadUsers();
  if (tab === 'schemas') loadSchemas();
  if (tab === 'backup') { loadBackupSettings(); loadBackupList(); }
  if (tab === 'email') loadEmailSettings();
  if (tab === 'ai-endpoints') loadAIEndpoints();
  if (tab === 'analytics') loadUmamiSettings();
  if (tab === 'telemetry') loadOTelSettings();
  if (tab === 'import') { loadImportSchemas(); loadImportLogs(); }
  if (tab === 'logs') { startLogAutoRefresh(); }
  else { stopLogAutoRefresh(); }
}
(window as any).showAdminTab = showAdminTab;

// ─── Logs ───

function startLogAutoRefresh() {
  loadLogLevel();
  loadLogs();
  if (logRefreshInterval) clearInterval(logRefreshInterval);
  logRefreshInterval = setInterval(() => {
    loadLogLevel();
    loadLogs();
  }, 5000);
}

function stopLogAutoRefresh() {
  if (logRefreshInterval) {
    clearInterval(logRefreshInterval);
    logRefreshInterval = null;
  }
}

async function loadLogs() {
  try {
    const levelFilter = (document.getElementById('logSourceFilter') as HTMLSelectElement)?.value || '';
    let url = '/api/admin/logs?limit=200';
    if (levelFilter) url += '&source=' + encodeURIComponent(levelFilter);
    const logs = await api('GET', url);
    const tbody = document.getElementById('logBody')!;
    if (!logs || logs.length === 0) {
      tbody.innerHTML = '<tr><td colspan="5" class="text-muted text-center">No log entries</td></tr>';
      document.getElementById('logCount')!.textContent = '0 entries';
      return;
    }
    const levelBadge: Record<string, string> = {
      debug: 'bg-secondary',
      info: 'bg-info text-dark',
      warn: 'bg-warning text-dark',
      error: 'bg-danger',
    };
    tbody.innerHTML = logs.map((l: any) => {
      const badge = levelBadge[l.level] || 'bg-secondary';
      const ts = l.timestamp ? new Date(l.timestamp).toLocaleString() : '-';
      const attrs = l.attributes && Object.keys(l.attributes).length > 0
        ? Object.entries(l.attributes).map(([k, v]) => `<span class="badge bg-light text-dark me-1" title="${esc(k)}">${esc(k)}=${esc(String(v))}</span>`).join('')
        : '';
      return `<tr>
        <td><span class="badge ${badge}">${esc(l.level)}</span></td>
        <td class="small">${esc(ts)}</td>
        <td><code class="small">${esc(l.source || '-')}</code></td>
        <td>${esc(l.message)}</td>
        <td class="small" style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${attrs}</td>
      </tr>`;
    }).join('');
    document.getElementById('logCount')!.textContent = logs.length + ' entries';
  } catch (e: any) {
    document.getElementById('logBody')!.innerHTML = '<tr><td colspan="5" class="text-danger text-center">Failed to load logs</td></tr>';
  }
}

async function loadLogLevel() {
  try {
    const res = await api('GET', '/api/admin/log-level');
    const sel = document.getElementById('logLevelSelect') as HTMLSelectElement;
    if (sel && res.level) sel.value = res.level;
  } catch {}
}

async function setLogLevel(level: string) {
  try {
    await api('PUT', '/api/admin/log-level', { level });
    loadLogs();
  } catch (e: any) {
    toast(e.message, true);
  }
}
(window as any).setLogLevel = setLogLevel;

function clearLogFilters() {
  const levelFilter = document.getElementById('logSourceFilter') as HTMLSelectElement;
  if (levelFilter) levelFilter.value = '';
  loadLogs();
}
(window as any).clearLogFilters = clearLogFilters;

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
        <td>${esc(u.email)}</td>
        <td><span class="badge ${u.role === 'admin' ? 'badge-blood' : 'badge-gold'}">${u.role}</span></td>
        <td>${u.created_at}</td>
        <td>
          <button class="btn btn-outline-primary btn-sm" onclick="editUser(${u.id},'${esc(u.username)}','${esc(u.display_name)}','${esc(u.email)}','${u.role}')"><i class="fa-solid fa-pen"></i></button>
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
    <div class="mb-3"><label class="form-label">Email</label><input class="form-control" type="email" id="addEmail"></div>
    <div class="mb-3">
      <label class="form-label">Role</label>
      <select class="form-select" id="addRole"><option value="user">User</option><option value="dm">DM</option><option value="admin">Admin</option></select>
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
      email: (document.getElementById('addEmail') as HTMLInputElement).value,
      role: (document.getElementById('addRole') as HTMLSelectElement).value,
    });
    hideModal();
    loadUsers();
    toast('User created');
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).editUser = function (id: number, username: string, display: string, email: string, role: string) {
  showModal('Edit User', `
    <div class="mb-3"><label class="form-label">Username</label><input class="form-control" id="editUsername" value="${esc(username)}"></div>
    <div class="mb-3"><label class="form-label">Display Name</label><input class="form-control" id="editDisplay" value="${esc(display)}"></div>
    <div class="mb-3"><label class="form-label">Email</label><input class="form-control" type="email" id="editEmail" value="${esc(email)}"></div>
    <div class="mb-3">
      <label class="form-label">Role</label>
      <select class="form-select" id="editRole"><option value="user" ${role === 'user' ? 'selected' : ''}>User</option><option value="dm" ${role === 'dm' ? 'selected' : ''}>DM</option><option value="admin" ${role === 'admin' ? 'selected' : ''}>Admin</option></select>
    </div>
    <button class="btn btn-primary w-100" onclick="saveEditUser(${id})">Save</button>
  `);
};

(window as any).saveEditUser = async function (id: number) {
  try {
    await api('PUT', `/api/admin/users/${id}`, {
      username: (document.getElementById('editUsername') as HTMLInputElement).value,
      display_name: (document.getElementById('editDisplay') as HTMLInputElement).value,
      email: (document.getElementById('editEmail') as HTMLInputElement).value,
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
    case 'monsters':
      return `
        <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="compName"></div>
        <div class="row g-3 mb-3">
          <div class="col-6"><label class="form-label">Type</label><input class="form-control" id="compTypeMonster" placeholder="e.g. beast, humanoid"></div>
          <div class="col-6"><label class="form-label">Size</label><input class="form-control" id="compSizeMonster" value="Medium"></div>
        </div>
        <div class="row g-3 mb-3">
          <div class="col-4"><label class="form-label">AC</label><input class="form-control" id="compAC" type="number" value="10"></div>
          <div class="col-4"><label class="form-label">HP</label><input class="form-control" id="compHP" type="number" value="10"></div>
          <div class="col-4"><label class="form-label">CR</label><input class="form-control" id="compCR" value="0"></div>
        </div>
        <div class="row g-3 mb-3">
          <div class="col-4"><label class="form-label">STR</label><input class="form-control" id="compStr" type="number" value="10"></div>
          <div class="col-4"><label class="form-label">DEX</label><input class="form-control" id="compDex" type="number" value="10"></div>
          <div class="col-4"><label class="form-label">CON</label><input class="form-control" id="compCon" type="number" value="10"></div>
        </div>
        <div class="row g-3 mb-3">
          <div class="col-4"><label class="form-label">INT</label><input class="form-control" id="compInt" type="number" value="10"></div>
          <div class="col-4"><label class="form-label">WIS</label><input class="form-control" id="compWis" type="number" value="10"></div>
          <div class="col-4"><label class="form-label">CHA</label><input class="form-control" id="compCha" type="number" value="10"></div>
        </div>
        <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="compDesc" rows="3"></textarea></div>`;
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
  } else if (type === 'monsters') {
    entry.type = (document.getElementById('compTypeMonster') as HTMLInputElement).value;
    entry.size = (document.getElementById('compSizeMonster') as HTMLInputElement).value || 'Medium';
    entry.ac = +(document.getElementById('compAC') as HTMLInputElement).value || 10;
    entry.hp = +(document.getElementById('compHP') as HTMLInputElement).value || 10;
    entry.cr = (document.getElementById('compCR') as HTMLInputElement).value || '0';
    entry.str = +(document.getElementById('compStr') as HTMLInputElement).value || 10;
    entry.dex = +(document.getElementById('compDex') as HTMLInputElement).value || 10;
    entry.con = +(document.getElementById('compCon') as HTMLInputElement).value || 10;
    entry.int = +(document.getElementById('compInt') as HTMLInputElement).value || 10;
    entry.wis = +(document.getElementById('compWis') as HTMLInputElement).value || 10;
    entry.cha = +(document.getElementById('compCha') as HTMLInputElement).value || 10;
    entry.description = (document.getElementById('compDesc') as HTMLTextAreaElement).value;
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

// ─── Schemas ───

async function loadSchemas() {
  try {
    const schemas = await api('GET', '/api/admin/compendium-schemas');
    const tbody = document.querySelector('#schemaTable tbody')!;
    tbody.innerHTML = schemas.map((s: any) => `
      <tr>
        <td style="font-size:1.3rem">📖</td>
        <td><strong>${esc(s.display_name)}</strong></td>
        <td><code>${esc(s.type_name)}</code></td>
        <td style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="${esc(s.display_name || '')}"></td>
        <td>${s.fields ? s.fields.length : 0}</td>
        <td><span class="badge badge-primary">${s.entry_count || 0}</span></td>
        <td>-</td>
        <td>
          <button class="btn btn-outline-primary btn-sm" onclick="editSchema(${s.id})" title="Edit"><i class="fa-solid fa-pen"></i></button>
          <button class="btn btn-outline-info btn-sm" onclick="browseSchemaEntries(${s.id},'${esc(s.display_name)}')" title="Entries"><i class="fa-solid fa-list"></i></button>
          <button class="btn btn-outline-danger btn-sm" onclick="deleteSchema(${s.id})" title="Delete"><i class="fa-solid fa-trash"></i></button>
        </td>
      </tr>
    `).join('');
  } catch (e: any) {
    toast(e.message, true);
  }
}
(window as any).loadSchemas = loadSchemas;

let schemaEditId: number | null = null;

(window as any).showAddSchema = function () {
  schemaEditId = null;
  showModal('New Compendium Schema', getSchemaFormHtml(null));
};

(window as any).editSchema = async function (id: number) {
  schemaEditId = id;
  try {
    const s = await api('GET', `/api/admin/compendium-schemas/${id}`);
    showModal('Edit Schema: ' + esc(s.display_name), getSchemaFormHtml(s));
  } catch (e: any) {
    toast(e.message, true);
  }
};

function getSchemaFormHtml(schema: any): string {
  const s = schema || {};
  const fields = s.fields || [{ name: '', label: '', type: 'text', required: false }];
  return `
    <input type="hidden" id="schemaId" value="${s.id || ''}">
    <div class="mb-2"><label class="form-label">Display Name</label><input class="form-control" id="schemaDisplayName" value="${esc(s.display_name || '')}" placeholder="e.g. Magic Items"></div>
    <div class="row g-2 mb-2">
      <div class="col-12"><label class="form-label">Type Name (slug)</label><input class="form-control" id="schemaTypeName" value="${esc(s.type_name || '')}" placeholder="e.g. magic-items"></div>
    </div>
    <hr>
    <label class="form-label fw-bold">Fields</label>
    <div id="schemaFields">${fields.map((f: any, i: number) => getSchemaFieldHtml(f, i)).join('')}</div>
    <button type="button" class="btn btn-sm btn-outline-secondary mt-1" onclick="addSchemaField()"><i class="fa-solid fa-plus me-1"></i>Add Field</button>
    <hr>
    <button class="btn btn-primary w-100" onclick="saveSchema()">${s.id ? 'Update' : 'Create'} Schema</button>
  `;
}

function getSchemaFieldHtml(field: any, index: number): string {
  return `<div class="schema-field-row row g-1 mb-1 align-items-end" id="sf-${index}">
    <div class="col-3"><input class="form-control form-control-sm" placeholder="Key" id="sf-name-${index}" value="${esc(field.name || '')}"></div>
    <div class="col-3"><input class="form-control form-control-sm" placeholder="Label" id="sf-label-${index}" value="${esc(field.label || '')}"></div>
    <div class="col-3">
      <select class="form-select form-select-sm" id="sf-type-${index}">
        <option value="text" ${field.type === 'text' ? 'selected' : ''}>Text</option>
        <option value="textarea" ${field.type === 'textarea' ? 'selected' : ''}>Textarea</option>
        <option value="number" ${field.type === 'number' ? 'selected' : ''}>Number</option>
        <option value="richtext" ${field.type === 'richtext' ? 'selected' : ''}>Rich Text</option>
        <option value="boolean" ${field.type === 'boolean' ? 'selected' : ''}>Yes/No</option>
        <option value="list" ${field.type === 'list' ? 'selected' : ''}>List</option>
      </select>
    </div>
    <div class="col-2">
      <div class="form-check form-switch mb-1">
        <input class="form-check-input" type="checkbox" id="sf-req-${index}" ${field.required ? 'checked' : ''}>
        <label class="form-check-label" style="font-size:0.75rem" for="sf-req-${index}">Req</label>
      </div>
    </div>
    <div class="col-1">
      <button class="btn btn-sm btn-outline-danger" onclick="removeSchemaField(${index})" title="Remove field"><i class="fa-solid fa-xmark"></i></button>
    </div>
  </div>`;
}

(window as any).addSchemaField = function () {
  const container = document.getElementById('schemaFields')!;
  const index = container.children.length;
  container.insertAdjacentHTML('beforeend', getSchemaFieldHtml({ key: '', label: '', type: 'text', required: false }, index));
};

(window as any).removeSchemaField = function (index: number) {
  const el = document.getElementById('sf-' + index);
  if (el) el.remove();
};

(window as any).saveSchema = async function () {
  const display_name = (document.getElementById('schemaDisplayName') as HTMLInputElement).value.trim();
  const type_name = (document.getElementById('schemaTypeName') as HTMLInputElement).value.trim();
  if (!display_name || !type_name) { toast('Display Name and Type Name are required', true); return; }

  const fields: any[] = [];
  const container = document.getElementById('schemaFields')!;
  for (let i = 0; i < container.children.length; i++) {
    const name = (document.getElementById('sf-name-' + i) as HTMLInputElement)?.value?.trim();
    if (!name) continue;
    fields.push({
      name,
      label: (document.getElementById('sf-label-' + i) as HTMLInputElement)?.value?.trim() || name,
      type: (document.getElementById('sf-type-' + i) as HTMLSelectElement)?.value || 'text',
      required: (document.getElementById('sf-req-' + i) as HTMLInputElement)?.checked || false,
    });
  }

  const body = { type_name, display_name, fields };
  try {
    if (schemaEditId) {
      await api('PUT', `/api/admin/compendium-schemas/${schemaEditId}`, body);
      toast('Schema updated');
    } else {
      await api('POST', '/api/admin/compendium-schemas', body);
      toast('Schema created');
    }
    hideModal();
    loadSchemas();
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).deleteSchema = async function (id: number) {
  if (!confirm('Delete this schema? This cannot be undone.')) return;
  try {
    await api('DELETE', `/api/admin/compendium-schemas/${id}`);
    loadSchemas();
    toast('Schema deleted');
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).browseSchemaEntries = async function (schemaId: number, schemaName: string) {
  try {
    const entries = await api('GET', `/api/admin/compendium-entries?schema_id=${schemaId}&limit=50`);
    const el = document.getElementById('schemaEntryBrowser')!;
    entryModalSchemaId = schemaId;
    selectedEntryIds.clear();
    if (!entries || entries.length === 0) {
      el.innerHTML = `<div class="mt-2">
        <button class="btn btn-success btn-sm mb-2" onclick="createSchemaEntry(${schemaId})"><i class="fa-solid fa-plus me-1"></i>Add Entry</button>
        <p class="text-muted">No entries in <strong>${esc(schemaName)}</strong></p>
      </div>`;
      return;
    }
    const fields = Object.keys(entries[0].data || {});
    const previewFields = fields.slice(0, 3);
    el.innerHTML = `
      <div class="card mt-2">
        <div class="card-header py-2 d-flex justify-content-between align-items-center">
          <span><strong>${esc(schemaName)}</strong> — ${entries.length} entries</span>
          <button class="btn btn-success btn-sm" onclick="createSchemaEntry(${schemaId})"><i class="fa-solid fa-plus me-1"></i>Add Entry</button>
        </div>
        <div id="bulkActions" class="px-3 py-2 bg-light border-bottom" style="display:none"></div>
        <div class="table-responsive">
          <table class="table table-sm table-hover mb-0">
            <thead><tr>
              <th style="width:40px"><input class="form-check-input" type="checkbox" id="selectAllEntries" onchange="toggleSelectAll(${schemaId})"></th>
              <th>Name</th>${previewFields.map(f => `<th>${esc(f)}</th>`).join('')}${fields.length > 3 ? '<th>...</th>' : ''}
              <th style="width:100px"></th>
            </tr></thead>
            <tbody>${entries.map((e: any) => `
              <tr>
                <td><input class="form-check-input entry-select-cb" type="checkbox" data-entry-id="${e.id}" onchange="toggleEntrySelect(${e.id})"></td>
                <td>${esc(e.name)}</td>
                ${previewFields.map(f => `<td style="max-width:120px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${esc(String(e.data?.[f] || ''))}</td>`).join('')}
                ${fields.length > 3 ? '<td>…</td>' : ''}
                <td class="text-nowrap">
                  <button class="btn btn-outline-primary btn-sm py-0" onclick="editSchemaEntry(${e.id},${schemaId})" title="Edit"><i class="fa-solid fa-pen"></i></button>
                  <button class="btn btn-outline-danger btn-sm py-0" onclick="deleteSchemaEntryById(${e.id})" title="Delete"><i class="fa-solid fa-trash"></i></button>
                </td>
              </tr>
            `).join('')}</tbody>
          </table>
        </div>
      </div>`;
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).deleteSchemaEntryById = async function (entryId: number) {
  if (!confirm('Delete this entry?')) return;
  try {
    await api('DELETE', `/api/admin/compendium-entries/${entryId}`);
    toast('Entry deleted');
    loadSchemas();
  } catch (e: any) {
    toast(e.message, true);
  }
};

// ─── Entry Create/Edit Modal ───

let entryModalSchemaId = 0;
let entryModalSchemaFields: any[] = [];
let entryModalEditId: number | null = null;

(window as any).createSchemaEntry = async function (schemaId: number) {
  const schemas = await api('GET', '/api/admin/compendium-schemas');
  const schema = schemas.find((s: any) => s.id === schemaId);
  if (!schema) { toast('Schema not found', true); return; }
  entryModalSchemaId = schemaId;
  entryModalSchemaFields = schema.fields || [];
  entryModalEditId = null;
  showModal('Create Entry', getEntryFormHtml(null));
};

(window as any).editSchemaEntry = async function (entryId: number, schemaId: number) {
  const [schemas, entry] = await Promise.all([
    api('GET', '/api/admin/compendium-schemas'),
    api('GET', `/api/admin/compendium-entries/${entryId}`)
  ]);
  const schema = schemas.find((s: any) => s.id === schemaId);
  if (!schema) { toast('Schema not found', true); return; }
  entryModalSchemaId = schemaId;
  entryModalSchemaFields = schema.fields || [];
  entryModalEditId = entryId;
  showModal('Edit Entry', getEntryFormHtml(entry.data || entry));
};

function getEntryFormHtml(data: any): string {
  const fields = entryModalSchemaFields;
  if (!fields || fields.length === 0) {
    return `<div class="text-muted">No fields defined for this schema.</div>`;
  }
  return fields.map((f: any) => {
    const val = data ? String(data[f.name] ?? '') : '';
    const requiredAttr = f.required ? 'required' : '';
    const requiredMark = f.required ? ' <span class="text-danger">*</span>' : '';
    let input = '';

    if (f.type === 'textarea') {
      input = `<textarea class="form-control" id="ef_${esc(f.name)}" ${requiredAttr}>${esc(val)}</textarea>`;
    } else if (f.type === 'checkbox') {
      const checked = data && data[f.name] ? 'checked' : '';
      input = `<div class="form-check"><input class="form-check-input" type="checkbox" id="ef_${esc(f.name)}" ${checked}></div>`;
    } else if (f.type === 'number') {
      input = `<input class="form-control" type="number" id="ef_${esc(f.name)}" value="${esc(val)}" ${requiredAttr}>`;
    } else {
      input = `<input class="form-control" type="text" id="ef_${esc(f.name)}" value="${esc(val)}" ${requiredAttr}>`;
    }

    return `<div class="mb-2">
      <label class="form-label">${esc(f.label || f.name)}${requiredMark}</label>
      ${input}
    </div>`;
  }).join('') + `
    <div class="mt-3">
      <button class="btn btn-primary" onclick="saveEntry()"><i class="fa-solid fa-floppy-disk me-1"></i>Save</button>
      <button class="btn btn-secondary" onclick="hideModal()">Cancel</button>
    </div>`;
}

(window as any).saveEntry = async function () {
  const data: Record<string, any> = {};
  let valid = true;

  for (const f of entryModalSchemaFields) {
    const el = document.getElementById('ef_' + f.name) as HTMLInputElement | HTMLTextAreaElement;
    if (!el) continue;
    let val: any;
    if (f.type === 'checkbox') {
      val = (el as HTMLInputElement).checked;
    } else if (f.type === 'number') {
      val = el.value ? parseFloat(el.value) : null;
    } else {
      val = el.value;
    }

    if (f.required && (val === null || val === '' || val === undefined)) {
      el.classList.add('is-invalid');
      valid = false;
    } else {
      el.classList.remove('is-invalid');
    }
    data[f.name] = val;
  }

  if (!valid) { toast('Please fill in all required fields', true); return; }

  try {
    if (entryModalEditId) {
      await api('PUT', `/api/admin/compendium-entries/${entryModalEditId}`, { data });
      toast('Entry updated');
    } else {
      await api('POST', `/api/admin/compendium-schemas/${entryModalSchemaId}/entries`, { data });
      toast('Entry created');
    }
    hideModal();
    loadSchemas();
  } catch (e: any) {
    toast(e.message, true);
  }
};

// ─── Entry Bulk Operations ───

let selectedEntryIds: Set<number> = new Set();

(window as any).toggleEntrySelect = function (id: number) {
  if (selectedEntryIds.has(id)) selectedEntryIds.delete(id);
  else selectedEntryIds.add(id);
  updateBulkActions();
};

(window as any).toggleSelectAll = function (schemaId: number) {
  const cb = document.getElementById('selectAllEntries') as HTMLInputElement;
  const checkboxes = document.querySelectorAll<HTMLInputElement>('.entry-select-cb');
  checkboxes.forEach(c => {
    c.checked = cb.checked;
    const eid = parseInt(c.dataset.entryId || '0', 10);
    if (cb.checked) selectedEntryIds.add(eid);
    else selectedEntryIds.delete(eid);
  });
  updateBulkActions();
};

function updateBulkActions() {
  const count = selectedEntryIds.size;
  const el = document.getElementById('bulkActions')!;
  if (count === 0) { el.style.display = 'none'; return; }
  el.style.display = 'block';
  el.innerHTML = `<span class="me-2 fw-bold">${count} selected</span>
    <button class="btn btn-outline-danger btn-sm me-1" onclick="batchDeleteEntries()"><i class="fa-solid fa-trash me-1"></i>Delete Selected</button>
    <button class="btn btn-outline-primary btn-sm" onclick="batchEditEntries(${count})"><i class="fa-solid fa-pen me-1"></i>Edit Selected</button>`;
}

(window as any).batchDeleteEntries = async function () {
  const ids = Array.from(selectedEntryIds);
  if (!confirm('Delete ' + ids.length + ' entries? This cannot be undone.')) return;
  try {
    const res = await api('POST', '/api/admin/compendium-entries/batch-delete', { ids });
    toast('Deleted ' + res.deleted + ' entries');
    selectedEntryIds.clear();
    loadSchemas();
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).batchEditEntries = async function (count: number) {
  const ids = Array.from(selectedEntryIds);
  // Get schema fields from the first entry's schema
  const schemas = await api('GET', '/api/admin/compendium-schemas');
  const firstEntry = await api('GET', `/api/admin/compendium-entries/${ids[0]}`);
  // Find schema that contains this entry (iterate schemas)
  const schema = schemas.find((s: any) => s.id === entryModalSchemaId) || schemas[0];
  const fields = schema?.fields || [];

  const fieldOpts = fields.map((f: any) =>
    `<option value="${esc(f.name)}">${esc(f.label || f.name)}</option>`
  ).join('');

  showModal('Bulk Edit ' + count + ' Entries', `
    <p>Set a field value for all ${count} selected entries.</p>
    <div class="mb-2">
      <label class="form-label">Field</label>
      <select class="form-select" id="bulkField">${fieldOpts}</select>
    </div>
    <div class="mb-2">
      <label class="form-label">Value</label>
      <input class="form-control" id="bulkValue" type="text">
    </div>
    <button class="btn btn-primary" onclick="saveBatchEdit()"><i class="fa-solid fa-floppy-disk me-1"></i>Update All</button>
    <button class="btn btn-secondary" onclick="hideModal()">Cancel</button>
  `);
};

(window as any).saveBatchEdit = async function () {
  const field = (document.getElementById('bulkField') as HTMLSelectElement).value;
  const value = (document.getElementById('bulkValue') as HTMLInputElement).value;
  if (!field) { toast('Select a field', true); return; }

  const ids = Array.from(selectedEntryIds);
  try {
    const res = await api('POST', '/api/admin/compendium-entries/batch-update', {
      ids,
      data: { [field]: value }
    });
    toast('Updated ' + res.updated + ' entries');
    hideModal();
    selectedEntryIds.clear();
    loadSchemas();
  } catch (e: any) {
    toast(e.message, true);
  }
};

// ─── Backup ───

async function loadBackupSettings() {
  try {
    const settings = await api('GET', '/api/backup/settings');
    (document.getElementById('backupEnabled') as HTMLInputElement).checked = settings.enabled;
    (document.getElementById('backupInterval') as HTMLInputElement).value = settings.interval_days || 7;
    (document.getElementById('backupKeepCount') as HTMLInputElement).value = settings.keep_count || 7;
  } catch {}
}

(window as any).saveBackupSettings = async function () {
  try {
    await api('PUT', '/api/backup/settings', {
      enabled: (document.getElementById('backupEnabled') as HTMLInputElement).checked,
      interval_days: +(document.getElementById('backupInterval') as HTMLInputElement).value || 7,
      keep_count: +(document.getElementById('backupKeepCount') as HTMLInputElement).value || 7,
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

// ─── Email Settings ───

async function loadEmailSettings() {
  try {
    const s = await api('GET', '/api/admin/email-settings');
    (document.getElementById('emailEnabled') as HTMLInputElement).checked = s.enabled;
    (document.getElementById('smtpHost') as HTMLInputElement).value = s.smtp_host || '';
    (document.getElementById('smtpPort') as HTMLInputElement).value = s.smtp_port || 587;
    (document.getElementById('smtpUsername') as HTMLInputElement).value = s.username || '';
    (document.getElementById('smtpFrom') as HTMLInputElement).value = s.from_addr || '';
    if (s.has_password) {
      (document.getElementById('smtpPassword') as HTMLInputElement).placeholder = 'Password is set (leave blank to keep)';
    }
  } catch {}
}

(window as any).saveEmailSettings = async function () {
  try {
    await api('POST', '/api/admin/email-settings', {
      smtp_host: (document.getElementById('smtpHost') as HTMLInputElement).value,
      smtp_port: +(document.getElementById('smtpPort') as HTMLInputElement).value || 587,
      username: (document.getElementById('smtpUsername') as HTMLInputElement).value,
      password: (document.getElementById('smtpPassword') as HTMLInputElement).value,
      from_addr: (document.getElementById('smtpFrom') as HTMLInputElement).value,
      enabled: (document.getElementById('emailEnabled') as HTMLInputElement).checked,
    });
    toast('Email settings saved');
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).testEmailSettings = async function () {
  try {
    await api('POST', '/api/admin/email-settings', {
      smtp_host: (document.getElementById('smtpHost') as HTMLInputElement).value,
      smtp_port: +(document.getElementById('smtpPort') as HTMLInputElement).value || 587,
      username: (document.getElementById('smtpUsername') as HTMLInputElement).value,
      password: (document.getElementById('smtpPassword') as HTMLInputElement).value,
      from_addr: (document.getElementById('smtpFrom') as HTMLInputElement).value,
      enabled: (document.getElementById('emailEnabled') as HTMLInputElement).checked,
      test: true,
    });
    toast('Test email sent! Check your inbox.');
  } catch (e: any) {
    toast(e.message, true);
  }
};

// ─── Umami Analytics ───

async function loadUmamiSettings() {
  try {
    const s = await api('GET', '/api/admin/umami-settings');
    (document.getElementById('umamiEnabled') as HTMLInputElement).checked = s.enabled;
    (document.getElementById('umamiHostname') as HTMLInputElement).value = s.tracker_hostname || '';
    (document.getElementById('umamiWebsiteID') as HTMLInputElement).value = s.website_id || '';
    (document.getElementById('umamiShareData') as HTMLInputElement).checked = s.share_data;
    (document.getElementById('umamiAdminTracking') as HTMLInputElement).checked = s.enable_admin_tracking;
  } catch {}
}

(window as any).saveUmamiSettings = async function () {
  try {
    await api('POST', '/api/admin/umami-settings', {
      enabled: (document.getElementById('umamiEnabled') as HTMLInputElement).checked,
      tracker_hostname: (document.getElementById('umamiHostname') as HTMLInputElement).value,
      website_id: (document.getElementById('umamiWebsiteID') as HTMLInputElement).value,
      share_data: (document.getElementById('umamiShareData') as HTMLInputElement).checked,
      enable_admin_tracking: (document.getElementById('umamiAdminTracking') as HTMLInputElement).checked,
    });
    toast('Analytics settings saved');
  } catch (e: any) {
    toast(e.message, true);
  }
};

// ─── OpenTelemetry ───

async function loadOTelSettings() {
  try {
    const s = await api('GET', '/api/admin/otel-settings');
    (document.getElementById('otelEnabled') as HTMLInputElement).checked = s.enabled;
    (document.getElementById('otelEndpoint') as HTMLInputElement).value = s.endpoint || '';
  } catch {}
}

(window as any).saveOTelSettings = async function () {
  try {
    await api('POST', '/api/admin/otel-settings', {
      enabled: (document.getElementById('otelEnabled') as HTMLInputElement).checked,
      endpoint: (document.getElementById('otelEndpoint') as HTMLInputElement).value,
    });
    toast('Telemetry settings saved');
  } catch (e: any) {
    toast(e.message, true);
  }
};

// ─── AI Endpoints ───

let aiEndpointEditId: number | null = null;

async function loadAIEndpoints() {
  try {
    const endpoints = await api('GET', '/api/admin/ai-endpoints');
    const tbody = document.querySelector('#aiEndpointTable tbody')!;
    tbody.innerHTML = endpoints.map((ep: any) => `
      <tr>
        <td>${esc(ep.name)}</td>
        <td><span class="badge ${ep.type === 'text' ? 'badge-primary' : 'badge-secondary'}">${ep.type}</span></td>
        <td>${esc(ep.model)}</td>
        <td style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="${esc(ep.base_url)}">${esc(ep.base_url)}</td>
        <td>${ep.tags && ep.tags.length ? ep.tags.map((t: string) => `<span class="badge bg-secondary me-1">${esc(t)}</span>`).join('') : '-'}</td>
        <td>${ep.enabled ? '<span class="text-success"><i class="fa-solid fa-check"></i></span>' : '<span class="text-danger"><i class="fa-solid fa-xmark"></i></span>'}</td>
        <td>
          <button class="btn btn-outline-primary btn-sm" onclick="editAIEndpoint(${ep.id})" title="Edit"><i class="fa-solid fa-pen"></i></button>
          <button class="btn btn-outline-danger btn-sm" onclick="deleteAIEndpoint(${ep.id})" title="Delete"><i class="fa-solid fa-trash"></i></button>
          <button class="btn btn-outline-info btn-sm" onclick="testAIEndpoint(${ep.id})" title="Test Connection"><i class="fa-solid fa-flask"></i></button>
        </td>
      </tr>
    `).join('');
  } catch (e: any) {
    toast(e.message, true);
  }
}
(window as any).loadAIEndpoints = loadAIEndpoints;

function toggleAIEndpointFields() {
  const type = (document.getElementById('aiEndpointType') as HTMLSelectElement).value;
  document.getElementById('aiEndpointTextFields')!.style.display = type === 'text' ? 'flex' : 'none';
  document.getElementById('aiEndpointImageSizeField')!.style.display = type === 'image' ? 'block' : 'none';
}
(window as any).toggleAIEndpointFields = toggleAIEndpointFields;

(window as any).showAddAIEndpoint = function () {
  aiEndpointEditId = null;
  document.getElementById('aiEndpointModalTitle')!.textContent = 'Add AI Endpoint';
  (document.getElementById('aiEndpointId') as HTMLInputElement).value = '';
  (document.getElementById('aiEndpointName') as HTMLInputElement).value = '';
  (document.getElementById('aiEndpointType') as HTMLSelectElement).value = 'text';
  (document.getElementById('aiEndpointBaseURL') as HTMLInputElement).value = '';
  (document.getElementById('aiEndpointAPIKey') as HTMLInputElement).value = '';
  (document.getElementById('aiEndpointTemperature') as HTMLInputElement).value = '';
  (document.getElementById('aiEndpointMaxTokens') as HTMLInputElement).value = '';
  (document.getElementById('aiEndpointImageSize') as HTMLSelectElement).value = '';
  (document.getElementById('aiEndpointTags') as HTMLInputElement).value = '';
  (document.getElementById('aiEndpointEnabled') as HTMLInputElement).checked = true;
  (document.getElementById('aiEndpointAPIKey') as HTMLInputElement).placeholder = 'sk-...';
  toggleAIEndpointFields();
  const modal = new (window as any).bootstrap.Modal(document.getElementById('aiEndpointModal')!);
  modal.show();
};

(window as any).editAIEndpoint = async function (id: number) {
  aiEndpointEditId = id;
  try {
    const ep = await api('GET', `/api/admin/ai-endpoints/${id}`);
    document.getElementById('aiEndpointModalTitle')!.textContent = 'Edit AI Endpoint';
    (document.getElementById('aiEndpointId') as HTMLInputElement).value = String(id);
    (document.getElementById('aiEndpointName') as HTMLInputElement).value = ep.name;
    (document.getElementById('aiEndpointType') as HTMLSelectElement).value = ep.type;
    (document.getElementById('aiEndpointBaseURL') as HTMLInputElement).value = ep.base_url;
    (document.getElementById('aiEndpointAPIKey') as HTMLInputElement).value = '';
    (document.getElementById('aiEndpointAPIKey') as HTMLInputElement).placeholder = 'Leave blank to keep current';
    (document.getElementById('aiEndpointModel') as HTMLInputElement).value = ep.model;
    (document.getElementById('aiEndpointTemperature') as HTMLInputElement).value = ep.temperature != null ? String(ep.temperature) : '';
    (document.getElementById('aiEndpointMaxTokens') as HTMLInputElement).value = ep.max_tokens != null ? String(ep.max_tokens) : '';
    (document.getElementById('aiEndpointImageSize') as HTMLSelectElement).value = ep.image_size || '';
    (document.getElementById('aiEndpointTags') as HTMLInputElement).value = ep.tags ? ep.tags.join(', ') : '';
    (document.getElementById('aiEndpointEnabled') as HTMLInputElement).checked = ep.enabled;
    toggleAIEndpointFields();
    const modal = new (window as any).bootstrap.Modal(document.getElementById('aiEndpointModal')!);
    modal.show();
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).saveAIEndpoint = async function () {
  const name = (document.getElementById('aiEndpointName') as HTMLInputElement).value.trim();
  const type = (document.getElementById('aiEndpointType') as HTMLSelectElement).value;
  const base_url = (document.getElementById('aiEndpointBaseURL') as HTMLInputElement).value.trim();
  const api_key = (document.getElementById('aiEndpointAPIKey') as HTMLInputElement).value;
  const model = (document.getElementById('aiEndpointModel') as HTMLInputElement).value.trim();
  const temperatureStr = (document.getElementById('aiEndpointTemperature') as HTMLInputElement).value;
  const maxTokensStr = (document.getElementById('aiEndpointMaxTokens') as HTMLInputElement).value;
  const imageSize = (document.getElementById('aiEndpointImageSize') as HTMLSelectElement).value;
  const tagsStr = (document.getElementById('aiEndpointTags') as HTMLInputElement).value;
  const enabled = (document.getElementById('aiEndpointEnabled') as HTMLInputElement).checked;

  if (!name || !type || !base_url || !model) {
    toast('Name, Type, Base URL, and Model are required', true);
    return;
  }
  if (!aiEndpointEditId && !api_key) {
    toast('API Key is required for new endpoints', true);
    return;
  }

  const body: any = { name, type, base_url, model, enabled };
  if (api_key) body.api_key = api_key;
  if (temperatureStr) body.temperature = parseFloat(temperatureStr);
  if (maxTokensStr) body.max_tokens = parseInt(maxTokensStr, 10);
  if (imageSize) body.image_size = imageSize;
  body.tags = tagsStr ? tagsStr.split(',').map((t: string) => t.trim()).filter((t: string) => t) : [];

  try {
    if (aiEndpointEditId) {
      await api('PUT', `/api/admin/ai-endpoints/${aiEndpointEditId}`, body);
      toast('Endpoint updated');
    } else {
      await api('POST', '/api/admin/ai-endpoints', body);
      toast('Endpoint created');
    }
    const modalEl = document.getElementById('aiEndpointModal')!;
    const modal = (window as any).bootstrap.Modal.getInstance(modalEl);
    if (modal) modal.hide();
    loadAIEndpoints();
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).deleteAIEndpoint = async function (id: number) {
  if (!confirm('Delete this AI endpoint? This cannot be undone.')) return;
  try {
    await api('DELETE', `/api/admin/ai-endpoints/${id}`);
    loadAIEndpoints();
    toast('Endpoint deleted');
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).testAIEndpoint = async function (id: number) {
  const btn = event?.target as HTMLElement;
  if (btn) btn.innerHTML = '<i class="fa-solid fa-spinner fa-spin"></i>';
  try {
    const result = await api('POST', `/api/admin/ai-endpoints/${id}/test`);
    if (result.success) {
      toast('Connection successful (status ' + result.status + ')');
    } else {
      toast('Test failed: ' + (result.error || 'Unknown error'), true);
    }
  } catch (e: any) {
    toast(e.message, true);
  }
  if (btn) btn.innerHTML = '<i class="fa-solid fa-flask"></i>';
};

// ─── Import Wizard ───

let importJsonData: { records: any[], filename: string } | null = null;
let importMapping: { jsonField: string, schemaField: string, schemaLabel: string, required: boolean, preview: string }[] = [];

async function loadImportSchemas() {
  try {
    const schemas = await api('GET', '/api/admin/compendium-schemas');
    const sel = document.getElementById('importSchema') as HTMLSelectElement;
    sel.innerHTML = '<option value="">— Select Schema —</option>' + schemas.map((s: any) => `<option value="${s.id}">${esc(s.display_name)} (${esc(s.type_name)})</option>`).join('');
  } catch (e: any) { toast(e.message, true); }
}

async function loadImportLogs() {
  try {
    const logs = await api('GET', '/api/admin/compendium-import-logs');
    const tbody = document.getElementById('importLogBody')!;
    if (!logs.length) {
      tbody.innerHTML = '<tr><td colspan="9" class="text-muted text-center">No imports yet</td></tr>';
      return;
    }
    tbody.innerHTML = logs.map((l: any) => {
      const summary = l.summary || {};
      const total = summary.total || l.total_entries || 0;
      const imported = summary.imported || l.imported_entries || 0;
      const skipped = summary.skipped || l.skipped_entries || 0;
      const errors = summary.errors || l.error_entries || 0;
      const statusBadge = l.status === 'completed' ? 'bg-success' : l.status === 'failed' ? 'bg-danger' : l.status === 'rolled_back' ? 'bg-warning text-dark' : 'bg-secondary';
      return `<tr>
        <td>${esc(l.schema_name || l.schema_id || '-')}</td>
        <td>${esc(l.filename || '-')}</td>
        <td style="white-space:nowrap">${l.created_at || '-'}</td>
        <td>${total}</td>
        <td>${imported}</td>
        <td>${skipped}</td>
        <td>${errors}</td>
        <td><span class="badge ${statusBadge}">${l.status}</span></td>
        <td>${l.status === 'completed' ? `<button class="btn btn-outline-warning btn-sm" onclick="rollbackImport(${l.id})">Rollback</button>` : '-'}</td>
      </tr>`;
    }).join('');
  } catch (e: any) { toast(e.message, true); }
}

(window as any).rollbackImport = async function (id: number) {
  if (!confirm('Roll back this import? This will delete imported entries. Cannot be undone.')) return;
  try {
    await api('POST', `/api/admin/compendium-import-logs/${id}/rollback`);
    toast('Import rolled back');
    loadImportLogs();
  } catch (e: any) { toast(e.message, true); }
};

(window as any).onImportSchemaChange = function () {
  // If data is already loaded, re-run auto-detection instead of hiding everything
  if (importJsonData && importJsonData.records.length > 0) {
    autoDetectMapping();
    return;
  }
  document.getElementById('importPreview')!.style.display = 'none';
  document.getElementById('importMapping')!.style.display = 'none';
  (document.getElementById('importStartBtn') as HTMLButtonElement).disabled = true;
};

(window as any).showImportPaste = function () {
  document.getElementById('importPasteArea')!.style.display = 'block';
  document.getElementById('importFetchArea')!.style.display = 'none';
};

(window as any).showImportFetch = function () {
  document.getElementById('importFetchArea')!.style.display = 'block';
  document.getElementById('importPasteArea')!.style.display = 'none';
};

function handleImportFile(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = function (e) {
    try {
      const data = JSON.parse(e.target?.result as string);
      setImportJsonData(data, file.name);
    } catch (err: any) {
      toast('Invalid JSON file: ' + err.message, true);
    }
  };
  reader.readAsText(file);
}
(window as any).handleImportFile = handleImportFile;

function useImportPaste() {
  const text = (document.getElementById('importPasteText') as HTMLTextAreaElement).value.trim();
  if (!text) { toast('Paste some JSON first', true); return; }
  try {
    const data = JSON.parse(text);
    setImportJsonData(data, 'pasted.json');
  } catch (err: any) {
    toast('Invalid JSON: ' + err.message, true);
  }
}
(window as any).useImportPaste = useImportPaste;

async function fetchImportUrl() {
  const url = (document.getElementById('importFetchUrl') as HTMLInputElement).value.trim();
  if (!url) { toast('Enter a URL', true); return; }

  // Detect and warn about GitHub blob URLs that won't work
  const githubBlobMatch = url.match(/^https?:\/\/github\.com\/([^\/]+\/[^\/]+)\/blob\/(.+)/);
  if (githubBlobMatch) {
    const rawUrl = `https://raw.githubusercontent.com/${githubBlobMatch[1]}/${githubBlobMatch[2]}`;
    toast(`GitHub blob URLs return HTML, not JSON. Try the raw URL instead: ${rawUrl}`, true);
    return;
  }

  const btn = document.querySelector('#importFetchArea .btn') as HTMLElement;
  if (btn) btn.innerHTML = '<i class="fa-solid fa-spinner fa-spin"></i> Fetching...';
  let errMsg = '';
  try {
    const res = await fetch(url);
    if (!res.ok) {
      const contentType = res.headers.get('content-type') || '';
      if (contentType.includes('text/html')) {
        errMsg = `URL returned HTML instead of JSON (HTTP ${res.status}). Make sure the URL points to a raw JSON file (e.g. raw.githubusercontent.com).`;
      } else if (res.status === 404) {
        errMsg = `URL not found (HTTP 404). Check that the path is correct.`;
      } else if (res.status === 403) {
        errMsg = `Access denied (HTTP 403). The server may be blocking cross-origin requests (CORS) or require authentication.`;
      } else {
        errMsg = `Server returned HTTP ${res.status}.`;
      }
      throw new Error(errMsg);
    }
    const data = await res.json();
    setImportJsonData(data, url.split('/').pop() || 'remote.json');
  } catch (err: any) {
    // Network/CORS errors don't have a status
    if (!errMsg) {
      if (err.message?.includes('Failed to fetch') || err.message?.includes('NetworkError')) {
        errMsg = `Cannot fetch from this URL due to CORS restrictions or network error. Try using a CORS-friendly URL like raw.githubusercontent.com, or paste the JSON content manually.`;
      } else {
        errMsg = `Fetch failed: ${err.message}`;
      }
    }
    toast(errMsg, true);
  }
  if (btn) btn.innerHTML = '<i class="fa-solid fa-download me-1"></i> Fetch';
}
(window as any).fetchImportUrl = fetchImportUrl;

function setImportJsonData(data: any, filename: string) {
  const arr = Array.isArray(data) ? data : [data];
  if (!arr.length) { toast('JSON has no records', true); return; }
  importJsonData = { records: arr, filename };
  const preview = document.getElementById('importPreview')!;
  preview.style.display = 'block';
  document.getElementById('importRecordCount')!.textContent = arr.length + ' records';
  const keys = [...new Set(arr.flatMap((r: any) => Object.keys(r)))];
  const thead = document.getElementById('importPreviewTable')!.querySelector('thead')!;
  const tbody = document.getElementById('importPreviewTable')!.querySelector('tbody')!;
  thead.innerHTML = '<tr>' + keys.slice(0, 6).map((k: string) => `<th>${esc(k)}</th>`).join('') + (keys.length > 6 ? '<th>…</th>' : '') + '</tr>';
  tbody.innerHTML = arr.slice(0, 5).map((r: any) =>
    '<tr>' + keys.slice(0, 6).map((k: string) => {
      const v = getNestedValue(r, k);
      return `<td style="max-width:150px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${esc(v != null ? String(v) : '-')}</td>`;
    }).join('') + (keys.length > 6 ? '<td>…</td>' : '') + '</tr>'
  ).join('');
  document.getElementById('importMapping')!.style.display = 'none';
  importMapping = [];
  (document.getElementById('importStartBtn') as HTMLButtonElement).disabled = true;
  toast('Loaded ' + arr.length + ' records from ' + filename);
}

function getNestedValue(obj: any, path: string): any {
  return path.split('.').reduce((o, k) => o != null ? o[k] : undefined, obj);
}

function discoverKeys(obj: any, prefix = ''): string[] {
  let keys: string[] = [];
  for (const k of Object.keys(obj)) {
    const full = prefix ? prefix + '.' + k : k;
    const v = obj[k];
    if (v !== null && typeof v === 'object' && !Array.isArray(v)) {
      keys.push(...discoverKeys(v, full));
    } else {
      keys.push(full);
    }
  }
  return keys;
}

function autoDetectMapping() {
  if (!importJsonData || !importJsonData.records.length) { toast('No data loaded', true); return; }
  const schemaId = parseInt((document.getElementById('importSchema') as HTMLSelectElement).value, 10);
  if (!schemaId) { toast('Select a target schema first', true); return; }
  api('GET', '/api/admin/compendium-schemas').then((schemas: any[]) => {
    const schema = schemas.find(s => s.id === schemaId);
    if (!schema) { toast('Schema not found', true); return; }
    const schemaFields = schema.fields || [];
    const sample = importJsonData!.records[0];
    const jsonKeys = discoverKeys(sample);
    const mapping = jsonKeys.map((jk: string) => {
      const lc = jk.toLowerCase();
      let match = schemaFields.find((f: any) => f.name.toLowerCase() === lc || f.label.toLowerCase() === lc);
      if (!match) match = schemaFields.find((f: any) => lc.includes(f.name.toLowerCase()) || f.name.toLowerCase().includes(lc));
      return {
        jsonField: jk,
        schemaField: match ? match.name : '',
        schemaLabel: match ? match.label : '(unmapped)',
        required: match ? match.required : false,
        preview: String(getNestedValue(sample, jk) ?? '').slice(0, 60)
      };
    });
    importMapping = mapping;
    renderMappingTable(schemaFields);
    document.getElementById('importMapping')!.style.display = 'block';
    (document.getElementById('importStartBtn') as HTMLButtonElement).disabled = false;
    toast('Detected ' + mapping.filter((m: any) => m.schemaField).length + ' mapped fields');
  }).catch((e: any) => toast(e.message, true));
}
(window as any).autoDetectMapping = autoDetectMapping;

function renderMappingTable(schemaFields: any[]) {
  const tbody = document.getElementById('importMappingTable')!.querySelector('tbody')!;
  tbody.innerHTML = importMapping.map((m, i) => {
    const options = '<option value="">(ignore)</option>' + schemaFields.map(f =>
      `<option value="${esc(f.name)}" ${m.schemaField === f.name ? 'selected' : ''}>${esc(f.label)}</option>`
    ).join('');
    return `<tr>
      <td><code>${esc(m.jsonField)}</code></td>
      <td>→</td>
      <td><select class="form-select form-select-sm" onchange="updateMapping(${i}, this.value)">${options}</select></td>
      <td>${m.required ? '<span class="text-danger">*</span>' : ''}</td>
      <td style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${esc(m.preview)}</td>
    </tr>`;
  }).join('');
}

(window as any).updateMapping = function (idx: number, schemaField: string) {
  importMapping[idx].schemaField = schemaField;
};

async function startImport() {
  const schemaId = parseInt((document.getElementById('importSchema') as HTMLSelectElement).value, 10);
  if (!schemaId) { toast('Select a schema', true); return; }
  if (!importJsonData || !importJsonData.records.length) { toast('No data to import', true); return; }
  const dedup = (document.getElementById('importDedup') as HTMLSelectElement).value;
  const mapping = importMapping.filter(m => m.schemaField).map(m => ({
    source_field: m.jsonField,
    schema_field: m.schemaField
  }));
  const bar = document.getElementById('importProgressBar')!;
  const pct = document.getElementById('importProgressPct')!;
  const text = document.getElementById('importProgressText')!;
  const results = document.getElementById('importResults')!;
  const btn = document.getElementById('importStartBtn') as HTMLButtonElement;
  document.getElementById('importProgressArea')!.style.display = 'block';
  btn.disabled = true;
  btn.innerHTML = '<i class="fa-solid fa-spinner fa-spin"></i> Importing...';
  const records = importJsonData.records;
  const batchSize = 50;
  let totalImported = 0, totalSkipped = 0, totalErrors = 0;
  for (let i = 0; i < records.length; i += batchSize) {
    const batch = records.slice(i, i + batchSize);
    const pctDone = Math.round((i / records.length) * 100);
    bar.style.width = pctDone + '%';
    pct.textContent = pctDone + '%';
    text.textContent = 'Importing records ' + (i + 1) + '-' + Math.min(i + batchSize, records.length) + ' of ' + records.length + '...';
    try {
      const res = await api('POST', '/api/admin/compendium-import', {
        schema_id: schemaId,
        entries: batch,
        dedup_action: dedup,
        field_mapping: mapping,
        filename: importJsonData.filename || 'import.json'
      });
      totalImported += res.imported || 0;
      totalSkipped += res.skipped || 0;
      totalErrors += (res.errors || []).length;
    } catch (e: any) {
      totalErrors += batch.length;
      results.innerHTML += `<div class="text-danger small">Batch ${Math.floor(i / batchSize) + 1} failed: ${esc(e.message)}</div>`;
    }
  }
  bar.style.width = '100%';
  pct.textContent = '100%';
  text.textContent = 'Import complete';
  results.innerHTML = `<div class="alert alert-success mb-0 py-2">
    <strong>Done!</strong> Imported ${totalImported}, Skipped ${totalSkipped}, Errors ${totalErrors} of ${records.length} records
  </div>`;
  btn.disabled = false;
  btn.innerHTML = '<i class="fa-solid fa-play me-1"></i>Start Import';
  loadImportLogs();
}
(window as any).startImport = startImport;

function resetImportForm() {
  importJsonData = null;
  importMapping = [];
  document.getElementById('importPreview')!.style.display = 'none';
  document.getElementById('importMapping')!.style.display = 'none';
  document.getElementById('importProgressArea')!.style.display = 'none';
  document.getElementById('importPasteArea')!.style.display = 'none';
  document.getElementById('importFetchArea')!.style.display = 'none';
  (document.getElementById('importPasteText') as HTMLTextAreaElement).value = '';
  (document.getElementById('importFetchUrl') as HTMLInputElement).value = '';
  (document.getElementById('importFileInput') as HTMLInputElement).value = '';
  (document.getElementById('importStartBtn') as HTMLButtonElement).disabled = true;
}
(window as any).resetImportForm = resetImportForm;

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

// ─── PDF Viewer ───

let pdfViewerDoc: any = null;
let pdfViewerPage = 1;
let pdfViewerScale = 1.5;
let pdfViewerUrl = '';
let pdfViewerTitle = '';
let pdfViewerLoaded = false;
let pdfViewerLoading = false;
const pdfViewerQueue: Array<() => void> = [];

function pdfViewerLoadLib(callback: () => void) {
  if (pdfViewerLoaded) { callback(); return; }
  if (pdfViewerLoading) { pdfViewerQueue.push(callback); return; }
  pdfViewerLoading = true;
  const s = document.createElement('script');
  s.src = 'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.min.js';
  s.onload = () => {
    (window as any).pdfjsLib.GlobalWorkerOptions.workerSrc = 'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.worker.min.js';
    pdfViewerLoaded = true;
    pdfViewerLoading = false;
    callback();
    pdfViewerQueue.forEach(fn => fn());
    pdfViewerQueue.length = 0;
  };
  s.onerror = () => {
    pdfViewerLoading = false;
    const errEl = document.getElementById('pdfViewerError');
    if (errEl) { errEl.textContent = 'Failed to load PDF viewer library. Check your internet connection.'; errEl.style.display = 'block'; }
    const loading = document.getElementById('pdfViewerLoading');
    if (loading) loading.style.display = 'none';
    pdfViewerQueue.length = 0;
  };
  document.head.appendChild(s);
}

function pdfViewerRenderPage(num: number) {
  const doc = pdfViewerDoc;
  if (!doc) return;
  const canvas = document.getElementById('pdfViewerCanvas') as HTMLCanvasElement;
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  if (!ctx) return;
  doc.getPage(num).then((page: any) => {
    const viewport = page.getViewport({ scale: pdfViewerScale });
    canvas.width = viewport.width;
    canvas.height = viewport.height;
    canvas.style.display = 'block';
    const loading = document.getElementById('pdfViewerLoading');
    if (loading) loading.style.display = 'none';
    const err = document.getElementById('pdfViewerError');
    if (err) err.style.display = 'none';
    page.render({ canvasContext: ctx, viewport: viewport });
    const info = document.getElementById('pdfViewerPageInfo');
    if (info) info.textContent = num + ' / ' + doc.numPages;
    const prev = document.getElementById('pdfViewerPrevBtn') as HTMLButtonElement;
    const next = document.getElementById('pdfViewerNextBtn') as HTMLButtonElement;
    if (prev) prev.disabled = num <= 1;
    if (next) next.disabled = num >= doc.numPages;
    const zoom = document.getElementById('pdfViewerZoomLevel');
    if (zoom) zoom.textContent = Math.round(pdfViewerScale * 100) + '%';
  });
}

function pdfViewerShowError(msg: string) {
  const el = document.getElementById('pdfViewerError');
  if (el) { el.textContent = msg; el.style.display = 'block'; }
  const loading = document.getElementById('pdfViewerLoading');
  if (loading) loading.style.display = 'none';
}

function pdfViewerFilenameFromUrl(url: string): string {
  const parts = url.split('/');
  const last = parts[parts.length - 1] || 'document.pdf';
  return decodeURIComponent(last);
}

(window as any).openPdfViewer = function (url: string, title?: string) {
  pdfViewerUrl = url;
  pdfViewerTitle = title || pdfViewerFilenameFromUrl(url);
  const modalEl = document.getElementById('pdfViewerModal');
  if (!modalEl) return;
  document.getElementById('pdfViewerTitle')!.textContent = pdfViewerTitle;
  const loading = document.getElementById('pdfViewerLoading');
  if (loading) {
    loading.style.display = 'block';
    loading.innerHTML = '<div class="spinner-border text-light mb-2" role="status"></div><p class="mb-0">Loading PDF...</p>';
  }
  const canvas = document.getElementById('pdfViewerCanvas') as HTMLCanvasElement;
  if (canvas) canvas.style.display = 'none';
  const err = document.getElementById('pdfViewerError');
  if (err) err.style.display = 'none';
  const info = document.getElementById('pdfViewerPageInfo');
  if (info) info.textContent = '- / -';
  const prev = document.getElementById('pdfViewerPrevBtn') as HTMLButtonElement;
  const next = document.getElementById('pdfViewerNextBtn') as HTMLButtonElement;
  if (prev) prev.disabled = true;
  if (next) next.disabled = true;
  const zoom = document.getElementById('pdfViewerZoomLevel');
  if (zoom) zoom.textContent = '100%';
  const modal = (window as any).bootstrap.Modal.getOrCreateInstance(modalEl);
  modal.show();
  pdfViewerScale = 1.5;
  pdfViewerPage = 1;
  if (pdfViewerDoc) { pdfViewerDoc.destroy(); pdfViewerDoc = null; }
  pdfViewerLoadLib(() => {
    (window as any).pdfjsLib.getDocument(url).promise.then((doc: any) => {
      pdfViewerDoc = doc;
      pdfViewerRenderPage(1);
    }).catch((err: any) => {
      pdfViewerShowError('Failed to load PDF: ' + (err.message || 'Unknown error'));
    });
  });
};

(window as any).pdfViewerPrevPage = function () {
  if (!pdfViewerDoc || pdfViewerPage <= 1) return;
  pdfViewerPage--;
  pdfViewerRenderPage(pdfViewerPage);
};

(window as any).pdfViewerNextPage = function () {
  if (!pdfViewerDoc || pdfViewerPage >= pdfViewerDoc.numPages) return;
  pdfViewerPage++;
  pdfViewerRenderPage(pdfViewerPage);
};

(window as any).pdfViewerZoomIn = function () {
  pdfViewerScale = Math.min(pdfViewerScale * 1.25, 5);
  if (pdfViewerDoc) pdfViewerRenderPage(pdfViewerPage);
};

(window as any).pdfViewerZoomOut = function () {
  pdfViewerScale = Math.max(pdfViewerScale / 1.25, 0.25);
  if (pdfViewerDoc) pdfViewerRenderPage(pdfViewerPage);
};

(window as any).pdfViewerFitToWidth = function () {
  if (!pdfViewerDoc) return;
  const canvas = document.getElementById('pdfViewerCanvas') as HTMLCanvasElement;
  if (!canvas) return;
  const container = canvas.parentElement;
  if (!container) return;
  const cw = container.clientWidth - 40;
  pdfViewerDoc.getPage(pdfViewerPage).then((page: any) => {
    const ov = page.getViewport({ scale: 1 });
    pdfViewerScale = cw / ov.width;
    pdfViewerRenderPage(pdfViewerPage);
  });
};

(window as any).pdfViewerDownload = function () {
  if (pdfViewerUrl) {
    const a = document.createElement('a');
    a.href = pdfViewerUrl;
    a.download = pdfViewerTitle.replace(/[^a-zA-Z0-9._-]/g, '_') + '.pdf';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
  }
};

// Cleanup on modal close
document.addEventListener('hidden.bs.modal', function (e: Event) {
  const target = e.target as HTMLElement;
  if (target && target.id === 'pdfViewerModal') {
    if (pdfViewerDoc) { pdfViewerDoc.destroy(); pdfViewerDoc = null; }
    pdfViewerPage = 1;
    pdfViewerScale = 1.5;
  }
});

// Keyboard navigation
document.addEventListener('keydown', function (e: KeyboardEvent) {
  const modalEl = document.getElementById('pdfViewerModal');
  if (modalEl && modalEl.classList.contains('show')) {
    if (e.key === 'ArrowLeft') { (window as any).pdfViewerPrevPage(); e.preventDefault(); }
    else if (e.key === 'ArrowRight') { (window as any).pdfViewerNextPage(); e.preventDefault(); }
  }
});

init();

})();
