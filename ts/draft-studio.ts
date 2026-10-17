// Draft Studio — one integrated workflow for AI drafting, tweaking and import.
//
// Replaces the separate chat (ai-draft.ts) and paste-import (draft-import.ts)
// modules: a chat pane on the left, an editable JSON draft pane on the right,
// with section-scoped AI tweaks and one-click import into the app.
import { esc, toast } from './lib/dom';
import { api } from './lib/api';
import { expose } from './lib/expose';
import { currentCampaign } from './lib/state';

interface DraftMessage { role: string; content: string }
interface DraftReply {
  id?: string;
  entity_type?: string;
  status: string;
  message?: string;
  draft?: unknown;
  messages?: DraftMessage[];
}
interface ImportReply {
  id: number;
  url: string;
  entity_type: string;
  name: string;
  updated: boolean;
  counts?: Record<string, number>;
}

const STORAGE_KEY = 'villum.aiDraft';
const WHOLE = 'Whole draft';

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

interface StudioOpts {
  type?: string;
  campaignId?: number | null;
  parentId?: number | null;
  entityId?: number | null;
  json?: string;
  mode?: 'chat' | 'paste';
}

let entityType = 'oneshot';
let campaignId: number | null = null;
let parentId: number | null = null;
let entityId: number | null = null;
let sessionId: string | null = null;
let serverReady = false;
let currentJson = '';
let selectedSection = WHOLE;
let busy = false;

function el<T extends HTMLElement>(id: string): T {
  return document.getElementById(id) as T;
}

function draftModal(): any {
  return (window as any).bootstrap?.Modal?.getOrCreateInstance(el('aiDraftModal'));
}

function typeLabel(type: string): string {
  return ENTITY_TYPES.find(([id]) => id === type)?.[1] ?? type;
}

function setError(msg: string): void {
  const box = el('aiDraftError');
  box.textContent = msg;
  box.classList.toggle('d-none', !msg);
}

function setNotice(msg: string): void {
  const box = el('aiDraftNotice');
  box.textContent = msg;
  box.classList.toggle('d-none', !msg);
}

function setStatus(text: string, cls: string): void {
  const badge = el('aiDraftStatus');
  badge.textContent = text;
  badge.className = 'badge ' + cls;
}

function setBusy(b: boolean): void {
  busy = b;
  el('aiDraftSendBtn').toggleAttribute('disabled', b);
  el('aiDraftTweakBtn').toggleAttribute('disabled', b);
  el('aiDraftImportBtn').toggleAttribute('disabled', b);
}

function endpointId(): number {
  const raw = (el<HTMLSelectElement>('aiDraftEndpoint').value || '').trim();
  return raw ? Number(raw) : 0;
}

function refreshEndpointOptions(endpoints: Array<{ id: number; name: string; model: string }>): void {
  const select = el<HTMLSelectElement>('aiDraftEndpoint');
  if (!endpoints.length) {
    select.innerHTML = '<option value="">No text endpoint configured</option>';
    return;
  }
  select.innerHTML = endpoints
    .map(e => `<option value="${e.id}">${esc(e.name)} (${esc(e.model)})</option>`)
    .join('');
}

async function loadEndpoints(): Promise<void> {
  try {
    const endpoints = await api<Array<{ id: number; name: string; model: string }>>('GET', '/api/ai/endpoints?type=text');
    refreshEndpointOptions(endpoints);
  } catch {
    refreshEndpointOptions([]);
  }
}

function addBubble(role: string, text: string): void {
  const wrap = el('aiDraftMessages');
  const isUser = role === 'user';
  const bubble = document.createElement('div');
  bubble.className = isUser ? 'align-self-end' : 'align-self-start';
  bubble.style.maxWidth = '92%';
  bubble.innerHTML = `
    <div class="p-2 rounded ${isUser ? 'bg-primary text-white' : 'bg-light border'}">
      <div class="small ${isUser ? 'text-white-50' : 'text-muted'} mb-1">${isUser ? 'You' : 'AI'}</div>
      <div style="white-space:pre-wrap">${esc(text)}</div>
    </div>`;
  wrap.appendChild(bubble);
  wrap.scrollTop = wrap.scrollHeight;
}

function renderMessages(messages: DraftMessage[]): void {
  const wrap = el('aiDraftMessages');
  wrap.innerHTML = '';
  for (const m of messages) {
    if (m.role === 'system') continue;
    addBubble(m.role, m.content);
  }
}

// extractJsonObject salvages the first JSON object from a reply that the
// provider did not deliver in the structured envelope.
function extractJsonObject(text: string): string | null {
  const start = text.indexOf('{');
  const end = text.lastIndexOf('}');
  if (start < 0 || end <= start) return null;
  const candidate = text.slice(start, end + 1);
  try {
    const parsed = JSON.parse(candidate);
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) return candidate;
  } catch { /* not valid JSON */ }
  return null;
}

function draftFromReply(reply: DraftReply): string | null {
  if (reply.draft && typeof reply.draft === 'object') return JSON.stringify(reply.draft);
  if (typeof reply.draft === 'string' && reply.draft.trim()) {
    try { JSON.parse(reply.draft); return reply.draft; } catch { /* fall through */ }
  }
  const messages = reply.messages ?? [];
  for (let i = messages.length - 1; i >= 0; i--) {
    if (messages[i].role !== 'assistant') continue;
    const salvaged = extractJsonObject(messages[i].content);
    if (salvaged) return salvaged;
  }
  return null;
}

function sectionLabels(json: string): string[] {
  let d: any;
  try { d = JSON.parse(json); } catch { return []; }
  if (!d || typeof d !== 'object') return [];
  const labels: string[] = [];
  (d.acts ?? []).forEach((a: any, i: number) => {
    labels.push(`Act ${i + 1}: ${a?.title || 'untitled'}`);
  });
  (d.npcs ?? []).forEach((n: any) => { if (n?.name) labels.push(`NPC: ${n.name}`); });
  (d.locations ?? []).forEach((l: any) => { if (l?.name) labels.push(`Location: ${l.name}`); });
  (d.encounters ?? []).forEach((e: any) => { if (e?.name) labels.push(`Encounter: ${e.name}`); });
  (d.clues ?? []).forEach((c: any) => { if (c?.title) labels.push(`Clue: ${c.title}`); });
  return labels;
}

function renderSections(): void {
  const box = el('aiDraftSections');
  const labels = [WHOLE, ...sectionLabels(currentJson)];
  if (!currentJson.trim()) {
    box.innerHTML = '';
    return;
  }
  if (!labels.includes(selectedSection)) selectedSection = WHOLE;
  box.innerHTML = labels
    .map(label => `<button type="button" class="btn btn-sm ${label === selectedSection ? 'btn-primary' : 'btn-outline-secondary'}"
        onclick="selectDraftSection(this)" data-section="${esc(label)}">${esc(label)}</button>`)
    .join('');
}

function setDraftJson(value: string): void {
  currentJson = value;
  const area = el<HTMLTextAreaElement>('aiDraftJson');
  if (area.value !== value) {
    try { area.value = JSON.stringify(JSON.parse(value), null, 2); }
    catch { area.value = value; }
  }
  renderSections();
  if (value.trim()) {
    setStatus(serverReady ? 'ready' : 'draft', serverReady ? 'bg-success' : 'bg-info');
  } else {
    setStatus('empty', 'bg-secondary');
  }
  el('aiDraftImportBtn').innerHTML = entityId
    ? '<i class="fa-solid fa-file-import me-1"></i>Apply changes'
    : '<i class="fa-solid fa-file-import me-1"></i>Import into app';
}

function persistSession(): void {
  if (sessionId) localStorage.setItem(STORAGE_KEY, JSON.stringify({ id: sessionId, entityType }));
  else localStorage.removeItem(STORAGE_KEY);
}

function applyReply(reply: DraftReply): void {
  if (reply.id) sessionId = reply.id;
  if (reply.entity_type) entityType = reply.entity_type;
  if (reply.messages) renderMessages(reply.messages);
  persistSession();
  const draft = draftFromReply(reply);
  if (draft) {
    serverReady = reply.status === 'ready';
    setDraftJson(draft);
    if (reply.status === 'ready') setNotice('Draft ready. Tweak any part or import it.');
  } else if (reply.message) {
    setNotice('');
  }
}

function renderStudio(): void {
  el('aiDraftModalTitle').textContent = `Draft Studio — ${typeLabel(entityType)}`;
  el('aiDraftEntityType').setAttribute('value', entityType);
  el<HTMLTextAreaElement>('aiDraftJson').value = '';
  el('aiDraftMessages').innerHTML = '';
  setError('');
  setNotice('');
  setStatus(currentJson.trim() ? 'draft' : 'empty', currentJson.trim() ? 'bg-info' : 'bg-secondary');
  if (currentJson.trim()) {
    setDraftJson(currentJson);
  } else {
    renderSections();
  }
}

async function reconnectSession(): Promise<void> {
  try {
    const reply = await api<DraftReply>('GET', `/api/ai/draft/${sessionId}`);
    applyReply(reply);
  } catch {
    sessionId = null;
    persistSession();
  }
}

export function openDraftStudio(opts: StudioOpts = {}): void {
  entityType = opts.type || 'oneshot';
  campaignId = opts.campaignId ?? (currentCampaign as any)?.id ?? null;
  parentId = opts.parentId ?? null;
  entityId = opts.entityId ?? null;
  serverReady = false;
  currentJson = opts.json || '';
  selectedSection = WHOLE;

  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored) {
    try {
      const parsed = JSON.parse(stored);
      if (parsed?.entityType === entityType && parsed?.id) sessionId = parsed.id;
    } catch { /* ignore */ }
  } else {
    sessionId = null;
  }

  renderStudio();
  draftModal()?.show();
  void loadEndpoints();
  if (opts.mode === 'chat' && sessionId) void reconnectSession();
  if (opts.json) setNotice(`Loaded ${typeLabel(entityType)}. Tweak any part or import it.`);
}

export function openAIDraft(type = 'oneshot', opts: StudioOpts = {}): void {
  openDraftStudio({ ...opts, type, mode: 'chat' });
}

export function openDraftImport(type = 'oneshot', opts: StudioOpts = {}): void {
  openDraftStudio({ ...opts, type, mode: 'paste' });
}

// tweakOneShotWithAI loads an existing adventure as a draft and opens the
// studio so individual parts can be revised and applied back.
export async function tweakOneShotWithAI(id: number): Promise<void> {
  try {
    const draft = await api<Record<string, unknown>>('GET', `/api/oneshot-adventures/${id}/draft`);
    openDraftStudio({ type: 'oneshot', entityId: id, json: JSON.stringify(draft), mode: 'paste' });
    setNotice('Loaded the existing adventure. Tweak any part or apply changes.');
  } catch (e: any) {
    toast(e?.message || 'Failed to load the adventure draft', true);
  }
}

export function selectDraftSection(btn: HTMLElement): void {
  selectedSection = btn.getAttribute('data-section') || WHOLE;
  renderSections();
}

export async function sendAIDraftMessage(): Promise<void> {
  if (busy) return;
  const input = el<HTMLTextAreaElement>('aiDraftInput');
  const message = input.value.trim();
  if (!message) return;
  setError('');
  setBusy(true);
  input.value = '';
  addBubble('user', message);
  try {
    const payload: Record<string, unknown> = { endpoint_id: endpointId(), message };
    let reply: DraftReply;
    if (sessionId) {
      reply = await api<DraftReply>('POST', `/api/ai/draft/${sessionId}/turn`, payload);
    } else {
      payload.entity_type = entityType;
      if (campaignId) payload.campaign_id = campaignId;
      if (parentId) payload.parent_id = parentId;
      reply = await api<DraftReply>('POST', '/api/ai/draft', payload);
    }
    applyReply(reply);
  } catch (e: any) {
    setError(e?.message || 'The AI request failed');
    toast(e?.message || 'The AI request failed', true);
  } finally {
    setBusy(false);
  }
}

export async function tweakSelectedDraftSection(): Promise<void> {
  if (busy) return;
  const instruction = el<HTMLInputElement>('aiDraftTweakInput').value.trim();
  if (!currentJson.trim()) { setError('There is no draft to tweak yet.'); return; }
  if (!instruction) { setError('Describe what to change first.'); return; }
  setError('');
  setNotice('');
  setBusy(true);
  try {
    const payload: Record<string, unknown> = { entity_type: entityType, json: currentJson, instruction, section: selectedSection };
    if (endpointId()) payload.endpoint_id = endpointId();
    const reply = await api<DraftReply>('POST', '/api/ai/revise', payload);
    const draft = draftFromReply(reply);
    if (draft) {
      serverReady = reply.status === 'ready';
      setDraftJson(draft);
      setNotice(`Tweaked ${selectedSection}.`);
      el<HTMLInputElement>('aiDraftTweakInput').value = '';
    } else {
      setNotice(reply.message || 'The AI did not return a revised draft.');
    }
  } catch (e: any) {
    setError(e?.message || 'The AI revision failed');
  } finally {
    setBusy(false);
  }
}

export async function commitAIDraft(): Promise<void> {
  if (busy) return;
  if (!currentJson.trim()) { setError('There is no draft to import yet.'); return; }
  try { JSON.parse(currentJson); } catch { setError('The draft JSON is not valid.'); return; }
  setError('');
  setBusy(true);
  try {
    const payload: Record<string, unknown> = { entity_type: entityType, json: currentJson };
    if (campaignId) payload.campaign_id = campaignId;
    if (parentId) payload.parent_id = parentId;
    if (entityId) payload.entity_id = entityId;
    const res = await api<ImportReply>('POST', '/api/ai/import', payload);
    entityId = res.id;
    serverReady = false;
    setDraftJson(currentJson);
    setStatus('imported', 'bg-success');
    setNotice(res.updated
      ? `Updated ${res.name}. Further edits apply to the same ${typeLabel(entityType)}.`
      : `Imported ${res.name}. Further edits apply to the same ${typeLabel(entityType)}.`);
    toast(res.updated ? `Updated ${res.name}` : `Imported ${res.name}`);
  } catch (e: any) {
    setError(e?.message || 'Import failed');
  } finally {
    setBusy(false);
  }
}

export function finishAIDraft(): void {
  draftModal()?.hide();
  if (entityType === 'oneshot' && typeof (window as any).showOneShots === 'function') {
    (window as any).showOneShots();
  }
}

export function copyAIDraftJson(): void {
  const value = el<HTMLTextAreaElement>('aiDraftJson').value;
  if (!value) return;
  navigator.clipboard?.writeText(value).then(
    () => toast('Draft JSON copied'),
    () => toast('Copy failed', true),
  );
}

export function resetAIDraft(): void {
  sessionId = null;
  entityId = null;
  serverReady = false;
  currentJson = '';
  selectedSection = WHOLE;
  localStorage.removeItem(STORAGE_KEY);
  renderStudio();
  el<HTMLTextAreaElement>('aiDraftInput').value = '';
  el<HTMLInputElement>('aiDraftTweakInput').value = '';
  setNotice('Started over.');
}

// Manual edits in the textarea become the draft that tweaks and imports use.
document.addEventListener('input', (e) => {
  const target = e.target as HTMLElement;
  if (target?.id === 'aiDraftJson') {
    currentJson = (target as HTMLTextAreaElement).value;
    renderSections();
  }
});

expose('openDraftStudio', openDraftStudio);
expose('openAIDraft', openAIDraft);
expose('openDraftImport', openDraftImport);
expose('tweakOneShotWithAI', tweakOneShotWithAI);
expose('sendAIDraftMessage', sendAIDraftMessage);
expose('tweakSelectedDraftSection', tweakSelectedDraftSection);
expose('selectDraftSection', selectDraftSection);
expose('commitAIDraft', commitAIDraft);
expose('finishAIDraft', finishAIDraft);
expose('copyAIDraftJson', copyAIDraftJson);
expose('resetAIDraft', resetAIDraft);
