import { api } from './lib/api';
import { esc } from './lib/dom';
import { toast } from './lib/dom';
import { expose } from './lib/expose';
import { currentChar, currentCampaign, currentUser } from './lib/state';

let resolvedChar: any = null;
let activeCid = 0;

function isTableVisible(): boolean {
  const el = document.getElementById('tableView');
  return !!el && el.style.display !== 'none';
}

async function resolveCharacter(campaignId: number): Promise<any | null> {
  if (currentChar && (!currentChar.campaign_id || currentChar.campaign_id === campaignId)) return currentChar;
  if (resolvedChar && resolvedChar.campaign_id === campaignId) return resolvedChar;
  try {
    const chars: any[] = await api('GET', '/api/characters');
    const found = chars.find((c: any) => c.campaign_id === campaignId) || chars[0] || null;
    if (found) resolvedChar = found;
    return found;
  } catch {
    return null;
  }
}

function charSummaryHtml(ch: any): string {
  if (!ch) return '<div class="small text-muted">No character found</div>';
  return `<div class="mb-2"><strong>${esc(ch.name)}</strong> ${esc(ch.class || '')} Lv ${esc(String(ch.level ?? ''))} · HP ${esc(String(ch.hp_current ?? ''))}/${esc(String(ch.hp_max ?? ''))} · AC ${esc(String(ch.ac ?? ''))}</div>`;
}

async function renderInitiative(campaignId: number): Promise<string> {
  try {
    const data: any = await api('GET', `/api/combat/current-turn?campaign_id=${campaignId}`);
    const cur = data?.current;
    if (!cur) return '<div class="small text-muted">—</div>';
    const name = esc(cur.name || cur.character_name || 'Unknown');
    const init = esc(String(cur.initiative_roll ?? cur.initiative_mod ?? ''));
    const order = esc(String(cur.turn_order ?? ''));
    return `<div>${name} · ${init} / ${order}</div>`;
  } catch {
    return '<div class="small text-muted">—</div>';
  }
}

async function renderHandouts(campaignId: number): Promise<string> {
  try {
    const entries: any[] = await api('GET', `/api/campaigns/${campaignId}/knowledge`);
    const shared = entries.filter((e: any) => e.shared);
    if (!shared.length) return '<div class="small text-muted">No shared handouts yet</div>';
    return shared.map((e: any) => `<div class="card mb-1"><div class="card-body py-1 px-2"><strong>${esc(e.title)}</strong><div class="small text-muted">${esc(e.content || '')}</div></div></div>`).join('');
  } catch {
    return '<div class="small text-muted">No shared handouts yet</div>';
  }
}

export async function showTable(campaignId?: number): Promise<void> {
  let cid = campaignId || (currentCampaign as any)?.id;
  if (!cid) {
    try {
      const camps: any[] = await api('GET', '/api/campaigns');
      cid = camps[0]?.id;
    } catch {}
  }
  if (!cid) {
    toast('Select a campaign and character first');
    return;
  }
  const ch = await resolveCharacter(cid);
  if (!ch) {
    toast('Select a campaign and character first');
    return;
  }
  activeCid = cid;
  const view = document.getElementById('tableView');
  const body = document.getElementById('tableBody');
  if (!view || !body) return;
  view.style.display = '';
  const initiativeHtml = await renderInitiative(cid);
  const handoutsHtml = await renderHandouts(cid);
  body.innerHTML = `
    ${charSummaryHtml(ch)}
    <div class="d-flex gap-2 mb-2">
      <input class="form-control form-control-sm" data-testid="table-dice-input" placeholder="e.g. 1d20+5" value="1d20">
      <button class="btn btn-primary btn-sm" data-testid="table-roll-btn" onclick="rollFromTable()">Roll</button>
      <button class="btn btn-outline-secondary btn-sm" onclick="deactivateSessionMode()">Exit</button>
    </div>
    <div data-testid="table-roll-feed" class="mb-2" style="min-height:1.5rem;border:1px solid var(--border, #ddd);border-radius:4px;padding:0.5rem;max-height:120px;overflow:auto"></div>
    <div class="mb-2"><strong>Initiative</strong><div data-testid="table-initiative">${initiativeHtml}</div></div>
    <div class="mb-2"><strong>Handouts</strong><div data-testid="table-handouts">${handoutsHtml}</div></div>
  `;
}

export function hideTable(): void {
  const el = document.getElementById('tableView');
  if (el) el.style.display = 'none';
}

export async function rollFromTable(): Promise<void> {
  const input = document.querySelector('[data-testid="table-dice-input"]') as HTMLInputElement | null;
  const expr = input?.value?.trim();
  if (!expr) return;
  const cid = activeCid || (currentCampaign as any)?.id;
  const ch = cid ? await resolveCharacter(cid) : null;
  if (!ch) { toast('Select a campaign and character first'); return; }
  try {
    const res: any = await api('POST', '/api/roll', { expression: expr, character_id: ch.id });
    // Append locally; handleLiveRoll will skip own user rolls to avoid duplicate
    const feed = document.querySelector('[data-testid="table-roll-feed"]');
    if (feed) {
      const total = res?.total ?? res?.result ?? '';
      const text = res?.text ?? '';
      const line = document.createElement('div');
      line.textContent = `${expr} = ${total}${text ? ' ' + text : ''}`;
      feed.appendChild(line);
    }
  } catch (e: any) {
    toast(e.message, true);
  }
}

export function handleLiveRoll(payload: any): void {
  if (!isTableVisible()) return;
  const uid = (currentUser as any)?.id;
  if (uid && payload?.user_id === uid) return;
  const feed = document.querySelector('[data-testid="table-roll-feed"]');
  if (!feed) return;
  const line = document.createElement('div');
  line.textContent = `${payload.username} · ${payload.expression} = ${payload.total}`;
  // esc not needed for textContent but keep consistent
  feed.appendChild(line);
}

export async function refreshTableInitiative(): Promise<void> {
  if (!isTableVisible()) return;
  const cid = activeCid || (currentCampaign as any)?.id;
  if (!cid) return;
  const el = document.querySelector('[data-testid="table-initiative"]');
  if (!el) return;
  el.innerHTML = await renderInitiative(cid);
}

export async function refreshTableHandouts(): Promise<void> {
  if (!isTableVisible()) return;
  const cid = activeCid || (currentCampaign as any)?.id;
  if (!cid) return;
  const el = document.querySelector('[data-testid="table-handouts"]');
  if (!el) return;
  el.innerHTML = await renderHandouts(cid);
}

expose('showTable', showTable);
expose('hideTable', hideTable);
expose('rollFromTable', rollFromTable);
expose('handleLiveRoll', handleLiveRoll);
expose('refreshTableInitiative', refreshTableInitiative);
expose('refreshTableHandouts', refreshTableHandouts);
