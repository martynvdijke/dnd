// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, toast } from '../lib/dom';
import { api, importJsonData, importMapping, setImportJsonData, setImportMapping } from './state';
import { renderError } from '../lib/errors';
async function loadImportSchemas() {
  try {
    const schemas = await api('GET', '/api/admin/compendium-schemas');
    const sel = document.getElementById('importSchema') as HTMLSelectElement;
    sel.innerHTML = '<option value="">— Select Schema —</option>' + schemas.map((s: any) => `<option value="${s.id}">${esc(s.display_name)} (${esc(s.type_name)})</option>`).join('');
  } catch (e: any) { renderError(e); }
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
  } catch (e: any) { renderError(e); }
}
expose('loadImportSchemas', loadImportSchemas);
expose('loadImportLogs', loadImportLogs);
expose('rollbackImport', async function (id: number) {
  if (!confirm('Roll back this import? This will delete imported entries. Cannot be undone.')) return;
  try {
    await api('POST', `/api/admin/compendium-import-logs/${id}/rollback`);
    toast('Import rolled back');
    loadImportLogs();
  } catch (e: any) { renderError(e); }
});
expose('onImportSchemaChange', function () {
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
      setImportJsonDataInternal(data, file.name);
    } catch (err: any) { toast('Invalid JSON file: ' + err.message, true); }
  };
  reader.readAsText(file);
}
expose('handleImportFile', handleImportFile);
function useImportPaste() {
  const text = (document.getElementById('importPasteText') as HTMLTextAreaElement).value.trim();
  if (!text) { toast('Paste some JSON first', true); return; }
  try {
    const data = JSON.parse(text);
    setImportJsonDataInternal(data, 'pasted.json');
  } catch (err: any) { toast('Invalid JSON: ' + err.message, true); }
}
expose('useImportPaste', useImportPaste);
async function fetchImportUrl() {
  const url = (document.getElementById('importFetchUrl') as HTMLInputElement).value.trim();
  if (!url) { toast('Enter a URL', true); return; }
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
      if (contentType.includes('text/html')) errMsg = `URL returned HTML instead of JSON (HTTP ${res.status}). Make sure the URL points to a raw JSON file (e.g. raw.githubusercontent.com).`;
      else if (res.status === 404) errMsg = `URL not found (HTTP 404). Check that the path is correct.`;
      else if (res.status === 403) errMsg = `Access denied (HTTP 403). The server may be blocking cross-origin requests (CORS) or require authentication.`;
      else errMsg = `Server returned HTTP ${res.status}.`;
      throw new Error(errMsg);
    }
    const data = await res.json();
    setImportJsonDataInternal(data, url.split('/').pop() || 'remote.json');
  } catch (err: any) {
    if (!errMsg) {
      if (err.message?.includes('Failed to fetch') || err.message?.includes('NetworkError')) errMsg = `Cannot fetch from this URL due to CORS restrictions or network error. Try using a CORS-friendly URL like raw.githubusercontent.com, or paste the JSON content manually.`;
      else errMsg = `Fetch failed: ${err.message}`;
    }
    toast(errMsg, true);
  }
  if (btn) btn.innerHTML = '<i class="fa-solid fa-download me-1"></i> Fetch';
}
expose('fetchImportUrl', fetchImportUrl);
function setImportJsonDataInternal(data: any, filename: string) {
  const arr = Array.isArray(data) ? data : [data];
  if (!arr.length) { toast('JSON has no records', true); return; }
  setImportJsonData({ records: arr, filename });
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
  setImportMapping([]);
  (document.getElementById('importStartBtn') as HTMLButtonElement).disabled = true;
  (document.getElementById('detectSchemaBtn') as HTMLButtonElement).disabled = false;
  (document.getElementById('importPreviewPlanBtn') as HTMLButtonElement).disabled = false;
  toast('Loaded ' + arr.length + ' records from ' + filename);
}
function getNestedValue(obj: any, path: string): any { return path.split('.').reduce((o, k) => o != null ? o[k] : undefined, obj); }
function discoverKeys(obj: any, prefix = ''): string[] {
  let keys: string[] = [];
  for (const k of Object.keys(obj)) {
    const full = prefix ? prefix + '.' + k : k;
    const v = obj[k];
    if (v !== null && typeof v === 'object' && !Array.isArray(v)) keys.push(...discoverKeys(v, full));
    else keys.push(full);
  }
  return keys;
}
function autoDetectMapping() {
  const cur = (globalThis as any).__importJsonData || importJsonData;
  // use current importJsonData from state
  const data = importJsonData;
  if (!data || !data.records.length) { toast('No data loaded', true); return; }
  const schemaId = parseInt((document.getElementById('importSchema') as HTMLSelectElement).value, 10);
  if (!schemaId) { toast('Select a target schema first', true); return; }
  api('GET', '/api/admin/compendium-schemas').then((schemas: any[]) => {
    const schema = schemas.find(s => s.id === schemaId);
    if (!schema) { toast('Schema not found', true); return; }
    const schemaFields = schema.fields || [];
    const sample = data.records[0];
    const jsonKeys = discoverKeys(sample);
    const mapping = jsonKeys.map((jk: string) => {
      const lc = jk.toLowerCase();
      let match = schemaFields.find((f: any) => f.name.toLowerCase() === lc || f.label.toLowerCase() === lc);
      if (!match) match = schemaFields.find((f: any) => lc.includes(f.name.toLowerCase()) || f.name.toLowerCase().includes(lc));
      return { jsonField: jk, schemaField: match ? match.name : '', schemaLabel: match ? match.label : '(unmapped)', required: match ? match.required : false, preview: String(getNestedValue(sample, jk) ?? '').slice(0, 60) };
    });
    setImportMapping(mapping);
    renderMappingTable(schemaFields);
    document.getElementById('importMapping')!.style.display = 'block';
    (document.getElementById('importStartBtn') as HTMLButtonElement).disabled = false;
    toast('Detected ' + mapping.filter((m: any) => m.schemaField).length + ' mapped fields');
  }).catch((e: any) => renderError(e));
}
expose('autoDetectMapping', autoDetectMapping);
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
      const schemas = await api('GET', '/api/admin/compendium-schemas');
      sel.innerHTML = '<option value="">— Select Schema —</option>' + schemas.map((s: any) => `<option value="${s.id}">${esc(s.display_name)} (${esc(s.type_name)})</option>`).join('');
      sel.value = String(best.schema_id);
    }
    (window as any).onImportSchemaChange?.();
    const confBadge = best.confidence === 'high' ? 'bg-success' : best.confidence === 'medium' ? 'bg-warning text-dark' : 'bg-secondary';
    toast(`Detected schema: ${best.display_name} (${best.confidence} match, ${best.matched_fields.length} fields)`);
    const results = document.getElementById('importResults')!;
    results.innerHTML = `<div class="alert alert-info mb-0 py-2">
      <strong>Schema detected:</strong> ${esc(best.display_name)} (${esc(best.type_name)})
      <span class="badge ${confBadge} ms-1">${esc(best.confidence)}</span>
      <span class="text-muted ms-2">${best.matched_fields.length}/${best.matched_fields.length + best.unmatched_schema_fields.length} fields matched</span>
    </div>`;
  } catch (e: any) { renderError(e); }
});
expose('previewImportPlan', async function () {
  const schemaId = parseInt((document.getElementById('importSchema') as HTMLSelectElement).value, 10);
  if (!schemaId) { toast('Select a schema', true); return; }
  if (!importJsonData || !importJsonData.records.length) { toast('No data to import', true); return; }
  const dedup = (document.getElementById('importDedup') as HTMLSelectElement).value;
  const mapping = importMapping.filter(m => m.schemaField).map(m => ({ source_field: m.jsonField, schema_field: m.schemaField }));
  try {
    const res = await api('POST', '/api/admin/compendium-import?dry_run=true', { schema_id: schemaId, entries: importJsonData.records, dedup_action: dedup, field_mapping: mapping, filename: importJsonData.filename || 'import.json' });
    const results = document.getElementById('importResults')!;
    results.innerHTML = `<div class="alert alert-secondary mb-0 py-2">
      <strong>Dry-run plan</strong> (nothing imported yet) — create: <span class="text-success fw-bold">${res.would_create ?? 0}</span>,
      update: <span class="text-warning fw-bold">${res.would_update ?? 0}</span>,
      skip: <span class="text-muted fw-bold">${res.would_skip ?? 0}</span>,
      validation errors: <span class="text-danger fw-bold">${res.validation_errors ?? 0}</span>
      of ${res.total ?? importJsonData.records.length} records
    </div>`;
    (document.getElementById('importStartBtn') as HTMLButtonElement).disabled = false;
  } catch (e: any) { renderError(e); }
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
expose('updateMapping', function (idx: number, schemaField: string) { importMapping[idx].schemaField = schemaField; });
async function startImport() {
  const schemaId = parseInt((document.getElementById('importSchema') as HTMLSelectElement).value, 10);
  if (!schemaId) { toast('Select a schema', true); return; }
  if (!importJsonData || !importJsonData.records.length) { toast('No data to import', true); return; }
  const dedup = (document.getElementById('importDedup') as HTMLSelectElement).value;
  const mapping = importMapping.filter(m => m.schemaField).map(m => ({ source_field: m.jsonField, schema_field: m.schemaField }));
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
      const res = await api('POST', '/api/admin/compendium-import', { schema_id: schemaId, entries: batch, dedup_action: dedup, field_mapping: mapping, filename: importJsonData.filename || 'import.json' });
      totalImported += res.imported || 0;
      totalSkipped += res.skipped || 0;
      totalErrors += (res.errors || []).length;
    } catch (e: any) {
      totalErrors += batch.length;
      results.innerHTML += `<div class="text-danger small">Batch ${Math.floor(i / batchSize) + 1} failed: ${esc(e.message)}</div>`;
    }
  }
  bar.style.width = '100%'; pct.textContent = '100%'; text.textContent = 'Import complete';
  results.innerHTML = `<div class="alert alert-success mb-0 py-2"><strong>Done!</strong> Imported ${totalImported}, Skipped ${totalSkipped}, Errors ${totalErrors} of ${records.length} records</div>`;
  btn.disabled = false; btn.innerHTML = '<i class="fa-solid fa-play me-1"></i>Start Import';
  loadImportLogs();
}
expose('startImport', startImport);
function resetImportForm() {
  setImportJsonData(null); setImportMapping([]);
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
