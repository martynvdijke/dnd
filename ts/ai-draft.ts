/**
 * AI Drafting module — a conversational, steerable generator.
 *
 * The DM describes what they want, the assistant asks questions and suggests
 * ideas, the DM steers, and once the DM approves the assistant returns a
 * structured JSON draft. The UI loads that draft and can create the entity
 * with one click. Sessions are persisted server-side, so closing the modal
 * and reopening it reconnects to the same conversation.
 */
import { esc, toast } from './lib/dom';
import { api } from './lib/api';
import { expose } from './lib/expose';

interface DraftMessage {
  role: string;
  content: string;
}

interface DraftReply {
  id: string;
  entity_type: string;
  status: string;
  message: string;
  draft: unknown;
  messages: DraftMessage[];
}

const STORAGE_KEY = 'villum.aiDraft';

let sessionId: string | null = null;
let entityType = 'oneshot';
let campaignId: number | null = null;
let parentId: number | null = null;
let busy = false;
let lastDraft: unknown = null;

function el<T extends HTMLElement>(id: string): T {
  return document.getElementById(id) as T;
}

function draftModal(): any {
  const b = (window as any).bootstrap;
  return b.Modal.getOrCreateInstance(el('aiDraftModal'));
}

async function loadEndpointOptions(): Promise<void> {
  const select = el<HTMLSelectElement>('aiDraftEndpoint');
  try {
    const endpoints = await api<any[]>('GET', '/api/ai/endpoints?type=text');
    if (!endpoints || endpoints.length === 0) {
      select.innerHTML = '<option value="">No text endpoints configured</option>';
      return;
    }
    select.innerHTML = endpoints
      .map(ep => `<option value="${ep.id}">${esc(ep.name)} (${esc(ep.model)})</option>`)
      .join('');
  } catch {
    select.innerHTML = '<option value="">Could not load endpoints</option>';
  }
}

function addBubble(role: string, text: string): void {
  const box = el('aiDraftMessages');
  const wrap = document.createElement('div');
  const isUser = role === 'user';
  wrap.className = 'd-flex ' + (isUser ? 'justify-content-end' : 'justify-content-start');
  const bubble = document.createElement('div');
  bubble.className = 'p-2 px-3 rounded';
  bubble.style.maxWidth = '85%';
  bubble.style.whiteSpace = 'pre-wrap';
  bubble.style.background = isUser ? 'var(--bs-primary)' : 'var(--bs-light)';
  bubble.style.color = isUser ? '#fff' : 'inherit';
  bubble.textContent = text;
  wrap.appendChild(bubble);
  box.appendChild(wrap);
  box.scrollTop = box.scrollHeight;
}

function renderSession(reply: DraftReply): void {
  sessionId = reply.id;
  entityType = reply.entity_type;
  el('aiDraftMessages').innerHTML = '';
  for (const m of reply.messages || []) {
    if (m.role === 'system') continue;
    addBubble(m.role, m.content);
  }
  lastDraft = reply.draft || null;
  const commitBtn = el<HTMLButtonElement>('aiDraftCommitBtn');
  const draftBox = el('aiDraftDraftBox');
  if (reply.status === 'ready' && lastDraft) {
    commitBtn.style.display = 'block';
    draftBox.style.display = 'block';
    el('aiDraftDraftJson').textContent = JSON.stringify(lastDraft, null, 2);
  } else {
    commitBtn.style.display = 'none';
    draftBox.style.display = 'none';
  }
  localStorage.setItem(STORAGE_KEY, JSON.stringify({ id: sessionId, entityType }));
}

function setBusy(value: boolean): void {
  busy = value;
  const btn = el<HTMLButtonElement>('aiDraftSendBtn');
  btn.disabled = value;
  btn.innerHTML = value
    ? '<i class="fa-solid fa-spinner fa-spin"></i>'
    : '<i class="fa-solid fa-paper-plane"></i>';
}

function endpointId(): number {
  const v = el<HTMLSelectElement>('aiDraftEndpoint').value;
  return v ? parseInt(v, 10) : 0;
}

/**
 * Open the drafting modal. `opts.campaignId` scopes created entities to a
 * campaign; `opts.parentId` supplies the parent for character- or
 * adventure-scoped types (quests, items).
 */
export function openAIDraft(
  type = 'oneshot',
  opts: { campaignId?: number | null; parentId?: number | null; reconnect?: boolean } = {}
): void {
  entityType = type;
  campaignId = opts.campaignId ?? null;
  parentId = opts.parentId ?? null;
  el('aiDraftModalTitle').textContent = 'Draft with AI';
  el('aiDraftInput').setAttribute('placeholder', 'Describe what you want, or type "looks good, generate it"...');

  const start = () => {
    sessionId = null;
    lastDraft = null;
    el('aiDraftMessages').innerHTML =
      '<div class="text-muted small">Tell the assistant what you have in mind. It will ask questions and suggest ideas before writing the draft.</div>';
    el('aiDraftCommitBtn').style.display = 'none';
    el('aiDraftDraftBox').style.display = 'none';
  };

  draftModal().show();
  void loadEndpointOptions();

  const stored = opts.reconnect === false ? null : localStorage.getItem(STORAGE_KEY);
  if (stored) {
    try {
      const parsed = JSON.parse(stored);
      if (parsed && parsed.entityType === type && parsed.id) {
        api<DraftReply>('GET', `/api/ai/draft/${parsed.id}`)
          .then(renderSession)
          .catch(start);
        return;
      }
    } catch {
      /* ignore malformed storage */
    }
  }
  start();
}

/** Send the current input as the next turn (or the opening message). */
export async function sendAIDraftMessage(): Promise<void> {
  if (busy) return;
  const input = el<HTMLTextAreaElement>('aiDraftInput');
  const text = input.value.trim();
  if (!text) return;
  const ep = endpointId();
  if (!ep) {
    toast('Select an AI endpoint first', true);
    return;
  }
  const wasNew = !sessionId;
  addBubble('user', text);
  input.value = '';
  setBusy(true);
  try {
    const body: any = { endpoint_id: ep, message: text };
    let reply: DraftReply;
    if (wasNew) {
      body.entity_type = entityType;
      if (campaignId != null) body.campaign_id = campaignId;
      if (parentId != null) body.parent_id = parentId;
      reply = await api<DraftReply>('POST', '/api/ai/draft', body);
    } else {
      reply = await api<DraftReply>('POST', `/api/ai/draft/${sessionId}/turn`, body);
    }
    renderSession(reply);
  } catch (e: any) {
    toast(e?.message || 'AI request failed', true);
    if (wasNew) {
      // Remove the optimistic bubble so the user can retry.
      el('aiDraftMessages').lastElementChild?.remove();
    }
  } finally {
    setBusy(false);
  }
}

/** Create the entity from the approved draft. */
export async function commitAIDraft(): Promise<void> {
  if (!sessionId) return;
  setBusy(true);
  try {
    const result = await api<{ entity_id: number; url: string; entity_type: string }>(
      'POST',
      `/api/ai/draft/${sessionId}/commit`
    );
    localStorage.removeItem(STORAGE_KEY);
    draftModal().hide();
    toast('Created successfully');
    if (result.entity_type === 'oneshot' && typeof (window as any).showOneShots === 'function') {
      (window as any).showOneShots();
    }
  } catch (e: any) {
    toast(e?.message || 'Failed to create', true);
  } finally {
    setBusy(false);
  }
}

/** Start a fresh conversation in the open modal. */
export function resetAIDraft(): void {
  localStorage.removeItem(STORAGE_KEY);
  openAIDraft(entityType, { campaignId, parentId, reconnect: false });
}

expose('openAIDraft', openAIDraft);
expose('sendAIDraftMessage', sendAIDraftMessage);
expose('commitAIDraft', commitAIDraft);
expose('resetAIDraft', resetAIDraft);
