// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, attrEscape, toast, showModal, hideModal } from '../lib/dom';
import { api, currentSchemaId, currentSchemaFields, currentSchemaPage, currentSchemaQuery, currentSchemaName, selectedEntryIds, setCurrentSchemaId, setCurrentSchemaFields, setCurrentSchemaPage, setCurrentSchemaQuery, setCurrentSchemaName } from './state';
import { renderError } from '../lib/errors';

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
          <button class="btn btn-outline-primary btn-sm py-0 js-browse-schema" data-schema-id="${s.id}" data-schema-name="${attrEscape(s.display_name)}" title="Browse entries"><i class="fa-solid fa-list"></i></button>
          <button class="btn btn-outline-info btn-sm py-0" onclick="editSchema(${s.id})" title="Edit schema"><i class="fa-solid fa-pen"></i></button>
          <button class="btn btn-outline-warning btn-sm py-0" onclick="openImportForSchemaId(${s.id})" title="Import"><i class="fa-solid fa-upload"></i></button>
          <button class="btn btn-outline-danger btn-sm py-0" onclick="deleteSchema(${s.id})" title="Delete"><i class="fa-solid fa-trash"></i></button>
        </td>
      </tr>`;
    }).join('');
    tbody.querySelectorAll<HTMLButtonElement>('.js-browse-schema').forEach(btn => {
      btn.addEventListener('click', () => browseSchema(Number(btn.dataset.schemaId), btn.dataset.schemaName || ''));
    });
  } catch (e: any) {
    tbody.innerHTML = '<tr><td colspan="6" class="text-danger text-center">Failed to load schemas: ' + esc(e.message) + '</td></tr>';
  }
}
expose('loadUnifiedCompendium', loadUnifiedCompendium);

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
  api('GET', '/api/admin/compendium-entries/' + entryId).then(entry => {
    if (entry && entry.schema_id) {
      (window as any).editSchemaEntryById(entryId, entry.schema_id);
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

function browseSchema(schemaId: number, schemaName: string) {
  setCurrentSchemaId(schemaId);
  setCurrentSchemaName(schemaName);
  setCurrentSchemaPage(1);
  setCurrentSchemaQuery('');
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
  setCurrentSchemaId(0);
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
    const schemas = await api('GET', '/api/admin/compendium-schemas');
    const schema = schemas.find((s: any) => s.id === schemaId);
    const fields = schema?.fields || [];
    setCurrentSchemaFields(fields);
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
  setCurrentSchemaPage(page);
  loadSchemaEntries();
}
expose('goToPage', goToPage);
let entrySearchTimeout: any = null;
function onEntrySearchInput(value: string) {
  clearTimeout(entrySearchTimeout);
  setCurrentSchemaQuery(value.trim());
  setCurrentSchemaPage(1);
  entrySearchTimeout = setTimeout(() => loadSchemaEntries(), 300);
}
expose('onEntrySearchInput', onEntrySearchInput);
function addEntryCurrentSchema() {
  if (currentSchemaId) (window as any).createSchemaEntry(currentSchemaId);
}
expose('addEntryCurrentSchema', addEntryCurrentSchema);
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
    renderError(e);
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
    renderError(e);
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
expose('loadSchemaEntries', loadSchemaEntries);
