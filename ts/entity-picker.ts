/**
 * Entity picker — reusable modal for searching and selecting entities to link.
 *
 * Fetches results from /api/search/v2 and emits the selected (type, id, title)
 * via a callback. Used by manual link creation and mention insertion.
 */
import { esc } from './lib/dom';

export interface EntityRef {
  entity_type: string;
  entity_id: number;
  title: string;
}

type PickerCallback = (ref: EntityRef) => void;

let pickerOverlay: HTMLElement | null = null;
let pickerCallback: PickerCallback | null = null;
let pickerFilter = '';
let pickerSearchTimeout: number | undefined;

// Map entity types to display icons.
const TYPE_ICONS: Record<string, string> = {
  character: 'fa-users',
  npc: 'fa-user-group',
  note: 'fa-note-sticky',
  quest: 'fa-scroll',
  journal: 'fa-book-open',
  session: 'fa-calendar',
  campaign: 'fa-flag',
  location: 'fa-map-pin',
  faction: 'fa-flag',
  shop: 'fa-store',
  oneshot: 'fa-scroll',
  encounter: 'fa-crosshairs',
  monster: 'fa-dragon',
  spell: 'fa-wand-sparkles',
  equipment: 'fa-backpack',
  race: 'fa-person',
  class: 'fa-graduation-cap',
  feat: 'fa-star',
  background: 'fa-address-card',
  item: 'fa-box',
};

function entityIcon(type: string): string {
  return TYPE_ICONS[type] || 'fa-file-lines';
}

function entityColor(type: string): string {
  const colors: Record<string, string> = {
    character: 'var(--blood)',
    npc: 'var(--gold)',
    note: 'var(--ink-muted)',
    quest: 'var(--primary)',
    campaign: 'var(--gold)',
    location: 'var(--success)',
    encounter: 'var(--danger)',
    monster: 'var(--danger)',
    spell: 'var(--primary)',
    faction: 'var(--gold)',
    shop: 'var(--success)',
  };
  return colors[type] || 'var(--ink-muted)';
}

export function showEntityPicker(
  title: string,
  callback: PickerCallback,
  initialFilter = '',
): void {
  pickerCallback = callback;
  pickerFilter = initialFilter;

  if (!pickerOverlay || !document.body.contains(pickerOverlay)) {
    pickerOverlay = document.createElement('div');
    pickerOverlay.id = 'entityPickerOverlay';
    pickerOverlay.className = 'entity-picker-overlay';
    pickerOverlay.addEventListener('click', (e) => {
      if (e.target === pickerOverlay) hideEntityPicker();
    });
    document.body.appendChild(pickerOverlay);

    const panel = document.createElement('div');
    panel.id = 'entityPickerPanel';
    panel.className = 'entity-picker-panel';
    pickerOverlay.appendChild(panel);
  }

  pickerOverlay.style.display = 'flex';

  const panel = document.getElementById('entityPickerPanel')!;
  const typeChips = buildTypeChips();
  panel.innerHTML = `
    <div class="entity-picker-header">
      <h5 class="entity-picker-title">${esc(title)}</h5>
      <button class="entity-picker-close" id="entityPickerClose">&times;</button>
    </div>
    <div class="entity-picker-search">
      <input
        type="text"
        class="form-control"
        id="entityPickerInput"
        placeholder="Search entities..."
        autocomplete="off"
      />
    </div>
    <div class="entity-picker-chips" id="entityPickerChips">${typeChips}</div>
    <div class="entity-picker-results" id="entityPickerResults">
      <div class="entity-picker-empty">
        <i class="fa-solid fa-search fa-2x d-block mb-2 text-muted"></i>
        <p class="text-muted small">Start typing to search</p>
      </div>
    </div>
  `;

  document.getElementById('entityPickerClose')!.addEventListener('click', hideEntityPicker);

  const input = document.getElementById('entityPickerInput') as HTMLInputElement;
  input.focus();
  input.addEventListener('input', () => {
    clearTimeout(pickerSearchTimeout);
    const q = input.value.trim();
    if (q.length < 2) {
      document.getElementById('entityPickerResults')!.innerHTML = `
        <div class="entity-picker-empty">
          <i class="fa-solid fa-search fa-2x d-block mb-2 text-muted"></i>
          <p class="text-muted small">Type at least 2 characters to search</p>
        </div>`;
      return;
    }
    pickerSearchTimeout = window.setTimeout(() => doPickerSearch(q), 250);
  });

  // Re-bind chip clicks
  document.querySelectorAll('.entity-picker-chip').forEach((chip) => {
    chip.addEventListener('click', () => {
      const type = (chip as HTMLElement).dataset.type || '';
      updatePickerFilter(type);
    });
  });

  if (initialFilter) {
    updatePickerFilter(initialFilter);
  }
}

export function hideEntityPicker(): void {
  if (pickerOverlay) {
    pickerOverlay.style.display = 'none';
  }
  pickerCallback = null;
}

function buildTypeChips(): string {
  const ENTITY_TYPES = [
    { type: '', label: 'All', icon: 'fa-search' },
    { type: 'character', label: 'Characters', icon: 'fa-users' },
    { type: 'npc', label: 'NPCs', icon: 'fa-user-group' },
    { type: 'campaign', label: 'Campaigns', icon: 'fa-flag' },
    { type: 'location', label: 'Locations', icon: 'fa-map-pin' },
    { type: 'quest', label: 'Quests', icon: 'fa-scroll' },
    { type: 'note', label: 'Notes', icon: 'fa-note-sticky' },
    { type: 'session', label: 'Sessions', icon: 'fa-calendar' },
    { type: 'encounter', label: 'Encounters', icon: 'fa-crosshairs' },
    { type: 'faction', label: 'Factions', icon: 'fa-flag' },
    { type: 'shop', label: 'Shops', icon: 'fa-store' },
    { type: 'oneshot', label: 'One-Shots', icon: 'fa-scroll' },
    { type: 'monster', label: 'Monsters', icon: 'fa-dragon' },
    { type: 'spell', label: 'Spells', icon: 'fa-wand-sparkles' },
    { type: 'item', label: 'Items', icon: 'fa-box' },
  ];
  return ENTITY_TYPES.map(
    (t) =>
      `<button class="entity-picker-chip ${pickerFilter === t.type ? 'active' : ''}" data-type="${t.type}">
        <i class="fa-solid ${t.icon}"></i>
        <span>${t.label}</span>
      </button>`,
  ).join('');
}

function updatePickerFilter(type: string): void {
  pickerFilter = type;
  document.querySelectorAll('.entity-picker-chip').forEach((chip) => {
    chip.classList.toggle('active', (chip as HTMLElement).dataset.type === type);
  });
  const input = document.getElementById('entityPickerInput') as HTMLInputElement;
  const q = input.value.trim();
  if (q.length >= 2) {
    doPickerSearch(q);
  }
}

async function doPickerSearch(q: string): Promise<void> {
  const resultsEl = document.getElementById('entityPickerResults');
  if (!resultsEl) return;

  try {
    let url = `/api/search/v2?q=${encodeURIComponent(q)}&limit=20`;
    if (pickerFilter) {
      url += `&types=${encodeURIComponent(pickerFilter)}`;
    }
    const res = await fetch(url, { credentials: 'include' });
    if (!res.ok) throw new Error('Search failed');
    const data = await res.json();
    const items = data.results || data || [];

    if (items.length === 0) {
      resultsEl.innerHTML = `
        <div class="entity-picker-empty">
          <i class="fa-solid fa-search fa-2x d-block mb-2 text-muted"></i>
          <p class="text-muted small">No results found</p>
        </div>`;
      return;
    }

    resultsEl.innerHTML = items
      .map(
        (item: any) => `
        <div class="entity-picker-item" data-type="${esc(item.entity_type || item.type)}" data-id="${item.entity_id || item.id}" data-title="${esc(item.title || item.name)}">
          <i class="fa-solid ${entityIcon(item.entity_type || item.type)}" style="color:${entityColor(item.entity_type || item.type)}"></i>
          <div class="entity-picker-item-info">
            <div class="entity-picker-item-title">${esc(item.title || item.name)}</div>
            <div class="entity-picker-item-type">${esc(item.entity_type || item.type)}</div>
          </div>
          <button class="btn btn-sm btn-outline-primary entity-picker-select">Select</button>
        </div>`,
      )
      .join('');

    // Bind select buttons
    resultsEl.querySelectorAll('.entity-picker-select').forEach((btn) => {
      btn.addEventListener('click', () => {
        const item = (btn as HTMLElement).closest('.entity-picker-item') as HTMLElement;
        if (!item) return;
        pickerCallback?.({
          entity_type: item.dataset.type || '',
          entity_id: parseInt(item.dataset.id || '0', 10),
          title: item.dataset.title || '',
        });
        hideEntityPicker();
      });
    });
  } catch (e: any) {
    resultsEl.innerHTML = `
      <div class="entity-picker-empty text-danger">
        <i class="fa-solid fa-exclamation-triangle fa-2x d-block mb-2"></i>
        <p class="small">${esc(e.message)}</p>
      </div>`;
  }
}
