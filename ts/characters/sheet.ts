// @ts-nocheck
/**
 * Character sheet rendering — tab bar, section switching, steppers, auto-save.
 * Extracted from app.ts (address-tech-debt-and-ux).
 */
import { expose } from '../lib/expose';
import { currentChar, currentTab, setCurrentTab } from '../lib/state';
import { esc, capitalize, toast } from '../lib/dom';
import { api } from '../lib/api';
import { markDirty, isDirty, isSaving, saveCharacter } from '../lib/save';
import { renderDiceTab } from '../dice';
import { renderStats, renderXPBar } from './stats';
import { renderCombat } from './combat';

declare const htmx: any;

export const sections = ['stats', 'combat', 'spells', 'inventory', 'features', 'feats', 'companions', 'crafting', 'locations', 'npcs', 'sessions', 'quests', 'journal', 'notes', 'graph', 'analytics', 'details', 'dice'];

const htmxTabs = ['spells', 'features', 'feats', 'companions', 'crafting', 'notes'];

export function renderSheet() {
  if (!currentChar) return;
  const c = currentChar;
  const multi = c.classes && c.classes.length > 0;
  const classStr = multi
    ? c.classes.map((cc: any) => `${cc.class}${cc.subclass ? ' (' + cc.subclass + ')' : ''} ${cc.level}`).join(' / ')
    : `${c.class}${c.subclass ? ' (' + c.subclass + ')' : ''}`;
  document.getElementById('sheetName')!.innerHTML = c.portrait_url
    ? `<img src="${esc(c.portrait_url)}" class="character-portrait me-2" alt="">${esc(c.name)}`
    : esc(c.name);
  document.getElementById('sheetSubtitle')!.textContent =
    `${c.race} ${classStr} · Level ${c.level}`;

  const tabBar = document.getElementById('tabBar')!;
  tabBar.innerHTML = sections.map(s => `
    <li class="nav-item"><button class="nav-link ${s === currentTab ? 'active' : ''}" onclick="switchTab('${s}')">${capitalize(s)}</button></li>
  `).join('');

  sections.forEach(s => {
    const el = document.getElementById(s + 'Section')!;
    el.style.display = s === currentTab ? 'block' : 'none';
  });

  renderStats();
  renderCombat();
  (window as any).renderGraph?.();
  (window as any).renderAnalytics?.();
  (window as any).renderCrafting?.();
  (window as any).renderDetails?.();
  renderDiceTab();
  applySheetReadonly();
  (window as any).updateSaveBtnState?.();
}

// Read-only mode for characters without edit rights (linked characters, campaign members)
function sheetCanEdit(): boolean {
  return (window as any).canEditCharacter !== false;
}

function applySheetReadonly() {
  const el = document.getElementById('sheetView');
  if (!el) return;
  const ro = !sheetCanEdit();
  el.classList.toggle('readonly', ro);
  if (ro) {
    el.querySelectorAll('input, textarea, select').forEach((i: any) => { i.disabled = true; });
  }
}

export function switchTab(tab: string) {
  // Save-on-tab-switch: persist any unsaved changes before rendering the next tab
  if (isDirty() && !isSaving()) {
    saveCharacter();
  }
  setCurrentTab(tab);
  renderSheet();
  if (htmxTabs.includes(tab) && currentChar) {
    const el = document.getElementById(tab + 'Section');
    if (el) {
      el.setAttribute('hx-get', `/htmx/${tab}?character_id=${currentChar.id}`);
      el.setAttribute('hx-trigger', 'load');
      el.setAttribute('hx-swap', 'innerHTML');
      el.innerHTML = '<div class="ornament">✧ Loading... ✧</div>';
      htmx.process(el);
    }
  }
  // Client-side rendering for tabs
  if (currentChar) {
    switch (tab) {
      case 'inventory': (window as any).renderInventory?.(); break;
      case 'locations': (window as any).renderLocations?.(); break;
      case 'npcs': (window as any).renderNPCs?.(); break;
      case 'sessions': (window as any).renderSessions?.(); break;
      case 'quests': (window as any).renderQuests?.(); break;
      case 'journal': (window as any).renderJournal?.(); break;
    }
    applySheetReadonly();
  }
}

// ─── Helper: Render a stepper control ───

export function renderStepper(field: string, value: number, delta: number, min?: number, max?: number, label?: string, size?: string): string {
  const sizeClass = size === 'lg' ? 'stepper-lg' : size === 'sm' ? 'currency-stepper' : '';
  const ariaInc = label ? `Increase ${label}` : `Increase ${field}`;
  const ariaDec = label ? `Decrease ${label}` : `Decrease ${field}`;
  return `<span class="stepper ${sizeClass}">
    <button class="stepper-btn" onclick="stepperField('${field}', -${Math.abs(delta)}, ${min ?? 'undefined'}, ${max ?? 'undefined'})" aria-label="${ariaDec}">−</button>
    <span class="stepper-value" onclick="editStepperValue('${field}', this)">${value}</span>
    <button class="stepper-btn" onclick="stepperField('${field}', ${Math.abs(delta)}, ${min ?? 'undefined'}, ${max ?? 'undefined'})" aria-label="${ariaInc}">+</button>
  </span>`;
}

// ─── Auto-save (thin wrappers → centralized ts/lib/save.ts) ───

export function autoSaveField(field: string, el: HTMLElement) {
  const input = el as HTMLInputElement;
  const isCheckbox = input.type === 'checkbox';
  const isTextarea = el.tagName === 'TEXTAREA';
  const raw = isCheckbox ? input.checked : (el as any).value;
  const num = parseFloat(String(raw));
  const finalVal = !isNaN(num) && !isCheckbox && !isTextarea ? num : raw;
  if (!currentChar) return;
  currentChar[field] = finalVal;
  markDirty();
}

export function stepperField(field: string, delta: number, min?: number, max?: number) {
  if (!currentChar) return;
  let val = currentChar[field] ?? 0;
  val += delta;
  if (min !== undefined) val = Math.max(min, val);
  if (max !== undefined) val = Math.min(max, val);
  currentChar[field] = val;
  markDirty();
  renderSheet();
}

export function editStepperValue(field: string, el: HTMLElement) {
  if (!currentChar) return;
  const current = currentChar[field] ?? 0;
  el.innerHTML = `<input type="number" class="form-control stepper-inline-input" value="${current}">`;
  const input = el.querySelector('input')!;
  input.focus();
  input.select();
  const save = () => {
    const parsed = parseInt(input.value);
    if (!isNaN(parsed)) {
      currentChar[field] = parsed;
      markDirty();
    }
    renderSheet();
  };
  input.addEventListener('blur', save);
  input.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') { e.preventDefault(); save(); }
    if (e.key === 'Escape') { renderSheet(); }
  });
}

export async function coinStepper(coin: string, delta: number) {
  if (!currentChar) return;
  const currency = currentChar.currency || {};
  const current = currency[coin] || 0;
  const newVal = Math.max(0, current + delta);
  currency[coin] = newVal;
  currentChar.currency = currency;
  const updates: Record<string, number> = {};
  ['cp','sp','ep','gp','pp'].forEach(c => { updates[c] = currency[c] || 0; });
  try {
    await api('PUT', `/api/characters/${currentChar.id}/currency`, updates);
    toast(`${coin.toUpperCase()} ${delta > 0 ? '+' : ''}${delta}`);
  } catch (e: any) { toast(e.message, true); }
  renderSheet();
}

export function updateXPBar() {
  const container = document.getElementById('xpBarContainer');
  if (container && currentChar) container.innerHTML = renderXPBar(currentChar);
}

export function updateField(field: string, value: any) {
  if (!currentChar) return;
  currentChar[field] = value;
  markDirty();
}

// Window registrations (centralized)
expose('renderSheet', renderSheet);
expose('switchTab', switchTab);
expose('renderStepper', renderStepper);
expose('autoSaveField', autoSaveField);
expose('stepperField', stepperField);
expose('editStepperValue', editStepperValue);
expose('coinStepper', coinStepper);
expose('updateField', updateField);
expose('updateXPBar', updateXPBar);
