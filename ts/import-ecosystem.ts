import { expose } from './lib/expose';
import { esc, showModal, toast } from './lib/dom';
import { api } from './lib/api';

let lastRows: any[] = [];
let lastCounts: any = null;
let lastLogId: string | number | null = null;

function sourceOptions(kind: string): string {
  if (kind === 'compendium') {
    return '<option value="5etools">5e.tools</option><option value="foundry-pack">Foundry JSON pack</option>';
  }
  return '<option value="dndbeyond">D&D Beyond</option><option value="foundry-actor">Foundry VTT actor</option>';
}

function buildModalHtml(kind: string): string {
  const schemaGroup = kind === 'compendium'
    ? `<div class="mb-3" id="ext-schema-group"><label class="form-label">Schema</label><select class="form-select" data-testid="ext-import-schema" id="ext-import-schema"><option value="">— none —</option></select></div>`
    : `<div class="mb-3" id="ext-schema-group" style="display:none"><label class="form-label">Schema</label><select class="form-select" data-testid="ext-import-schema" id="ext-import-schema"><option value="">— none —</option></select></div>`;
  return `
    <div class="mb-3"><label class="form-label">Kind</label>
      <select class="form-select" data-testid="ext-import-kind" id="ext-import-kind" onchange="extImportKindChanged()">
        <option value="character" ${kind === 'character' ? 'selected' : ''}>character</option>
        <option value="compendium" ${kind === 'compendium' ? 'selected' : ''}>compendium</option>
      </select></div>
    <div class="mb-3"><label class="form-label">Source</label>
      <select class="form-select" data-testid="ext-import-source" id="ext-import-source">${sourceOptions(kind)}</select></div>
    <div class="mb-3"><label class="form-label">JSON payload</label>
      <textarea class="form-control" data-testid="ext-import-payload" id="ext-import-payload" rows="6" style="font-family:monospace;font-size:0.8rem" placeholder="Paste JSON"></textarea></div>
    <div class="mb-3"><label class="form-label">URL</label>
      <input class="form-control" data-testid="ext-import-url" id="ext-import-url" placeholder="https://..."></div>
    ${schemaGroup}
    <div class="mb-3"><label class="form-label">Dedup</label>
      <select class="form-select" data-testid="ext-import-dedup" id="ext-import-dedup">
        <option value="skip" selected>skip</option><option value="overwrite">overwrite</option><option value="create-new">create-new</option><option value="force">force</option>
      </select></div>
    <div class="d-flex gap-2 mb-3">
      <button class="btn btn-outline-primary flex-fill" data-testid="ext-import-preview" onclick="previewExternalImport()">Preview</button>
      <button class="btn btn-primary flex-fill" data-testid="ext-import-commit" onclick="commitExternalImport()">Import</button>
    </div>
    <div class="mb-3"><button class="btn btn-outline-secondary btn-sm" data-testid="ext-import-open-logs" onclick="showExternalImportLogs()">View logs</button></div>
    <div data-testid="ext-import-rows" id="ext-import-rows" style="min-height:20px;border:1px dashed transparent">&nbsp;</div>
  `;
}

function renderRows(res: any): void {
  const container = document.getElementById('ext-import-rows');
  if (!container) return;
  const rows: any[] = res.rows || [];
  const counts = `created=${res.created ?? 0} skipped=${res.skipped ?? 0} duplicates=${res.duplicates ?? 0}`;
  let html = `<div class="small text-muted mb-2">${esc(counts)}</div>`;
  if (!rows.length) {
    html += '<div class="small text-muted">No rows</div>';
  } else {
    html += '<table class="table table-sm"><thead><tr><th>#</th><th>Name</th><th>Status</th></tr></thead><tbody>';
    for (const r of rows) {
      html += `<tr><td>${esc(String(r.index ?? ''))}</td><td>${esc(r.name ?? '')}</td><td>${esc(r.status ?? '')}${r.error ? ' - ' + esc(r.error) : ''}</td></tr>`;
    }
    html += '</tbody></table>';
  }
  container.innerHTML = html;
}

async function populateSchemas(): Promise<void> {
  const sel = document.getElementById('ext-import-schema') as HTMLSelectElement | null;
  if (!sel) return;
  try {
    const schemas: any[] = await api('GET', '/api/admin/compendium-schemas');
    const opts = '<option value=""></option>' + schemas.map((s: any) => `<option value="${esc(String(s.id))}">${esc(s.display_name)}</option>`).join('');
    sel.innerHTML = opts;
  } catch {
    // ignore for non-admin
  }
}

function readForm(): any {
  const kind = (document.getElementById('ext-import-kind') as HTMLSelectElement)?.value || 'character';
  const source = (document.getElementById('ext-import-source') as HTMLSelectElement)?.value || '';
  const payloadRaw = (document.getElementById('ext-import-payload') as HTMLTextAreaElement)?.value?.trim() || '';
  const url = (document.getElementById('ext-import-url') as HTMLInputElement)?.value?.trim() || '';
  const dedup = (document.getElementById('ext-import-dedup') as HTMLSelectElement)?.value || 'skip';
  const schemaVal = (document.getElementById('ext-import-schema') as HTMLSelectElement)?.value || '';
  const body: any = { source, kind, dedup_action: dedup, dry_run: true };
  if (payloadRaw) {
    try { body.payload = JSON.parse(payloadRaw); } catch { body.payload = payloadRaw; }
  }
  if (url) body.url = url;
  if (schemaVal) body.schema_id = isNaN(Number(schemaVal)) ? schemaVal : Number(schemaVal);
  return body;
}

expose('extImportKindChanged', function () {
  const kind = (document.getElementById('ext-import-kind') as HTMLSelectElement)?.value || 'character';
  const sourceSel = document.getElementById('ext-import-source') as HTMLSelectElement | null;
  if (sourceSel) sourceSel.innerHTML = sourceOptions(kind);
  const group = document.getElementById('ext-schema-group');
  if (group) group.style.display = kind === 'compendium' ? 'block' : 'none';
  if (kind === 'compendium') populateSchemas();
});

expose('showExternalImport', function () {
  showModal('Import from External Source', buildModalHtml('character'));
  populateSchemas();
});

expose('previewExternalImport', async function () {
  const body = readForm();
  body.dry_run = true;
  try {
    const res = await api('POST', '/api/import/external', body);
    lastRows = res.rows || [];
    lastCounts = { created: res.created, skipped: res.skipped, duplicates: res.duplicates };
    if (res.log_id) lastLogId = res.log_id;
    renderRows(res);
  } catch (e: any) {
    toast(e.message || 'Preview failed', true);
  }
});

expose('commitExternalImport', async function () {
  const body = readForm();
  body.dry_run = false;
  try {
    const res = await api('POST', '/api/import/external', body);
    lastRows = res.rows || [];
    if (res.log_id) lastLogId = res.log_id;
    toast(`Created ${res.created ?? 0} item(s)`);
    renderRows(res);
    if (res.log_id) {
      // refresh logs silently
    }
  } catch (e: any) {
    toast(e.message || 'Import failed', true);
  }
});

expose('showExternalImportLogs', async function () {
  showModal('Import Logs', '<div id="ext-logs-body" class="small text-muted">Loading...</div>');
  try {
    const logs: any[] = await api('GET', '/api/import/external/logs');
    const body = document.getElementById('ext-logs-body');
    if (!body) return;
    if (!logs.length) {
      body.innerHTML = '<div class="text-muted">No logs</div>';
      return;
    }
    body.innerHTML = logs.map((l: any) => {
      const created = Array.isArray(l.created_ids) ? l.created_ids.length : (l.counts?.created ?? 0);
      const status = esc(l.status || '');
      const source = esc(l.source || '');
      const kind = esc(l.kind || '');
      return `<div class="d-flex justify-content-between align-items-center border-bottom py-2" data-testid="ext-import-log">
        <span>${source} ${kind} ${status} created=${created}</span>
        <button class="btn btn-sm btn-outline-danger" data-testid="ext-import-rollback" onclick="rollbackExternalImport(${l.id})">Rollback</button>
      </div>`;
    }).join('');
  } catch (e: any) {
    const body = document.getElementById('ext-logs-body');
    if (body) body.innerHTML = `<div class="text-danger">${esc(e.message)}</div>`;
  }
});

expose('rollbackExternalImport', async function (id: number) {
  if (!confirm('Rollback this import?')) return;
  try {
    const res = await api('POST', `/api/import/external/logs/${id}/rollback`);
    toast(`Rolled back ${res.rolled_back ?? 0} item(s)`);
    (window as any).showExternalImportLogs();
  } catch (e: any) {
    toast(e.message || 'Rollback failed', true);
  }
});
