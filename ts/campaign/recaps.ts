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

export async function generateAIRecap(): Promise<void> {
  const cid = getCid();
  if (!cid) return;
  try {
    const r: any = await api('POST', `/api/campaigns/${cid}/recaps/generate-ai`);
    showRecapFormPrefilled(r.title || '', r.content || '');
  } catch (e: any) { toast(e.message, true); }
}

function showRecapFormPrefilled(title: string, content: string): void {
  editingId = null;
  showModal('New Recap (AI Generated)', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="recapTitle" value="${esc(title)}" placeholder="Session Recap"></div>
    <div class="mb-3"><label class="form-label">Notes</label><div class="editor-toolbar" id="recapToolbar"></div><div id="recapEditor" class="journal-editor" style="min-height:150px;border:1px solid var(--border-light);border-radius:6px;padding:8px"></div></div>
    <button class="ai-generate-btn btn btn-outline-secondary btn-sm mb-3" data-ai-mode="text" data-ai-target="recapEditor" data-ai-hint="Summarize these session notes into a concise campaign recap"><i class="fa-solid fa-wand-magic-sparkles me-1"></i>Generate with AI</button>
    <button class="btn btn-primary w-100" onclick="saveRecap()"><i class="fa-solid fa-save me-1"></i>Save</button>
  `);
  initEditor(content);
}

export async function renderRecaps(campaignId?: number): Promise<void> {
  const cid = campaignId ?? getCid();
  if (!cid) { toast('No campaign selected', true); return; }
  showView('recaps');
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
      recaps.map((r: any) => `<div class="card mb-2"><div class="card-body py-2 px-3"><div class="d-flex justify-content-between align-items-start"><div><span class="fw-bold">${esc(r.title)}</span>${r.is_sent ? ' <span class="badge bg-success">Sent</span>' : ''}${r.is_edited ? ' <span class="badge bg-secondary">Edited</span>' : ''}<br><small class="text-muted">${r.word_count || 0} words</small><div class="small mt-1">${r.content}</div></div><div class="d-flex gap-1 flex-shrink-0"><button class="btn btn-sm btn-outline-primary" onclick="showRecapForm(${r.id})"><i class="fa-solid fa-pen"></i></button><button class="btn btn-sm btn-outline-success" onclick="markRecapSent(${r.id})" ${r.is_sent ? 'disabled' : ''}><i class="fa-solid fa-paper-plane"></i></button><button class="btn btn-sm btn-outline-danger" onclick="deleteRecap(${r.id})"><i class="fa-solid fa-trash"></i></button></div></div></div></div>`).join('');
  } catch (e: any) { el.innerHTML = `<div class="alert alert-danger">${esc(e.message)}</div>`; }
}

export async function showRecapForm(id?: number): Promise<void> {
  const cid = getCid();
  editingId = id ?? null;
  let title = '', content = '';
  if (editingId) {
    try { const r = await api('GET', `/api/recaps/${editingId}`); title = r.title || ''; content = r.content || ''; } catch { toast('Failed to load recap', true); return; }
  }
  showModal(editingId ? 'Edit Recap' : 'New Recap', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="recapTitle" value="${esc(title)}" placeholder="Session Recap"></div>
    <div class="mb-3"><label class="form-label">Notes</label><div class="editor-toolbar" id="recapToolbar"></div><div id="recapEditor" class="journal-editor" style="min-height:150px;border:1px solid var(--border-light);border-radius:6px;padding:8px"></div></div>
    <button class="ai-generate-btn btn btn-outline-secondary btn-sm mb-3" data-ai-mode="text" data-ai-target="recapEditor" data-ai-hint="Summarize these session notes into a concise campaign recap"><i class="fa-solid fa-wand-magic-sparkles me-1"></i>Generate with AI</button>
    <button class="btn btn-primary w-100" onclick="saveRecap()"><i class="fa-solid fa-save me-1"></i>${editingId ? 'Update' : 'Save'}</button>
  `);
  initEditor(content);
}

export async function saveRecap(): Promise<void> {
  const cid = getCid();
  if (!cid) return;
  const title = (document.getElementById('recapTitle') as HTMLInputElement)?.value || '';
  const content = recapEditor?.getHTML() || '';
  try {
    if (editingId) await api('PUT', `/api/recaps/${editingId}`, { title, content, is_edited: true });
    else await api('POST', `/api/campaigns/${cid}/recaps`, { title, content });
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

expose('generateAIRecap', generateAIRecap);
expose('renderRecaps', renderRecaps);
expose('showRecapForm', showRecapForm);
expose('saveRecap', saveRecap);
expose('deleteRecap', deleteRecap);
expose('markRecapSent', markRecapSent);
expose('showRecaps', showRecaps);
