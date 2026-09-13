// ts/knowledge.ts — Party Knowledge panel (campaign sub-view).
// Server-rendered JSON API lives in handlers/knowledge.go; this module renders
// the list/detail and wires mentions, known-by, entity links and bulk reveal.
import { Editor } from '@tiptap/core';
import StarterKit from '@tiptap/starter-kit';
import Placeholder from '@tiptap/extension-placeholder';
import { expose } from './lib/expose';
import { esc, showModal, hideModal, toast } from './lib/dom';
import { api } from './lib/api';
import { showView } from './navigation';
import { currentCampaign, currentUser, setCurrentCampaign } from './lib/state';
import { filterKnowledgeByStatus } from './knowledge-filter';
import type { KnowledgeStatus } from './knowledge-filter';
import { showEntityPicker } from './entity-picker';
import { Mention, parseMentions, renderMentionChips, serializeMentions } from './mentions';

let activeCampaignId = 0;
let activeCampaign: any = null;
let activeKnowledgeId = 0;
let knowledgeEditor: Editor | null = null;

const STATUS_LABEL: Record<string, string> = {
  rumor: 'Rumor', confirmed: 'Confirmed', revealed: 'Revealed', false: 'False',
};
const STATUS_BADGE: Record<string, string> = {
  rumor: 'text-bg-secondary', confirmed: 'text-bg-info', revealed: 'text-bg-success', false: 'text-bg-danger',
};

function canManage(): boolean {
  const u = currentUser as any;
  const c = activeCampaign as any;
  if (!u || !c) return false;
  return u.role === 'admin' || c.user_id === u.id || c.my_role === 'dm';
}

function statusBadge(status: string): string {
  return `<span class="badge ${STATUS_BADGE[status] || 'text-bg-secondary'}">${esc(STATUS_LABEL[status] || status)}</span>`;
}

// ─── Panel ───

expose('showKnowledge', async function (campaignId?: number) {
  activeCampaignId = campaignId || (currentCampaign as any)?.id || 0;
  if (!activeCampaignId) {
    const camps: any[] = await api('GET', '/api/campaigns').catch(() => []);
    activeCampaign = camps[0] || null;
    activeCampaignId = activeCampaign?.id || 0;
  } else if ((currentCampaign as any)?.id === activeCampaignId) {
    activeCampaign = currentCampaign;
  } else {
    const camps: any[] = await api('GET', '/api/campaigns').catch(() => []);
    activeCampaign = camps.find((c: any) => c.id === activeCampaignId) || null;
  }
  // Explicit campaignId deep-links must satisfy the view gate for members,
  // who otherwise have no selected campaign yet.
  if (activeCampaign && (currentCampaign as any)?.id !== activeCampaignId) {
    setCurrentCampaign(activeCampaign);
  }
  showView('knowledge');
  await refreshKnowledgePanel();
});

async function refreshKnowledgePanel() {
  const el = document.getElementById('knowledgeContent');
  if (!el) return;
  el.innerHTML = '<div class="ornament">✧ Loading knowledge... ✧</div>';
  let all: any[] = [];
  try {
    all = await api('GET', `/api/campaigns/${activeCampaignId}/knowledge`);
  } catch (e: any) {
    el.innerHTML = `<div class="empty-state"><p class="small text-muted">${esc(e.message)}</p></div>`;
    return;
  }
  const manage = canManage();
  const options = (['', 'rumor', 'confirmed', 'revealed', 'false'] as KnowledgeStatus[])
    .map(s => `<option value="${esc(s)}" ${s === knowledgeFilter ? 'selected' : ''}>${s === '' ? 'All statuses' : STATUS_LABEL[s]}</option>`)
    .join('');
  const shown = filterKnowledgeByStatus(all, knowledgeFilter);
  el.innerHTML = `
    <div class="d-flex justify-content-between align-items-center flex-wrap gap-2 mb-2">
      <select class="form-select form-select-sm w-auto" data-testid="knowledge-status-filter" onchange="setKnowledgeFilter(this.value)">
        ${options}
      </select>
      ${manage ? '<button class="btn btn-gold btn-sm" onclick="showAddKnowledge()"><i class="fa-solid fa-plus me-1"></i>New Knowledge</button>' : ''}
    </div>
    <div class="row g-3">
      <div class="col-12 col-md-5">
        <div data-testid="knowledge-list">
          ${shown.length ? shown.map(k => `
            <div class="knowledge-card ${k.id === activeKnowledgeId ? 'active' : ''}" data-testid="knowledge-card" data-id="${k.id}" data-status="${esc(k.status)}" onclick="loadKnowledge(${k.id})" style="cursor:pointer">
              <div class="d-flex justify-content-between align-items-start gap-2">
                <span class="fw-bold">${esc(k.title)}</span>
                ${statusBadge(k.status)}
              </div>
              <div class="small text-muted">${k.source ? esc(k.source) : ''} ${manage ? (k.shared ? '· shared' : '· DM only') : ''}</div>
            </div>`).join('') : '<div class="empty-state"><i class="fa-solid fa-lightbulb fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">No knowledge entries.</p></div>'}
        </div>
      </div>
      <div class="col-12 col-md-7" id="knowledgeDetailPane">
        <div class="p-3 text-center text-muted"><i class="fa-solid fa-lightbulb fa-2x mb-2 d-block"></i><p>Select an entry to read it</p></div>
      </div>
    </div>`;
  if (activeKnowledgeId && shown.some(k => k.id === activeKnowledgeId)) {
    await renderKnowledgeDetail(activeKnowledgeId);
  }
}
let knowledgeFilter: KnowledgeStatus = '';

expose('setKnowledgeFilter', function (status: KnowledgeStatus) {
  knowledgeFilter = status || '';
  refreshKnowledgePanel();
});

// ─── Detail ───

expose('loadKnowledge', async function (kid: number) {
  activeKnowledgeId = kid;
  await renderKnowledgeDetail(kid);
});

async function renderKnowledgeDetail(kid: number) {
  const pane = document.getElementById('knowledgeDetailPane');
  if (!pane) return;
  let k: any;
  try {
    k = await api('GET', `/api/knowledge/${kid}`);
  } catch (e: any) {
    pane.innerHTML = `<div class="empty-state"><p class="small text-muted">${esc(e.message)}</p></div>`;
    return;
  }
  const manage = canManage();
  const characters: any[] = manage ? await api('GET', `/api/campaigns/${activeCampaignId}/characters`).catch(() => []) : [];
  const nameOf = (id: number) => characters.find(c => c.id === id)?.name || `Character #${id}`;
  const knownBy: number[] = manage ? await api('GET', `/api/knowledge/${kid}/known-by`).catch(() => []) : [];
  const links: any = manage ? await api('GET', `/api/links/knowledge/${kid}`).catch(() => ({ outgoing: [] })) : { outgoing: [] };

  pane.innerHTML = `
    <div class="p-3" data-testid="knowledge-detail" data-id="${k.id}">
      <div class="d-flex justify-content-between align-items-start flex-wrap gap-2">
        <h4 class="mb-0">${esc(k.title)}</h4>
        <div class="d-flex gap-1 align-items-center">
          ${statusBadge(k.status)}
          ${manage ? `<button class="btn btn-sm btn-outline-primary" onclick="showEditKnowledge(${k.id})"><i class="fa-solid fa-pen"></i></button>
          <button class="btn btn-sm btn-outline-danger" onclick="deleteKnowledge(${k.id})"><i class="fa-solid fa-trash"></i></button>` : ''}
        </div>
      </div>
      ${k.source ? `<div class="small text-muted mt-1">Source: ${esc(k.source)}</div>` : ''}
      <hr>
      <div class="knowledge-body">${renderMentionChips(k.content || '')}</div>
      ${manage ? `
        <hr>
        <div class="d-flex flex-wrap gap-2 align-items-center mb-2">
          <select class="form-select form-select-sm w-auto" onchange="setKnowledgeStatus(${k.id},this.value)">
            ${['rumor', 'confirmed', 'revealed', 'false'].map(s => `<option value="${s}" ${s === k.status ? 'selected' : ''}>${STATUS_LABEL[s]}</option>`).join('')}
          </select>
          <button class="btn btn-sm ${k.shared ? 'btn-success' : 'btn-outline-secondary'}" data-testid="share-toggle" onclick="toggleKnowledgeShare(${k.id},${!k.shared})">
            <i class="fa-solid fa-eye me-1"></i>${k.shared ? 'Shared with party' : 'DM only'}
          </button>
          <button class="btn btn-sm btn-outline-gold" data-testid="bulk-reveal-btn" onclick="bulkRevealKnowledge(${k.id})">
            <i class="fa-solid fa-bullhorn me-1"></i>Reveal to party
          </button>
        </div>
        <div data-testid="known-by-picker" class="mb-2">
          <span class="small fw-bold text-muted me-1">Known by:</span>
          ${knownBy.map(cid => `<span class="badge text-bg-light border me-1">${esc(nameOf(cid))} <a href="#" class="text-danger ms-1" onclick="removeKnownBy(${k.id},${cid});return false">&times;</a></span>`).join('') || '<span class="small text-muted">nobody yet</span>'}
          <button class="btn btn-sm btn-outline-primary ms-1" onclick="addKnownBy(${k.id})"><i class="fa-solid fa-plus"></i></button>
        </div>
        <div data-testid="entity-link-picker" class="mb-2">
          <span class="small fw-bold text-muted me-1">Links:</span>
          ${(links.outgoing || []).map((l: any) => `<a class="badge text-bg-light border me-1" href="${esc(l.target_url || '#')}">${esc(l.target_title || l.target_type)}</a> <a href="#" class="text-danger ms-1" onclick="removeKnowledgeLink(${l.id});return false">&times;</a>`).join('') || '<span class="small text-muted">none</span>'}
          <button class="btn btn-sm btn-outline-primary ms-1" onclick="addKnowledgeLink(${k.id})"><i class="fa-solid fa-link"></i></button>
        </div>` : ''}
    </div>`;
}

expose('setKnowledgeStatus', async function (kid: number, status: string) {
  try {
    await api('PUT', `/api/knowledge/${kid}`, { status });
    toast('Status updated');
    await refreshKnowledgePanel();
  } catch (e: any) { toast(e.message, true); }
});

expose('toggleKnowledgeShare', async function (kid: number, shared: boolean) {
  try {
    await api('PUT', `/api/knowledge/${kid}`, { shared });
    toast(shared ? 'Shared with party' : 'Hidden from party');
    await refreshKnowledgePanel();
  } catch (e: any) { toast(e.message, true); }
});

expose('bulkRevealKnowledge', async function (kid: number) {
  if (!confirm('Reveal this to the whole party? Every character is marked as knowing it.')) return;
  try {
    await api('POST', `/api/knowledge/${kid}/reveal`, {});
    toast('Revealed to party');
    await refreshKnowledgePanel();
  } catch (e: any) { toast(e.message, true); }
});

expose('addKnownBy', function (kid: number) {
  showEntityPicker('Mark as known by', (ref) => {
    if (ref.entity_type !== 'character') { toast('Pick a character', true); return; }
    api('POST', `/api/knowledge/${kid}/known-by`, { character_id: ref.entity_id })
      .then(() => { toast('Added'); return refreshKnowledgePanel(); })
      .catch((e: any) => toast(e.message, true));
  }, 'character');
});

expose('removeKnownBy', async function (kid: number, characterId: number) {
  try {
    await api('DELETE', `/api/knowledge/${kid}/known-by/${characterId}`);
    await refreshKnowledgePanel();
  } catch (e: any) { toast(e.message, true); }
});

expose('addKnowledgeLink', function (kid: number) {
  showEntityPicker('Link entity', (ref) => {
    api('POST', '/api/links', { source_type: 'knowledge', source_id: kid, target_type: ref.entity_type, target_id: ref.entity_id, context: 'manual' })
      .then(() => { toast('Linked'); return refreshKnowledgePanel(); })
      .catch((e: any) => toast(e.message, true));
  });
});

expose('removeKnowledgeLink', async function (linkId: number) {
  try {
    await api('DELETE', `/api/links/${linkId}`);
    await refreshKnowledgePanel();
  } catch (e: any) { toast(e.message, true); }
});

// ─── Create / edit ───

expose('showAddKnowledge', function () {
  openKnowledgeModal();
});

expose('showEditKnowledge', async function (kid: number) {
  try {
    const k = await api('GET', `/api/knowledge/${kid}`);
    openKnowledgeModal(k);
  } catch (e: any) { toast(e.message, true); }
});

function openKnowledgeModal(k?: any) {
  showModal(k ? 'Edit Knowledge' : 'New Knowledge', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="knowledgeTitle" value="${esc(k?.title || '')}"></div>
    <div class="mb-3"><label class="form-label">What is known</label>
      <div class="editor-toolbar" id="knowledgeToolbar"></div>
      <div id="knowledgeEditor" class="journal-editor"></div>
    </div>
    <div class="row g-2 mb-3">
      <div class="col-7"><label class="form-label">Source</label><input class="form-control" id="knowledgeSource" value="${esc(k?.source || '')}" placeholder="Where the party heard it"></div>
      <div class="col-5"><label class="form-label">Status</label>
        <select class="form-select" id="knowledgeStatusInput">
          ${['rumor', 'confirmed', 'revealed', 'false'].map(s => `<option value="${s}" ${s === (k?.status || 'rumor') ? 'selected' : ''}>${STATUS_LABEL[s]}</option>`).join('')}
        </select>
      </div>
    </div>
    <button class="btn btn-outline-primary w-100 mb-2" onclick="knowledgeInsertMention()"><i class="fa-solid fa-at me-1"></i>Insert mention</button>
    <button class="btn btn-primary w-100" onclick="saveKnowledge(${k ? k.id : ''})"><i class="fa-solid fa-save me-1"></i>${k ? 'Update' : 'Create'}</button>
  `);
  initKnowledgeEditor(k ? parseMentions(k.content || '') : undefined);
}

function initKnowledgeEditor(content?: string) {
  setTimeout(() => {
    const el = document.getElementById('knowledgeEditor');
    if (!el) return;
    knowledgeEditor = new Editor({
      element: el,
      extensions: [
        StarterKit.configure({ heading: { levels: [1, 2, 3] } }),
        Placeholder.configure({ placeholder: 'What did the party learn? Use Insert mention to link entities...' }),
        Mention,
      ],
      content: content || '<p></p>',
    });
    const modal = document.getElementById('genericModal');
    modal?.addEventListener('hidden.bs.modal', destroyKnowledgeEditor, { once: true });
  }, 50);
}

function destroyKnowledgeEditor() {
  if (knowledgeEditor) { knowledgeEditor.destroy(); knowledgeEditor = null; }
}

expose('knowledgeInsertMention', function () {
  showEntityPicker('Insert mention', (ref) => {
    knowledgeEditor?.chain().focus().insertContent({
      type: 'mention',
      attrs: { id: ref.entity_id, type: ref.entity_type, name: ref.title },
    }).run();
  });
});

expose('saveKnowledge', async function (editId?: number) {
  const title = (document.getElementById('knowledgeTitle') as HTMLInputElement)?.value?.trim();
  if (!title) { toast('Title required', true); return; }
  const body = {
    title,
    content: serializeMentions(knowledgeEditor?.getHTML() || ''),
    source: (document.getElementById('knowledgeSource') as HTMLInputElement)?.value || '',
    status: (document.getElementById('knowledgeStatusInput') as HTMLSelectElement)?.value || 'rumor',
  };
  try {
    if (editId) {
      await api('PUT', `/api/knowledge/${editId}`, body);
    } else {
      await api('POST', `/api/campaigns/${activeCampaignId}/knowledge`, body);
    }
    destroyKnowledgeEditor();
    hideModal();
    await refreshKnowledgePanel();
    toast(editId ? 'Knowledge updated' : 'Knowledge created');
  } catch (e: any) { toast(e.message, true); }
});

expose('deleteKnowledge', async function (kid: number) {
  if (!confirm('Delete this knowledge entry?')) return;
  try {
    await api('DELETE', `/api/knowledge/${kid}`);
    if (activeKnowledgeId === kid) activeKnowledgeId = 0;
    await refreshKnowledgePanel();
    toast('Knowledge deleted');
  } catch (e: any) { toast(e.message, true); }
});
