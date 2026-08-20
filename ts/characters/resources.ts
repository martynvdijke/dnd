// @ts-nocheck
/**
 * Character resources & wealth rendering — coin counters (gold, silver, etc.)
 * and tracked resources (rations, arrows, ki points, spell charges...).
 * Backed by the character_resources API; the backend existed but had no UI.
 */
import { expose } from '../lib/expose';
import { currentChar } from '../lib/state';
import { esc, toast, showModal, hideModal } from '../lib/dom';
import { api } from '../lib/api';

const COINS = ['cp', 'sp', 'ep', 'gp', 'pp'];
const GP_VALUES: Record<string, number> = { cp: 0.01, sp: 0.1, ep: 0.5, gp: 1, pp: 10 };

const ICONS = [
  'fa-bolt', 'fa-fire', 'fa-droplet', 'fa-heart', 'fa-star', 'fa-coins',
  'fa-drumstick-bite', 'fa-mug-hot', 'fa-campground', 'fa-box', 'fa-flask',
  'fa-shield-halved', 'fa-wand-sparkles', 'fa-gem', 'fa-scroll', 'fa-crosshairs',
];

let resources: any[] = [];
let editingId: number | null = null;

export function wealthTotalGp(): number {
  const cur = (currentChar && currentChar.currency) || {};
  return COINS.reduce((s, coin) => s + ((cur[coin] || 0) * GP_VALUES[coin]), 0);
}

async function loadResources() {
  if (!currentChar) return;
  try {
    resources = await api('GET', `/api/characters/${currentChar.id}/resources`);
  } catch { resources = []; }
}

export async function renderResources() {
  if (!currentChar) return;
  const el = document.getElementById('resourcesSection');
  if (!el) return;
  await loadResources();
  const cur = currentChar.currency || {};
  const total = wealthTotalGp();
  el.innerHTML = `
    <div class="d-flex justify-content-between align-items-center">
      <h5><i class="fa-solid fa-coins me-2 text-gold"></i>Wealth</h5>
      <span class="badge badge-gold fs-6" data-testid="resources-total" title="Total value in gold pieces">${total.toLocaleString(undefined, { maximumFractionDigits: 2 })} gp</span>
    </div>
    <div class="row g-3">
      ${COINS.map(coin => `
        <div class="col-4 col-md-2">
          <label class="form-label small">${coin.toUpperCase()}</label>
          <div class="currency-stepper">
            <button class="stepper-btn" onclick="coinStepper('${coin}', -1)" aria-label="Decrease ${coin.toUpperCase()}">−</button>
            <span class="stepper-value" id="coin${coin}">${cur[coin] || 0}</span>
            <button class="stepper-btn" onclick="coinStepper('${coin}', 1)" aria-label="Increase ${coin.toUpperCase()}">+</button>
          </div>
        </div>`).join('')}
    </div>
    <div class="d-flex justify-content-between align-items-center mt-4 mb-2">
      <h5 class="mb-0"><i class="fa-solid fa-boxes-stacked me-2 text-gold"></i>Resources</h5>
      <button class="btn btn-primary btn-sm" onclick="openResourceForm()" data-testid="resource-add"><i class="fa-solid fa-plus me-1"></i>Add Resource</button>
    </div>
    <div class="small text-muted mb-2" id="resourceHint"></div>
    <div id="resourceList">
      ${resources.length
        ? resources.map(r => renderResourceRow(r)).join('')
        : '<div class="empty-state"><i class="fa-solid fa-box-open fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Resources</p><p class="small text-muted">Track rations, arrows, ki points, spell charges and more.</p></div>'}
    </div>`;
  document.getElementById('resourceHint')!.innerHTML = resources.some(r => r.short_rest_recovery > 0 || r.long_rest_recovery > 0)
    ? '<i class="fa-solid fa-mug-hot me-1"></i>Resources with rest recovery refill automatically on Short/Long Rest.'
    : '';
}

function renderResourceRow(r: any): string {
  const hasMax = r.max > 0;
  return `
    <div class="resource-row" data-testid="resource-row">
      <span class="resource-icon" title="${esc(r.name)}"><i class="fa-solid ${esc(r.icon) || 'fa-bolt'}"></i></span>
      <div class="resource-info">
        <span class="fw-bold">${esc(r.name)}</span>
        <span class="resource-rest-badges">
          ${r.short_rest_recovery > 0 ? `<span class="badge badge-muted" title="Recovers ${r.short_rest_recovery} on a short rest"><i class="fa-solid fa-mug-hot me-1"></i>+${r.short_rest_recovery} SR</span>` : ''}
          ${r.long_rest_recovery > 0 ? `<span class="badge badge-muted" title="Recovers ${r.long_rest_recovery} on a long rest"><i class="fa-solid fa-moon me-1"></i>+${r.long_rest_recovery} LR</span>` : ''}
        </span>
      </div>
      <div class="resource-value">
        <span class="stepper">
          <button class="stepper-btn" onclick="resourceStepper(${r.id}, -1)" aria-label="Decrease ${esc(r.name)}">−</button>
          <span class="stepper-value" onclick="resourceSetValue(${r.id}, this)">${r.current}</span>
          <button class="stepper-btn" onclick="resourceStepper(${r.id}, 1)" aria-label="Increase ${esc(r.name)}">+</button>
        </span>
        ${hasMax ? `<span class="text-muted small">/ ${r.max}</span>` : ''}
      </div>
      <div class="d-flex gap-1">
        <button class="btn btn-sm btn-outline-primary" onclick="openResourceForm(${r.id})" title="Edit"><i class="fa-solid fa-pen"></i></button>
        <button class="btn btn-sm btn-outline-danger" onclick="deleteResource(${r.id})" title="Remove"><i class="fa-solid fa-trash"></i></button>
      </div>
    </div>`;
}

async function saveResource(r: any) {
  await api('PUT', `/api/resources/${r.id}`, {
    name: r.name,
    current: r.current,
    max: r.max,
    short_rest_recovery: r.short_rest_recovery || 0,
    long_rest_recovery: r.long_rest_recovery === undefined ? 1 : r.long_rest_recovery,
    icon: r.icon || 'fa-bolt',
    sort_order: r.sort_order || 0,
  });
}

export async function resourceStepper(id: number, delta: number) {
  const r = resources.find(x => x.id === id);
  if (!r) return;
  const max = r.max > 0 ? r.max : Infinity;
  r.current = Math.max(0, Math.min(max, (r.current || 0) + delta));
  try {
    await saveResource(r);
  } catch (e: any) { toast(e.message, true); }
  await renderResources();
}

export function resourceSetValue(id: number, el: HTMLElement) {
  const r = resources.find(x => x.id === id);
  if (!r) return;
  el.innerHTML = `<input type="number" class="form-control stepper-inline-input" value="${r.current}" style="width:70px">`;
  const input = el.querySelector('input')!;
  input.focus();
  input.select();
  const save = async () => {
    const parsed = parseInt(input.value);
    if (!isNaN(parsed)) {
      r.current = Math.max(0, r.max > 0 ? Math.min(r.max, parsed) : parsed);
      try {
        await saveResource(r);
      } catch (e: any) { toast(e.message, true); }
    }
    await renderResources();
  };
  input.addEventListener('blur', save);
  input.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') { e.preventDefault(); save(); }
    if (e.key === 'Escape') { renderResources(); }
  });
}

export function openResourceForm(id?: number) {
  const r = id ? resources.find(x => x.id === id) : null;
  editingId = id ?? null;
  showModal(r ? 'Edit Resource' : 'Add Resource', `
    <form onsubmit="saveResourceForm();return false">
      <div class="mb-3">
        <label class="form-label">Name</label>
        <input type="text" class="form-control" id="resourceName" data-testid="resource-name-input" value="${esc(r?.name || '')}" placeholder="e.g. Rations, Arrows, Ki Points" required>
      </div>
      <div class="row g-3">
        <div class="col-6">
          <label class="form-label">Current</label>
          <input type="number" min="0" class="form-control" id="resourceCurrent" data-testid="resource-current-input" value="${r?.current ?? 0}">
        </div>
        <div class="col-6">
          <label class="form-label">Max <span class="text-muted small fw-normal">(0 = no max)</span></label>
          <input type="number" min="0" class="form-control" id="resourceMax" data-testid="resource-max-input" value="${r?.max ?? 0}">
        </div>
      </div>
      <div class="row g-3 mt-1">
        <div class="col-6">
          <label class="form-label" title="Amount restored when taking a short rest">Short Rest Recover</label>
          <input type="number" min="0" class="form-control" id="resourceShortRecovery" data-testid="resource-short-recovery-input" value="${r?.short_rest_recovery ?? 0}">
        </div>
        <div class="col-6">
          <label class="form-label" title="Amount restored when taking a long rest">Long Rest Recover</label>
          <input type="number" min="0" class="form-control" id="resourceLongRecovery" data-testid="resource-long-recovery-input" value="${r?.long_rest_recovery ?? (r ? 0 : 1)}">
        </div>
      </div>
      <div class="mb-3 mt-3">
        <label class="form-label">Icon</label>
        <div class="d-flex flex-wrap gap-2" id="resourceIconPicker">
          ${ICONS.map(ic => `
            <span class="resource-icon-option ${(r?.icon || 'fa-bolt') === ic ? 'selected' : ''}" data-icon="${ic}" role="button" tabindex="0" title="${ic.replace('fa-', '')}"><i class="fa-solid ${ic}"></i></span>`).join('')}
        </div>
        <input type="hidden" id="resourceIcon" value="${esc(r?.icon || 'fa-bolt')}">
      </div>
      <button type="submit" class="btn btn-gold w-100" data-testid="resource-save"><i class="fa-solid fa-floppy-disk me-1"></i>${r ? 'Save Changes' : 'Add Resource'}</button>
    </form>
  `);
  const picker = document.getElementById('resourceIconPicker');
  picker?.addEventListener('click', (e) => {
    const opt = (e.target as HTMLElement).closest('.resource-icon-option') as HTMLElement | null;
    if (!opt) return;
    picker.querySelectorAll('.resource-icon-option').forEach(o => o.classList.remove('selected'));
    opt.classList.add('selected');
    (document.getElementById('resourceIcon') as HTMLInputElement).value = opt.dataset.icon || 'fa-bolt';
  });
  setTimeout(() => {
    // The form may have been removed before the delayed focus runs (for
    // example while navigating away or when the DOM realm is torn down).
    if (typeof document === 'undefined') return;
    (document.getElementById('resourceName') as HTMLInputElement | null)?.focus();
  }, 50);
}

export async function saveResourceForm() {
  const nameEl = document.getElementById('resourceName') as HTMLInputElement;
  const name = nameEl?.value.trim();
  if (!name) { toast('Resource name is required', true); return; }
  const max = parseInt((document.getElementById('resourceMax') as HTMLInputElement).value || '0') || 0;
  const hasMax = max > 0;
  const body = {
    name,
    current: parseInt((document.getElementById('resourceCurrent') as HTMLInputElement).value || '0') || 0,
    max,
    // Resources without a max are consumables — they never refill on a rest.
    short_rest_recovery: hasMax ? (parseInt((document.getElementById('resourceShortRecovery') as HTMLInputElement).value || '0') || 0) : 0,
    long_rest_recovery: hasMax ? (parseInt((document.getElementById('resourceLongRecovery') as HTMLInputElement).value || '0') || 0) : 0,
    icon: (document.getElementById('resourceIcon') as HTMLInputElement).value || 'fa-bolt',
    sort_order: 0,
  };
  try {
    if (editingId) {
      await api('PUT', `/api/resources/${editingId}`, body);
      toast('Resource updated');
    } else {
      await api('POST', `/api/characters/${currentChar.id}/resources`, body);
      toast(`Added ${name}`);
    }
    hideModal();
    await renderResources();
  } catch (e: any) { toast(e.message, true); }
}

export async function deleteResource(id: number) {
  if (!confirm('Remove this resource?')) return;
  try {
    await api('DELETE', `/api/resources/${id}`);
    toast('Resource removed');
    await renderResources();
  } catch (e: any) { toast(e.message, true); }
}

// Window registrations (centralized)
expose('renderResources', renderResources);
expose('wealthTotalGp', wealthTotalGp);
expose('resourceStepper', resourceStepper);
expose('resourceSetValue', resourceSetValue);
expose('openResourceForm', openResourceForm);
expose('saveResourceForm', saveResourceForm);
expose('deleteResource', deleteResource);
