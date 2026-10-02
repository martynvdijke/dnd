import { esc, hideModal, showModal, toast } from './lib/dom';
import { api } from './lib/api';
import { expose } from './lib/expose';
import { currentCampaign } from './lib/state';

interface DraftReply {
  status?: string;
  message?: string;
  draft?: unknown;
}

interface ImportReply {
  id: number;
  url: string;
  entity_type: string;
  name: string;
  updated: boolean;
  counts?: Record<string, number>;
}

interface DraftImportOptions {
  campaignId?: number | null;
  parentId?: number | null;
  entityId?: number | null;
  json?: string;
}

const ENTITY_TYPES: Array<[string, string]> = [
  ['oneshot', 'One-Shot Adventure'],
  ['npc', 'NPC'],
  ['location', 'Location'],
  ['encounter', 'Encounter'],
  ['faction', 'Faction'],
  ['campaign', 'Campaign'],
  ['quest', 'Quest'],
  ['item', 'Item'],
];

let entityType = 'oneshot';
let campaignId: number | null = null;
let parentId: number | null = null;
let entityId: number | null = null;
let busy = false;

function el<T extends HTMLElement = HTMLElement>(id: string): T | null {
  return document.getElementById(id) as T | null;
}

function selectedTypeLabel(): string {
  return ENTITY_TYPES.find(([value]) => value === entityType)?.[1] ?? entityType;
}

function setError(message: string): void {
  const box = el<HTMLDivElement>('draftImportError');
  if (!box) return;
  box.textContent = message;
  box.classList.toggle('d-none', !message);
}

function setNotice(message: string): void {
  const box = el<HTMLDivElement>('draftImportNotice');
  if (!box) return;
  box.textContent = message;
  box.classList.toggle('d-none', !message);
}

function setBusy(value: boolean): void {
  busy = value;
  for (const id of ['draftImportApplyBtn', 'draftImportReviseBtn']) {
    const btn = el<HTMLButtonElement>(id);
    if (btn) btn.disabled = value;
  }
}

function refreshList(): void {
  if (entityType === 'oneshot') (window as any).showOneShots?.();
}

function currentJson(): string {
  return el<HTMLTextAreaElement>('draftImportJson')?.value?.trim() ?? '';
}

function renderBody(): string {
  const options = ENTITY_TYPES.map(([value, label]) =>
    `<option value="${value}"${value === entityType ? ' selected' : ''}>${esc(label)}</option>`).join('');
  const applyLabel = entityId ? 'Apply changes' : 'Import';
  return `
    <div class="mb-2">
      <label class="form-label small mb-1" for="draftImportType">Entity type</label>
      <select class="form-select form-select-sm" id="draftImportType">${options}</select>
    </div>
    <div class="mb-2">
      <label class="form-label small mb-1" for="draftImportJson">Draft JSON</label>
      <textarea class="form-control font-monospace" id="draftImportJson" rows="10" spellcheck="false"
        placeholder="Paste a draft object or the assistant's full reply"></textarea>
    </div>
    <div class="mb-2">
      <label class="form-label small mb-1" for="draftImportInstruction">Ask AI to change it (optional)</label>
      <div class="input-group input-group-sm">
        <input type="text" class="form-control" id="draftImportInstruction" placeholder="e.g. make act 2 a heist">
        <button class="btn btn-outline-primary" type="button" id="draftImportReviseBtn">Ask AI</button>
      </div>
    </div>
    <div class="alert alert-danger py-2 d-none" id="draftImportError"></div>
    <div class="alert alert-info py-2 d-none" id="draftImportNotice"></div>
    <div class="d-flex justify-content-end gap-2">
      <button class="btn btn-outline-secondary btn-sm" type="button" id="draftImportCancelBtn">Cancel</button>
      <button class="btn btn-outline-success btn-sm d-none" type="button" id="draftImportDoneBtn">Done</button>
      <button class="btn btn-primary btn-sm" type="button" id="draftImportApplyBtn">${applyLabel}</button>
    </div>`;
}

function wire(): void {
  el<HTMLSelectElement>('draftImportType')?.addEventListener('change', (e) => {
    entityType = (e.target as HTMLSelectElement).value;
  });
  el('draftImportApplyBtn')?.addEventListener('click', () => { void importDraft(); });
  el('draftImportReviseBtn')?.addEventListener('click', () => { void reviseDraft(); });
  el('draftImportCancelBtn')?.addEventListener('click', () => hideModal());
  el('draftImportDoneBtn')?.addEventListener('click', () => { hideModal(); refreshList(); });
  el<HTMLInputElement>('draftImportInstruction')?.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      void reviseDraft();
    }
  });
}

/** Open the paste-JSON import modal, optionally scoped to an entity type. */
function openDraftImport(type = 'oneshot', opts: DraftImportOptions = {}): void {
  entityType = type;
  campaignId = opts.campaignId ?? (currentCampaign as any)?.id ?? null;
  parentId = opts.parentId ?? null;
  entityId = opts.entityId ?? null;
  busy = false;
  showModal('Import JSON Draft', renderBody());
  const textarea = el<HTMLTextAreaElement>('draftImportJson');
  if (textarea && opts.json) textarea.value = opts.json;
  wire();
}

async function importDraft(): Promise<void> {
  if (busy) return;
  await applyDraft();
}

async function applyDraft(): Promise<void> {
  const json = currentJson();
  if (!json) {
    setError('Paste a JSON draft first.');
    return;
  }
  setBusy(true);
  setError('');
  try {
    const body: Record<string, unknown> = { entity_type: entityType, json };
    if (campaignId != null) body.campaign_id = campaignId;
    if (parentId != null) body.parent_id = parentId;
    if (entityId != null) body.entity_id = entityId;
    const res = await api<ImportReply>('POST', '/api/ai/import', body);
    entityId = res.id;
    const action = res.updated ? 'Updated' : 'Imported';
    setNotice(`${action} ${res.name}. Further edits apply to this ${selectedTypeLabel().toLowerCase()}.`);
    const applyBtn = el<HTMLButtonElement>('draftImportApplyBtn');
    if (applyBtn) applyBtn.textContent = 'Apply changes';
    el('draftImportDoneBtn')?.classList.remove('d-none');
    toast(`${action} ${res.name}`);
  } catch (e: any) {
    setError(e?.message || 'Import failed.');
  } finally {
    setBusy(false);
  }
}

async function reviseDraft(): Promise<void> {
  if (busy) return;
  const json = currentJson();
  const instruction = el<HTMLInputElement>('draftImportInstruction')?.value?.trim() ?? '';
  if (!json) {
    setError('Paste a JSON draft first.');
    return;
  }
  if (!instruction) {
    setError('Describe the change you want first.');
    return;
  }
  setBusy(true);
  setError('');
  setNotice('');
  try {
    const reply = await api<DraftReply>('POST', '/api/ai/revise', { entity_type: entityType, json, instruction });
    if (reply.message) setNotice(reply.message);
    if (reply.draft) {
      const textarea = el<HTMLTextAreaElement>('draftImportJson');
      if (textarea) textarea.value = JSON.stringify(reply.draft, null, 2);
      if (entityId != null) await applyDraft();
    }
  } catch (e: any) {
    setError(e?.message || 'Revision failed.');
  } finally {
    setBusy(false);
  }
}

expose('openDraftImport', openDraftImport);
