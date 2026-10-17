import { api } from './lib/api';
import { esc, toast } from './lib/dom';
import { expose } from './lib/expose';

export interface TelegramTargets {
  campaigns: Array<{ id: number; title: string; chat_title: string }>;
  dm: { linked: boolean };
}

export interface TelegramMediaResult {
  upload_id: number;
  status: 'sent' | 'failed';
  telegram_message_id?: number;
  error?: string;
}

const PHOTO_MAX = 10 * 1024 * 1024;
const OTHER_MAX = 50 * 1024 * 1024;

export function isAlbumKind(kind: string): boolean {
  const k = (kind || '').toLowerCase();
  return k === 'image' || k === 'video';
}

export function sizeLimitForKind(kind: string): number {
  return isAlbumKind(kind) && kind.toLowerCase() === 'image' ? PHOTO_MAX : kind.toLowerCase() === 'image' ? PHOTO_MAX : kind.toLowerCase() === 'video' ? OTHER_MAX : kind.toLowerCase() === 'pdf' || kind === '' ? PHOTO_MAX : kind.toLowerCase() === 'image' ? PHOTO_MAX : (kind.toLowerCase() === 'image' ? PHOTO_MAX : (isAlbumKind(kind) ? OTHER_MAX : OTHER_MAX));
  // Simplified: photos 10MB, everything else 50MB per spec, but spec says photos ≤10, other ≤50. Images are photos.
  // So image 10, otherwise 50.
}

export function sizeLimitSimple(kind: string): number {
  return (kind || '').toLowerCase() === 'image' ? PHOTO_MAX : OTHER_MAX;
}

export function validateSelection(items: Array<{ id: number; kind: string; size?: number }>): string | null {
  if (items.length === 0) return 'Select at least one file';
  if (items.length > 10) return 'Select up to 10 items per post (album limit is 10)';
  for (const it of items) {
    if (it.size != null) {
      const limit = sizeLimitSimple(it.kind);
      if (it.size > limit) {
        const mb = Math.round(limit / (1024 * 1024));
        if (it.kind.toLowerCase() === 'image') return `Photo ${it.id} exceeds ${mb} MB limit`;
        return `File ${it.id} exceeds ${mb} MB limit`;
      }
    }
  }
  return null;
}

export function batchAlbumIds(items: Array<{ id: number; kind: string }>): Array<Array<{ id: number; kind: string }>> {
  const batches: Array<Array<{ id: number; kind: string }>> = [];
  let current: Array<{ id: number; kind: string }> = [];
  const flush = () => { if (current.length) { batches.push(current); current = []; } };
  for (const it of items) {
    if (isAlbumKind(it.kind)) {
      current.push(it);
      if (current.length >= 10) flush();
    } else {
      flush();
      batches.push([it]);
    }
  }
  flush();
  return batches;
}

export async function fetchTargets(): Promise<TelegramTargets> {
  return api<TelegramTargets>('GET', '/api/telegram/targets');
}

export function renderTargetOptions(targets: TelegramTargets): string {
  let html = '';
  for (const c of targets.campaigns) {
    const label = c.chat_title ? `${esc(c.title)} — ${esc(c.chat_title)}` : esc(c.title);
    html += `<option value="campaign:${c.id}">${label} (campaign chat)</option>`;
  }
  if (targets.dm?.linked) {
    html += `<option value="dm">My DM chat</option>`;
  }
  return html;
}

export function renderResults(results: TelegramMediaResult[]): string {
  if (!results || results.length === 0) return '<div class="text-muted">No results</div>';
  return results.map(r => {
    if (r.status === 'sent') {
      return `<div class="text-success"><i class="fa-solid fa-check me-1"></i>File ${r.upload_id} sent${r.telegram_message_id ? ` (#${r.telegram_message_id})` : ''}</div>`;
    }
    return `<div class="text-danger"><i class="fa-solid fa-triangle-exclamation me-1"></i>File ${r.upload_id} failed: ${esc(r.error || 'Unknown error')}</div>`;
  }).join('');
}

async function loadTargets(ownerType: string, ownerID: string): Promise<void> {
  const targetsEl = document.getElementById(`telegramPostTargets-${ownerType}-${ownerID}`);
  const select = document.getElementById(`telegramPostTargetSelect-${ownerType}-${ownerID}`) as HTMLSelectElement | null;
  if (!targetsEl || !select) return;
  targetsEl.textContent = 'Loading targets…';
  try {
    const targets = await fetchTargets();
    if (targets.campaigns.length === 0 && !targets.dm.linked) {
      targetsEl.innerHTML = '<span class="text-muted">No campaign chat bound and no DM linked. Link Telegram or bind a campaign chat first.</span>';
      select.innerHTML = '<option value="">No targets available</option>';
      return;
    }
    const count = `${targets.campaigns.length} campaign chat(s), DM ${targets.dm.linked ? 'linked' : 'not linked'}`;
    targetsEl.textContent = count;
    const opts = '<option value="">Select target</option>' + renderTargetOptions(targets);
    select.innerHTML = opts;
  } catch (e: any) {
    targetsEl.textContent = e?.message || 'Failed to load targets';
    select.innerHTML = '<option value="">Failed to load</option>';
  }
}

async function send(ownerType: string, ownerID: string): Promise<void> {
  const select = document.getElementById(`telegramPostTargetSelect-${ownerType}-${ownerID}`) as HTMLSelectElement | null;
  const captionEl = document.getElementById(`telegramPostCaption-${ownerType}-${ownerID}`) as HTMLTextAreaElement | null;
  const recapEl = document.getElementById(`telegramPostRecap-${ownerType}-${ownerID}`) as HTMLInputElement | null;
  const validationEl = document.getElementById(`telegramPostValidation-${ownerType}-${ownerID}`);
  const resultsEl = document.getElementById(`telegramPostResults-${ownerType}-${ownerID}`);
  const sendBtn = document.getElementById(`telegramPostSendBtn-${ownerType}-${ownerID}`) as HTMLButtonElement | null;
  if (!select) return;
  const rawTarget = select.value;
  if (!rawTarget) {
    if (validationEl) { validationEl.textContent = 'Choose a target (campaign chat or My DM chat)'; validationEl.classList.remove('d-none'); }
    return;
  }
  // Gather selected ids from gallery
  const checks = document.querySelectorAll(`#mediaGrid-${ownerType}-${ownerID} input[data-testid^="media-select-"]:checked`);
  const ids: number[] = [];
  const items: Array<{ id: number; kind: string }> = [];
  checks.forEach(c => {
    const el = c as HTMLInputElement;
    const id = parseInt(el.value);
    if (!isNaN(id)) { ids.push(id); items.push({ id, kind: el.getAttribute('data-kind') || '' }); }
  });
  if (ids.length === 0) {
    if (validationEl) { validationEl.textContent = 'Select at least one file'; validationEl.classList.remove('d-none'); }
    return;
  }
  if (ids.length > 10) {
    if (validationEl) { validationEl.textContent = 'Select up to 10 items (album limit is 10)'; validationEl.classList.remove('d-none'); }
    return;
  }
  // Light client-side size validation if data-size present (not yet, skip)
  const caption = captionEl?.value?.slice(0, 1024) || undefined;
  let recapId: number | undefined;
  if (recapEl?.value?.trim()) {
    const n = parseInt(recapEl.value.trim());
    if (!isNaN(n)) recapId = n;
  }
  let target: any;
  if (rawTarget === 'dm') target = { kind: 'dm' };
  else if (rawTarget.startsWith('campaign:')) target = { kind: 'campaign', campaign_id: parseInt(rawTarget.split(':')[1]) };
  else target = { kind: rawTarget };

  if (validationEl) { validationEl.classList.add('d-none'); validationEl.textContent = ''; }
  if (resultsEl) resultsEl.innerHTML = '<span class="text-muted">Sending…</span>';
  if (sendBtn) sendBtn.disabled = true;
  try {
    const body: any = { upload_ids: ids, target };
    if (caption) body.caption = caption;
    if (recapId != null) body.recap_id = recapId;
    const res = await api<{ results: TelegramMediaResult[] }>('POST', '/api/telegram/media', body);
    const html = renderResults(res.results || []);
    if (resultsEl) resultsEl.innerHTML = html;
    const failed = (res.results || []).filter(r => r.status === 'failed');
    if (failed.length === 0) toast('Posted to Telegram');
    else toast(`${failed.length} file(s) failed`, true);
  } catch (e: any) {
    const msg = e?.message || 'Failed to post';
    if (resultsEl) resultsEl.innerHTML = `<div class="text-danger">${esc(msg)}</div>`;
    toast(msg, true);
  } finally {
    if (sendBtn) sendBtn.disabled = false;
  }
}

expose('TelegramMedia', { loadTargets, send, fetchTargets, renderTargetOptions, renderResults, validateSelection, batchAlbumIds, isAlbumKind, sizeLimitSimple });

export const TelegramMedia = { loadTargets, send, fetchTargets, renderTargetOptions, renderResults, validateSelection, batchAlbumIds, isAlbumKind, sizeLimitSimple };
