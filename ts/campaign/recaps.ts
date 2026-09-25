import { Editor } from '@tiptap/core';
import StarterKit from '@tiptap/starter-kit';
import Placeholder from '@tiptap/extension-placeholder';
import { api } from '../lib/api';
import { esc, showModal, hideModal, toast } from '../lib/dom';
import { expose } from '../lib/expose';
import { showView } from '../navigation';
import { currentCampaign } from '../lib/state';

let recapEditor: Editor | null = null;
let editingId: number | null = null;

function getCid(): number | null { return (currentCampaign as any)?.id ?? null; }

function destroyEditor() { if (recapEditor) { recapEditor.destroy(); recapEditor = null; } }

function initEditor(content?: string) {
  setTimeout(() => {
    const el = document.getElementById('recapEditor');
    if (!el) return;
    recapEditor = new Editor({
      element: el,
      extensions: [StarterKit.configure({ heading: { levels: [1, 2, 3] } }), Placeholder.configure({ placeholder: 'Session notes...' })],
      content: content || '<p></p>',
    });
    const modal = document.getElementById('genericModal');
    modal?.addEventListener('hidden.bs.modal', destroyEditor, { once: true });
  }, 50);
}

function dateVal(v: any): string { return typeof v === 'string' && v ? v.substring(0, 10) : ''; }

function dateRangeHtml(r: any): string {
  const s = dateVal(r.session_start_date);
  const e = dateVal(r.session_end_date);
  if (!s && !e) return '';
  const txt = s && e ? `${esc(s)} – ${esc(e)}` : esc(s || e);
  return ` <small class="text-muted">${txt}</small>`;
}

function getSelectValues(id: string): number[] {
  const sel = document.getElementById(id) as HTMLSelectElement | null;
  if (!sel) return [];
  return [...sel.selectedOptions].map(o => +o.value).filter(n => !isNaN(n));
}

async function syncRecapLinks(recapId: number, placeIds: number[], eventIds: number[]) {
  try {
    let data: any = {};
    try { data = await api('GET', `/api/links/recap/${recapId}`); } catch { data = {}; }
    const outgoing: any[] = data.outgoing || data.links || [];
    for (const l of outgoing) {
      if (l.source_type === 'recap' && l.source_id === recapId) {
        try { await api('DELETE', `/api/links/${l.id}`); } catch {}
      }
    }
    for (const tid of placeIds) {
      try { await api('POST', '/api/links', { source_type: 'recap', source_id: recapId, target_type: 'location', target_id: tid, context: 'manual' }); } catch (e: any) {
        if (!String(e.message).includes('already exists')) throw e;
      }
    }
    for (const tid of eventIds) {
      try { await api('POST', '/api/links', { source_type: 'recap', source_id: recapId, target_type: 'timeline', target_id: tid, context: 'manual' }); } catch (e: any) {
        if (!String(e.message).includes('already exists')) throw e;
      }
    }
  } catch (e: any) {
    toast('Recap saved, but linking failed: ' + e.message, true);
  }
}

async function populateRecapSelectors(recapId: number | null, cid: number) {
  let locs: any[] = [];
  let events: any[] = [];
  try { locs = await api('GET', '/api/locations'); } catch {}
  try {
    const raw: any = await api('GET', `/api/timeline?campaign_id=${cid}`);
    events = Array.isArray(raw) ? raw : (raw.events ?? []);
  } catch {}
  let selectedPlaceIds = new Set<number>();
  let selectedEventIds = new Set<number>();
  if (recapId) {
    try {
      const data: any = await api('GET', `/api/links/recap/${recapId}`);
      const outgoing: any[] = data.outgoing || data.links || [];
      for (const l of outgoing) {
        if (l.source_type !== 'recap' || l.source_id !== recapId) continue;
        if (l.target_type === 'location') selectedPlaceIds.add(l.target_id);
        if (l.target_type === 'timeline') selectedEventIds.add(l.target_id);
      }
    } catch {}
  }
  const placeSel = document.getElementById('recapPlaces') as HTMLSelectElement | null;
  if (placeSel) {
    placeSel.innerHTML = locs.map((l: any) => `<option value="${l.id}" ${selectedPlaceIds.has(l.id) ? 'selected' : ''}>${esc(l.name)} (${esc(l.type || '')})</option>`).join('') || '<option disabled>No places</option>';
  }
  const eventSel = document.getElementById('recapEvents') as HTMLSelectElement | null;
  if (eventSel) {
    eventSel.innerHTML = events.map((e: any) => `<option value="${e.id}" ${selectedEventIds.has(e.id) ? 'selected' : ''}>${esc(e.title)}${e.event_date ? ` — ${esc(e.event_date)}` : ''}</option>`).join('') || '<option disabled>No timeline events</option>';
  }
}

function recapFormExtras(startDate: string, endDate: string): string {
  return `
    <div class="row g-2 mb-3">
      <div class="col-6"><label class="form-label">Start date</label><input class="form-control" type="date" id="recapStartDate" value="${esc(startDate)}"></div>
      <div class="col-6"><label class="form-label">End date</label><input class="form-control" type="date" id="recapEndDate" value="${esc(endDate)}"></div>
    </div>
    <div class="mb-3"><label class="form-label">Linked places</label><select multiple class="form-select" id="recapPlaces" size="4"><option disabled>Loading...</option></select><small class="text-muted">Ctrl/Cmd+click to select multiple</small></div>
    <div class="mb-3"><label class="form-label">Linked timeline events</label><select multiple class="form-select" id="recapEvents" size="4"><option disabled>Loading...</option></select></div>
  `;
}

export async function generateAIRecap(): Promise<void> {
  const cid = getCid();
  if (!cid) return;
  try {
    const r: any = await api('POST', `/api/campaigns/${cid}/recaps/generate-ai`);
    showRecapFormPrefilled(r.title || '', r.content || '', dateVal(r.session_start_date), dateVal(r.session_end_date));
  } catch (e: any) { toast(e.message, true); }
}

function showRecapFormPrefilled(title: string, content: string, startDate = '', endDate = ''): void {
  const cid = getCid();
  editingId = null;
  showModal('New Recap (AI Generated)', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="recapTitle" value="${esc(title)}" placeholder="Session Recap"></div>
    ${recapFormExtras(startDate, endDate)}
    <div class="mb-3"><label class="form-label">Notes</label><div class="editor-toolbar" id="recapToolbar"></div><div id="recapEditor" class="journal-editor" style="min-height:150px;border:1px solid var(--border-light);border-radius:6px;padding:8px"></div></div>
    <button class="ai-generate-btn btn btn-outline-secondary btn-sm mb-3" data-ai-mode="text" data-ai-target="recapEditor" data-ai-hint="Summarize these session notes into a concise campaign recap"><i class="fa-solid fa-wand-magic-sparkles me-1"></i>Generate with AI</button>
    <button class="btn btn-primary w-100" onclick="saveRecap()"><i class="fa-solid fa-save me-1"></i>Save</button>
  `);
  initEditor(content);
  if (cid) populateRecapSelectors(null, cid);
}

export async function renderRecaps(campaignId?: number): Promise<void> {
  const cid = campaignId ?? getCid();
  if (!cid) {
    // Sessions are campaign-scoped. Route to the picker instead of dead-ending
    // on a toast — showView's campaign gate exempts admins, so do it explicitly.
    (window as any).loadCampaignPicker?.();
    return;
  }
  showView('recaps' as any);
  const el = document.getElementById('recapsContent')!;
  el.innerHTML = '<div class="ornament">Loading sessions...</div>';
  try {
    const recaps = await api('GET', `/api/campaigns/${cid}/recaps`);
    const headerBtns = `<div class="d-flex gap-2"><button class="btn btn-outline-secondary btn-sm" onclick="generateAIRecap()"><i class="fa-solid fa-wand-magic-sparkles me-1"></i>Generate AI Recap</button><button class="btn btn-primary btn-sm" onclick="showRecapForm()"><i class="fa-solid fa-plus me-1"></i>New Recap</button></div>`;
    if (!recaps.length) {
      el.innerHTML = `<div class="d-flex justify-content-between align-items-center mb-3"><h5 class="mb-0">Session Recaps</h5>${headerBtns}</div><div class="empty-state"><p class="text-muted">No session recaps yet.</p></div>`;
      return;
    }
    el.innerHTML = `<div class="d-flex justify-content-between align-items-center mb-3"><h5 class="mb-0">Session Recaps (${recaps.length})</h5>${headerBtns}</div>` +
      recaps.map((r: any) => `<div class="card mb-2"><div class="card-body py-2 px-3"><div class="d-flex justify-content-between align-items-start"><div><span class="fw-bold">${esc(r.title)}</span>${r.is_sent ? ' <span class="badge bg-success">Sent</span>' : ''}${r.is_edited ? ' <span class="badge bg-secondary">Edited</span>' : ''}${r.ai_used ? ' <span class="badge bg-info text-dark">AI</span>' : ''}${dateRangeHtml(r)}<br><small class="text-muted">${r.word_count || 0} words</small><div class="small mt-1">${r.content}</div></div><div class="d-flex gap-1 flex-shrink-0"><button class="btn btn-sm btn-outline-info" onclick="shareRecap(${r.id})" title="Share recap"><i class="fa-solid fa-share-nodes"></i></button><button class="btn btn-sm btn-outline-primary" onclick="showRecapForm(${r.id})"><i class="fa-solid fa-pen"></i></button><button class="btn btn-sm btn-outline-success" onclick="markRecapSent(${r.id})" ${r.is_sent ? 'disabled' : ''}><i class="fa-solid fa-paper-plane"></i></button><button class="btn btn-sm btn-outline-danger" onclick="deleteRecap(${r.id})"><i class="fa-solid fa-trash"></i></button></div></div></div></div>`).join('');
  } catch (e: any) { el.innerHTML = `<div class="alert alert-danger">${esc(e.message)}</div>`; }
}

export async function showRecapForm(id?: number): Promise<void> {
  const cid = getCid();
  editingId = id ?? null;
  let title = '', content = '', startDate = '', endDate = '';
  if (editingId) {
    try { const r = await api('GET', `/api/recaps/${editingId}`); title = r.title || ''; content = r.content || ''; startDate = dateVal(r.session_start_date); endDate = dateVal(r.session_end_date); } catch { toast('Failed to load recap', true); return; }
  }
  showModal(editingId ? 'Edit Recap' : 'New Recap', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="recapTitle" value="${esc(title)}" placeholder="Session Recap"></div>
    ${recapFormExtras(startDate, endDate)}
    <div class="mb-3"><label class="form-label">Notes</label><div class="editor-toolbar" id="recapToolbar"></div><div id="recapEditor" class="journal-editor" style="min-height:150px;border:1px solid var(--border-light);border-radius:6px;padding:8px"></div></div>
    <button class="ai-generate-btn btn btn-outline-secondary btn-sm mb-3" data-ai-mode="text" data-ai-target="recapEditor" data-ai-hint="Summarize these session notes into a concise campaign recap"><i class="fa-solid fa-wand-magic-sparkles me-1"></i>Generate with AI</button>
    <button class="btn btn-primary w-100" onclick="saveRecap()"><i class="fa-solid fa-save me-1"></i>${editingId ? 'Update' : 'Save'}</button>
  `);
  initEditor(content);
  if (cid) populateRecapSelectors(editingId, cid);
}

export async function saveRecap(): Promise<void> {
  const cid = getCid();
  if (!cid) return;
  const title = (document.getElementById('recapTitle') as HTMLInputElement)?.value || '';
  const content = recapEditor?.getHTML() || '';
  const session_start_date = (document.getElementById('recapStartDate') as HTMLInputElement)?.value || null;
  const session_end_date = (document.getElementById('recapEndDate') as HTMLInputElement)?.value || null;
  const placeIds = getSelectValues('recapPlaces');
  const eventIds = getSelectValues('recapEvents');
  try {
    let recapId = editingId;
    if (editingId) {
      await api('PUT', `/api/recaps/${editingId}`, { title, content, is_edited: true, session_start_date, session_end_date });
    } else {
      const res: any = await api('POST', `/api/campaigns/${cid}/recaps`, { title, content, session_start_date, session_end_date });
      recapId = res.id ?? res.ID ?? null;
    }
    if (recapId) await syncRecapLinks(recapId, placeIds, eventIds);
    destroyEditor(); hideModal(); toast(editingId ? 'Recap updated' : 'Recap saved');
    editingId = null; renderRecaps(cid);
  } catch (e: any) { toast(e.message, true); }
}

export async function deleteRecap(id: number): Promise<void> {
  if (!confirm('Delete this recap?')) return;
  await api('DELETE', `/api/recaps/${id}`);
  toast('Recap deleted'); renderRecaps();
}

export async function markRecapSent(id: number): Promise<void> {
  await api('POST', `/api/recaps/${id}/mark-sent`);
  toast('Marked as sent'); renderRecaps();
}

export function showRecaps(): Promise<void> { return renderRecaps(); }

// shareRecap creates a public share link for a recap and shows the URL modal.
export async function shareRecap(id: number): Promise<void> {
  try {
    const result: any = await api('POST', '/api/share', { entity_type: 'recap', entity_id: id });
    showModal('Share Recap', `
      <p>Anyone with this link can read this recap.</p>
      <div class="input-group mb-3">
        <input class="form-control" id="shareUrl" value="${esc(result.url)}" readonly onclick="this.select()">
        <button class="btn btn-gold" onclick="copyShareUrl()"><i class="fa-solid fa-copy"></i></button>
      </div>
      <div class="d-flex gap-2">
        <button class="btn btn-primary flex-grow-1" onclick="window.open('mailto:?subject=${encodeURIComponent('Campaign recap')}&body=${encodeURIComponent(result.url)}','_blank')"><i class="fa-solid fa-envelope me-1"></i>Email</button>
        <button class="btn btn-outline-secondary" onclick="hideModal()">Close</button>
      </div>
    `);
  } catch (e: any) { toast(e.message, true); }
}

expose('generateAIRecap', generateAIRecap);
expose('renderRecaps', renderRecaps);
expose('showRecapForm', showRecapForm);
expose('saveRecap', saveRecap);
expose('deleteRecap', deleteRecap);
expose('markRecapSent', markRecapSent);
expose('showRecaps', showRecaps);
expose('shareRecap', shareRecap);
