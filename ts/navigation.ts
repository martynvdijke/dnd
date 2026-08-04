import type { ViewState } from './types';
import { openBottomSheet } from './bottom-sheet';
import { updateFabForView } from './fab';
import { navigate as routerNavigate } from './router';
import { expose } from './lib/expose';

export interface ViewItem {
  id: ViewState;
  divId: string;
}

const views: ViewItem[] = [
  { id: 'characters', divId: 'charactersView' },
  { id: 'sheet', divId: 'sheetView' },
  { id: 'dice', divId: 'diceView' },
  { id: 'compendium', divId: 'compendiumView' },
  { id: 'party', divId: 'partyView' },
  { id: 'encounter', divId: 'encounterView' },
  { id: 'timeline', divId: 'timelineView' },
  { id: 'combatTracker', divId: 'combatTrackerView' },
  { id: 'wiki', divId: 'wikiView' },
  { id: 'oneshot', divId: 'oneshotView' },
  { id: 'factions', divId: 'factionsView' },
  { id: 'shops', divId: 'shopsView' },
  { id: 'singleEncounter', divId: 'singleEncounterView' },
  { id: 'campaignOverview', divId: 'campaignOverviewView' },
];

export let currentView: ViewState = 'characters';
let navigatingFromRouter = false;
let initialLoad = true;

export function setCurrentView(view: ViewState): void {
  currentView = view;
}

export function getCurrentView(): ViewState {
  return currentView;
}

/**
 * Internal implementation of view switching.
 */
function switchView(view: ViewState): void {
  currentView = view;
  views.forEach(v => {
    const el = document.getElementById(v.divId);
    if (el) {
      const isVisible = v.id === view ||
        (view === 'sheet' && v.id === 'characters');
      el.style.display = isVisible ? 'block' : 'none';
    }
  });
  updateActiveTab(view);
  updateFabForView(view);
}

/**
 * Show a view and update the URL hash.
 */
export function showView(view: ViewState): void {
  switchView(view);
  if (!navigatingFromRouter && !initialLoad) {
    routerNavigate(view);
  }
  initialLoad = false;
}

/**
 * Called by the router on hashchange — updates the view without
 * touching the hash (to avoid loops).
 */
export function showViewFromRouter(view: ViewState): void {
  navigatingFromRouter = true;
  switchView(view);
  navigatingFromRouter = false;
}

export function updateActiveTab(view: ViewState): void {
  document.body.setAttribute('data-active-tab', view);
  const tabs = document.querySelectorAll<HTMLElement>('[data-nav]');
  tabs.forEach(tab => {
    const navView = tab.getAttribute('data-nav');
    tab.classList.toggle('active', navView === view || (view === 'sheet' && navView === 'characters'));
  });
}

expose('showView', showView);
expose('getCurrentView', getCurrentView);
expose('setCurrentView', setCurrentView);

export function toggleSidebar(): void {
  const sidebar = document.getElementById('appSidebar');
  if (sidebar) {
    sidebar.classList.toggle('collapsed');
    document.body.classList.toggle('sidebar-collapsed');
    document.body.classList.toggle('sidebar-expanded');
  }
}
expose('toggleSidebar', toggleSidebar);

export function showMoreNav(): void {
  // Mirror visibility from corresponding top nav items (set by init based on role)
  const show = (id: string) => document.getElementById(id)?.style.display !== 'none' ? '' : 'display:none';
  const content = `
    <div class="d-flex flex-column gap-1">
      <button class="btn btn-outline-primary w-100 text-start" onclick="showEncounterBuilder();closeBottomSheet()"><i class="fa-solid fa-crosshairs me-2" aria-hidden="true"></i>Encounters</button>
      <button class="btn btn-outline-primary w-100 text-start" onclick="showCombatTracker();closeBottomSheet()" style="${show('combatNavItem')}"><i class="fa-solid fa-swords me-2" aria-hidden="true"></i>Combat</button>
      <button class="btn btn-outline-primary w-100 text-start" onclick="showTimeline();closeBottomSheet()"><i class="fa-solid fa-timeline me-2" aria-hidden="true"></i>Timeline</button>
      <button class="btn btn-outline-primary w-100 text-start" onclick="showOneShots();closeBottomSheet()" style="${show('oneshotNavItem')}"><i class="fa-solid fa-scroll me-2" aria-hidden="true"></i>One-Shots</button>
      <button class="btn btn-outline-primary w-100 text-start" onclick="showFactions();closeBottomSheet()" style="${show('factionsNavItem')}"><i class="fa-solid fa-flag me-2" aria-hidden="true"></i>Factions</button>
      <button class="btn btn-outline-primary w-100 text-start" onclick="showShops();closeBottomSheet()" style="${show('shopsNavItem')}"><i class="fa-solid fa-store me-2" aria-hidden="true"></i>Shops</button>
      <button class="btn btn-outline-primary w-100 text-start" onclick="window.location.href='/admin'" style="${show('adminNavItem')}"><i class="fa-solid fa-shield-halved me-2" aria-hidden="true"></i>Admin</button>
    </div>
  `;
  openBottomSheet({ id: 'more-nav', title: 'More', content });
}
expose('showMoreNav', showMoreNav);
