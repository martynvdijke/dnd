/**
 * Universal command palette / search overlay.
 *
 * Uses the FTS5 search API (/api/search) for fast, ranked results.
 * Features:
 *  - Command palette overlay (open with Cmd+K / Ctrl+K or click search icon)
 *  - Type filter chips to narrow results
 *  - Recent searches tracked in localStorage
 *  - Entity-specific navigation on result click
 *  - Keyboard navigation (arrow keys + Enter)
 */

import { esc, attrEscape, toast } from './lib/dom';
import { api, getCsrfToken, getApiToken } from './lib/api';
import { expose } from './lib/expose';
import { currentCampaign } from './lib/state';

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

// Keys MUST match entity_search_index.entity_type values (singular).
// See db/search_index.go — these are the only filter values the API accepts.
const SEARCH_TYPES: SearchTypeDef[] = [
  { key: '',            label: 'All',         icon: 'fa-magnifying-glass' },
  { key: 'character',   label: 'Characters',  icon: 'fa-users' },
  { key: 'npc',         label: 'NPCs',        icon: 'fa-user-group' },
  { key: 'campaign',    label: 'Campaigns',   icon: 'fa-flag' },
  { key: 'note',        label: 'Notes',       icon: 'fa-note-sticky' },
  { key: 'quest',       label: 'Quests',      icon: 'fa-scroll' },
  { key: 'session',     label: 'Sessions',    icon: 'fa-calendar' },
  { key: 'journal',     label: 'Journal',     icon: 'fa-book-open' },
  { key: 'location',    label: 'Locations',   icon: 'fa-map' },
  { key: 'encounter',   label: 'Encounters',  icon: 'fa-crosshairs' },
  { key: 'monster',     label: 'Monsters',    icon: 'fa-dragon' },
  { key: 'shop',        label: 'Shops',       icon: 'fa-store' },
  { key: 'faction',     label: 'Factions',    icon: 'fa-flag' },
  { key: 'adventure',   label: 'Adventures',  icon: 'fa-map' },
  { key: 'wiki',        label: 'Wiki',        icon: 'fa-book' },
  { key: 'recap',       label: 'Recaps',      icon: 'fa-feather' },
  { key: 'timeline',    label: 'Timeline',    icon: 'fa-timeline' },
  { key: 'knowledge',   label: 'Knowledge',   icon: 'fa-lightbulb' },
  { key: 'item',        label: 'Items',       icon: 'fa-backpack' },
  { key: 'compendium',  label: 'Compendium',  icon: 'fa-spell-book' },
];

// Icon lookup for search results; keyed by singular entity_type.
const ENTITY_ICONS: Record<string, string> = Object.fromEntries(
  SEARCH_TYPES.filter(t => t.key).map(t => [t.key, t.icon]),
);

// ─── State ───

let currentTypeFilter = '';
let searchTimeout: ReturnType<typeof setTimeout> | null = null;
let selectedIndex = -1;
let lastQuery = '';
let aiMode = false;
let aiLoading = false;

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
  aiMode = false;
  aiLoading = false;
  const input = document.getElementById('cpSearchInput') as HTMLInputElement | null;
  if (input) {
    input.value = '';
    input.focus();
  }
  updateTypeFilter(''); // reset filter
  lastQuery = '';
  syncAIToggleState();
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
        <input type="text" class="cp-input" id="cpSearchInput" placeholder="Search everything..." autocomplete="off" spellcheck="false" aria-label="Search">
        <kbd class="cp-kbd">ESC</kbd>
      </div>
      <div class="cp-filters" id="cpFilters"></div>
    </div>
    <div class="cp-results" id="cpResults">
      <div class="cp-recents" id="cpRecents"></div>
    </div>
    <div class="cp-footer" id="cpFooter">
      <div class="cp-footer-hints">
        <span><kbd>↑↓</kbd> navigate</span>
        <span><kbd>Enter</kbd> open</span>
        <span><kbd>Esc</kbd> close</span>
      </div>
      <button type="button" class="cp-ai-toggle" id="cpAskAiBtn" data-testid="search-ai-toggle"
        aria-label="Ask AI about your query" onclick="window.__searchAskAI()">
        <i class="fa-solid fa-wand-magic-sparkles" aria-hidden="true"></i>
        <span>Ask AI</span>
        <kbd class="cp-ai-kbd">⌘↵</kbd>
      </button>
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
  const input = document.getElementById('cpSearchInput') as HTMLInputElement;
  input.addEventListener('input', () => onSearchInput(input.value));
  input.addEventListener('keydown', (e) => onSearchKeydown(e, input));

  // Render recents
  renderRecents();
  syncAIToggleState();
}

expose('__searchSetType', function (type: string) {
  updateTypeFilter(type);
  const input = document.getElementById('cpSearchInput') as HTMLInputElement;
  if (input && input.value.trim()) {
    doSearch(input.value);
  } else {
    renderRecents();
  }
});

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
      <div class="cp-result-item" data-index="${i}" onclick="window.__searchRecent('${attrEscape(q)}')">
        <i class="fa-solid fa-clock-rotate-left cp-result-icon text-muted"></i>
        <div class="cp-result-body">
          <div class="cp-result-name">${esc(q)}</div>
        </div>
      </div>
    `).join('')}`;
}

expose('__clearRecents', function () {
  clearRecents();
  renderRecents();
});

expose('__searchRecent', function (query: string) {
  const input = document.getElementById('cpSearchInput') as HTMLInputElement;
  if (input) {
    input.value = query;
    doSearch(query);
  }
});

// ─── AI Search helpers (pure, testable) ───

export interface AISource {
  kind: string;
  id: number | string;
  title: string;
  subtitle?: string;
  url?: string;
  snippet?: string;
}

export interface AISearchPayload {
  query: string;
  campaign_id?: number;
  type_filter?: string;
}

export function getAISearchPayload(query: string, typeFilter: string): AISearchPayload {
  const q = query.trim();
  const payload: AISearchPayload = { query: q };
  if (typeFilter) payload.type_filter = typeFilter;
  const cid = (currentCampaign as any)?.id;
  if (typeof cid === 'number' && cid > 0) payload.campaign_id = cid;
  return payload;
}

export function formatAIAnswer(text: string): string {
  if (!text) return '';
  return esc(text).replace(/\r\n/g, '\n').replace(/\n/g, '<br>');
}

export function buildAISourcesHtml(sources: AISource[]): string {
  if (!sources || sources.length === 0) return '';
  return sources.map((s, idx) => {
    const title = esc(s.title || `${s.kind} #${s.id}`);
    const subtitle = s.subtitle ? `<span class="cp-ai-source-subtitle">${esc(s.subtitle)}</span>` : '';
    const snippet = s.snippet ? `<span class="cp-ai-source-snippet">${esc(s.snippet)}</span>` : '';
    const kindLabel = esc(s.kind || 'source');
    // Use a button-like link that navigates via __searchNavigate / showView
    return `<li class="cp-ai-source-item">
      <a href="#" class="cp-ai-source-link" data-testid="search-ai-source-link" data-kind="${attrEscape(s.kind)}" data-id="${attrEscape(String(s.id))}" data-url="${attrEscape(s.url || '')}" onclick="window.__searchAISource(event, ${idx})">
        <span class="cp-ai-source-kind badge badge-muted">${kindLabel}</span>
        <span class="cp-ai-source-title">${title}</span>
        ${subtitle}
        ${snippet}
      </a>
    </li>`;
  }).join('');
}

export function getAINoAnswerMessage(scope: string, sources: AISource[]): string {
  if (!sources || sources.length === 0) return 'No direct matches found for this question.';
  if (scope === 'campaign') return 'No campaign context matched — here are the closest compendium and campaign sources.';
  return 'AI is not configured; showing direct matches instead.';
}

function syncAIToggleState(): void {
  const btn = document.getElementById('cpAskAiBtn') as HTMLButtonElement | null;
  const input = document.getElementById('cpSearchInput') as HTMLInputElement | null;
  if (!btn) return;
  const hasQuery = !!(input?.value.trim() || lastQuery);
  btn.disabled = aiLoading || !hasQuery;
  btn.setAttribute('aria-disabled', String(btn.disabled));
  btn.title = !hasQuery ? 'Type a question first' : aiLoading ? 'Waiting for answer…' : 'Ask AI (⌘↵)';
}

function renderAILoading(query: string): void {
  const el = document.getElementById('cpResults');
  if (!el) return;
  el.innerHTML = `
    <div class="cp-ai-loading" data-testid="search-ai-loading" role="status" aria-live="polite" aria-busy="true">
      <div class="cp-ai-loading-spinner" aria-hidden="true"><i class="fa-solid fa-circle-notch fa-spin"></i></div>
      <div class="cp-ai-loading-text">
        <div class="fw-bold">Asking the archives…</div>
        <div class="small text-muted">Looking up "${esc(query)}"</div>
      </div>
    </div>`;
}

function renderAIAnswerState(opts: { answer: string; sources: AISource[]; scope: string; errorMsg?: string; query: string }): void {
  const el = document.getElementById('cpResults');
  if (!el) return;
  const hasAnswer = !!opts.answer?.trim();
  const sourcesHtml = buildAISourcesHtml(opts.sources);
  const scopeBadge = opts.scope ? `<span class="cp-ai-scope badge ${opts.scope === 'campaign' ? 'badge-gold' : 'badge-muted'}">${esc(opts.scope)}</span>` : '';
  const answerHtml = hasAnswer
    ? `<div class="cp-ai-answer-text" data-testid="search-ai-answer">${formatAIAnswer(opts.answer)}</div>`
    : `<div class="cp-ai-no-answer" data-testid="search-ai-answer"><p class="small text-muted mb-0">${esc(opts.errorMsg || getAINoAnswerMessage(opts.scope, opts.sources))}</p></div>`;
  const errorBanner = opts.errorMsg && hasAnswer ? '' : opts.errorMsg && !hasAnswer ? '' : '';
  // When disabled, the message is already in errorMsg via getAINoAnswerMessage; render friendly banner if we came from 503
  const banner = opts.errorMsg ? `<div class="cp-ai-banner" role="status">${esc(opts.errorMsg)}</div>` : '';
  el.innerHTML = `
    <div class="cp-ai-panel" role="region" aria-label="AI answer">
      <div class="cp-ai-panel-header">
        <button type="button" class="btn btn-sm btn-outline-secondary cp-ai-back" data-testid="search-ai-back" onclick="window.__searchAIBack()" aria-label="Back to search results">
          <i class="fa-solid fa-arrow-left me-1" aria-hidden="true"></i> Back
        </button>
        <span class="cp-ai-panel-title"><i class="fa-solid fa-wand-magic-sparkles me-1" aria-hidden="true"></i>Answer ${scopeBadge}</span>
      </div>
      ${banner}
      ${answerHtml}
      ${opts.sources.length ? `<div class="cp-ai-sources" data-testid="search-ai-sources"><div class="cp-section-label">Sources · ${opts.sources.length}</div><ol class="cp-ai-source-list">${sourcesHtml}</ol></div>` : ''}
      <div class="cp-ai-actions">
        <button type="button" class="btn btn-sm btn-outline-secondary" data-testid="search-ai-back" onclick="window.__searchAIBack()">Back to results</button>
      </div>
    </div>`;
}

function renderAIError(message: string, query: string): void {
  const el = document.getElementById('cpResults');
  if (!el) return;
  el.innerHTML = `
    <div class="cp-ai-panel" role="alert">
      <div class="cp-ai-panel-header">
        <button type="button" class="btn btn-sm btn-outline-secondary cp-ai-back" data-testid="search-ai-back" onclick="window.__searchAIBack()" aria-label="Back to search results">
          <i class="fa-solid fa-arrow-left me-1" aria-hidden="true"></i> Back
        </button>
        <span class="cp-ai-panel-title">Ask AI</span>
      </div>
      <div class="cp-ai-banner cp-ai-banner-error">${esc(message)}</div>
      <p class="small text-muted mt-2 mb-0">Try a different phrasing or browse the lexical results instead.</p>
      <div class="cp-ai-actions">
        <button type="button" class="btn btn-sm btn-outline-secondary" data-testid="search-ai-back" onclick="window.__searchAIBack()">Back to results</button>
      </div>
    </div>`;
}

// ─── Search Execution ───

function onSearchInput(value: string): void {
  if (searchTimeout) clearTimeout(searchTimeout);
  aiMode = false;
  syncAIToggleState();
  const q = value.trim();

  if (!q) {
    document.getElementById('cpResults')!.innerHTML = '';
    renderRecents();
    syncAIToggleState();
    return;
  }

  searchTimeout = setTimeout(() => doSearch(q), 200);
  syncAIToggleState();
}

export async function doSearch(query?: string): Promise<void> {
  if (aiMode) return;
  const input = document.getElementById('cpSearchInput') as HTMLInputElement | null;
  const q = query || input?.value?.trim() || '';
  if (!q) return;

  lastQuery = q;
  selectedIndex = -1;
  aiMode = false;
  aiLoading = false;
  syncAIToggleState();

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
           onclick="window.__searchNavigate('${r.entity_type}',${r.entity_id},'${attrEscape(r.name)}')"
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

export async function doAISearch(query?: string): Promise<void> {
  if (searchTimeout) { clearTimeout(searchTimeout); searchTimeout = null; }
  const input = document.getElementById('cpSearchInput') as HTMLInputElement | null;
  const q = (query || input?.value?.trim() || lastQuery || '').trim();
  if (!q) {
    toast('Type a question first', true);
    input?.focus();
    return;
  }
  lastQuery = q;
  aiMode = true;
  aiLoading = true;
  selectedIndex = -1;
  syncAIToggleState();
  renderAILoading(q);
  const payload = getAISearchPayload(q, currentTypeFilter);
  // Build headers consistent with api helper
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  const csrf = getCsrfToken();
  const tok = getApiToken();
  if (csrf) headers['X-CSRF-Token'] = csrf;
  if (tok) headers['Authorization'] = `Bearer ${tok}`;
  try {
    const res = await fetch('/api/search/ai', {
      method: 'POST',
      headers,
      credentials: 'include',
      body: JSON.stringify(payload),
    });
    const data = await res.json().catch(() => ({}));
    aiLoading = false;
    syncAIToggleState();
    if (!res.ok) {
      // 503 disabled / not configured -> show friendly message plus sources if present
      if (res.status === 503) {
        const msg = data.error || 'AI is not configured; showing direct matches instead';
        const sources: AISource[] = Array.isArray(data.sources) ? data.sources : [];
        const scope: string = data.scope || 'compendium';
        renderAIAnswerState({ answer: '', sources, scope, errorMsg: msg, query: q });
        return;
      }
      throw new Error(data.error || `AI search failed (${res.status})`);
    }
    const answer: string = data.answer || '';
    const sources: AISource[] = Array.isArray(data.sources) ? data.sources : [];
    const scope: string = data.scope || 'compendium';
    if (!answer && sources.length === 0) {
      renderAIAnswerState({ answer: '', sources: [], scope, errorMsg: 'No answer and no sources found. Try a different question.', query: q });
      return;
    }
    renderAIAnswerState({ answer, sources, scope, query: q });
  } catch (e: any) {
    aiLoading = false;
    syncAIToggleState();
    renderAIError(e?.message || 'AI search failed', q);
  }
}

expose('__searchAskAI', function () { doAISearch(); });
expose('__searchAIBack', function () {
  aiMode = false;
  aiLoading = false;
  syncAIToggleState();
  const input = document.getElementById('cpSearchInput') as HTMLInputElement | null;
  const q = input?.value?.trim() || lastQuery;
  if (q) {
    doSearch(q);
  } else {
    const el = document.getElementById('cpResults');
    if (el) { el.innerHTML = ''; renderRecents(); }
  }
  input?.focus();
});
expose('__searchAISource', function (ev: Event, idx: number) {
  ev.preventDefault();
  const link = (ev.currentTarget as HTMLElement) || (ev.target as HTMLElement).closest('a');
  const kind = link?.dataset.kind || '';
  const idRaw = link?.dataset.id || '';
  const url = link?.dataset.url || '';
  // If URL is present, honor it (some sources may carry a direct URL)
  if (url) {
    hideSearchOverlay();
    // Let the URL drive navigation if it's an app route; otherwise open externally
    if (url.startsWith('/')) {
      window.location.hash = url;
      const navFn = (window as any).navigateSearchResult;
      // Fall through to generic handling
    } else {
      window.open(url, '_blank', 'noopener');
      hideSearchOverlay();
      return;
    }
  }
  const id = parseInt(idRaw, 10);
  const numId = Number.isNaN(id) ? 0 : id;
  // Reuse the same navigation mapping as lexical results
  const navFn = (window as any).navigateSearchResult;
  if (typeof navFn === 'function' && kind && numId) {
    hideSearchOverlay();
    navFn(kind, numId, kind);
  } else if (kind && numId) {
    hideSearchOverlay();
    // Minimal fallback: at least close the palette and hint
    import('./navigation').then(({ showView }) => {
      if (kind === 'compendium' || kind === 'monster' || kind === 'spell' || kind === 'item') {
        (window as any).showCompendium?.();
      } else {
        showView('characters');
      }
    });
  } else {
    hideSearchOverlay();
  }
});

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
  // Ctrl/Cmd+Enter triggers Ask AI (explicit, never on debounce)
  if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
    e.preventDefault();
    doAISearch(input.value);
    return;
  }
  // In AI mode, plain Enter should not navigate a non-existent lexical list
  if (aiMode && e.key === 'Enter' && !e.metaKey && !e.ctrlKey) {
    // Allow Enter to re-trigger Ask AI if query present
    if (input.value.trim()) {
      e.preventDefault();
      doAISearch(input.value);
    }
    return;
  }
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
  }
}

expose('__searchHover', function (index: number) {
  selectedIndex = index;
  updateSelection(document.querySelectorAll('#cpResults .cp-result-item'));
});

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

expose('__searchNavigate', function (type: string, id: number, name: string) {
  hideSearchOverlay();

  // Use the same navigation as the existing legacy system
  const navFn = (window as any).navigateSearchResult;
  if (typeof navFn === 'function') {
    navFn(type, id, name);
  } else {
    // Fallback: basic navigation (entity_type values are singular)
    import('./navigation').then(({ showView }) => {
      if (type === 'character') {
        (window as any).openChar?.(id);
      } else if (type === 'monster' || type === 'compendium') {
        (window as any).showCompendium?.();
      } else {
        showView('characters');
      }
    });
  }
});

// ─── Init ───

export function initSearch(): void {
  // Add Cmd+K / Ctrl+K shortcut for command palette
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      const overlay = document.getElementById('searchOverlay');
      const isOpen = !!overlay && overlay.style.display !== 'none';
      if (isOpen) {
        e.preventDefault();
        if (aiMode) (window as any).__searchAIBack?.();
        else hideSearchOverlay();
        return;
      }
    }
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
expose('doSearch', doSearch);
expose('doAISearch', doAISearch);
