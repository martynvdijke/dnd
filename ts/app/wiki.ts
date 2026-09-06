// @ts-nocheck — split from monolith
import * as bootstrap from 'bootstrap';
import { expose } from '../lib/expose';
import { esc, attrEscape, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';
import { showView } from '../navigation';
import { renderMarkdown } from '../lib/markdown';

// ─── Wiki ───

expose('showWiki', async function (campaignId?: number) {
  showView('wiki');
  const el = document.getElementById('wikiContent')!;
  el.innerHTML = '<div class="ornament">✧ Loading wiki... ✧</div>';
  try {
    const campaigns = await api('GET', '/api/campaigns');
    if (!campaigns.length) {
      el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-book fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Campaigns</p><p class="small text-muted">Create a campaign to start building your campaign wiki.</p></div>';
      return;
    }
    const cid = campaignId || campaigns[0].id;
    const camp = campaigns.find((c: any) => c.id === cid);
    const pages = await api('GET', `/api/campaigns/${cid}/wiki`);

    const rootPages = pages.filter((p: any) => !p.parent_id);
    const childMap: Record<number, any[]> = {};
    pages.forEach((p: any) => {
      if (p.parent_id) {
        if (!childMap[p.parent_id]) childMap[p.parent_id] = [];
        childMap[p.parent_id].push(p);
      }
    });

    let sidebarHtml = '<div class="list-group list-group-flush">';
    for (const p of rootPages) {
      sidebarHtml += `<a href="#" class="list-group-item list-group-item-action py-1" onclick="loadWikiPage(${p.id});if(window.innerWidth<768){const o=document.getElementById('wikiOffcanvas');if(o){bootstrap.Offcanvas.getInstance(o)?.hide()}}">${attrEscape(p.title)}</a>
        ${buildWikiChildren(p.id, childMap, 1)}`;
    }
    sidebarHtml += '</div>';

    if (!rootPages.length) {
      el.innerHTML = `
        <div class="d-flex justify-content-between align-items-center mb-3">
          <h4 class="mb-0"><i class="fa-solid fa-book me-2"></i>${esc(camp?.name || 'Wiki')}</h4>
          <div class="d-flex gap-1">
            <button class="btn btn-gold btn-sm" onclick="showAddWikiPage(${cid})"><i class="fa-solid fa-plus me-1"></i>New Page</button>
            <button class="btn btn-outline-gold btn-sm" onclick="showCampaignGraph(${cid})"><i class="fa-solid fa-project-diagram me-1"></i>Graph</button>
          </div>
        </div>
        <div class="empty-state"><i class="fa-solid fa-book-open fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">Empty Wiki</p><p class="small text-muted">Start building your campaign lore by creating pages.</p></div>`;
      return;
    }

    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center mb-3 flex-wrap gap-2">
        <h4 class="mb-0"><i class="fa-solid fa-book me-2"></i>${esc(camp?.name || 'Wiki')}</h4>
        <div class="d-flex gap-1">
          <button class="btn btn-gold btn-sm" onclick="showAddWikiPage(${cid})"><i class="fa-solid fa-plus me-1"></i>New Page</button>
          <button class="btn btn-outline-gold btn-sm" onclick="showCampaignGraph(${cid})"><i class="fa-solid fa-project-diagram me-1"></i>Graph</button>
        </div>
      </div>
      <div class="row g-0" style="min-height:500px">
        <div class="col-md-3 d-none d-md-block" style="overflow-y:auto;max-height:70vh;border-right:1px solid var(--border)">
          <div class="p-2"><small class="fw-bold text-muted">PAGES</small></div>
          ${sidebarHtml}
        </div>
        <div class="offcanvas offcanvas-start" id="wikiOffcanvas" tabindex="-1">
          <div class="offcanvas-header border-bottom">
            <h5 class="offcanvas-title">${esc(camp?.name || 'Wiki')} Pages</h5>
            <button type="button" class="btn-close" data-bs-dismiss="offcanvas"></button>
          </div>
          <div class="offcanvas-body p-0">
            <div class="p-2 border-bottom"><small class="fw-bold text-muted">PAGES</small></div>
            ${sidebarHtml}
          </div>
        </div>
        <div class="col-12 col-md-9" id="wikiPageContent">
          <div class="d-flex d-md-none gap-1 mb-2">
            <button class="btn btn-outline-primary btn-sm" onclick="toggleWikiSidebar()"><i class="fa-solid fa-bars me-1"></i> Pages</button>
          </div>
          <div class="p-3 text-center text-muted"><i class="fa-solid fa-book-open fa-2x mb-2 d-block"></i><p>Select a page from the sidebar</p></div>
        </div>
      </div>`;
  } catch (e: any) { el.innerHTML = `<div class="empty-state"><p class="small text-muted">Error: ${esc(e.message)}</p></div>`; }
});

expose('toggleWikiSidebar', function () {
  const offcanvas = document.getElementById('wikiOffcanvas');
  if (offcanvas) bootstrap.Offcanvas.getOrCreateInstance(offcanvas).toggle();
});

function buildWikiChildren(parentId: number, childMap: Record<number, any[]>, depth: number): string {
  const children = childMap[parentId] || [];
  if (!children.length) return '';
  const pad = depth * 16;
  return children.map((c: any) =>
    `<a href="#" class="list-group-item list-group-item-action py-1 ps-${3 + depth}" style="padding-left:${pad + 16}px!important;font-size:0.9rem" onclick="loadWikiPage(${c.id});if(window.innerWidth<768){const o=document.getElementById('wikiOffcanvas');if(o){bootstrap.Offcanvas.getInstance(o)?.hide()}}">↳ ${attrEscape(c.title)}</a>
    ${buildWikiChildren(c.id, childMap, depth + 1)}`
  ).join('');
}

expose('loadWikiPage', async function (pageId: number) {
  try {
    const page = await api('GET', `/api/wiki/${pageId}`);
    const el = document.getElementById('wikiPageContent')!;
    const renderContent = renderMarkdown(page.content);
    el.innerHTML = `
      <div class="p-3">
        <div class="d-flex justify-content-between align-items-start flex-wrap gap-2">
          <h3 class="mb-0">${esc(page.title)}</h3>
          <div class="d-flex gap-1">
            <button class="btn btn-sm btn-outline-primary js-edit-wiki" data-id="${page.id}" data-visibility="${esc(page.visibility)}"><i class="fa-solid fa-pen"></i></button>
            <button class="btn btn-sm btn-outline-danger" onclick="deleteWikiPage(${page.id})"><i class="fa-solid fa-trash"></i></button>
          </div>
        </div>
        <hr>
        <div class="wiki-content">${renderContent}</div>
        <div class="small text-muted mt-3">Updated: ${page.updated_at}</div>
      </div>`;
    // cache page content to avoid interpolation XSS
    (el as any)._wikiPage = page;
    const editBtn = el.querySelector<HTMLButtonElement>('.js-edit-wiki');
    if (editBtn) {
      editBtn.addEventListener('click', () => {
        const p = (el as any)._wikiPage;
        (window as any).showEditWikiPage(p.id, p.title, p.content, p.visibility);
      });
    }
  } catch (e: any) { toast(e.message, true); }
});

expose('showAddWikiPage', function (campaignId: number) {
  showModal('New Wiki Page', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="wikiTitle"></div>
    <div class="mb-3"><label class="form-label">Content (Markdown)</label><textarea class="form-control" id="wikiContent" rows="8" placeholder="Write in Markdown..."></textarea></div>
    <div class="mb-3"><label class="form-label">Visibility</label>
      <select class="form-select" id="wikiVis"><option value="public">Public</option><option value="dm-only">DM Only</option></select></div>
    <button class="btn btn-primary w-100" onclick="saveWikiPage(${campaignId})">Create Page</button>
  `);
});

expose('saveWikiPage', async function (campaignId: number) {
  try {
    await api('POST', `/api/campaigns/${campaignId}/wiki`, {
      campaign_id: campaignId,
      title: (document.getElementById('wikiTitle') as HTMLInputElement).value,
      content: (document.getElementById('wikiContent') as HTMLTextAreaElement).value,
      visibility: (document.getElementById('wikiVis') as HTMLSelectElement).value,
      tags: '[]',
      sort_order: 0,
    });
    hideModal();
    (window as any).showWiki(campaignId);
    toast('Wiki page created');
  } catch (e: any) { toast(e.message, true); }
});

expose('showEditWikiPage', function (id: number, title: string, content: string, visibility: string) {
  showModal('Edit Wiki Page', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="wikiTitle" value="${esc(title)}"></div>
    <div class="mb-3"><label class="form-label">Content (Markdown)</label><textarea class="form-control" id="wikiContent" rows="8">${esc(content)}</textarea></div>
    <div class="mb-3"><label class="form-label">Visibility</label>
      <select class="form-select" id="wikiVis"><option value="public" ${visibility === 'public' ? 'selected' : ''}>Public</option><option value="dm-only" ${visibility === 'dm-only' ? 'selected' : ''}>DM Only</option></select></div>
    <button class="btn btn-primary w-100" onclick="saveEditWikiPage(${id})">Save</button>
  `);
});

expose('saveEditWikiPage', async function (id: number) {
  try {
    const page = await api('GET', `/api/wiki/${id}`);
    await api('PUT', `/api/wiki/${id}`, {
      ...page,
      title: (document.getElementById('wikiTitle') as HTMLInputElement).value,
      content: (document.getElementById('wikiContent') as HTMLTextAreaElement).value,
      visibility: (document.getElementById('wikiVis') as HTMLSelectElement).value,
    });
    hideModal();
    (window as any).loadWikiPage(id);
    toast('Wiki page updated');
  } catch (e: any) { toast(e.message, true); }
});

expose('deleteWikiPage', async function (id: number) {
  if (!confirm('Delete this wiki page?')) return;
  try {
    await api('DELETE', `/api/wiki/${id}`);
    const cid = await api('GET', '/api/campaigns').then((cs: any[]) => cs[0]?.id);
    (window as any).showWiki(cid);
    toast('Wiki page deleted');
  } catch (e: any) { toast(e.message, true); }
});

// ─── Campaign Graph ───

expose('showCampaignGraph', async function (campaignId: number) {
  const modalEl = document.getElementById('genericModal')!;
  const dialogEl = modalEl.querySelector('.modal-dialog') as HTMLElement;
  const origClass = dialogEl.className;
  dialogEl.className = 'modal-dialog modal-xl modal-dialog-scrollable';
  showModal('Campaign Web', `
    <div id="campaignGraphContainer" style="width:100%;height:600px;border:1px solid var(--border);border-radius:4px;background:var(--parchment-light)"></div>
    <div class="text-center mt-2"><small class="text-muted" id="campaignGraphStats">Loading all connections...</small></div>
  `);
  try {
    const data = await api('GET', `/api/campaigns/${campaignId}/graph`);
    const container = document.getElementById('campaignGraphContainer')!;
    (window as any).createForceGraph(container, data, {
      campaign: { shape: 'ellipse', color: '#8b0000' },
      character: { shape: 'ellipse', color: '#8b0000' },
      location: { shape: 'square', color: '#b8963e' },
      npc: { shape: 'diamond', color: '#2d6a2d' },
      quest: { shape: 'star', color: '#8b4513' },
      session: { shape: 'dot', color: '#5c3a2a' },
      wiki: { shape: 'hexagon', color: '#b8963e' },
      faction: { shape: 'triangle', color: '#9b59b6' },
      encounter: { shape: 'dot', color: '#e67e22' },
      timeline: { shape: 'dot', color: '#5c3a2a' },
      calendar: { shape: 'dot', color: '#b8963e' },
    }, { linkDistance: 250, chargeStrength: -400 });
    document.getElementById('campaignGraphStats')!.innerHTML =
      `${data.nodes.length} entities &middot; ${data.edges.length} connections`;
  } catch (e:any) {
    const container = document.getElementById('campaignGraphContainer');
    if (container) container.innerHTML = `<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">${esc(e.message)}</p></div>`;
  }
  modalEl.addEventListener('hidden.bs.modal', function restore() {
    dialogEl.className = origClass;
    modalEl.removeEventListener('hidden.bs.modal', restore);
  }, { once: true });
});
