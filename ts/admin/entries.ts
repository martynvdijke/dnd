// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, toast, showModal, hideModal } from '../lib/dom';
import { api, currentSchemaId, entryModalSchemaId, entryModalSchemaFields, entryModalEditId, selectedEntryIds, setEntryModalSchemaId, setEntryModalSchemaFields, setEntryModalEditId } from './state';
import { renderError } from '../lib/errors';

function createSchemaEntry(schemaId: number) {
  api('GET', '/api/admin/compendium-schemas').then((schemas: any[]) => {
    const schema = schemas.find((s: any) => s.id === schemaId);
    if (!schema) { toast('Schema not found', true); return; }
    setEntryModalSchemaId(schemaId);
    setEntryModalSchemaFields(schema.fields || []);
    setEntryModalEditId(null);
    showModal('Create Entry in ' + esc(schema.display_name), getEntryFormHtml(null));
  }).catch((e: any) => renderError(e));
}
expose('createSchemaEntry', createSchemaEntry);
function editSchemaEntryById(entryId: number, schemaId: number) {
  Promise.all([
    api('GET', '/api/admin/compendium-schemas'),
    api('GET', '/api/admin/compendium-entries/' + entryId)
  ]).then(([schemas, entry]) => {
    const schema = schemas.find((s: any) => s.id === schemaId);
    if (!schema) { toast('Schema not found', true); return; }
    setEntryModalSchemaId(schemaId);
    setEntryModalSchemaFields(schema.fields || []);
    setEntryModalEditId(entryId);
    showModal('Edit Entry', getEntryFormHtml(entry.data || entry));
  }).catch((e: any) => renderError(e));
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
  }).catch((e: any) => renderError(e));
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
    (window as any).loadSchemaEntries?.();
  } catch (e: any) {
    renderError(e);
  }
}
expose('duplicateSchemaEntry', duplicateSchemaEntry);
function deleteSchemaEntryById(entryId: number) {
  if (!confirm('Delete this entry?')) return;
  api('DELETE', '/api/admin/compendium-entries/' + entryId).then(() => {
    toast('Entry deleted');
    selectedEntryIds.delete(entryId);
    (window as any).loadSchemaEntries?.();
  }).catch((e: any) => renderError(e));
}
expose('deleteSchemaEntryById', deleteSchemaEntryById);
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
          try { val = JSON.parse(el.value); } catch { el.classList.add('is-invalid'); valid = false; continue; }
        } else { val = null; }
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
    (window as any).loadSchemaEntries?.();
  } catch (e: any) {
    renderError(e);
  }
});
function openImportForSchema() {
  if (currentSchemaId) openImportForSchemaId(currentSchemaId);
}
expose('openImportForSchema', openImportForSchema);
function openImportForSchemaId(schemaId: number) {
  (window as any).showAdminTab('import');
  const sel = document.getElementById('importSchema') as HTMLSelectElement;
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
