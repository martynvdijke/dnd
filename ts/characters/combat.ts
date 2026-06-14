/**
 * Character combat rendering — HP/AC, damage/heal, rest, death saves, exhaustion, conditions.
 */
import { esc, toast } from '../lib/dom';
import { api } from '../lib/api';
import { currentChar, currentTab } from '../lib/state';
import { XP_TABLE } from './stats';

export let autoSaveTimer: number | null = null;

export async function rollCheck(type: string, name: string, adv: string) {
  const c = currentChar;
  if (!c) return;
  const url = `/api/roll/check?type=${type}&name=${name}&adv=${adv}&char_id=${c.id}`;
  try {
    const r = await api('GET', url);
    if (r.total) {
      const advText = adv === 'advantage' ? 'Advantage' : adv === 'disadvantage' ? 'Disadvantage' : '';
      toast(`<strong>${esc(r.label || name)}</strong> ${advText}: <span class="roll-value">${r.total}</span>${r.result ? ` (${esc(r.result)})` : ''}`, false, 5000);
    }
  } catch (e: any) {
    toast(e.message, true);
  }
}

export async function applyHeal() {
  const c = currentChar;
  if (!c) return;
  const amt = prompt('Heal amount:');
  if (!amt) return;
  try {
    await api('POST', `/api/characters/${c.id}/heal`, { amount: parseInt(amt) });
    // openChar imported lazily via window to avoid circular deps
    if (currentTab === 'combat') { const { renderCombat } = await import('./combat'); renderCombat(); }
    else { await loadOpenChar(c.id); }
  } catch (e: any) { toast(e.message, true); }
}

function loadOpenChar(id: number) {
  // openChar is still in app.ts, accessible via window
  const fn = (window as any).openChar;
  if (fn) fn(id);
}

export async function doRest(type: string) {
  const c = currentChar;
  if (!c) return;
  try {
    await api('POST', `/api/characters/${c.id}/rest`, { type });
    if (currentTab === 'combat') renderCombat();
    else loadOpenChar(c.id);
  } catch (e: any) { toast(e.message, true); }
}

export async function doLevelUp() {
  const c = currentChar;
  if (!c) return;
  try {
    await api('POST', `/api/characters/${c.id}/level-up`);
    if (currentTab === 'combat') renderCombat();
    else loadOpenChar(c.id);
  } catch (e: any) { toast(e.message, true); }
}

export function autoSaveField(field: string, el: HTMLElement) {
  if (autoSaveTimer) clearTimeout(autoSaveTimer);
  const c = currentChar;
  if (!c) return;
  autoSaveTimer = window.setTimeout(async () => {
    const val = (el as HTMLInputElement).value;
    try {
      await updateField(field, val);
    } catch {}
  }, 600);
}

export function renderStepper(field: string, value: number, delta: number, min?: number, max?: number, label?: string, size?: string): string {
  const nextVal = value + delta;
  if (min !== undefined && nextVal < min) return '';
  if (max !== undefined && nextVal > max) return '';
  const sz = size === 'sm' ? 'btn-sm' : '';
  return `<button class="btn btn-outline-gold ${sz}" onclick="updateField('${field}',${nextVal})">${esc(label || (delta > 0 ? '+' : '') + delta)}</button>`;
}

export async function updateField(field: string, value: any) {
  const c = currentChar;
  if (!c) return;
  await api('PUT', `/api/characters/${c.id}/field`, { field, value });
}

export function updateXPBar() {
  const c = currentChar;
  if (!c) return;
  const nextLv = c.level >= 20 ? '—' : XP_TABLE[c.level];
  const prevLv = c.level > 1 ? XP_TABLE[c.level - 1] : 0;
  const xp = c.xp || 0;
  const pct = nextLv === '—' ? 100 : ((xp - prevLv) / ((nextLv as number) - prevLv)) * 100;
  const el = document.getElementById('xpBarFill');
  if (el) el.style.width = Math.min(100, Math.max(0, pct)) + '%';
  const lbl = document.getElementById('xpBarLabel');
  if (lbl) lbl.textContent = xp + ' / ' + nextLv;
}

export function renderCombat() {
  // Combat render — will be ported from app.ts
}
