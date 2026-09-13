import { expose } from './lib/expose';
import { esc, toast } from './lib/dom';
import { api } from './lib/api';
import { showView } from './navigation';
import { currentCampaign, setCurrentCampaign } from './lib/state';

let activeCampaignId = 0;
let activeConversationId: number | string | null = null;
let lastAnswerSources: any[] = [];
let lastTranscriptId: number | null = null;

function renderSources(sources: any[]): string {
  if (!sources || !sources.length) return '';
  return sources.map((s: any) => `<a data-testid="copilot-source-link" href="${esc(s.url || '#')}">${esc(s.title || s.entity_type || 'source')}</a>`).join(' ');
}

function sourcesHtml(): string {
  if (!lastAnswerSources.length) return '';
  return lastAnswerSources.map((s: any) => `<a data-testid="copilot-source-link" href="${esc(s.url || '#')}">${esc(s.title || s.entity_type || 'source')}</a>`).join(' ');
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
      </div>
    </div>
  `;
  // restore last answer if any
  const ansEl = document.getElementById('copilotAnswer');
  const srcEl = document.getElementById('copilotSources');
  if (ansEl && (ansEl as any)._copilotAnswer) ansEl.innerHTML = (ansEl as any)._copilotAnswer;
  if (srcEl) srcEl.innerHTML = sourcesHtml();
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
  await refreshCopilotPanel();
  const ansEl = document.getElementById('copilotAnswer');
  const srcEl = document.getElementById('copilotSources');
  if (ansEl && savedAnswer) { ansEl.innerHTML = savedAnswer; (ansEl as any)._copilotAnswer = savedAnswer; }
  if (srcEl) srcEl.innerHTML = renderSources(savedSources);
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
