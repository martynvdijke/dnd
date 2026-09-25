/**
 * Battlemap: the campaign's active map as a tactical surface with draggable
 * tokens. Tokens linked to a combat entry mirror its live HP/AC, so the board
 * is fed by the encounter tracker.
 */
import { expose } from './lib/expose';
import { api } from './lib/api';
import { esc, showModal, hideModal, toast } from './lib/dom';
import { showView } from './navigation';
import { currentCampaign } from './lib/state';

interface BattlemapToken {
  id: number;
  combat_entry_id?: number | null;
  name: string;
  x: number;
  y: number;
  color: string;
  size: number;
  hp_current?: number;
  hp_max?: number;
  ac?: number;
  conditions?: string;
}

interface BattlemapData {
  map: {
    id: number; name: string; image_url: string; width: number; height: number;
    grid_size: number; grid_units: string;
  } | null;
  tokens: BattlemapToken[];
}

function campaignId(): number | null {
  return (currentCampaign as any)?.id ?? null;
}

function gridOverlay(map: BattlemapData['map']): string {
  const w = map?.width || 1000;
  const h = map?.height || 800;
  const g = map?.grid_size || 50;
  const cols = Math.max(1, Math.round(w / g));
  const rows = Math.max(1, Math.round(h / g));
  return `<div style="position:absolute;inset:0;pointer-events:none;background-image:linear-gradient(to right, rgba(255,255,255,.14) 1px, transparent 1px),linear-gradient(to bottom, rgba(255,255,255,.14) 1px, transparent 1px);background-size:${100 / cols}% ${100 / rows}%"></div>`;
}

function tokenHtml(t: BattlemapToken): string {
  const size = 46 * (t.size || 1);
  const hp = t.hp_max != null && t.hp_current != null
    ? `<div style="margin-top:3px;background:#000;border-radius:3px;width:${size}px;height:5px;overflow:hidden"><div style="height:100%;width:${Math.max(0, Math.min(100, Math.round((t.hp_current / Math.max(1, t.hp_max)) * 100)))}%;background:${t.hp_current / Math.max(1, t.hp_max) > 0.5 ? '#2f9e44' : t.hp_current / Math.max(1, t.hp_max) > 0.25 ? '#d9a441' : '#e03131'}"></div></div>
       <span class="bm-hp" style="font-size:10px;color:#fff;text-shadow:0 0 3px #000">${t.hp_current}/${t.hp_max}</span>`
    : '';
  return `
    <div class="bm-token" data-id="${t.id}" data-name="${esc(t.name)}"
         style="position:absolute;left:${t.x * 100}%;top:${t.y * 100}%;transform:translate(-50%,-50%);display:flex;flex-direction:column;align-items:center;cursor:grab;touch-action:none;z-index:2">
      <div style="width:${size}px;height:${size}px;border-radius:50%;background:${esc(t.color)};border:2px solid #111;box-shadow:0 2px 6px rgba(0,0,0,.5);display:flex;align-items:center;justify-content:center;color:#fff;font-weight:700;user-select:none" title="${esc(t.name)}${t.ac != null ? ' · AC ' + t.ac : ''}">${esc(t.name.slice(0, 2).toUpperCase())}</div>
      <span style="font-size:11px;color:#fff;text-shadow:0 0 3px #000;white-space:nowrap">${esc(t.name)}</span>
      ${hp}
      <button class="btn btn-sm btn-outline-danger bm-remove" style="position:absolute;top:-8px;right:-8px;--bs-btn-padding-y:0;--bs-btn-padding-x:4px;--bs-btn-font-size:9px;line-height:1" onclick="event.stopPropagation();battlemapRemoveToken(${t.id})" title="Remove token">×</button>
    </div>`;
}

export async function showBattlemap(campaignIdOverride?: number): Promise<void> {
  showView('battlemap');
  const el = document.getElementById('battlemapContent');
  if (!el) return;
  const cid = campaignIdOverride ?? campaignId();
  if (!cid) {
    el.innerHTML = '<p class="text-muted">Select a campaign first.</p>';
    return;
  }
  let data: BattlemapData;
  try {
    data = await api<BattlemapData>('GET', `/api/campaigns/${cid}/battlemap`);
  } catch (e: any) {
    el.innerHTML = `<p class="text-danger">Could not load the battlemap (${esc(e.message || 'error')}).</p>`;
    return;
  }
  const map = data.map;
  const bg = map?.image_url
    ? `background-image:url('${esc(map.image_url)}');background-size:cover;background-position:center`
    : 'background:#15130f';
  el.innerHTML = `
    <div class="d-flex gap-2 mb-2 align-items-center">
      <button class="btn btn-sm btn-gold" onclick="battlemapSync(${cid})"><i class="fa-solid fa-arrows-to-dot me-1"></i>Sync Combatants</button>
      <button class="btn btn-sm btn-outline-light" onclick="battlemapAddToken(${cid})"><i class="fa-solid fa-plus me-1"></i>Add Token</button>
      ${map ? `<span class="ms-auto small text-muted">${esc(map.name)} · grid ${map.grid_size}${esc(map.grid_units || '')}</span>` : ''}
    </div>
    ${!map ? '<p class="text-muted">No map yet — add one in the World view first.</p>' : ''}
    <div id="battlemapBoard" data-testid="battlemap-board"
         style="position:relative;width:100%;aspect-ratio:${map?.width || 1000}/${map?.height || 800};${bg};border:1px solid var(--bs-border-color);border-radius:8px;overflow:hidden">
      ${map ? gridOverlay(map) : ''}
      ${data.tokens.map(tokenHtml).join('')}
    </div>`;
  attachDrag();
}

function attachDrag(): void {
  const board = document.getElementById('battlemapBoard');
  if (!board) return;
  board.querySelectorAll<HTMLElement>('.bm-token').forEach(tok => {
    tok.addEventListener('pointerdown', (ev) => {
      if ((ev.target as HTMLElement).classList.contains('bm-remove')) return;
      ev.preventDefault();
      tok.setPointerCapture(ev.pointerId);
      tok.style.cursor = 'grabbing';
      const move = (e: PointerEvent) => {
        const rect = board.getBoundingClientRect();
        const x = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width));
        const y = Math.min(1, Math.max(0, (e.clientY - rect.top) / rect.height));
        tok.style.left = `${x * 100}%`;
        tok.style.top = `${y * 100}%`;
        tok.dataset.x = String(x);
        tok.dataset.y = String(y);
      };
      const up = async () => {
        tok.style.cursor = 'grab';
        tok.removeEventListener('pointermove', move);
        tok.removeEventListener('pointerup', up);
        if (tok.dataset.x === undefined) return;
        try {
          await api('PUT', `/api/battlemap-tokens/${tok.dataset.id}`, {
            x: Number(tok.dataset.x), y: Number(tok.dataset.y),
          });
        } catch { /* position stays local until next refresh */ }
      };
      tok.addEventListener('pointermove', move);
      tok.addEventListener('pointerup', up);
    });
  });
}

async function refresh(campaignId: number): Promise<void> {
  await showBattlemap(campaignId);
}

expose('showBattlemap', showBattlemap);

expose('battlemapSync', async function (campaignId: number) {
  try {
    const res = await api<{ created: number }>('POST', `/api/campaigns/${campaignId}/battlemap/sync`, {});
    toast(res.created > 0 ? `Added ${res.created} token(s)` : 'All combatants already placed');
    await refresh(campaignId);
  } catch (e: any) { toast(e.message, true); }
});

expose('battlemapAddToken', async function (campaignId: number) {
  let entries: any[] = [];
  try { entries = await api<any[]>('GET', `/api/combat?campaign_id=${campaignId}`); } catch { }
  const options = entries.map(e => `<option value="${e.id}">${esc(e.name)}${e.hp_max != null ? ` (${e.hp_current}/${e.hp_max})` : ''}</option>`).join('');
  showModal('Add Token', `
    <label class="form-label small">Combatant</label>
    <select id="bmTokenEntry" class="form-select mb-2"><option value="">— free token —</option>${options}</select>
    <label class="form-label small">Name</label>
    <input id="bmTokenName" class="form-control mb-2" placeholder="Optional for a combatant">
    <label class="form-label small">Colour</label>
    <input id="bmTokenColor" type="color" class="form-control form-control-color mb-3" value="#b8963e">
    <button class="btn btn-gold w-100" onclick="battlemapSaveToken(${campaignId})">Add</button>`);
});

expose('battlemapSaveToken', async function (campaignId: number) {
  const entry = (document.getElementById('bmTokenEntry') as HTMLSelectElement)?.value;
  const name = (document.getElementById('bmTokenName') as HTMLInputElement)?.value || '';
  const color = (document.getElementById('bmTokenColor') as HTMLInputElement)?.value || '#b8963e';
  try {
    await api('POST', `/api/campaigns/${campaignId}/battlemap/tokens`, {
      combat_entry_id: entry ? Number(entry) : null, name, color,
      x: 0.5 + (Math.random() - 0.5) * 0.2, y: 0.5 + (Math.random() - 0.5) * 0.2,
    });
    hideModal();
    await refresh(campaignId);
  } catch (e: any) { toast(e.message, true); }
});

expose('battlemapRemoveToken', async function (tokenId: number) {
  if (!confirm('Remove this token? The combat entry is not deleted.')) return;
  try {
    await api('DELETE', `/api/battlemap-tokens/${tokenId}`);
    const cid = campaignId();
    if (cid) await refresh(cid);
  } catch (e: any) { toast(e.message, true); }
});
