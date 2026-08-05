import { expose } from './lib/expose';
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
expose('toggleTheme', toggleTheme);

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
  const tabBtn = document.getElementById('tab' + capitalize(tab) + 'Btn');
  if (tabBtn) tabBtn.classList.add('active');
  const allTabs = ['users', 'unified-compendium', 'backup', 'email', 'ai-endpoints', 'analytics', 'telemetry', 'events', 'import', 'e-ink', 'logs'];
  allTabs.forEach(s => {
    const parts = s.split('-').map((p, i) => i === 0 ? capitalize(p) : capitalize(p));
    const id = 'admin' + parts.join('');
    const el = document.getElementById(id);
    if (el) el.style.display = s === tab ? 'block' : 'none';
  });
  if (tab === 'users') loadUsers();
  if (tab === 'unified-compendium') loadUnifiedCompendium();
  if (tab === 'backup') { loadBackupSettings(); loadBackupList(); }
  if (tab === 'email') loadEmailSettings();
  if (tab === 'ai-endpoints') loadAIEndpoints();
  if (tab === 'analytics') loadUmamiSettings();
  if (tab === 'telemetry') loadOTelSettings();
  if (tab === 'events') { loadEventsSettings(); loadCampaignEventSettings(); loadEventsPublicLink(); }
  if (tab === 'import') { loadImportSchemas(); loadImportLogs(); }
  if (tab === 'e-ink') loadEinkSetting();
  if (tab === 'logs') { startLogAutoRefresh(); }
  else { stopLogAutoRefresh(); }
}
expose('showAdminTab', showAdminTab);

// ─── E-ink Mode ───

async function loadEinkSetting() {
  try {
    const res = await api('GET', '/api/admin/settings/eink');
    const el = document.getElementById('einkEnabled') as HTMLInputElement | null;
    if (el) el.checked = !!res.enabled;
  } catch {
    toast('Failed to load e-ink setting', true);
  }
}

async function saveEinkSetting() {
  const el = document.getElementById('einkEnabled') as HTMLInputElement | null;
  const enabled = !!el && el.checked;
  try {
    await api('PUT', '/api/admin/settings/eink', { enabled });
    toast(enabled ? 'E-ink mode enabled site-wide' : 'E-ink mode disabled');
  } catch {
    toast('Failed to save e-ink setting', true);
  }
}
expose('loadEinkSetting', loadEinkSetting);
expose('saveEinkSetting', saveEinkSetting);

// ─── Unified Compendium ───

let currentSchemaId: number = 0;
let currentSchemaFields: any[] = [];
let currentSchemaPage: number = 1;
let currentSchemaQuery: string = '';
let currentSchemaName: string = '';
let selectedEntryIds: Set<number> = new Set();
let entryModalSchemaId = 0;
let entryModalSchemaFields: any[] = [];
let entryModalEditId: number | null = null;

async function loadUnifiedCompendium() {
  const tbody = document.getElementById('unifiedSchemaBody')!;
  try {
    const schemas = await api('GET', '/api/admin/compendium-schemas');
    document.getElementById('schemaCount')!.textContent = schemas.length + ' schemas';
    tbody.innerHTML = schemas.map((s: any) => {
      const icon = s.type_name === 'monster' ? '🐉' : s.type_name === 'spell' ? '🔮' : s.type_name === 'race' ? '🧝' : s.type_name === 'class' ? '⚔️' : s.type_name === 'equipment' ? '🛡️' : s.type_name === 'feat' ? '💪' : s.type_name === 'background' ? '📜' : '📖';
      return `<tr>
        <td style="font-size:1.2rem">${icon}</td>
        <td><strong>${esc(s.display_name)}</strong></td>
        <td><code>${esc(s.type_name)}</code></td>
        <td>${s.fields ? s.fields.length : 0}</td>
        <td><span class="badge badge-primary">${s.entry_count || 0}</span></td>
        <td class="text-nowrap">
          <button class="btn btn-outline-primary btn-sm py-0" onclick="browseSchema(${s.id},'${esc(s.display_name)}')" title="Browse entries"><i class="fa-solid fa-list"></i></button>
          <button class="btn btn-outline-info btn-sm py-0" onclick="editSchema(${s.id})" title="Edit schema"><i class="fa-solid fa-pen"></i></button>
          <button class="btn btn-outline-warning btn-sm py-0" onclick="openImportForSchemaId(${s.id})" title="Import"><i class="fa-solid fa-upload"></i></button>
          <button class="btn btn-outline-danger btn-sm py-0" onclick="deleteSchema(${s.id})" title="Delete"><i class="fa-solid fa-trash"></i></button>
        </td>
      </tr>`;
    }).join('');
  } catch (e: any) {
    tbody.innerHTML = '<tr><td colspan="6" class="text-danger text-center">Failed to load schemas: ' + esc(e.message) + '</td></tr>';
  }
}

expose('loadUnifiedCompendium', loadUnifiedCompendium);

// ─── Global Cross-Schema Search ───

let globalSearchTimeout: any = null;

function onGlobalSearchInput(value: string) {
  clearTimeout(globalSearchTimeout);
  const q = value.trim();
  if (q.length < 2) {
    document.getElementById('globalSearchResults')!.style.display = 'none';
    document.getElementById('unifiedSchemaListView')!.style.display = 'block';
    document.getElementById('unifiedEntryBrowser')!.style.display = 'none';
    return;
  }
  globalSearchTimeout = setTimeout(() => doGlobalSearch(q), 300);
}
expose('onGlobalSearchInput', onGlobalSearchInput);

async function doGlobalSearch(q: string) {
  try {
    document.getElementById('unifiedSchemaListView')!.style.display = 'none';
    document.getElementById('unifiedEntryBrowser')!.style.display = 'none';
    const results = await api('GET', '/api/admin/compendium-search?q=' + encodeURIComponent(q));
    const container = document.getElementById('globalSearchResults')!;
    container.style.display = 'block';
    if (!results || results.length === 0) {
      document.getElementById('globalSearchEmpty')!.style.display = 'block';
      document.getElementById('globalSearchResultGroups')!.innerHTML = '';
      return;
    }
    document.getElementById('globalSearchEmpty')!.style.display = 'none';
    // Group by type_name
    const groups: Record<string, any[]> = {};
    for (const r of results) {
      const key = r.type_name || r.type || 'Other';
      if (!groups[key]) groups[key] = [];
      groups[key].push(r);
    }
    const groupsHtml = Object.entries(groups).map(([typeName, items]) => `
      <div class="mb-2">
        <div class="d-flex align-items-center gap-2 mb-1">
          <span class="badge bg-secondary">${esc(typeName)}</span>
          <small class="text-muted">${items.length} result${items.length > 1 ? 's' : ''}</small>
        </div>
        <div class="list-group list-group-flush small">
          ${items.map((r: any) => `
            <button class="list-group-item list-group-item-action py-1 d-flex align-items-center gap-2" onclick="openEntryFromSearch(${r.id},${r.type || 0})">
              <i class="fa-solid fa-file-lines text-muted" style="font-size:0.75rem"></i>
              <span>${esc(r.name || '(unnamed)')}</span>
              <small class="text-muted ms-auto">${r.snippet ? esc(r.snippet.slice(0, 60)) : ''}</small>
            </button>
          `).join('')}
        </div>
      </div>
    `).join('');
    document.getElementById('globalSearchResultGroups')!.innerHTML = groupsHtml;
  } catch (e: any) {
    toast('Search failed: ' + e.message, true);
  }
}

function openEntryFromSearch(entryId: number, typeId: number) {
  // We need to figure out the schema ID. The search doesn't return it directly.
  // Fetch the entry to get schema_id
  api('GET', '/api/admin/compendium-entries/' + entryId).then(entry => {
    if (entry && entry.schema_id) {
      editSchemaEntryById(entryId, entry.schema_id);
    }
  }).catch(() => {
    toast('Could not open entry', true);
  });
}
expose('openEntryFromSearch', openEntryFromSearch);

function clearGlobalSearch() {
  (document.getElementById('compendiumGlobalSearch') as HTMLInputElement).value = '';
  document.getElementById('globalSearchResults')!.style.display = 'none';
  document.getElementById('unifiedSchemaListView')!.style.display = 'block';
  document.getElementById('unifiedEntryBrowser')!.style.display = 'none';
}
expose('clearGlobalSearch', clearGlobalSearch);

// ─── Entry Browser ───

function browseSchema(schemaId: number, schemaName: string) {
  currentSchemaId = schemaId;
  currentSchemaName = schemaName;
  currentSchemaPage = 1;
  currentSchemaQuery = '';
  selectedEntryIds.clear();
  document.getElementById('unifiedSchemaListView')!.style.display = 'none';
  document.getElementById('globalSearchResults')!.style.display = 'none';
  document.getElementById('unifiedEntryBrowser')!.style.display = 'block';
  document.getElementById('entryBrowserTitle')!.textContent = schemaName + ' Entries';
  loadSchemaEntries();
}
expose('browseSchema', browseSchema);

function backToSchemaList() {
  document.getElementById('unifiedEntryBrowser')!.style.display = 'none';
  document.getElementById('unifiedSchemaListView')!.style.display = 'block';
  currentSchemaId = 0;
}
expose('backToSchemaList', backToSchemaList);

async function loadSchemaEntries() {
  const schemaId = currentSchemaId;
  if (!schemaId) return;
  const page = currentSchemaPage;
  const q = currentSchemaQuery;
  const tbody = document.getElementById('entryTableBody')!;
  const thead = document.getElementById('entryTableHead')!;
  try {
    const params = new URLSearchParams({ page: String(page), page_size: '50' });
    if (q) params.set('q', q);
    const result = await api('GET', '/api/admin/compendium-schemas/' + schemaId + '/entries?' + params.toString());
    const entries = result.entries || result || [];
    const total = result.total || entries.length;
    const totalPages = result.total_pages || Math.max(1, Math.ceil(total / 50));

    // Get schema fields for column headers
    const schemas = await api('GET', '/api/admin/compendium-schemas');
    const schema = schemas.find((s: any) => s.id === schemaId);
    const fields = schema?.fields || [];
    currentSchemaFields = fields;

    // Build header - first 5 fields
    const previewFields = fields.slice(0, 5);
    thead.innerHTML = '<tr>' +
      '<th style="width:36px"><input class="form-check-input" type="checkbox" id="selectAllEntries" onchange="toggleSelectAll()"></th>' +
      previewFields.map((f: any) => `<th>${esc(f.label || f.name)}</th>`).join('') +
      (fields.length > 5 ? '<th style="width:30px">…</th>' : '') +
      '<th style="width:160px">Actions</th>' +
    '</tr>';

    if (!entries || entries.length === 0) {
      tbody.innerHTML = '<tr><td colspan="' + (previewFields.length + 3) + '" class="text-muted text-center py-3">No entries found</td></tr>';
      document.getElementById('entryCountInfo')!.textContent = '0 entries';
      document.getElementById('entryPagination')!.style.display = 'none';
      updateBulkActionsToolbar();
      return;
    }

    tbody.innerHTML = entries.map((e: any) => {
      const data = e.data || {};
      return '<tr>' +
        '<td><input class="form-check-input entry-select-cb" type="checkbox" data-entry-id="' + e.id + '" onchange="toggleEntrySelect(' + e.id + ')" ' + (selectedEntryIds.has(e.id) ? 'checked' : '') + '></td>' +
        previewFields.map((f: any) => {
          let val = data[f.name];
          if (val === null || val === undefined) val = '';
          if (f.type === 'boolean') val = val ? '✓' : '—';
          else if (typeof val === 'object') val = JSON.stringify(val).slice(0, 50) + '…';
          else val = String(val).slice(0, 80);
          return '<td style="max-width:150px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">' + esc(val) + '</td>';
        }).join('') +
        (fields.length > 5 ? '<td title="' + fields.slice(5).map((f: any) => esc(f.label || f.name) + ': ' + esc(String(data[f.name] ?? ''))).join(' | ') + '">…</td>' : '') +
        '<td class="text-nowrap">' +
          '<button class="btn btn-outline-primary btn-sm py-0" onclick="editSchemaEntryById(' + e.id + ',' + schemaId + ')" title="Edit"><i class="fa-solid fa-pen"></i></button> ' +
          '<button class="btn btn-outline-info btn-sm py-0" onclick="viewSchemaEntry(' + e.id + ')" title="View"><i class="fa-solid fa-eye"></i></button> ' +
          '<button class="btn btn-outline-secondary btn-sm py-0" onclick="duplicateSchemaEntry(' + e.id + ',' + schemaId + ')" title="Duplicate"><i class="fa-solid fa-copy"></i></button> ' +
          '<button class="btn btn-outline-danger btn-sm py-0" onclick="deleteSchemaEntryById(' + e.id + ')" title="Delete"><i class="fa-solid fa-trash"></i></button>' +
        '</td>' +
      '</tr>';
    }).join('');

    document.getElementById('entryCountInfo')!.textContent = total + ' entries (page ' + page + ' of ' + totalPages + ')';
    // Pagination
    const paginationEl = document.getElementById('entryPagination')!;
    if (totalPages <= 1) {
      paginationEl.style.display = 'none';
    } else {
      paginationEl.style.display = 'block';
      const ul = document.getElementById('entryPaginationUl')!;
      let pagHtml = '';
      if (page > 1) pagHtml += '<li class="page-item"><button class="page-link" onclick="goToPage(' + (page - 1) + ')">«</button></li>';
      for (let p = Math.max(1, page - 2); p <= Math.min(totalPages, page + 2); p++) {
        pagHtml += '<li class="page-item' + (p === page ? ' active' : '') + '"><button class="page-link" onclick="goToPage(' + p + ')">' + p + '</button></li>';
      }
      if (page < totalPages) pagHtml += '<li class="page-item"><button class="page-link" onclick="goToPage(' + (page + 1) + ')">»</button></li>';
      ul.innerHTML = pagHtml;
    }
    updateBulkActionsToolbar();
  } catch (e: any) {
    tbody.innerHTML = '<tr><td colspan="10" class="text-danger text-center">Failed: ' + esc(e.message) + '</td></tr>';
  }
}

function goToPage(page: number) {
  currentSchemaPage = page;
  loadSchemaEntries();
}
expose('goToPage', goToPage);

let entrySearchTimeout: any = null;
function onEntrySearchInput(value: string) {
  clearTimeout(entrySearchTimeout);
  currentSchemaQuery = value.trim();
  currentSchemaPage = 1;
  entrySearchTimeout = setTimeout(() => loadSchemaEntries(), 300);
}
expose('onEntrySearchInput', onEntrySearchInput);

function addEntryCurrentSchema() {
  if (currentSchemaId) createSchemaEntry(currentSchemaId);
}
expose('addEntryCurrentSchema', addEntryCurrentSchema);

// ─── Entry Selection / Bulk Ops ───

function toggleEntrySelect(id: number) {
  if (selectedEntryIds.has(id)) selectedEntryIds.delete(id);
  else selectedEntryIds.add(id);
  updateBulkActionsToolbar();
}
expose('toggleEntrySelect', toggleEntrySelect);

function toggleSelectAll() {
  const cb = document.getElementById('selectAllEntries') as HTMLInputElement;
  const checkboxes = document.querySelectorAll<HTMLInputElement>('.entry-select-cb');
  checkboxes.forEach(c => {
    c.checked = cb.checked;
    const eid = parseInt(c.dataset.entryId || '0', 10);
    if (cb.checked) selectedEntryIds.add(eid);
    else selectedEntryIds.delete(eid);
  });
  updateBulkActionsToolbar();
}
expose('toggleSelectAll', toggleSelectAll);

function updateBulkActionsToolbar() {
  const count = selectedEntryIds.size;
  const el = document.getElementById('bulkActionsToolbar')!;
  if (count === 0) { el.style.display = 'none'; return; }
  el.style.display = 'flex';
  document.getElementById('bulkSelectedCount')!.textContent = count + ' selected';
}

function clearEntrySelection() {
  selectedEntryIds.clear();
  document.querySelectorAll<HTMLInputElement>('.entry-select-cb').forEach(c => c.checked = false);
  const selAll = document.getElementById('selectAllEntries') as HTMLInputElement;
  if (selAll) selAll.checked = false;
  updateBulkActionsToolbar();
}
expose('clearEntrySelection', clearEntrySelection);

async function batchDeleteEntries() {
  const ids = Array.from(selectedEntryIds);
  if (ids.length === 0) return;
  if (!confirm('Delete ' + ids.length + ' entries? This cannot be undone.')) return;
  try {
    const res = await api('POST', '/api/admin/compendium-entries/batch-delete', { ids });
    toast('Deleted ' + res.deleted + ' entries');
    selectedEntryIds.clear();
    loadSchemaEntries();
  } catch (e: any) {
    toast(e.message, true);
  }
}
expose('batchDeleteEntries', batchDeleteEntries);

async function batchEditEntries() {
  const ids = Array.from(selectedEntryIds);
  if (ids.length === 0) return;
  const schemas = await api('GET', '/api/admin/compendium-schemas');
  const schema = schemas.find((s: any) => s.id === currentSchemaId) || schemas[0];
  const fields = schema?.fields || [];
  const fieldOpts = fields.map((f: any) =>
    `<option value="${esc(f.name)}">${esc(f.label || f.name)}</option>`
  ).join('');
  showModal('Bulk Edit ' + ids.length + ' Entries', `
    <p>Set a field value for all ${ids.length} selected entries.</p>
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
}
expose('batchEditEntries', batchEditEntries);

async function saveBatchEdit() {
  const field = (document.getElementById('bulkField') as HTMLSelectElement).value;
  const value = (document.getElementById('bulkValue') as HTMLInputElement).value;
  if (!field) { toast('Select a field', true); return; }
  const ids = Array.from(selectedEntryIds);
  try {
    const res = await api('POST', '/api/admin/compendium-entries/batch-update', { ids, data: { [field]: value } });
    toast('Updated ' + res.updated + ' entries');
    hideModal();
    selectedEntryIds.clear();
    loadSchemaEntries();
  } catch (e: any) {
    toast(e.message, true);
  }
}
expose('saveBatchEdit', saveBatchEdit);

async function batchExportEntries() {
  const ids = Array.from(selectedEntryIds);
  if (ids.length === 0) return;
  try {
    const entries: any[] = [];
    for (const id of ids) {
      const entry = await api('GET', '/api/admin/compendium-entries/' + id);
      if (entry) entries.push(entry);
    }
    const blob = new Blob([JSON.stringify({ schema_id: currentSchemaId, entries: entries }, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = currentSchemaName.replace(/\s+/g, '-').toLowerCase() + '-export.json';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    toast('Exported ' + entries.length + ' entries');
  } catch (e: any) {
    toast('Export failed: ' + e.message, true);
  }
}
expose('batchExportEntries', batchExportEntries);

// ─── Entry CRUD ───

function createSchemaEntry(schemaId: number) {
  api('GET', '/api/admin/compendium-schemas').then((schemas: any[]) => {
    const schema = schemas.find((s: any) => s.id === schemaId);
    if (!schema) { toast('Schema not found', true); return; }
    entryModalSchemaId = schemaId;
    entryModalSchemaFields = schema.fields || [];
    entryModalEditId = null;
    showModal('Create Entry in ' + esc(schema.display_name), getEntryFormHtml(null));
  }).catch((e: any) => toast(e.message, true));
}
expose('createSchemaEntry', createSchemaEntry);

function editSchemaEntryById(entryId: number, schemaId: number) {
  Promise.all([
    api('GET', '/api/admin/compendium-schemas'),
    api('GET', '/api/admin/compendium-entries/' + entryId)
  ]).then(([schemas, entry]) => {
    const schema = schemas.find((s: any) => s.id === schemaId);
    if (!schema) { toast('Schema not found', true); return; }
    entryModalSchemaId = schemaId;
    entryModalSchemaFields = schema.fields || [];
    entryModalEditId = entryId;
    showModal('Edit Entry', getEntryFormHtml(entry.data || entry));
  }).catch((e: any) => toast(e.message, true));
}
expose('editSchemaEntryById', editSchemaEntryById);

function viewSchemaEntry(entryId: number) {
  api('GET', '/api/admin/compendium-entries/' + entryId).then(entry => {
    if (!entry) { toast('Entry not found', true); return; }
    const data = entry.data || {};
    const schemaId = entry.schema_id;
    api('GET', '/api/admin/compendium-schemas').then((schemas: any[]) => {
      const schema = schemas.find((s: any) => s.id === schemaId);
      const fields = schema?.fields || [];
      const bodyHtml = fields.map((f: any) => {
        let val = data[f.name];
        if (val === null || val === undefined) val = '';
        if (f.type === 'boolean') {
          val = val ? '<span class="text-success"><i class="fa-solid fa-check-circle me-1"></i>Yes</span>' : '<span class="text-muted"><i class="fa-solid fa-circle-xmark me-1"></i>No</span>';
        } else if (f.type === 'json' && val) {
          try {
            const formatted = typeof val === 'string' ? JSON.stringify(JSON.parse(val), null, 2) : JSON.stringify(val, null, 2);
            val = '<pre class="mb-0" style="max-height:200px;overflow:auto"><code>' + esc(formatted) + '</code></pre>';
          } catch { val = esc(String(val)); }
        } else if (typeof val === 'object') {
          val = '<pre class="mb-0" style="max-height:200px;overflow:auto"><code>' + esc(JSON.stringify(val, null, 2)) + '</code></pre>';
        } else {
          val = esc(String(val));
        }
        return `<div class="mb-2">
          <label class="form-label text-muted small mb-0">${esc(f.label || f.name)}</label>
          <div class="p-2 bg-light rounded">${val}</div>
        </div>`;
      }).join('');
      showModal('View: ' + esc(data.name || data.Name || 'Entry'), bodyHtml + `
        <div class="mt-3">
          <button class="btn btn-primary" onclick="hideModal();editSchemaEntryById(${entry.id},${schemaId})"><i class="fa-solid fa-pen me-1"></i>Edit</button>
          <button class="btn btn-secondary" onclick="hideModal()">Close</button>
        </div>
      `);
    });
  }).catch((e: any) => toast(e.message, true));
}
expose('viewSchemaEntry', viewSchemaEntry);

async function duplicateSchemaEntry(entryId: number, schemaId: number) {
  try {
    const entry = await api('GET', '/api/admin/compendium-entries/' + entryId);
    if (!entry) { toast('Entry not found', true); return; }
    const data = entry.data || {};
    const name = data.name || data.Name || '';
    if (name) data.name = name + ' (copy)';
    await api('POST', '/api/admin/compendium-schemas/' + schemaId + '/entries', { data });
    toast('Entry duplicated');
    loadSchemaEntries();
  } catch (e: any) {
    toast(e.message, true);
  }
}
expose('duplicateSchemaEntry', duplicateSchemaEntry);

function deleteSchemaEntryById(entryId: number) {
  if (!confirm('Delete this entry?')) return;
  api('DELETE', '/api/admin/compendium-entries/' + entryId).then(() => {
    toast('Entry deleted');
    selectedEntryIds.delete(entryId);
    loadSchemaEntries();
  }).catch((e: any) => toast(e.message, true));
}
expose('deleteSchemaEntryById', deleteSchemaEntryById);

// ─── Schema-Aware Entry Editor ───

function getEntryFormHtml(data: any): string {
  const fields = entryModalSchemaFields;
  if (!fields || fields.length === 0) {
    return '<div class="text-muted">No fields defined for this schema.</div>';
  }
  return fields.map((f: any) => {
    const val = data ? data[f.name] : undefined;
    const requiredAttr = f.required ? 'required' : '';
    const requiredMark = f.required ? ' <span class="text-danger">*</span>' : '';
    let input = '';

    switch (f.type) {
      case 'text':
      case 'textarea':
        input = `<textarea class="form-control" id="ef_${esc(f.name)}" rows="3" ${requiredAttr}>${esc(String(val ?? ''))}</textarea>`;
        break;
      case 'integer':
        input = `<input class="form-control" type="number" step="1" id="ef_${esc(f.name)}" value="${esc(String(val ?? ''))}" ${requiredAttr}>`;
        break;
      case 'float':
        input = `<input class="form-control" type="number" step="0.01" id="ef_${esc(f.name)}" value="${esc(String(val ?? ''))}" ${requiredAttr}>`;
        break;
      case 'boolean':
        const checked = val ? 'checked' : '';
        input = `<div class="form-check"><input class="form-check-input" type="checkbox" id="ef_${esc(f.name)}" ${checked}></div>`;
        break;
      case 'select':
        const options = (f.options || []).map((o: string) =>
          `<option value="${esc(o)}" ${String(val) === o ? 'selected' : ''}>${esc(o)}</option>`
        ).join('');
        input = `<select class="form-select" id="ef_${esc(f.name)}" ${requiredAttr}>${options}</select>`;
        break;
      case 'multi-select':
        const selectedVals = Array.isArray(val) ? val : (val ? String(val).split(',') : []);
        const multiOpts = (f.options || []).map((o: string) =>
          `<option value="${esc(o)}" ${selectedVals.includes(o) ? 'selected' : ''}>${esc(o)}</option>`
        ).join('');
        input = `<select class="form-select" multiple id="ef_${esc(f.name)}" ${requiredAttr}>${multiOpts}</select>`;
        break;
      case 'json':
        let jsonVal = '';
        if (val) {
          try { jsonVal = typeof val === 'string' ? val : JSON.stringify(val, null, 2); }
          catch { jsonVal = String(val); }
        }
        input = `<textarea class="form-control font-monospace" id="ef_${esc(f.name)}" rows="4" ${requiredAttr} placeholder="Enter JSON...">${esc(jsonVal)}</textarea>`;
        break;
      default:
        input = `<input class="form-control" type="text" id="ef_${esc(f.name)}" value="${esc(String(val ?? ''))}" ${requiredAttr}>`;
    }

    return `<div class="mb-2">
      <label class="form-label">${esc(f.label || f.name)}${requiredMark}</label>
      ${input}
      ${f.type === 'json' ? '<small class="text-muted">Must be valid JSON</small>' : ''}
    </div>`;
  }).join('') + `
    <div class="mt-3">
      <button class="btn btn-primary" onclick="saveEntry()"><i class="fa-solid fa-floppy-disk me-1"></i>Save</button>
      <button class="btn btn-secondary" onclick="hideModal()">Cancel</button>
    </div>`;
}

expose('saveEntry', async function () {
  const data: Record<string, any> = {};
  let valid = true;

  for (const f of entryModalSchemaFields) {
    const el = document.getElementById('ef_' + f.name) as HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement;
    if (!el) continue;
    let val: any;

    switch (f.type) {
      case 'boolean':
        val = (el as HTMLInputElement).checked;
        break;
      case 'integer':
        val = el.value ? parseInt(el.value, 10) : null;
        if (val !== null && isNaN(val)) { val = null; }
        break;
      case 'float':
        val = el.value ? parseFloat(el.value) : null;
        if (val !== null && isNaN(val)) { val = null; }
        break;
      case 'multi-select':
        val = Array.from((el as HTMLSelectElement).selectedOptions).map(opt => opt.value);
        if (val.length === 0) val = null;
        break;
      case 'json':
        if (el.value.trim()) {
          try {
            val = JSON.parse(el.value);
          } catch {
            el.classList.add('is-invalid');
            valid = false;
            continue;
          }
        } else {
          val = null;
        }
        break;
      case 'select':
        val = (el as HTMLSelectElement).value;
        break;
      default:
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
      await api('PUT', '/api/admin/compendium-entries/' + entryModalEditId, { data });
      toast('Entry updated');
    } else {
      await api('POST', '/api/admin/compendium-schemas/' + entryModalSchemaId + '/entries', { data });
      toast('Entry created');
    }
    hideModal();
    loadSchemaEntries();
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── Import Integration ───

function openImportForSchema() {
  if (currentSchemaId) openImportForSchemaId(currentSchemaId);
}
expose('openImportForSchema', openImportForSchema);

function openImportForSchemaId(schemaId: number) {
  // Show the import tab with the schema pre-selected
  showAdminTab('import');
  const sel = document.getElementById('importSchema') as HTMLSelectElement;
  // Wait for schemas to load
  setTimeout(() => {
    sel.value = String(schemaId);
    if (sel.value) {
      const event = new Event('change');
      sel.dispatchEvent(event);
      toast('Schema pre-selected for import');
    }
  }, 500);
}
expose('openImportForSchemaId', openImportForSchemaId);

// ─── Logs ───

function startLogAutoRefresh() {
  loadLogLevel();
  loadLogSources();
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

async function loadLogSources() {
  try {
    const sources: string[] = await api('GET', '/api/admin/log-sources');
    const sel = document.getElementById('logSourceFilter') as HTMLSelectElement;
    if (!sel) return;
    const current = sel.value;
    // Keep the "All" option, rebuild the rest
    sel.innerHTML = '<option value="">All</option>';
    const seen = new Set<string>();
    for (const s of sources) {
      if (!s || seen.has(s)) continue;
      seen.add(s);
      const opt = document.createElement('option');
      opt.value = s;
      opt.textContent = s.charAt(0).toUpperCase() + s.slice(1);
      sel.appendChild(opt);
    }
    if (current && [...sel.options].some(o => o.value === current)) {
      sel.value = current;
    }
  } catch { /* silently degrade to hardcoded options */ }
}

async function loadLogs() {
  try {
    const sourceFilter = (document.getElementById('logSourceFilter') as HTMLSelectElement)?.value || '';
    let url = '/api/admin/logs?limit=200';
    if (sourceFilter) url += '&source=' + encodeURIComponent(sourceFilter);

    const tableContainer = document.getElementById('adminLogs')?.querySelector('.table-responsive');
    const wasAtBottom = tableContainer
      ? tableContainer.scrollHeight - tableContainer.scrollTop - tableContainer.clientHeight < 50
      : false;

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
    tbody.innerHTML = logs.map((l: any, idx: number) => {
      const badge = levelBadge[l.level] || 'bg-secondary';
      const ts = l.timestamp ? new Date(l.timestamp).toLocaleString() : '-';
      const hasAttrs = l.attributes && Object.keys(l.attributes).length > 0;
      const attrSummary = hasAttrs
        ? Object.entries(l.attributes).map(([k, v]) => `<span class="badge bg-light text-dark me-1">${esc(k)}=${esc(String(v)).substring(0, 30)}</span>`).join('')
        : '';
      const detailId = `logDetail_${idx}`;
      const attrsHtml = hasAttrs
        ? '<dl class="log-detail-attrs mb-0">' +
          Object.entries(l.attributes).map(([k, v]) => `<dt>${esc(k)}</dt><dd><code>${esc(JSON.stringify(v))}</code></dd>`).join('') +
          '</dl>'
        : '<span class="text-muted">No attributes</span>';
      return `<tr class="log-row" style="cursor:pointer" onclick="toggleLogDetail('${detailId}')">
        <td><span class="badge ${badge}">${esc(l.level)}</span></td>
        <td class="small">${esc(ts)}</td>
        <td><code class="small">${esc(l.source || '-')}</code></td>
        <td>${esc(l.message)}</td>
        <td class="small" style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${attrSummary}</td>
      </tr>
      <tr id="${detailId}" class="log-detail-row" style="display:none">
        <td colspan="5">${attrsHtml}</td>
      </tr>`;
    }).join('');

    const entryCount = logs.length;
    const totalCount = (logs[0] as any)._total !== undefined ? (logs[0] as any)._total : entryCount;
    document.getElementById('logCount')!.textContent = entryCount + ' entries' + (totalCount > entryCount ? ` (showing last ${entryCount})` : '');

    // Auto-scroll if user was at bottom
    if (wasAtBottom && tableContainer) {
      tableContainer.scrollTop = tableContainer.scrollHeight;
    }
  } catch (e: any) {
    document.getElementById('logBody')!.innerHTML = '<tr><td colspan="5" class="text-danger text-center">Failed to load logs</td></tr>';
  }
}

function toggleLogDetail(detailId: string) {
  const detailRow = document.getElementById(detailId);
  if (!detailRow) return;
  const isVisible = detailRow.style.display !== 'none';
  detailRow.style.display = isVisible ? 'none' : 'table-row';
}
expose('toggleLogDetail', toggleLogDetail);

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
expose('setLogLevel', setLogLevel);

function clearLogFilters() {
  const sourceFilter = document.getElementById('logSourceFilter') as HTMLSelectElement;
  if (sourceFilter) sourceFilter.value = '';
  loadLogs();
}
expose('clearLogFilters', clearLogFilters);

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

expose('showAddUser', function () {
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
});

expose('saveNewUser', async function () {
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
});

expose('editUser', function (id: number, username: string, display: string, email: string, role: string) {
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
});

expose('saveEditUser', async function (id: number) {
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
});

expose('deleteUser', async function (id: number) {
  if (!confirm('Delete this user?')) return;
  try {
    await api('DELETE', `/api/admin/users/${id}`);
    loadUsers();
    toast('User deleted');
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('resetPass', function (id: number) {
  showModal('Reset Password', `
    <div class="mb-3"><label class="form-label">New Password</label><input class="form-control" type="password" id="resetPass"></div>
    <button class="btn btn-primary w-100" onclick="doResetPass(${id})">Reset</button>
  `);
});

expose('doResetPass', async function (id: number) {
  try {
    await api('PUT', `/api/admin/users/${id}/password`, {
      password: (document.getElementById('resetPass') as HTMLInputElement).value,
    });
    hideModal();
    toast('Password reset');
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── Schemas ───

let schemaEditId: number | null = null;

expose('showAddSchema', function () {
  schemaEditId = null;
  showModal('New Compendium Schema', getSchemaFormHtml(null));
});

expose('editSchema', async function (id: number) {
  schemaEditId = id;
  try {
    const s = await api('GET', `/api/admin/compendium-schemas/${id}`);
    showModal('Edit Schema: ' + esc(s.display_name), getSchemaFormHtml(s));
  } catch (e: any) {
    toast(e.message, true);
  }
});

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

expose('addSchemaField', function () {
  const container = document.getElementById('schemaFields')!;
  const index = container.children.length;
  container.insertAdjacentHTML('beforeend', getSchemaFieldHtml({ key: '', label: '', type: 'text', required: false }, index));
});

expose('removeSchemaField', function (index: number) {
  const el = document.getElementById('sf-' + index);
  if (el) el.remove();
});

expose('saveSchema', async function () {
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
    loadUnifiedCompendium();
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('deleteSchema', async function (id: number) {
  if (!confirm('Delete this schema? This cannot be undone.')) return;
  try {
    await api('DELETE', `/api/admin/compendium-schemas/${id}`);
    loadUnifiedCompendium();
    toast('Schema deleted');
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── Backup ───

async function loadBackupSettings() {
  try {
    const settings = await api('GET', '/api/backup/settings');
    (document.getElementById('backupEnabled') as HTMLInputElement).checked = settings.enabled;
    (document.getElementById('backupInterval') as HTMLInputElement).value = settings.interval_days || 7;
    (document.getElementById('backupKeepCount') as HTMLInputElement).value = settings.keep_count || 7;
  } catch {}
}

expose('saveBackupSettings', async function () {
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
});

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

expose('triggerBackup', async function () {
  try {
    const result = await api('POST', '/api/backup/trigger');
    toast('Backup created: ' + result.path);
    loadBackupList();
  } catch (e: any) {
    toast(e.message, true);
  }
});

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

expose('saveEmailSettings', async function () {
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
});

expose('testEmailSettings', async function () {
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
});

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

expose('saveUmamiSettings', async function () {
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
});

// ─── OpenTelemetry ───

async function loadOTelSettings() {
  try {
    const s = await api('GET', '/api/admin/otel-settings');
    (document.getElementById('otelEnabled') as HTMLInputElement).checked = s.enabled;
    (document.getElementById('otelEndpoint') as HTMLInputElement).value = s.endpoint || '';
  } catch {}
}

expose('saveOTelSettings', async function () {
  try {
    await api('POST', '/api/admin/otel-settings', {
      enabled: (document.getElementById('otelEnabled') as HTMLInputElement).checked,
      endpoint: (document.getElementById('otelEndpoint') as HTMLInputElement).value,
    });
    toast('Telemetry settings saved');
  } catch (e: any) {
    toast(e.message, true);
  }
});

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
expose('loadAIEndpoints', loadAIEndpoints);

function toggleAIEndpointFields() {
  const type = (document.getElementById('aiEndpointType') as HTMLSelectElement).value;
  document.getElementById('aiEndpointTextFields')!.style.display = type === 'text' ? 'flex' : 'none';
  document.getElementById('aiEndpointImageSizeField')!.style.display = type === 'image' ? 'block' : 'none';
}
expose('toggleAIEndpointFields', toggleAIEndpointFields);

expose('showAddAIEndpoint', function () {
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
});

expose('editAIEndpoint', async function (id: number) {
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
});

expose('saveAIEndpoint', async function () {
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
});

expose('deleteAIEndpoint', async function (id: number) {
  if (!confirm('Delete this AI endpoint? This cannot be undone.')) return;
  try {
    await api('DELETE', `/api/admin/ai-endpoints/${id}`);
    loadAIEndpoints();
    toast('Endpoint deleted');
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('testAIEndpoint', async function (id: number) {
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
});

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

expose('rollbackImport', async function (id: number) {
  if (!confirm('Roll back this import? This will delete imported entries. Cannot be undone.')) return;
  try {
    await api('POST', `/api/admin/compendium-import-logs/${id}/rollback`);
    toast('Import rolled back');
    loadImportLogs();
  } catch (e: any) { toast(e.message, true); }
});

expose('onImportSchemaChange', function () {
  // If data is already loaded, re-run auto-detection instead of hiding everything
  if (importJsonData && importJsonData.records.length > 0) {
    autoDetectMapping();
    return;
  }
  document.getElementById('importPreview')!.style.display = 'none';
  document.getElementById('importMapping')!.style.display = 'none';
  (document.getElementById('importStartBtn') as HTMLButtonElement).disabled = true;
});

expose('showImportPaste', function () {
  document.getElementById('importPasteArea')!.style.display = 'block';
  document.getElementById('importFetchArea')!.style.display = 'none';
});

expose('showImportFetch', function () {
  document.getElementById('importFetchArea')!.style.display = 'block';
  document.getElementById('importPasteArea')!.style.display = 'none';
});

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
expose('handleImportFile', handleImportFile);

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
expose('useImportPaste', useImportPaste);

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
expose('fetchImportUrl', fetchImportUrl);

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
  (document.getElementById('detectSchemaBtn') as HTMLButtonElement).disabled = false;
  (document.getElementById('importPreviewPlanBtn') as HTMLButtonElement).disabled = false;
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
expose('autoDetectMapping', autoDetectMapping);

// Auto-detect the best-matching schema for the loaded JSON via server-side
// field-overlap scoring across all known schemas (compendium-overhaul 4.3).
expose('autoDetectSchema', async function () {
  if (!importJsonData || !importJsonData.records.length) { toast('No data loaded', true); return; }
  try {
    const res = await api('POST', '/api/admin/compendium/import/detect-schema', importJsonData.records);
    const matches = res.matches || [];
    if (!matches.length) { toast('No schema matched the loaded fields', true); return; }
    const best = matches[0];
    const sel = document.getElementById('importSchema') as HTMLSelectElement;
    sel.value = String(best.schema_id);
    if (!sel.value) {
      // schema select not yet populated — reload options first
      const schemas = await api('GET', '/api/admin/compendium-schemas');
      sel.innerHTML = '<option value="">— Select Schema —</option>' + schemas.map((s: any) => `<option value="${s.id}">${esc(s.display_name)} (${esc(s.type_name)})</option>`).join('');
      sel.value = String(best.schema_id);
    }
    // onImportSchemaChange is window-exposed (anonymous function) — call via window
    (window as any).onImportSchemaChange?.();
    const confBadge = best.confidence === 'high' ? 'bg-success' : best.confidence === 'medium' ? 'bg-warning text-dark' : 'bg-secondary';
    toast(`Detected schema: ${best.display_name} (${best.confidence} match, ${best.matched_fields.length} fields)`);
    // Show the detection result inline
    const results = document.getElementById('importResults')!;
    results.innerHTML = `<div class="alert alert-info mb-0 py-2">
      <strong>Schema detected:</strong> ${esc(best.display_name)} (${esc(best.type_name)})
      <span class="badge ${confBadge} ms-1">${esc(best.confidence)}</span>
      <span class="text-muted ms-2">${best.matched_fields.length}/${best.matched_fields.length + best.unmatched_schema_fields.length} fields matched</span>
    </div>`;
  } catch (e: any) { toast(e.message, true); }
});

// Dry-run preview: compute the import plan without writing anything (4.9).
expose('previewImportPlan', async function () {
  const schemaId = parseInt((document.getElementById('importSchema') as HTMLSelectElement).value, 10);
  if (!schemaId) { toast('Select a schema', true); return; }
  if (!importJsonData || !importJsonData.records.length) { toast('No data to import', true); return; }
  const dedup = (document.getElementById('importDedup') as HTMLSelectElement).value;
  const mapping = importMapping.filter(m => m.schemaField).map(m => ({
    source_field: m.jsonField,
    schema_field: m.schemaField
  }));
  try {
    const res = await api('POST', '/api/admin/compendium-import?dry_run=true', {
      schema_id: schemaId,
      entries: importJsonData.records,
      dedup_action: dedup,
      field_mapping: mapping,
      filename: importJsonData.filename || 'import.json'
    });
    const results = document.getElementById('importResults')!;
    results.innerHTML = `<div class="alert alert-secondary mb-0 py-2">
      <strong>Dry-run plan</strong> (nothing imported yet) — create: <span class="text-success fw-bold">${res.would_create ?? 0}</span>,
      update: <span class="text-warning fw-bold">${res.would_update ?? 0}</span>,
      skip: <span class="text-muted fw-bold">${res.would_skip ?? 0}</span>,
      validation errors: <span class="text-danger fw-bold">${res.validation_errors ?? 0}</span>
      of ${res.total ?? importJsonData.records.length} records
    </div>`;
    (document.getElementById('importStartBtn') as HTMLButtonElement).disabled = false;
  } catch (e: any) { toast(e.message, true); }
});

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

expose('updateMapping', function (idx: number, schemaField: string) {
  importMapping[idx].schemaField = schemaField;
});

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
expose('startImport', startImport);

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
  (document.getElementById('detectSchemaBtn') as HTMLButtonElement).disabled = true;
  (document.getElementById('importPreviewPlanBtn') as HTMLButtonElement).disabled = true;
}
expose('resetImportForm', resetImportForm);

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

expose('logout', async function () {
  await api('POST', '/api/logout');
  window.location.href = '/login';
});

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

expose('openPdfViewer', function (url: string, title?: string) {
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
});

expose('pdfViewerPrevPage', function () {
  if (!pdfViewerDoc || pdfViewerPage <= 1) return;
  pdfViewerPage--;
  pdfViewerRenderPage(pdfViewerPage);
});

expose('pdfViewerNextPage', function () {
  if (!pdfViewerDoc || pdfViewerPage >= pdfViewerDoc.numPages) return;
  pdfViewerPage++;
  pdfViewerRenderPage(pdfViewerPage);
});

expose('pdfViewerZoomIn', function () {
  pdfViewerScale = Math.min(pdfViewerScale * 1.25, 5);
  if (pdfViewerDoc) pdfViewerRenderPage(pdfViewerPage);
});

expose('pdfViewerZoomOut', function () {
  pdfViewerScale = Math.max(pdfViewerScale / 1.25, 0.25);
  if (pdfViewerDoc) pdfViewerRenderPage(pdfViewerPage);
});

expose('pdfViewerFitToWidth', function () {
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
});

expose('pdfViewerDownload', function () {
  if (pdfViewerUrl) {
    const a = document.createElement('a');
    a.href = pdfViewerUrl;
    a.download = pdfViewerTitle.replace(/[^a-zA-Z0-9._-]/g, '_') + '.pdf';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
  }
});

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

// ─── Events Settings (Global) ───

expose('toggleEventSourceFields', function () {
  const isIcal = (document.getElementById('sourceTypeIcal') as HTMLInputElement).checked;
  (document.getElementById('eventsIcalUrlField') as HTMLElement).style.display = isIcal ? '' : 'none';
  (document.getElementById('eventsGoogleFields') as HTMLElement).style.display = isIcal ? 'none' : '';
  (document.getElementById('eventsGoogleOnlyFields') as HTMLElement).style.display = isIcal ? 'none' : '';
});

expose('toggleEventsAuthFields', function () {
  const sa = (document.getElementById('authMethodServiceAccount') as HTMLInputElement).checked;
  (document.getElementById('eventsServiceAccountFields') as HTMLElement).style.display = sa ? '' : 'none';
  (document.getElementById('eventsOAuthFields') as HTMLElement).style.display = sa ? 'none' : '';
});

async function loadEventsSettings() {
  try {
    const s = await api('GET', '/api/admin/events-settings');
    (document.getElementById('eventsCalendarId') as HTMLInputElement).value = s.calendar_id || '';
    (document.getElementById('eventsTags') as HTMLInputElement).value = s.tags || '';
    (document.getElementById('eventsCacheTTL') as HTMLInputElement).value = s.cache_ttl_seconds || 300;
    (document.getElementById('eventsColorLabels') as HTMLInputElement).value = s.color_labels || '';
    (document.getElementById('eventsIcalUrl') as HTMLInputElement).value = s.ical_url || '';
    // Source type radio
    const sourceType = s.source_type || 'google_api';
    if (sourceType === 'ical') {
      (document.getElementById('sourceTypeIcal') as HTMLInputElement).checked = true;
    } else {
      (document.getElementById('sourceTypeGoogleApi') as HTMLInputElement).checked = true;
    }
    // Filter mode radio
    const filterMode = s.filter_mode || 'text';
    const fmRadio = document.getElementById('filterMode' + filterMode.charAt(0).toUpperCase() + filterMode.slice(1)) as HTMLInputElement;
    if (fmRadio) fmRadio.checked = true;
    // Auth method radio
    const authMethod = s.auth_method || 'service_account';
    if (authMethod === 'oauth') {
      (document.getElementById('authMethodOAuth') as HTMLInputElement).checked = true;
    } else {
      (document.getElementById('authMethodServiceAccount') as HTMLInputElement).checked = true;
    }
    (document.getElementById('eventsCredentialsJson') as HTMLTextAreaElement).value = s.credentials_json || '';
    (document.getElementById('eventsOAuthClientId') as HTMLInputElement).value = s.oauth_client_id || '';
    (document.getElementById('eventsOAuthClientSecret') as HTMLInputElement).value = s.oauth_client_secret || '';
    (document.getElementById('eventsOAuthRefreshToken') as HTMLInputElement).value = s.oauth_refresh_token || '';
    // Toggle field visibility
    (window as any).toggleEventSourceFields();
    (window as any).toggleEventsAuthFields();
  } catch (e: any) {
    toast(e.message, true);
  }
}
expose('loadEventsSettings', loadEventsSettings);

expose('saveEventsSettings', async function () {
  const sourceTypeEl = document.querySelector('input[name="eventsSourceType"]:checked') as HTMLInputElement;
  const filterModeEl = document.querySelector('input[name="eventsFilterMode"]:checked') as HTMLInputElement;
  const authMethodEl = document.querySelector('input[name="eventsAuthMethod"]:checked') as HTMLInputElement;
  const body: any = {
    source_type: sourceTypeEl ? sourceTypeEl.value : 'google_api',
    ical_url: (document.getElementById('eventsIcalUrl') as HTMLInputElement).value.trim(),
    calendar_id: (document.getElementById('eventsCalendarId') as HTMLInputElement).value.trim(),
    tags: (document.getElementById('eventsTags') as HTMLInputElement).value.trim(),
    cache_ttl_seconds: parseInt((document.getElementById('eventsCacheTTL') as HTMLInputElement).value) || 300,
    color_labels: (document.getElementById('eventsColorLabels') as HTMLInputElement).value.trim(),
    filter_mode: filterModeEl ? filterModeEl.value : 'text',
    auth_method: authMethodEl ? authMethodEl.value : 'service_account',
    credentials_json: (document.getElementById('eventsCredentialsJson') as HTMLTextAreaElement).value.trim(),
    oauth_client_id: (document.getElementById('eventsOAuthClientId') as HTMLInputElement).value.trim(),
    oauth_client_secret: (document.getElementById('eventsOAuthClientSecret') as HTMLInputElement).value.trim(),
    oauth_refresh_token: (document.getElementById('eventsOAuthRefreshToken') as HTMLInputElement).value.trim(),
  };
  try {
    await api('PUT', '/api/admin/events-settings', body);
    toast('Events settings saved');
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('clearEventsCache', async function () {
  try {
    await api('POST', '/api/admin/events-cache/clear');
    toast('Events cache cleared');
  } catch (e: any) {
    toast(e.message, true);
  }
});

async function loadEventsPublicLink() {
  try {
    const res = await api('GET', '/api/admin/events/public-link');
    const input = document.getElementById('eventsPublicLink') as HTMLInputElement;
    if (input) input.value = res.url || '';
    const img = document.getElementById('eventsQRImg') as HTMLImageElement;
    if (img) img.src = '/api/admin/events/qr';
  } catch (e: any) {
    toast(e.message, true);
  }
}
expose('loadEventsPublicLink', loadEventsPublicLink);

expose('copyPublicLink', async function () {
  const input = document.getElementById('eventsPublicLink') as HTMLInputElement;
  try {
    await navigator.clipboard.writeText(input.value);
    const label = document.getElementById('copyLinkLabel');
    if (label) label.textContent = 'Copied!';
    setTimeout(() => { const l = document.getElementById('copyLinkLabel'); if (l) l.textContent = 'Copy link'; }, 2000);
  } catch {
    input.select();
    document.execCommand('copy');
  }
});

expose('openEventsPage', function () {
  const input = document.getElementById('eventsPublicLink') as HTMLInputElement;
  window.open(input.value, '_blank');
});

expose('downloadQR', function () {
  const img = document.getElementById('eventsQRImg') as HTMLImageElement;
  const a = document.createElement('a');
  a.href = img.src;
  a.download = 'events-qr.png';
  document.body.appendChild(a);
  a.click();
  a.remove();
});

// ─── Campaign Event Share (link + QR) ───

expose('shareCampaignEvent', async function (id: number, slug: string) {
  try {
    const res = await api('GET', '/api/admin/events/public-link?slug=' + encodeURIComponent(slug));
    const url = res.url || '';
    const qrSrc = '/api/admin/events/qr?slug=' + encodeURIComponent(slug);
    showModal('Share Event Page', `
      <p class="text-muted">Share this campaign's public events page with players.</p>
      <div class="mb-3">
        <label class="form-label">Public URL</label>
        <div class="input-group">
          <input type="text" id="campaignShareUrl" class="form-control font-monospace" readonly value="${esc(url)}" style="font-size:0.85rem">
          <button class="btn btn-outline-primary" id="campaignCopyLinkBtn" title="Copy to clipboard"><i class="fa-solid fa-copy me-1"></i><span id="campaignCopyLinkLabel">Copy link</span></button>
        </div>
      </div>
      <div class="text-center">
        <div class="border rounded p-2 d-inline-block bg-light">
          <img id="campaignShareQRImg" src="${qrSrc}" alt="QR Code" width="160" height="160" style="image-rendering:pixelated">
        </div>
        <div class="mt-2">
          <button class="btn btn-outline-secondary" id="campaignDownloadQRBtn"><i class="fa-solid fa-download me-1"></i>Download QR</button>
        </div>
      </div>
    `);
    document.getElementById('campaignCopyLinkBtn')!.addEventListener('click', async () => {
      const input = document.getElementById('campaignShareUrl') as HTMLInputElement;
      try {
        await navigator.clipboard.writeText(input.value);
        const label = document.getElementById('campaignCopyLinkLabel');
        if (label) label.textContent = 'Copied!';
        setTimeout(() => { const l = document.getElementById('campaignCopyLinkLabel'); if (l) l.textContent = 'Copy link'; }, 2000);
      } catch {
        input.select();
        document.execCommand('copy');
      }
    });
    document.getElementById('campaignDownloadQRBtn')!.addEventListener('click', () => {
      const img = document.getElementById('campaignShareQRImg') as HTMLImageElement;
      const a = document.createElement('a');
      a.href = img.src;
      a.download = 'events-qr-' + slug + '.png';
      document.body.appendChild(a);
      a.click();
      a.remove();
    });
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── Campaign Event Settings CRUD ───

let campaignEventEditId: number | null = null;

async function loadCampaignEventSettings() {
  try {
    const campaigns = await api('GET', '/api/admin/events-campaigns');
    const tbody = document.getElementById('campaignEventBody')!;
    if (!campaigns || campaigns.length === 0) {
      tbody.innerHTML = '<tr><td colspan="9" class="text-muted text-center py-3">No campaign event pages configured. Add one to create a public event page for a campaign.</td></tr>';
      return;
    }
    tbody.innerHTML = campaigns.map((c: any) => {
      const filterParts: string[] = [];
      if (c.filter_mode && c.filter_mode !== 'text') filterParts.push(c.filter_mode);
      if (c.color_labels) filterParts.push('color:' + c.color_labels);
      const sourceType = c.source_type || 'google_api';
      const sourceLabel = sourceType === 'ical' ? 'iCal' : 'GCal API';
      const sourceDetail = sourceType === 'ical' ? (c.ical_url || '(global)') : (c.calendar_id || '(global)');
      return `
      <tr>
        <td><strong>${esc(c.display_name)}</strong></td>
        <td><code>${esc(c.slug)}</code></td>
        <td style="max-width:180px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="${esc(sourceDetail)}"><small class="text-muted">${sourceLabel}</small> ${esc(sourceDetail)}</td>
        <td style="max-width:120px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="${esc(c.tags || '')}">${esc(c.tags || '(global)')}</td>
        <td><small class="text-muted">${filterParts.length ? esc(filterParts.join(', ')) : 'text'}</small></td>
        <td><small class="text-muted">${sourceType === 'ical' ? '—' : (c.auth_method === 'oauth' ? 'OAuth' : 'Service Acct')}</small></td>
        <td>${c.is_active ? '<span class="text-success"><i class="fa-solid fa-check"></i></span>' : '<span class="text-muted"><i class="fa-solid fa-xmark"></i></span>'}</td>
        <td class="text-nowrap">
          <button class="btn btn-outline-success btn-sm py-0" onclick="shareCampaignEvent(${c.id}, '${esc(c.slug)}')" title="Share link & QR"><i class="fa-solid fa-share-nodes"></i></button>
          <a href="/events/c/${esc(c.slug)}" class="btn btn-outline-info btn-sm py-0" target="_blank" title="View public page"><i class="fa-solid fa-eye"></i></a>
        </td>
        <td class="text-nowrap">
          <button class="btn btn-outline-warning btn-sm py-0" onclick="clearCampaignCache(${c.id})" title="Clear cache for this campaign page"><i class="fa-solid fa-eraser"></i></button>
          <button class="btn btn-outline-primary btn-sm py-0" onclick="editCampaignEventSetting(${c.id})" title="Edit"><i class="fa-solid fa-pen"></i></button>
          <button class="btn btn-outline-danger btn-sm py-0" onclick="deleteCampaignEventSetting(${c.id})" title="Delete"><i class="fa-solid fa-trash"></i></button>
        </td>
      </tr>`;
    }).join('');
  } catch (e: any) {
    const tbody = document.getElementById('campaignEventBody');
    if (tbody) tbody.innerHTML = '<tr><td colspan="9" class="text-danger text-center">Failed to load: ' + esc(e.message) + '</td></tr>';
  }
}
expose('loadCampaignEventSettings', loadCampaignEventSettings);

expose('toggleCampaignAuthFields', function () {
  const sa = (document.getElementById('campaignAuthMethodServiceAccount') as HTMLInputElement).checked;
  (document.getElementById('campaignServiceAccountFields') as HTMLElement).style.display = sa ? '' : 'none';
  (document.getElementById('campaignOAuthFields') as HTMLElement).style.display = sa ? 'none' : '';
});

expose('toggleCampaignEventSourceFields', function () {
  const isIcal = (document.getElementById('campaignSourceTypeIcal') as HTMLInputElement).checked;
  (document.getElementById('campaignIcalUrlField') as HTMLElement).style.display = isIcal ? '' : 'none';
  (document.getElementById('campaignGoogleFields') as HTMLElement).style.display = isIcal ? 'none' : '';
  (document.getElementById('campaignGoogleOnlyFields') as HTMLElement).style.display = isIcal ? 'none' : '';
});

expose('showAddCampaignEvent', function () {
  campaignEventEditId = null;
  campaignSlugAuto = true;
  document.getElementById('campaignEventModalTitle')!.textContent = 'Add Campaign Event Page';
  (document.getElementById('campaignEventId') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventDisplayName') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventSlug') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventCalendarId') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventTags') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventCacheTTL') as HTMLInputElement).value = '300';
  (document.getElementById('campaignEventColorLabels') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventIcalUrl') as HTMLInputElement).value = '';
  (document.getElementById('campaignSourceTypeGoogleApi') as HTMLInputElement).checked = true;
  (document.getElementById('campaignFilterModeText') as HTMLInputElement).checked = true;
  (document.getElementById('campaignAuthMethodServiceAccount') as HTMLInputElement).checked = true;
  (document.getElementById('campaignEventCredentialsJson') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventOAuthClientId') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventOAuthClientSecret') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventOAuthRefreshToken') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventIsActive') as HTMLInputElement).checked = true;
  updateCampaignSlugPreview();
  // Reset field visibility
  (document.getElementById('campaignServiceAccountFields') as HTMLElement).style.display = 'none';
  (document.getElementById('campaignOAuthFields') as HTMLElement).style.display = 'none';
  (window as any).toggleCampaignEventSourceFields();
  new (window as any).bootstrap.Modal(document.getElementById('campaignEventModal')!).show();
});

expose('editCampaignEventSetting', async function (id: number) {
  campaignEventEditId = id;
  campaignSlugAuto = false; // slug is user-controlled in edit mode
  try {
    const c = await api('GET', '/api/admin/events-campaigns/' + id);
    document.getElementById('campaignEventModalTitle')!.textContent = 'Edit Campaign Event Page';
    (document.getElementById('campaignEventId') as HTMLInputElement).value = String(id);
    (document.getElementById('campaignEventDisplayName') as HTMLInputElement).value = c.display_name || '';
    (document.getElementById('campaignEventSlug') as HTMLInputElement).value = c.slug || '';
    (document.getElementById('campaignEventCalendarId') as HTMLInputElement).value = c.calendar_id || '';
    (document.getElementById('campaignEventTags') as HTMLInputElement).value = c.tags || '';
    (document.getElementById('campaignEventCacheTTL') as HTMLInputElement).value = c.cache_ttl_seconds || 300;
    (document.getElementById('campaignEventColorLabels') as HTMLInputElement).value = c.color_labels || '';
    (document.getElementById('campaignEventIcalUrl') as HTMLInputElement).value = c.ical_url || '';
    // Set source type radio
    const sourceType = c.source_type || 'google_api';
    if (sourceType === 'ical') {
      (document.getElementById('campaignSourceTypeIcal') as HTMLInputElement).checked = true;
    } else {
      (document.getElementById('campaignSourceTypeGoogleApi') as HTMLInputElement).checked = true;
    }
    // Set filter mode radio
    const filterMode = c.filter_mode || 'text';
    const fmRadio = document.getElementById('campaignFilterMode' + filterMode.charAt(0).toUpperCase() + filterMode.slice(1)) as HTMLInputElement;
    if (fmRadio) fmRadio.checked = true;
    // Set auth method radio
    const authMethod = c.auth_method || 'service_account';
    if (authMethod === 'oauth') {
      (document.getElementById('campaignAuthMethodOAuth') as HTMLInputElement).checked = true;
    } else {
      (document.getElementById('campaignAuthMethodServiceAccount') as HTMLInputElement).checked = true;
    }
    (document.getElementById('campaignEventCredentialsJson') as HTMLInputElement).value = c.credentials_json || '';
    (document.getElementById('campaignEventOAuthClientId') as HTMLInputElement).value = c.oauth_client_id || '';
    (document.getElementById('campaignEventOAuthClientSecret') as HTMLInputElement).value = c.oauth_client_secret || '';
    (document.getElementById('campaignEventOAuthRefreshToken') as HTMLInputElement).value = c.oauth_refresh_token || '';
    (document.getElementById('campaignEventIsActive') as HTMLInputElement).checked = c.is_active;
    updateCampaignSlugPreview();
    // Show correct field visibility
    (window as any).toggleCampaignEventSourceFields();
    (window as any).toggleCampaignAuthFields();
    new (window as any).bootstrap.Modal(document.getElementById('campaignEventModal')!).show();
  } catch (e: any) {
    toast(e.message, true);
  }
});

function updateCampaignSlugPreview() {
  const slug = (document.getElementById('campaignEventSlug') as HTMLInputElement).value.trim() || 'your-slug';
  const preview = document.getElementById('campaignSlugPreview');
  if (preview) preview.textContent = slug;
}
function slugifyCampaignName(name: string): string {
  return name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '');
}
let campaignSlugAuto = false; // true while slug is auto-generated from display name
document.addEventListener('input', function (e: Event) {
  const el = e.target as HTMLElement;
  if (el && el.id === 'campaignEventSlug') {
    campaignSlugAuto = false;
    updateCampaignSlugPreview();
  }
  if (el && el.id === 'campaignEventDisplayName') {
    if (campaignSlugAuto) {
      (document.getElementById('campaignEventSlug') as HTMLInputElement).value = slugifyCampaignName((el as HTMLInputElement).value);
      updateCampaignSlugPreview();
    }
  }
});

expose('saveCampaignEventSetting', async function () {
  const displayName = (document.getElementById('campaignEventDisplayName') as HTMLInputElement).value.trim();
  const slug = (document.getElementById('campaignEventSlug') as HTMLInputElement).value.trim();
  if (!displayName || !slug) { toast('Display name and slug are required', true); return; }
  const sourceTypeEl = document.querySelector('input[name="campaignSourceType"]:checked') as HTMLInputElement;
  const filterModeEl = document.querySelector('input[name="campaignFilterMode"]:checked') as HTMLInputElement;
  const authMethodEl = document.querySelector('input[name="campaignAuthMethod"]:checked') as HTMLInputElement;
  const body: any = {
    display_name: displayName,
    slug: slug,
    source_type: sourceTypeEl ? sourceTypeEl.value : 'google_api',
    ical_url: (document.getElementById('campaignEventIcalUrl') as HTMLInputElement).value.trim(),
    calendar_id: (document.getElementById('campaignEventCalendarId') as HTMLInputElement).value.trim(),
    tags: (document.getElementById('campaignEventTags') as HTMLInputElement).value.trim(),
    color_labels: (document.getElementById('campaignEventColorLabels') as HTMLInputElement).value.trim(),
    filter_mode: filterModeEl ? filterModeEl.value : 'text',
    auth_method: authMethodEl ? authMethodEl.value : 'service_account',
    credentials_json: (document.getElementById('campaignEventCredentialsJson') as HTMLInputElement).value.trim(),
    oauth_client_id: (document.getElementById('campaignEventOAuthClientId') as HTMLInputElement).value.trim(),
    oauth_client_secret: (document.getElementById('campaignEventOAuthClientSecret') as HTMLInputElement).value.trim(),
    oauth_refresh_token: (document.getElementById('campaignEventOAuthRefreshToken') as HTMLInputElement).value.trim(),
    cache_ttl_seconds: parseInt((document.getElementById('campaignEventCacheTTL') as HTMLInputElement).value) || 300,
    is_active: (document.getElementById('campaignEventIsActive') as HTMLInputElement).checked
  };
  try {
    if (campaignEventEditId) {
      body.id = campaignEventEditId;
      await api('PUT', '/api/admin/events-campaigns/' + campaignEventEditId, body);
      toast('Campaign page updated');
    } else {
      await api('POST', '/api/admin/events-campaigns', body);
      toast('Campaign page created');
    }
    const modalEl = document.getElementById('campaignEventModal')!;
    const modal = (window as any).bootstrap.Modal.getInstance(modalEl);
    if (modal) modal.hide();
    loadCampaignEventSettings();
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('deleteCampaignEventSetting', async function (id: number) {
  if (!confirm('Delete this campaign event page? The public page will no longer be available.')) return;
  try {
    await api('DELETE', '/api/admin/events-campaigns/' + id);
    toast('Campaign page deleted');
    loadCampaignEventSettings();
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('clearCampaignCache', async function (id: number) {
  try {
    await api('POST', '/api/admin/events-campaigns/' + id + '/clear-cache');
    toast('Campaign cache cleared');
    loadCampaignEventSettings();
  } catch (e: any) {
    toast(e.message, true);
  }
});

init();

})();
