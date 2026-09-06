// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, toast, showModal, hideModal } from '../lib/dom';
import { api, schemaEditId, setSchemaEditId } from './state';
import { renderError } from '../lib/errors';
expose('showAddSchema', function () {
  setSchemaEditId(null);
  showModal('New Compendium Schema', getSchemaFormHtml(null));
});
expose('editSchema', async function (id: number) {
  setSchemaEditId(id);
  try {
    const s = await api('GET', `/api/admin/compendium-schemas/${id}`);
    showModal('Edit Schema: ' + esc(s.display_name), getSchemaFormHtml(s));
  } catch (e: any) {
    renderError(e);
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
    (window as any).loadUnifiedCompendium?.();
  } catch (e: any) {
    renderError(e);
  }
});
expose('deleteSchema', async function (id: number) {
  if (!confirm('Delete this schema? This cannot be undone.')) return;
  try {
    await api('DELETE', `/api/admin/compendium-schemas/${id}`);
    (window as any).loadUnifiedCompendium?.();
    toast('Schema deleted');
  } catch (e: any) {
    renderError(e);
  }
});
