// party-subtabs.ts — campaign tabs (locations/npcs/sessions/quests/graph/analytics)
// moved from the character sheet into the party view (reorganize-sheet-tabs change).
import { expose } from './lib/expose';
import { esc, toast } from './lib/dom';
import { api } from './lib/api';
import * as d3 from 'd3';

const PARTY_TABS = ['overview', 'locations', 'npcs', 'sessions', 'quests', 'graph', 'analytics'];
const TAB_LABELS: Record<string, string> = {
  overview: 'Overview', locations: 'Locations', npcs: 'NPCs', sessions: 'Sessions',
  quests: 'Quests', graph: 'Graph', analytics: 'Analytics',
};

async function campaignMembers(): Promise<any[]> {
  let campaigns: any = null;
  try { campaigns = await api('GET', '/api/party'); } catch { campaigns = null; }
  const groups = Array.isArray(campaigns) ? campaigns : (campaigns && campaigns.groups) || [];
  const seen = new Set<number>();
  const members: any[] = [];
  for (const g of groups || []) {
    for (const m of (g.members || [])) {
      if (m && m.id && !seen.has(m.id)) { seen.add(m.id); members.push(m); }
    }
  }
  return members;
}

async function fetchForMembers(resource: string): Promise<Array<{ member: any; items: any[] }>> {
  const members = await campaignMembers();
  const out: Array<{ member: any; items: any[] }> = [];
  for (const m of members) {
    let items: any = [];
    try { items = await api('GET', `/api/characters/${m.id}/${resource}`); } catch { items = []; }
    out.push({ member: m, items: Array.isArray(items) ? items : [] });
  }
  return out;
}

export function renderPartySubTabBar(active: string): void {
  const view = document.getElementById('partyView');
  if (!view) return;
  let bar = document.getElementById('partySubTabBar') as HTMLElement | null;
  if (!bar) {
    bar = document.createElement('div');
    bar.id = 'partySubTabBar';
    bar.className = 'd-flex flex-wrap gap-1 mb-3';
    view.insertBefore(bar, view.querySelector('#partyContent') || view.firstChild);
  }
  bar.innerHTML = PARTY_TABS.map((t) =>
    `<button class="btn btn-sm ${active === t ? 'btn-gold' : 'btn-outline-gold'}" onclick="partySubTab('${t}')">${TAB_LABELS[t]}</button>`,
  ).join('');
}

export async function partySubTab(tab: string): Promise<void> {
  const content = document.getElementById('partyContent') as HTMLElement | null;
  if (!content) return;
  renderPartySubTabBar(tab);
  if (tab === 'overview') { (window as any).showParty?.(); return; }
  content.innerHTML = '<div class="text-center py-4"><i class="fa-solid fa-spinner fa-spin me-1"></i>Loading...</div>';
  try {
    if (tab === 'locations') await renderPartyLocations();
    else if (tab === 'npcs') await renderPartyNPCs();
    else if (tab === 'sessions') await renderPartySessions();
    else if (tab === 'quests') await renderPartyQuests();
    else if (tab === 'graph') await renderPartyGraph();
    else if (tab === 'analytics') await renderPartyAnalytics();
  } catch (e: any) {
    toast(e.message || 'Failed to load', true);
  }
}

export async function renderPartyLocations(): Promise<void> {
  const content = document.getElementById('partyContent');
  if (!content) return;
  const rows = await fetchForMembers('locations');
  const withCoords: any[] = [];
  const blocks = rows
    .filter((r) => r.items.length)
    .map((r) => {
      for (const loc of r.items) {
        const lat = Number(loc.latitude || 0);
        const lng = Number(loc.longitude || 0);
        if (lat !== 0 && lng !== 0) withCoords.push({ name: loc.location_name || loc.name || 'Unknown', lat, lng, char: r.member.name });
      }
      return `<div class="mb-3"><h6 class="text-gold">${esc(r.member.name)}</h6><ul class="list-unstyled">${r.items
        .map((l: any) => {
          const locName = l.location_name || l.name;
          const locType = l.location_type || l.type;
          return `<li class="mb-1"><i class="fa-solid fa-location-dot me-1 text-muted"></i><strong>${esc(locName || 'Unnamed')}</strong>${locType ? ` <span class="badge bg-secondary">${esc(locType)}</span>` : ''}${l.description ? `<div class="small text-muted">${esc(String(l.description).slice(0, 120))}</div>` : ''}</li>`;
        })
        .join('')}</ul></div>`;
    })
    .join('');
  content.innerHTML =
    `<div class="d-flex justify-content-between align-items-center"><h5>Campaign Locations</h5></div>` +
    (withCoords.length
      ? `<div id="partyLocationMap" style="height:300px" class="rounded border mb-3"></div>`
      : '') +
    (blocks || '<p class="text-muted">No locations recorded.</p>');
  if (withCoords.length) {
    try {
      const L: any = (await import('leaflet')).default || (window as any).L;
      if (L && typeof L.map === 'function') {
        const map = L.map('partyLocationMap').setView([withCoords[0].lat, withCoords[0].lng], 5);
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', { maxZoom: 18 }).addTo(map);
        for (const p of withCoords) L.marker([p.lat, p.lng]).addTo(map).bindPopup(`<strong>${esc(p.name)}</strong><br>${esc(p.char)}`);
      }
    } catch { /* tiles unavailable — list is the fallback */ }
  }
}

export async function renderPartyNPCs(): Promise<void> {
  const content = document.getElementById('partyContent');
  if (!content) return;
  const rows = await fetchForMembers('npcs');
  const all: any[] = [];
  rows.forEach((r) => r.items.forEach((n) => all.push({ ...n, charName: r.member.name })));
  content.innerHTML =
    `<div class="d-flex justify-content-between align-items-center mb-2"><h5 class="mb-0">Campaign NPCs</h5><input id="partyNpcFilter" class="form-control form-control-sm" style="max-width:220px" placeholder="Filter by name..."></div><div id="partyNpcList">${renderNpcList(all, '')}</div>`;
  const input = document.getElementById('partyNpcFilter') as HTMLInputElement | null;
  if (input) input.addEventListener('input', () => {
    const list = document.getElementById('partyNpcList');
    if (list) list.innerHTML = renderNpcList(all, input.value.trim().toLowerCase());
  });
}

function renderNpcList(all: any[], q: string): string {
  const filtered = q ? all.filter((n) => (n.npc_name || n.name || '').toLowerCase().includes(q)) : all;
  if (!filtered.length) return '<p class="text-muted">No NPCs found.</p>';
  const byChar = new Map<string, any[]>();
  filtered.forEach((n) => {
    const k = n.charName || 'Unknown';
    if (!byChar.has(k)) byChar.set(k, []);
    byChar.get(k)!.push(n);
  });
  return Array.from(byChar.entries())
    .map(([char, npcs]) =>
      `<div class="mb-3"><h6 class="text-gold">${esc(char)}</h6><ul class="list-unstyled">${npcs
        .map((n) => `<li class="mb-1"><i class="fa-solid fa-user me-1 text-muted"></i><strong>${esc(n.npc_name || n.name || 'Unnamed')}</strong>${n.npc_race ? ` <span class="badge bg-secondary">${esc(n.npc_race)}</span>` : ''}${n.npc_class ? ` <span class="badge bg-secondary">${esc(n.npc_class)}</span>` : ''}</li>`)
        .join('')}</ul></div>`,
    )
    .join('');
}

export async function renderPartySessions(): Promise<void> {
  const content = document.getElementById('partyContent');
  if (!content) return;
  const rows = await fetchForMembers('sessions');
  const all: any[] = [];
  rows.forEach((r) => r.items.forEach((s) => all.push({ ...s, charName: r.member.name })));
  all.sort((a, b) => String(b.session_date || b.date || '').localeCompare(String(a.session_date || a.date || '')));
  content.innerHTML =
    `<h5>Session Log</h5>` +
    (all.length
      ? `<ul class="list-unstyled">${all
          .map((s) => `<li class="mb-2 p-2 border rounded"><i class="fa-solid fa-book me-1 text-gold"></i><strong>${esc(s.session_date || s.date || '')}</strong> — ${esc(s.title || 'Session')} <span class="badge bg-secondary">${esc(s.charName)}</span>${s.notes ? `<div class="small text-muted">${esc(String(s.notes).slice(0, 140))}</div>` : ''}</li>`)
          .join('')}</ul>`
      : '<p class="text-muted">No sessions logged.</p>');
}

export async function renderPartyQuests(): Promise<void> {
  const content = document.getElementById('partyContent');
  if (!content) return;
  const rows = await fetchForMembers('quests');
  const all: any[] = [];
  rows.forEach((r) => r.items.forEach((q) => all.push({ ...q, charName: r.member.name })));
  const byStatus = new Map<string, any[]>();
  all.forEach((q) => {
    const k = q.status || 'active';
    if (!byStatus.has(k)) byStatus.set(k, []);
    byStatus.get(k)!.push(q);
  });
  const order = ['active', 'completed', 'failed'];
  content.innerHTML =
    `<h5>Quests</h5>` +
    (all.length
      ? order
          .filter((s) => byStatus.has(s))
          .map((s) => `<h6 class="text-gold mt-2">${esc(s.charAt(0).toUpperCase() + s.slice(1))}</h6><ul class="list-unstyled">${byStatus
              .get(s)!
              .map((q) => `<li class="mb-1"><i class="fa-solid fa-scroll me-1 text-muted"></i><strong>${esc(q.title || q.name || 'Quest')}</strong> <span class="badge bg-secondary">${esc(q.charName)}</span>${q.notes ? `<div class="small text-muted">${esc(String(q.notes).slice(0, 120))}</div>` : ''}</li>`)
              .join('')}</ul>`)
          .join('') +
        (Array.from(byStatus.keys()).some((k) => !order.includes(k))
          ? `<h6 class="text-gold mt-2">Other</h6><ul class="list-unstyled">${Array.from(byStatus.entries())
              .filter(([k]) => !order.includes(k))
              .flatMap(([, qs]) => qs.map((q) => `<li class="mb-1">${esc(q.title || q.name || 'Quest')} (${esc(q.status || '')})</li>`))
              .join('')}</ul>`
          : '')
      : '<p class="text-muted">No quests yet.</p>');
}

export async function renderPartyGraph(): Promise<void> {
  const content = document.getElementById('partyContent');
  if (!content) return;
  const members = await campaignMembers();
  const npcRows = await fetchForMembers('npcs');
  const locRows = await fetchForMembers('locations');
  const nodes: any[] = [];
  const links: any[] = [];
  const nodeId = new Set<string>();
  const addNode = (id: string, label: string, group: string) => {
    if (!nodeId.has(id)) { nodeId.add(id); nodes.push({ id, label, group }); }
  };
  members.forEach((m) => addNode('c' + m.id, m.name || 'Char', 'char'));
  npcRows.forEach((r) => {
    r.items.forEach((n: any) => {
      const id = 'n' + (n.npc_id || n.id);
      addNode(id, n.npc_name || n.name || 'NPC', 'npc');
      links.push({ source: 'c' + r.member.id, target: id });
    });
  });
  locRows.forEach((r) => {
    r.items.forEach((l: any) => {
      const id = 'l' + (l.location_id || l.id);
      addNode(id, l.location_name || l.name || 'Loc', 'loc');
      links.push({ source: 'c' + r.member.id, target: id });
    });
  });
  content.innerHTML = `<h5>Campaign Graph</h5><div id="partyGraphSvg"></div>`;
  if (!nodes.length) { content.innerHTML += '<p class="text-muted">Nothing to graph yet.</p>'; return; }
  const width = Math.min(900, (document.getElementById('partyContent')?.clientWidth || 800) - 20);
  const height = 420;
  const svg = d3.select('#partyGraphSvg').append('svg').attr('width', width).attr('height', height);
  const color: Record<string, string> = { char: '#c9a227', npc: '#3b82f6', loc: '#22c55e' };
  const sim = d3.forceSimulation(nodes as any)
    .force('link', d3.forceLink(links as any).id((d: any) => d.id).distance(70))
    .force('charge', d3.forceManyBody().strength(-220))
    .force('center', d3.forceCenter(width / 2, height / 2));
  const link = svg.append('g').selectAll('line').data(links).join('line').attr('stroke', '#555').attr('stroke-opacity', 0.5);
  const node = svg.append('g').selectAll('circle').data(nodes).join('circle')
    .attr('r', 8).attr('fill', (d: any) => color[d.group] || '#888').attr('stroke', '#fff').attr('stroke-width', 1);
  node.append('title').text((d: any) => d.label);
  const label = svg.append('g').selectAll('text').data(nodes).join('text')
    .attr('x', 10).attr('y', 3).attr('font-size', 10).text((d: any) => d.label);
  sim.on('tick', () => {
    link.attr('x1', (d: any) => d.source.x).attr('y1', (d: any) => d.source.y).attr('x2', (d: any) => d.target.x).attr('y2', (d: any) => d.target.y);
    node.attr('cx', (d: any) => d.x).attr('cy', (d: any) => d.y);
    label.attr('x', (d: any) => d.x + 10).attr('y', (d: any) => d.y + 3);
  });
  (node as any).call(d3.drag().on('start', (ev: any, d: any) => { if (!ev.active) sim.alphaTarget(0.3).restart(); d.fx = d.x; d.fy = d.y; })
    .on('drag', (ev: any, d: any) => { d.fx = ev.x; d.fy = ev.y; })
    .on('end', (ev: any, d: any) => { if (!ev.active) sim.alphaTarget(0); d.fx = null; d.fy = null; }) as any);
}

export async function renderPartyAnalytics(): Promise<void> {
  const content = document.getElementById('partyContent');
  if (!content) return;
  const [members, sessions, quests, npcs, locs] = await Promise.all([
    campaignMembers(),
    fetchForMembers('sessions'),
    fetchForMembers('quests'),
    fetchForMembers('npcs'),
    fetchForMembers('locations'),
  ]);
  const sessionCount = sessions.reduce((a, r) => a + r.items.length, 0);
  const questCount = quests.reduce((a, r) => a + r.items.length, 0);
  const questCompleted = quests.reduce((a, r) => a + r.items.filter((q: any) => q.status === 'completed').length, 0);
  const npcCount = npcs.reduce((a, r) => a + r.items.length, 0);
  const locCount = locs.reduce((a, r) => a + r.items.length, 0);
  const avgLevel = members.length ? Math.round((members.reduce((a, m) => a + (Number(m.level) || 0), 0) / members.length) * 10) / 10 : 0;
  content.innerHTML =
    `<h5>Campaign Analytics</h5>` +
    `<div class="row g-3 mt-1">
      <div class="col-6 col-md-4"><div class="border rounded p-3 text-center"><div class="fs-3 fw-bold text-gold">${members.length}</div><div class="small text-muted">Characters</div></div></div>
      <div class="col-6 col-md-4"><div class="border rounded p-3 text-center"><div class="fs-3 fw-bold text-gold">${sessionCount}</div><div class="small text-muted">Sessions</div></div></div>
      <div class="col-6 col-md-4"><div class="border rounded p-3 text-center"><div class="fs-3 fw-bold text-gold">${questCount}</div><div class="small text-muted">Quests (${questCompleted} completed)</div></div></div>
      <div class="col-6 col-md-4"><div class="border rounded p-3 text-center"><div class="fs-3 fw-bold text-gold">${npcCount}</div><div class="small text-muted">NPCs</div></div></div>
      <div class="col-6 col-md-4"><div class="border rounded p-3 text-center"><div class="fs-3 fw-bold text-gold">${locCount}</div><div class="small text-muted">Locations</div></div></div>
      <div class="col-6 col-md-4"><div class="border rounded p-3 text-center"><div class="fs-3 fw-bold text-gold">${avgLevel}</div><div class="small text-muted">Avg level</div></div></div>
    </div>`;
}

expose('partySubTab', partySubTab);
expose('renderPartySubTabBar', renderPartySubTabBar);
expose('renderPartyLocations', renderPartyLocations);
expose('renderPartyNPCs', renderPartyNPCs);
expose('renderPartySessions', renderPartySessions);
expose('renderPartyQuests', renderPartyQuests);
expose('renderPartyGraph', renderPartyGraph);
expose('renderPartyAnalytics', renderPartyAnalytics);
