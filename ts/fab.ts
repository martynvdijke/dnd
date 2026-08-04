import type { FABAction, ViewState } from './types';
import { expose } from './lib/expose';

const actionMap: Record<string, FABAction[]> = {
  characters: [
    { id: 'new-char', label: 'New Character', icon: 'fa-plus', onclick: 'newChar()' },
    { id: 'import-char', label: 'Import Character', icon: 'fa-file-import', onclick: 'showImport()' },
  ],
  sheet: [
    { id: 'roll-dice', label: 'Roll Dice', icon: 'fa-dice', onclick: "showView('dice');renderDiceTab()" },
    { id: 'short-rest', label: 'Short Rest', icon: 'fa-campground', onclick: "doRest('short')" },
    { id: 'long-rest', label: 'Long Rest', icon: 'fa-moon', onclick: "doRest('long')" },
  ],
  dice: [
    { id: 'roll-d20', label: 'Roll d20', icon: 'fa-dice-d20', onclick: "document.getElementById('diceExpr')?.focus()" },
    { id: 'roll-custom', label: 'Roll Custom', icon: 'fa-calculator', onclick: "document.getElementById('diceExpr')?.focus()" },
  ],
  party: [
    { id: 'refresh', label: 'Refresh', icon: 'fa-rotate', onclick: 'showParty()' },
  ],
  compendium: [
    { id: 'search', label: 'Search', icon: 'fa-search', onclick: "document.getElementById('searchInput')?.focus()" },
    { id: 'browse', label: 'Browse All', icon: 'fa-list', onclick: "loadCompendiumTab('races')" },
  ],
  combatTracker: [
    { id: 'roll-init', label: 'Roll Initiative', icon: 'fa-gauge-high', onclick: 'rollAllInitiative()' },
    { id: 'end-turn', label: 'End Turn', icon: 'fa-forward', onclick: 'nextTurn()' },
  ],
  encounter: [
    { id: 'new-encounter', label: 'New Encounter', icon: 'fa-plus', onclick: 'showNewEncounter()' },
  ],
  timeline: [
    { id: 'new-event', label: 'New Event', icon: 'fa-plus', onclick: 'showNewTimelineEvent()' },
  ],
  oneshot: [
    { id: 'new-oneshot', label: 'New Adventure', icon: 'fa-plus', onclick: 'showNewOneShot()' },
  ],
  factions: [
    { id: 'new-faction', label: 'New Faction', icon: 'fa-plus', onclick: 'showNewFaction()' },
  ],
  shops: [
    { id: 'new-shop', label: 'New Shop', icon: 'fa-plus', onclick: 'showNewShop()' },
  ],
};

const sessionModeActions: FABAction[] = [
  { id: 'roll-dice', label: 'Roll Dice', icon: 'fa-dice', onclick: "showView('dice');renderDiceTab()" },
  { id: 'end-turn', label: 'End Turn', icon: 'fa-forward', onclick: 'nextTurn()' },
];

export function getActionsForView(view: ViewState, isSessionMode: boolean = false): FABAction[] {
  if (isSessionMode) return sessionModeActions;
  return actionMap[view] || actionMap.characters;
}

export function toggleFabMenu(): void {
  const menu = document.getElementById('fabMenu');
  if (menu) menu.classList.toggle('open');
}

expose('toggleFabMenu', toggleFabMenu);

export function updateFabForView(view: ViewState, isSessionMode: boolean = false): void {
  const menu = document.getElementById('fabMenu');
  if (!menu) return;
  const actions = getActionsForView(view, isSessionMode);
  menu.innerHTML = actions.map(a =>
    `<button class="fab-menu-item" onclick="${a.onclick};toggleFabMenu()">
      <i class="fa-solid ${a.icon} me-2" aria-hidden="true"></i>${a.label}
    </button>`
  ).join('');
  menu.classList.remove('open');
}
