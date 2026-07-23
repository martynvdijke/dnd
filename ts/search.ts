/**
 * Universal command palette / search overlay.
 *
 * Uses the FTS5 search v2 API (/api/search/v2) for fast, ranked results.
 * Features:
 *  - Command palette overlay (open with Cmd+K / Ctrl+K or click search icon)
 *  - Type filter chips to narrow results
 *  - Recent searches tracked in localStorage
 *  - Entity-specific navigation on result click
 *  - Keyboard navigation (arrow keys + Enter)
 */

import { esc, toast } from './lib/dom';
import { api } from './lib/api';

// ─── Types ───

interface SearchResultV2 {
  entity_type: string;
  entity_id: number;
  name: string;
  snippet?: string;
  rank: number;
}

interface SearchResponseV2 {
  results: SearchResultV2[];
}

// ─── Type Definitions ───

interface SearchTypeDef {
  key: string;
  label: string;
  icon: string;
}

const SEARCH_TYPES: SearchTypeDef[] = [
  { key: '',            label: 'All',      icon: 'fa-magnifying-glass' },
  { key: 'characters',  label: 'Characters', icon: 'fa-users' },
  { key: 'npcs',        label: 'NPCs',       icon: 'fa-user-group' },
  { key: 'campaigns',   label: 'Campaigns',  icon: 'fa-flag' },
  { key: 'notes',       label: 'Notes',      icon: 'fa-note-sticky' },
  { key: 'quests',      label: 'Quests',     icon: 'fa-scroll' },
  { key: 'sessions',    label: 'Sessions',   icon: 'fa-calendar' },
  { key: 'journal',     label: 'Journal',    icon: 'fa-book-open' },
  { key: 'spells',      label: 'Spells',     icon: 'fa-wand-sparkles' },
  { key: 'equipment',   label: 'Equipment',  icon: 'fa-backpack' },
  { key: 'monsters',    label: 'Monsters',   icon: 'fa-dragon' },
];

const ENTITY_ICONS: Record<string, string> = {
  characters: 'fa-users',
  npcs: 'fa-user-group',
  campaigns: 'fa-flag',
  notes: 'fa-note-sticky',
  quests: 'fa-scroll',
  sessions: 'fa-calendar',
  journal: 'fa-book-open',
  spells: 'fa-wand-sparkles',
  equipment: 'fa-backpack',
  races: 'fa-person',
  classes: 'fa-graduation-cap',
  feats: 'fa-star',
  backgrounds: 'fa-address-card',
  monsters: 'fa-dragon',
  adventures: 'fa-scroll',
  factions: 'fa-flag',
  shops: 'fa-store',
  encounters: 'fa-crosshairs',
  locations: 'fa-map',
};

// ─── State ───

let currentTypeFilter = '';
let searchTimeout: ReturnType<typeof setTimeout> | null = null;
let selectedIndex = -1;
let lastQuery = '';

const RECENTS_KEY = 'villum-search-recents';
const MAX_RECENTS = 10;

// ─── Recent Searches ───

export function getRecents(): string[] {
  try {
    return JSON.parse(localStorage.getItem(RECENTS_KEY) || '[]');
  } catch { return []; }
}

export function addRecent(query: string): void {
  if (!query.trim()) return;
  const recents = getRecents().filter(r => r !== query);
  recents.unshift(query);
  localStorage.setItem(RECENTS_KEY, JSON.stringify(recents.slice(0, MAX_RECENTS)));
}

export function clearRecents(): void {
  localStorage.removeItem(RECENTS_KEY);
}

// ─── Search Overlay ───

export function showSearchOverlay(): void {
  let overlay = document.getElementById('searchOverlay');
  if (!overlay) {
    overlay = document.createElement('div');
    overlay.id = 'searchOverlay';
    overlay.className = 'search-overlay';
    overlay.addEventListener('click', (e) => { if (e.target === overlay) hideSearchOverlay(); });
    document.body.appendChild(overlay);
    const panel = document.createElement('div');
    panel.id = 'searchPanel';
    panel.className = 'search-panel command-palette';
    overlay.appendChild(panel);
    // Build initial HTML
    buildCommandPalette(panel);
  }
  overlay.style.display = 'flex';
  selectedIndex = -1;
  const input = document.getElementById('searchInput') as HTMLInputElement | null;
  if (input) {
    input.value = '';
    input.focus();
  }
  updateTypeFilter(''); // reset filter
  lastQuery = '';
}

export function hideSearchOverlay(): void {
  const overlay = document.getElementById('searchOverlay');
  if (overlay) overlay.style.display = 'none';
}

function buildCommandPalette(panel: HTMLElement): void {
  panel.innerHTML = `
    <div class="cp-header">
      <div class="cp-input-wrapper">
        <i class="fa-solid fa-search cp-search-icon"></i>
        <input type="text" class="cp-input" id="searchInput" placeholder="Search everything..." autocomplete="off" spellcheck="false">
        <kbd class="cp-kbd">ESC</kbd>
      </div>
      <div class="cp-filters" id="cpFilters"></div>
    </div>
    <div class="cp-results" id="cpResults">
      <div class="cp-recents" id="cpRecents"></div>
    </div>
    <div class="cp-footer">
      <span><kbd>↑↓</kbd> navigate</span>
      <span><kbd>Enter</kbd> open</span>
      <span><kbd>Esc</kbd> close</span>
    </div>
  `;

  // Build filter chips
  const filtersEl = document.getElementById('cpFilters')!;
  filtersEl.innerHTML = SEARCH_TYPES.map(t =>
    `<button class="cp-filter-chip ${t.key === '' ? 'active' : ''}" data-type="${t.key}" onclick="window.__searchSetType('${t.key}')">
      <i class="fa-solid ${t.icon}"></i> ${t.label}
    </button>`
  ).join('');

  // Setup input handler
  const input = document.getElementById('searchInput') as HTMLInputElement;
  input.addEventListener('input', () => onSearchInput(input.value));
  input.addEventListener('keydown', (e) => onSearchKeydown(e, input));

  // Render recents
  renderRecents();
}

(window as any).__searchSetType = function (type: string) {
  updateTypeFilter(type);
  const input = document.getElementById('searchInput') as HTMLInputElement;
  if (input && input.value.trim()) {
    doSearch(input.value);
  } else {
    renderRecents();
  }
};

function updateTypeFilter(type: string): void {
  currentTypeFilter = type;
  document.querySelectorAll('#cpFilters .cp-filter-chip').forEach(el => {
    el.classList.toggle('active', (el as HTMLElement).dataset.type === type);
  });
}

function renderRecents(): void {
  const recentsEl = document.getElementById('cpRecents');
  if (!recentsEl) return;
  const recents = getRecents();
  if (recents.length === 0) {
    recentsEl.innerHTML = `
      <div class="cp-empty">
        <i class="fa-solid fa-magnifying-glass fa-2x mb-2 d-block text-muted"></i>
        <p class="fw-bold mb-1">Search Everything</p>
        <p class="small text-muted mb-0">Type to search characters, notes, compendium &amp; more</p>
      </div>`;
    return;
  }
  recentsEl.innerHTML = `
    <div class="cp-section-label">
      Recent Searches
      <button class="btn btn-sm btn-link text-muted p-0 ms-2" onclick="window.__clearRecents()">Clear</button>
    </div>
    ${recents.map((q, i) => `
      <div class="cp-result-item" data-index="${i}" onclick="window.__searchRecent('${esc(q)}')">
        <i class="fa-solid fa-clock-rotate-left cp-result-icon text-muted"></i>
        <div class="cp-result-body">
          <div class="cp-result-name">${esc(q)}</div>
        </div>
      </div>
    `).join('')}`;
}

(window as any).__clearRecents = function () {
  clearRecents();
  renderRecents();
};

(window as any).__searchRecent = function (query: string) {
  const input = document.getElementById('searchInput') as HTMLInputElement;
  if (input) {
    input.value = query;
    doSearch(query);
  }
};

// ─── Search Execution ───

function onSearchInput(value: string): void {
  if (searchTimeout) clearTimeout(searchTimeout);
  const q = value.trim();

  if (!q) {
    document.getElementById('cpResults')!.innerHTML = '';
    renderRecents();
    return;
  }

  searchTimeout = setTimeout(() => doSearch(q), 200);
}

export async function doSearch(query?: string): Promise<void> {
  const input = document.getElementById('searchInput') as HTMLInputElement | null;
  const q = query || input?.value?.trim() || '';
  if (!q) return;

  lastQuery = q;
  selectedIndex = -1;

  try {
    let url = '/api/search?q=' + encodeURIComponent(q);
    if (currentTypeFilter) {
      url += '&types=' + encodeURIComponent(currentTypeFilter);
    }

    const data = await api('GET', url);
    // Map backend response (title/score) to frontend expectations (name/rank)
    const rawResults = data.results || [];
    const results: SearchResultV2[] = rawResults.map((r: any) => ({
      entity_type: r.entity_type,
      entity_id: r.entity_id,
      name: r.title || r.name,
      snippet: r.snippet,
      rank: r.score || r.rank,
    }));

    const resultsEl = document.getElementById('cpResults');
    if (!resultsEl) return;

    if (results.length === 0) {
      resultsEl.innerHTML = `
        <div class="cp-empty">
          <i class="fa-solid fa-search fa-2x mb-2 d-block text-muted"></i>
          <p class="fw-bold mb-1">No Results</p>
          <p class="small text-muted">No matches found for "${esc(q)}"</p>
        </div>`;
      return;
    }

    resultsEl.innerHTML = results.map((r, i) => `
      <div class="cp-result-item search-result-item ${i === 0 ? 'selected' : ''}" data-index="${i}"
           onclick="window.__searchNavigate('${r.entity_type}',${r.entity_id},'${esc(r.name)}')"
           onmouseenter="__searchHover(${i})">
        <i class="fa-solid ${ENTITY_ICONS[r.entity_type] || 'fa-file'} cp-result-icon"></i>
        <div class="cp-result-body">
          <div class="cp-result-name">${highlightMatch(esc(r.name), q)}</div>
          ${r.snippet ? `<div class="cp-result-snippet">${r.snippet}</div>` : ''}
        </div>
        <span class="cp-result-type badge badge-muted">${r.entity_type.charAt(0).toUpperCase() + r.entity_type.slice(1)}</span>
      </div>
    `).join('');

    // Save search to recents
    addRecent(q);
  } catch (e: any) {
    toast('Search failed: ' + e.message, true);
  }
}

export function highlightMatch(text: string, query: string): string {
  if (!query) return text;
  try {
    const escaped = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const re = new RegExp(`(${escaped.split(/\s+/).join('|')})`, 'gi');
    return text.replace(re, '<mark>$1</mark>');
  } catch {
    return text;
  }
}

// ─── Keyboard Navigation ───

function onSearchKeydown(e: KeyboardEvent, input: HTMLInputElement): void {
  const items = document.querySelectorAll('#cpResults .cp-result-item');

  switch (e.key) {
    case 'ArrowDown':
      e.preventDefault();
      selectedIndex = Math.min(selectedIndex + 1, items.length - 1);
      updateSelection(items);
      break;
    case 'ArrowUp':
      e.preventDefault();
      selectedIndex = Math.max(selectedIndex - 1, 0);
      updateSelection(items);
      break;
    case 'Enter':
      e.preventDefault();
      if (selectedIndex >= 0 && selectedIndex < items.length) {
        (items[selectedIndex] as HTMLElement).click();
      } else if (input.value.trim()) {
        // If nothing selected, try opening first result
        const first = items[0] as HTMLElement | undefined;
        if (first) first.click();
      }
      break;
    case 'Escape':
      hideSearchOverlay();
      break;
  }
}

(window as any).__searchHover = function (index: number) {
  selectedIndex = index;
  updateSelection(document.querySelectorAll('#cpResults .cp-result-item'));
};

function updateSelection(items: NodeListOf<Element>): void {
  items.forEach((el, i) => {
    el.classList.toggle('selected', i === selectedIndex);
  });
  const selected = items[selectedIndex] as HTMLElement | undefined;
  if (selected) {
    selected.scrollIntoView({ block: 'nearest' });
  }
}

// ─── Result Navigation ───

(window as any).__searchNavigate = function (type: string, id: number, name: string) {
  hideSearchOverlay();

  // Use the same navigation as the existing legacy system
  const navFn = (window as any).navigateSearchResult;
  if (typeof navFn === 'function') {
    navFn(type, id, name);
  } else {
    // Fallback: basic navigation
    import('./navigation').then(({ showView }) => {
      if (type === 'characters') {
        (window as any).openChar?.(id);
      } else if (['spells', 'equipment', 'races', 'classes', 'feats', 'backgrounds', 'monsters'].includes(type)) {
        (window as any).showCompendium?.();
      } else {
        showView('characters');
      }
    });
  }
};

// ─── Init ───

export function initSearch(): void {
  // Add Cmd+K / Ctrl+K shortcut for command palette
  document.addEventListener('keydown', (e) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      showSearchOverlay();
      return;
    }
    // Forward-slash to open search (when not in an input)
    const target = e.target as HTMLElement;
    const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.tagName === 'SELECT';
    if (!isInput && e.key === '/' && !e.metaKey && !e.ctrlKey) {
      e.preventDefault();
      showSearchOverlay();
      return;
    }
  });

  // Backward compatibility: wire existing navbar search button
  const searchBtn = document.getElementById('searchBtn');
  if (searchBtn) {
    searchBtn.addEventListener('click', showSearchOverlay);
  }

  // The old searchInput in the navbar now opens the command palette on focus
  const navSearch = document.getElementById('searchInput') as HTMLInputElement | null;
  if (navSearch) {
    navSearch.addEventListener('focus', (e) => {
      // Only intercept if this is the navbar search (not our command palette input)
      if (navSearch.closest('.command-palette')) return;
      showSearchOverlay();
      navSearch.blur();
    });
    navSearch.addEventListener('click', (e) => {
      if (navSearch.closest('.command-palette')) return;
      showSearchOverlay();
      navSearch.blur();
    });
  }
}

// Backward compatibility: expose for e2e tests and legacy inline usage
(window as any).doSearch = doSearch;
