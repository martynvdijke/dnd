import { expose } from './lib/expose';
import { esc, toast } from './lib/dom';
import { api } from './lib/api';
import { showView } from './navigation';
import { currentCampaign, setCurrentCampaign } from './lib/state';

let activeCampaignId = 0;
let activeConversationId: number | string | null = null;
let lastAnswerSources: any[] = [];
let lastTranscriptId: number | null = null;

let lastExtractRecap: { title: string; content: string } | null = null;
let lastExtractDrafts: Array<{ id: string; entity_type: string; name: string; status?: string }> = [];
let lastExtractError = '';
let extractLoading = false;

function renderSources(sources: any[]): string {
  if (!sources || !sources.length) return '';
  return sources.map((s: any) => `<a data-testid="copilot-source-link" href="${esc(s.url || '#')}">${esc(s.title || s.entity_type || 'source')}</a>`).join(' ');
}

function sourcesHtml(): string {
  if (!lastAnswerSources.length) return '';
  return lastAnswerSources.map((s: any) => `<a data-testid="copilot-source-link" href="${esc(s.url || '#')}">${esc(s.title || s.entity_type || 'source')}</a>`).join(' ');
}

export function getTranscriptId(): number | null {
  return lastTranscriptId || (window as any).__copilotTranscriptId || null;
}

export function buildExtractErrorMessage(payload: any, fallback = 'Extraction failed'): string {
  if (!payload || typeof payload !== 'object') return fallback;
  const err = typeof payload.error === 'string' ? payload.error.trim() : '';
  const msg = typeof payload.message === 'string' ? payload.message.trim() : '';
  if (err && msg) return `${err}: ${msg}`;
  if (err) return err;
  if (msg) return msg;
  return fallback;
}

export function recapPreviewHtml(recap: { title?: string; content?: string } | null): string {
  if (!recap) return '';
  const title = esc(recap.title || 'Recap');
  const content = esc(recap.content || '').replace(/\n/g, '<br>');
  return `<strong>${title}</strong><div style="white-space:pre-wrap">${content}</div>`;
}

export function extractedDraftRowHtml(draft: { id: string; entity_type: string; name: string; status?: string }): string {
  const idEsc = esc(draft.id);
  const typeEsc = esc(draft.entity_type);
  const nameEsc = esc(draft.name);
  const status = draft.status || 'ready';
  const isDone = status === 'committed' || status === 'discarded';
  const statusBadge = status === 'committed'
    ? `<span class="badge bg-success ms-1">committed</span>`
    : status === 'discarded'
      ? `<span class="badge bg-secondary ms-1">discarded</span>`
      : '';
  // Use single quotes for onclick id arg; escape single quotes in id via replace
  const jsId = String(draft.id).replace(/\\/g, '\\\\').replace(/'/g, "\\'");
  return `<div data-testid="copilot-extracted-draft" data-id="${idEsc}" class="d-flex align-items-center justify-content-between p-2 border rounded mb-1">
    <span class="me-2"><small class="text-muted">${typeEsc}</small> ${nameEsc}${statusBadge}</span>
    <span class="d-flex gap-1 flex-shrink-0">
      <button class="btn btn-sm btn-primary" data-testid="copilot-draft-commit" onclick="copilotCommitExtractedDraft('${jsId}')" ${isDone ? 'disabled' : ''}>Commit</button>
      <button class="btn btn-sm btn-outline-secondary" data-testid="copilot-draft-discard" onclick="copilotDiscardExtractedDraft('${jsId}')" ${isDone ? 'disabled' : ''}>Discard</button>
    </span>
  </div>`;
}

function renderExtractSection(): void {
  const recapEl = document.getElementById('copilotRecapPreview');
  const draftsEl = document.getElementById('copilotExtractedDrafts');
  const errEl = document.getElementById('copilotExtractError');
  if (errEl) {
    if (lastExtractError) {
      errEl.textContent = lastExtractError;
      errEl.classList.remove('d-none');
    } else {
      errEl.textContent = '';
      errEl.classList.add('d-none');
    }
  }
  if (recapEl) {
    if (lastExtractRecap) {
      recapEl.classList.remove('d-none');
      recapEl.setAttribute('data-testid', 'copilot-recap-preview');
      recapEl.innerHTML = recapPreviewHtml(lastExtractRecap);
      // ensure testid persists
      recapEl.setAttribute('data-testid', 'copilot-recap-preview');
    } else {
      recapEl.classList.add('d-none');
      recapEl.innerHTML = '';
    }
  }
  if (draftsEl) {
    if (lastExtractDrafts.length) {
      draftsEl.innerHTML = lastExtractDrafts.map(d => extractedDraftRowHtml(d)).join('');
    } else {
      draftsEl.innerHTML = lastExtractDrafts.length === 0 && lastExtractRecap ? '<small class="text-muted">No entities extracted</small>' : '';
    }
  }
  const btn = document.getElementById('copilotExtractBtn') as HTMLButtonElement | null;
  if (btn) {
    const hasTid = !!(lastTranscriptId || (window as any).__copilotTranscriptId);
    btn.disabled = !hasTid || extractLoading;
    btn.textContent = extractLoading ? 'Extracting…' : 'Recap + extract';
  }
}

expose('showCopilot', async function (campaignId?: number) {
  activeCampaignId = campaignId || (currentCampaign as any)?.id || 0;
  if (!activeCampaignId) {
    const camps: any[] = await api('GET', '/api/campaigns').catch(() => []);
    const c = camps[0] || null;
    activeCampaignId = c?.id || 0;
    if (c && !(currentCampaign as any)?.id) setCurrentCampaign(c);
  } else if ((currentCampaign as any)?.id === activeCampaignId) {
    // already set
  } else {
    const camps: any[] = await api('GET', '/api/campaigns').catch(() => []);
    const c = camps.find((x: any) => x.id === activeCampaignId) || null;
    if (c && (currentCampaign as any)?.id !== activeCampaignId) setCurrentCampaign(c);
  }
  showView('copilot');
  await refreshCopilotPanel();
});

async function refreshCopilotPanel() {
  const el = document.getElementById('copilotContent');
  if (!el) return;
  let conversations: any[] = [];
  try {
    conversations = await api('GET', `/api/campaigns/${activeCampaignId}/copilot/conversations`);
    if (!Array.isArray(conversations)) conversations = [];
  } catch {
    conversations = [];
  }
  const hasTid = !!(lastTranscriptId || (window as any).__copilotTranscriptId);
  el.innerHTML = `
    <div class="mb-3 d-flex gap-2">
      <button class="btn btn-sm btn-outline-primary" data-testid="copilot-new" onclick="copilotNewConversation()">New Conversation</button>
      <button class="btn btn-sm btn-outline-secondary" data-testid="copilot-prep" onclick="showCopilotPrep()">Prep Session</button>
    </div>
    <div data-testid="copilot-conversations" id="copilotConversations" class="mb-3">
      ${conversations.length ? conversations.map((c: any) => `<div data-testid="copilot-conversation" data-id="${c.id}" onclick="copilotLoadConversation(${JSON.stringify(c.id)})" style="cursor:pointer" class="p-1 border rounded mb-1">${esc(c.title || 'Conversation ' + c.id)} <small class="text-muted">(${c.message_count ?? ''})</small></div>`).join('') : '<small class="text-muted">No conversations yet</small>'}
    </div>
    <div data-testid="copilot-history" id="copilotHistory" class="mb-3"></div>
    <div class="mb-3">
      <textarea class="form-control mb-2" id="copilotQuery" data-testid="copilot-query" rows="2" placeholder="Ask the copilot..."></textarea>
      <button class="btn btn-primary btn-sm" data-testid="copilot-ask" id="copilotAsk" onclick="copilotAsk()">Ask</button>
    </div>
    <div data-testid="copilot-answer" id="copilotAnswer" class="p-2 border rounded mb-2" style="white-space:pre-wrap;min-height:2rem"></div>
    <div data-testid="copilot-sources" id="copilotSources" class="mb-3">${sourcesHtml()}</div>
    <div class="mb-3">
      <label class="form-label">Transcript</label>
      <textarea class="form-control mb-2" id="copilotTranscript" data-testid="copilot-transcript" rows="4" placeholder="Paste transcript text..."></textarea>
      <div class="d-flex gap-2">
        <button class="btn btn-sm btn-outline-primary" data-testid="copilot-ingest" onclick="copilotIngestTranscript()">Ingest</button>
        <button class="btn btn-sm btn-outline-secondary" data-testid="copilot-summarize" onclick="copilotSummarizeTranscript()">Summarize</button>
        <button class="btn btn-sm btn-outline-primary" data-testid="copilot-generate-recap-extract" id="copilotExtractBtn" onclick="copilotGenerateRecapExtract()" ${hasTid ? '' : 'disabled'}>${extractLoading ? 'Extracting…' : 'Recap + extract'}</button>
      </div>
      <div id="copilotExtractError" class="mt-2 small text-danger ${lastExtractError ? '' : 'd-none'}">${esc(lastExtractError)}</div>
      <div data-testid="copilot-recap-preview" id="copilotRecapPreview" class="mt-3 p-2 border rounded ${lastExtractRecap ? '' : 'd-none'}" style="white-space:pre-wrap">${lastExtractRecap ? recapPreviewHtml(lastExtractRecap) : ''}</div>
      <div data-testid="copilot-extracted-drafts" id="copilotExtractedDrafts" class="mt-2">${lastExtractDrafts.length ? lastExtractDrafts.map(d => extractedDraftRowHtml(d)).join('') : (lastExtractRecap ? '<small class="text-muted">No entities extracted</small>' : '')}</div>
    </div>
  `;
  // restore last answer if any
  const ansEl = document.getElementById('copilotAnswer');
  const srcEl = document.getElementById('copilotSources');
  if (ansEl && (ansEl as any)._copilotAnswer) ansEl.innerHTML = (ansEl as any)._copilotAnswer;
  if (srcEl) srcEl.innerHTML = sourcesHtml();
  // re-apply extract section state (ensures disabled etc)
  renderExtractSection();
}

expose('refreshCopilotPanel', refreshCopilotPanel);

expose('copilotAsk', async function () {
  const qEl = document.getElementById('copilotQuery') as HTMLTextAreaElement | HTMLInputElement | null;
  const query = qEl?.value?.trim() || '';
  if (!query) { toast('Enter a query', true); return; }
  try {
    const body: any = { query };
    if (activeConversationId) body.conversation_id = activeConversationId;
    const res: any = await api('POST', `/api/campaigns/${activeCampaignId}/copilot/chat`, body);
    if (res.conversation_id) activeConversationId = res.conversation_id;
    lastAnswerSources = res.sources || [];
    const ansEl = document.getElementById('copilotAnswer');
    const srcEl = document.getElementById('copilotSources');
    if (ansEl) {
      const html = esc(res.answer || '').replace(/\n/g, '<br>');
      ansEl.innerHTML = html;
      (ansEl as any)._copilotAnswer = html;
    }
    if (srcEl) {
      srcEl.innerHTML = renderSources(lastAnswerSources);
    }
    await refreshCopilotPanelWrapper();
    // re-apply answer after refresh
    const ans2 = document.getElementById('copilotAnswer');
    const src2 = document.getElementById('copilotSources');
    if (ans2) { const html = esc(res.answer || '').replace(/\n/g, '<br>'); ans2.innerHTML = html; (ans2 as any)._copilotAnswer = html; }
    if (src2) src2.innerHTML = renderSources(lastAnswerSources);
  } catch (e: any) {
    toast(e.message, true);
  }
});

async function refreshCopilotPanelWrapper() {
  // refresh but preserve answer
  const savedAnswer = (document.getElementById('copilotAnswer') as any)?._copilotAnswer || '';
  const savedSources = [...lastAnswerSources];
  const savedRecap = lastExtractRecap;
  const savedDrafts = [...lastExtractDrafts];
  const savedErr = lastExtractError;
  const savedLoading = extractLoading;
  await refreshCopilotPanel();
  const ansEl = document.getElementById('copilotAnswer');
  const srcEl = document.getElementById('copilotSources');
  if (ansEl && savedAnswer) { ansEl.innerHTML = savedAnswer; (ansEl as any)._copilotAnswer = savedAnswer; }
  if (srcEl) srcEl.innerHTML = renderSources(savedSources);
  lastExtractRecap = savedRecap;
  lastExtractDrafts = savedDrafts;
  lastExtractError = savedErr;
  extractLoading = savedLoading;
  renderExtractSection();
}

expose('copilotNewConversation', function () {
  activeConversationId = null;
  lastAnswerSources = [];
  const hist = document.getElementById('copilotHistory');
  if (hist) hist.innerHTML = '';
  const ans = document.getElementById('copilotAnswer');
  if (ans) { ans.innerHTML = ''; (ans as any)._copilotAnswer = ''; }
  const src = document.getElementById('copilotSources');
  if (src) src.innerHTML = '';
  toast('New conversation');
});

expose('copilotLoadConversation', async function (id: number | string) {
  try {
    const conv: any = await api('GET', `/api/campaigns/${activeCampaignId}/copilot/conversations/${id}`);
    activeConversationId = conv.id ?? id;
    const hist = document.getElementById('copilotHistory');
    if (hist) {
      const msgs = conv.messages || [];
      hist.innerHTML = msgs.map((m: any) => `<div class="mb-1"><strong>${esc(m.role)}:</strong> <span style="white-space:pre-wrap">${esc(m.content)}</span></div>`).join('') || '<small class="text-muted">No messages</small>';
    }
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('copilotIngestTranscript', async function () {
  const el = document.getElementById('copilotTranscript') as HTMLTextAreaElement | null;
  const text = el?.value?.trim() || '';
  if (!text) { toast('Enter transcript text', true); return; }
  try {
    const res: any = await api('POST', `/api/campaigns/${activeCampaignId}/copilot/transcript`, { text });
    lastTranscriptId = res.id ?? null;
    (window as any).__copilotTranscriptId = lastTranscriptId;
    // enable extract button without full refresh
    const btn = document.getElementById('copilotExtractBtn') as HTMLButtonElement | null;
    if (btn) btn.disabled = !lastTranscriptId;
    toast('Transcript ingested');
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('copilotSummarizeTranscript', async function () {
  if (!lastTranscriptId && !(window as any).__copilotTranscriptId) {
    toast('Ingest a transcript first', true);
    return;
  }
  const id = lastTranscriptId || (window as any).__copilotTranscriptId;
  try {
    const res: any = await api('POST', `/api/campaigns/${activeCampaignId}/copilot/transcript/summarize`, { id });
    lastAnswerSources = res.sources || [];
    const ansEl = document.getElementById('copilotAnswer');
    const srcEl = document.getElementById('copilotSources');
    if (ansEl) {
      const html = esc(res.answer || '').replace(/\n/g, '<br>');
      ansEl.innerHTML = html;
      (ansEl as any)._copilotAnswer = html;
    }
    if (srcEl) srcEl.innerHTML = renderSources(lastAnswerSources);
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('copilotGenerateRecapExtract', async function () {
  const tid = lastTranscriptId || (window as any).__copilotTranscriptId;
  if (!tid) {
    toast('Ingest a transcript first', true);
    return;
  }
  if (extractLoading) return;
  extractLoading = true;
  lastExtractError = '';
  renderExtractSection();
  // also set button loading text directly for immediate feedback before re-render
  const btn = document.getElementById('copilotExtractBtn') as HTMLButtonElement | null;
  if (btn) { btn.disabled = true; btn.textContent = 'Extracting…'; }
  const path = `/api/campaigns/${activeCampaignId}/copilot/transcript/${tid}/extract`;
  try {
    const data: any = await api('POST', path, {});
    // success shape: {recap:{title,content}, drafts:[{id, entity_type, name}], source:{...}}
    lastExtractRecap = data?.recap && typeof data.recap === 'object' ? { title: String(data.recap.title || ''), content: String(data.recap.content || '') } : null;
    const rawDrafts = Array.isArray(data?.drafts) ? data.drafts : [];
    lastExtractDrafts = rawDrafts.map((d: any) => ({
      id: String(d.id ?? ''),
      entity_type: String(d.entity_type ?? ''),
      name: String(d.name ?? ''),
      status: 'ready',
    })).filter((d: any) => d.id);
    // Empty result without recap and without drafts should not leave blank; keep empty drafts list
    toast('Extraction ready');
  } catch (e: any) {
    const payload = (e as any)?.payload;
    let msg: string;
    if (payload && typeof payload === 'object' && (payload.error || payload.message)) {
      msg = buildExtractErrorMessage(payload, e?.message || 'Extraction failed');
    } else {
      msg = e?.message || 'Extraction failed';
    }
    lastExtractError = msg;
    toast(msg, true);
    // keep any previous recap/drafts cleared on error? Keep prior drafts but show error
    // If error came with empty extraction, ensure drafts cleared
    // Don't clear recap/drafts on 503/502 — leave prior results intact
  } finally {
    extractLoading = false;
    renderExtractSection();
  }
});

expose('copilotCommitExtractedDraft', async function (draftId: string) {
  const id = String(draftId || '').trim();
  if (!id) return;
  try {
    await api('POST', `/api/ai/draft/${encodeURIComponent(id)}/commit`, {});
    lastExtractDrafts = lastExtractDrafts.map(d => d.id === id ? { ...d, status: 'committed' } : d);
    toast('Draft committed');
    renderExtractSection();
  } catch (e: any) {
    toast(e?.message || 'Commit failed', true);
    lastExtractError = e?.message || 'Commit failed';
    renderExtractSection();
  }
});

expose('copilotDiscardExtractedDraft', async function (draftId: string) {
  const id = String(draftId || '').trim();
  if (!id) return;
  try {
    await api('DELETE', `/api/ai/draft/${encodeURIComponent(id)}`, undefined);
    lastExtractDrafts = lastExtractDrafts.map(d => d.id === id ? { ...d, status: 'discarded' } : d);
    toast('Draft discarded');
    renderExtractSection();
  } catch (e: any) {
    toast(e?.message || 'Discard failed', true);
    lastExtractError = e?.message || 'Discard failed';
    renderExtractSection();
  }
});

expose('showCopilotPrep', async function () {
  try {
    const res: any = await api('POST', `/api/campaigns/${activeCampaignId}/copilot/prep`, {});
    lastAnswerSources = res.sources || [];
    const ansEl = document.getElementById('copilotAnswer');
    const srcEl = document.getElementById('copilotSources');
    if (ansEl) {
      const html = esc(res.answer || '').replace(/\n/g, '<br>');
      ansEl.innerHTML = html;
      (ansEl as any)._copilotAnswer = html;
    }
    if (srcEl) srcEl.innerHTML = renderSources(lastAnswerSources);
  } catch (e: any) {
    toast(e.message, true);
  }
});

// test-only helpers to reset internal state between vitest cases
expose('_copilotTestReset', function () {
  lastExtractRecap = null;
  lastExtractDrafts = [];
  lastExtractError = '';
  extractLoading = false;
  lastTranscriptId = null;
  (window as any).__copilotTranscriptId = null;
});
expose('_copilotTestSetTranscript', function (id: number | null) {
  lastTranscriptId = id;
  (window as any).__copilotTranscriptId = id;
});
expose('_copilotTestSetExtract', function (recap: any, drafts: any[], err: string) {
  lastExtractRecap = recap;
  lastExtractDrafts = drafts || [];
  lastExtractError = err || '';
});
