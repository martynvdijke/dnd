import { api } from '../lib/api';
import { esc, toast } from '../lib/dom';
import { expose } from '../lib/expose';
import { currentCampaign } from '../lib/state';

function getCid(): number | null { return (currentCampaign as any)?.id ?? null; }

function statusBadge(status: string): string {
  switch (status) {
    case 'promoted': return '<span class="badge bg-success">Promoted</span>';
    case 'reviewed': return '<span class="badge bg-primary">Reviewed</span>';
    case 'transcribed': return '<span class="badge bg-info text-dark">Transcribed</span>';
    case 'failed': return '<span class="badge bg-danger">Failed</span>';
    default: return '<span class="badge bg-secondary">Needs text</span>';
  }
}

function captureCard(cap: any): string {
  const pages = (cap.pages || []).map((p: any) => {
    const img = p.thumbnail_url || p.url || '';
    const page = `<a href="${esc(p.url || '')}" target="_blank" rel="noopener"><img src="${esc(img)}" alt="Page ${p.page_index + 1}" class="ink-page-thumb" style="height:96px;max-width:140px;object-fit:contain;border:1px solid #ccc;background:#fff"></a>`;
    const text = `<textarea class="form-control form-control-sm mb-1" rows="3" data-ink-page="${p.id}" data-testid="ink-page-text" placeholder="Type this page's text, or share a Boox text export">${esc(p.text || '')}</textarea>`;
    return `<div class="d-flex gap-2 mb-1 align-items-start">${page}<div class="flex-grow-1">${text}</div></div>`;
  }).join('');
  const journalBtn = cap.character_id
    ? `<button class="btn btn-sm btn-outline-primary" onclick="inkPromote(${cap.id},'journal')"><i class="fa-solid fa-book me-1"></i>Journal</button>`
    : '';
  return `<div class="card mb-3" data-ink-id="${cap.id}" data-testid="ink-capture">
    <div class="card-body py-2 px-3">
      <div class="d-flex justify-content-between align-items-start mb-2">
        <div><span class="fw-bold">${esc(cap.title || 'Untitled capture')}</span> ${statusBadge(cap.status)}
          <span class="badge bg-light text-dark">${esc(cap.source)}</span>
          ${cap.session_key ? `<small class="text-muted ms-1">${esc(cap.session_key)}</small>` : ''}
          ${cap.error ? `<div class="small text-danger">${esc(cap.error)}</div>` : ''}
        </div>
        <div class="d-flex gap-1 flex-wrap justify-content-end">
          <button class="btn btn-sm btn-outline-secondary" onclick="inkSave(${cap.id})"><i class="fa-solid fa-save me-1"></i>Save text</button>
          <button class="btn btn-sm btn-outline-info" onclick="inkOCR(${cap.id})"><i class="fa-solid fa-wand-magic-sparkles me-1"></i>Retry OCR</button>
          <button class="btn btn-sm btn-outline-success" onclick="inkPromote(${cap.id},'recap')"><i class="fa-solid fa-book-open me-1"></i>Recap</button>
          ${journalBtn}
          <button class="btn btn-sm btn-outline-danger" onclick="inkDelete(${cap.id})" aria-label="Delete capture"><i class="fa-solid fa-trash"></i></button>
        </div>
      </div>
      ${pages}
    </div>
  </div>`;
}

export async function renderInkInbox(campaignId?: number): Promise<void> {
  const cid = campaignId ?? getCid();
  const host = document.getElementById('inkInbox');
  if (!host || !cid) return;
  host.innerHTML = '<div class="text-muted small">Loading ink inbox...</div>';
  try {
    const captures = await api('GET', `/api/ink/captures?campaign_id=${cid}`);
    if (!captures.length) {
      host.innerHTML = `<div class="empty-state" data-testid="ink-empty"><p class="text-muted mb-1">No handwritten pages yet.</p><small class="text-muted">On the Boox, export pages as PNG and share them to Villum, or use the ink canvas in the Android app. They land here for review before anything is published.</small></div>`;
      return;
    }
    host.innerHTML = `<div class="d-flex justify-content-between align-items-center mb-2"><h5 class="mb-0">Ink Inbox (${captures.length})</h5><button class="btn btn-sm btn-outline-secondary" onclick="renderInkInbox()"><i class="fa-solid fa-rotate me-1"></i>Refresh</button></div>` +
      captures.map(captureCard).join('');
  } catch (e: any) {
    host.innerHTML = `<div class="alert alert-danger">${esc(e.message)}</div>`;
  }
}

export async function inkSave(id: number): Promise<void> {
  const card = document.querySelector(`[data-ink-id="${id}"]`);
  if (!card) return;
  const pages = [...card.querySelectorAll<HTMLTextAreaElement>('[data-ink-page]')].map((el) => ({
    id: Number(el.dataset.inkPage),
    text: el.value,
  }));
  try {
    await api('PATCH', `/api/ink/captures/${id}`, { pages });
    toast('Ink text saved');
    renderInkInbox();
  } catch (e: any) { toast(e.message, true); }
}

export async function inkOCR(id: number): Promise<void> {
  try {
    await api('POST', `/api/ink/captures/${id}/ocr`, {});
    toast('Transcription ready for review');
    renderInkInbox();
  } catch (e: any) {
    toast(e.message + ' — type the text manually instead.', true);
  }
}

export async function inkPromote(id: number, target: 'recap' | 'journal'): Promise<void> {
  try {
    const res: any = await api('POST', `/api/ink/captures/${id}/promote`, { target });
    toast(target === 'recap' ? 'Added as an unpublished recap' : 'Added to the character journal');
    if (target === 'recap' && res.id) {
      (window as any).renderRecaps?.();
      return;
    }
    renderInkInbox();
  } catch (e: any) { toast(e.message, true); }
}

export async function inkDelete(id: number): Promise<void> {
  if (!confirm('Delete this ink capture? The page images stay in the media library.')) return;
  try {
    await api('DELETE', `/api/ink/captures/${id}`);
    toast('Capture deleted');
    renderInkInbox();
  } catch (e: any) { toast(e.message, true); }
}

expose('renderInkInbox', renderInkInbox);
expose('inkSave', inkSave);
expose('inkOCR', inkOCR);
expose('inkPromote', inkPromote);
expose('inkDelete', inkDelete);
