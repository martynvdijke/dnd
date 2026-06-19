/**
 * Global search module — search overlay, query handling, results display.
 */
import { esc, toast } from './lib/dom';
import { api } from './lib/api';
import { showView, getCurrentView } from './navigation';

// ─── Search Overlay ───

export function showSearchOverlay() {
  let overlay = document.getElementById('searchOverlay');
  if (!overlay) {
    overlay = document.createElement('div');
    overlay.id = 'searchOverlay';
    overlay.className = 'search-overlay';
    overlay.addEventListener('click', (e) => { if (e.target === overlay) hideSearchOverlay(); });
    document.body.appendChild(overlay);
    const panel = document.createElement('div');
    panel.id = 'searchPanel';
    panel.className = 'search-panel';
    overlay.appendChild(panel);
  }
  overlay.style.display = 'flex';
}

export function hideSearchOverlay() {
  const overlay = document.getElementById('searchOverlay');
  if (overlay) overlay.style.display = 'none';
}

export async function doSearch() {
  const q = (document.getElementById('searchInput') as HTMLInputElement)?.value?.trim();
  if (!q) return;
  try {
    const results = await api('GET', '/api/search?q=' + encodeURIComponent(q));
    let html = '';
    let total = 0;
    const sections: Record<string, { label: string; icon: string; items: any[] }> = {
      characters:  { label: 'Characters',  icon: 'fa-users',     items: results.characters },
      npcs:        { label: 'NPCs',         icon: 'fa-user-group', items: results.npcs },
      notes:       { label: 'Notes',        icon: 'fa-note-sticky', items: results.notes },
      quests:      { label: 'Quests',       icon: 'fa-scroll',    items: results.quests },
      journal:     { label: 'Journal',      icon: 'fa-book-open', items: results.journal },
      sessions:    { label: 'Sessions',     icon: 'fa-calendar',  items: results.sessions },
      campaigns:   { label: 'Campaigns',    icon: 'fa-flag',      items: results.campaigns },
      spells:      { label: 'Spells',       icon: 'fa-wand-sparkles', items: results.spells },
      equipment:   { label: 'Equipment',    icon: 'fa-backpack',  items: results.equipment },
      races:       { label: 'Races',        icon: 'fa-person',    items: results.races },
      classes:     { label: 'Classes',      icon: 'fa-graduation-cap', items: results.classes },
      feats:       { label: 'Feats',        icon: 'fa-star',      items: results.feats },
      backgrounds: { label: 'Backgrounds',  icon: 'fa-address-card', items: results.backgrounds },
      monsters:    { label: 'Monsters',     icon: 'fa-dragon',    items: results.monsters },
    };
    for (const [key, sec] of Object.entries(sections)) {
      if (sec.items.length === 0) continue;
      total += sec.items.length;
      html += `<h6 class="mt-3 mb-2"><i class="fa-solid ${sec.icon} me-2 text-muted"></i>${sec.label} (${sec.items.length})</h6>`;
      for (const item of sec.items) {
        html += `<div class="search-result-item" onclick="navigateSearchResult('${key}',${item.id},'${esc(item.name)}');hideSearchOverlay()">
          <div class="fw-bold small">${esc(item.name)}</div>
          ${item.snippet ? `<div class="text-muted small">${item.snippet}</div>` : ''}
        </div>`;
      }
    }
    if (total === 0) {
      html = `<div class="empty-state"><i class="fa-solid fa-search fa-2x mb-2 d-block text-muted"></i><p class="fw-bold">No Results</p><p class="small text-muted">No matches found for "${esc(q)}".</p></div>`;
    }
    showSearchOverlay();
    const panel = document.getElementById('searchPanel');
    if (panel) {
      panel.innerHTML = `<div class="d-flex justify-content-between align-items-center mb-2"><h5 class="mb-0">Search Results${total > 0 ? ` (${total})` : ''}</h5><button class="btn btn-sm btn-outline-secondary" onclick="hideSearchOverlay()"><i class="fa-solid fa-xmark"></i></button></div>${html}`;
    }
  } catch (e: any) {
    toast(e.message, true);
  }
};

export function initSearch() {
  const input = document.getElementById('searchInput');
  if (!input) return;
  input.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') doSearch();
  });
  const btn = document.getElementById('searchBtn');
  if (btn) btn.addEventListener('click', doSearch);
}
