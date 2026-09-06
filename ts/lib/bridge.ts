/**
 * Central registration point for all (window as any) assignments.
 *
 * Every function that needs to be accessible from HTML onclick attributes
 * or inline event handlers must be registered here. This keeps the global
 * namespace pollution explicit and auditable.
 *
 * As the app migrates away from inline event handlers, entries here can be
 * removed one by one.
 */
import { showView } from '../navigation';

export { expose } from './expose';

import {
  toggleTheme,
  toggleEink,
} from './theme';

import {
  showModal,
  hideModal,
  toast,
} from './dom';

import {
  setDiceExpr,
  rollWithAdvantage,
  renderDiceTab,
  doRoll,
  loadDiceHistory,
  toggleDiceSound,
} from '../dice';

import {
  showSearchOverlay,
  hideSearchOverlay,
  doSearch,
  initSearch,
} from '../search';

import {
  openPdfViewer,
  pdfViewerPrevPage,
  pdfViewerNextPage,
  pdfViewerZoomIn,
  pdfViewerZoomOut,
  pdfViewerFitToWidth,
  pdfViewerDownload,
} from '../pdf-viewer';

import {
  openAIGenModal,
  generateWithAI,
  regenerateWithAI,
  insertAIGenResult,
  replaceAIGenResult,
} from '../ai';

export function initBridge(): void {
  const w = window as unknown as Record<string, unknown>;

  // Navigation — showView is the single source of truth.
  // Real view handlers (showCompendium/showParty/showCombatTracker/showWiki
  // showEncounterBuilder/showFactions/showTimeline/showDice/showShops/
  // showOneShots) are registered via side-effect imports (./compendium,
  // ./party, ./combat-tracker, ./encounter, ./factions, ./timeline,
  // ./dice, ./app/shops, ./app/oneshot) which do showView PLUS data
  // loading. Do NOT overwrite them with stubs — leave them intact.
  // Only provide fallbacks for views that have no dedicated module.
  w.showView = showView;
  const maybe = (k: string, fn: () => void) => {
    if (typeof w[k] !== 'function') w[k] = fn;
  };
  maybe('showCompendium', () => showView('compendium'));
  maybe('showParty', () => showView('party'));
  maybe('showCombatTracker', () => showView('combatTracker'));
  maybe('showWiki', () => showView('wiki'));
  maybe('showEncounterBuilder', () => showView('encounter'));
  maybe('showFactions', () => showView('factions'));
  maybe('showTimeline', () => showView('timeline'));
  maybe('showDice', () => showView('dice'));
  maybe('showShops', () => showView('shops'));
  maybe('showOneShots', () => showView('oneshot'));

  // Theme
  w.toggleTheme = toggleTheme;
  w.toggleEink = toggleEink;

  // DOM / UI
  w.showModal = showModal;
  w.hideModal = hideModal;
  w.toast = toast;

  // Bootstrap (external lib)
  w.bootstrap = window.bootstrap;

  // Dice
  w.setDiceExpr = setDiceExpr;
  w.rollWithAdvantage = rollWithAdvantage;
  w.renderDiceTab = renderDiceTab;
  w.doRoll = doRoll;
  w.loadDiceHistory = loadDiceHistory;
  w.toggleDiceSound = toggleDiceSound;

  // Search
  w.showSearchOverlay = showSearchOverlay;
  w.hideSearchOverlay = hideSearchOverlay;
  w.doSearch = doSearch;

  // PDF Viewer
  w.openPdfViewer = openPdfViewer;
  w.pdfViewerPrevPage = pdfViewerPrevPage;
  w.pdfViewerNextPage = pdfViewerNextPage;
  w.pdfViewerZoomIn = pdfViewerZoomIn;
  w.pdfViewerZoomOut = pdfViewerZoomOut;
  w.pdfViewerFitToWidth = pdfViewerFitToWidth;
  w.pdfViewerDownload = pdfViewerDownload;

  // AI Generation
  w.openAIGenModal = openAIGenModal;
  w.generateWithAI = generateWithAI;
  w.regenerateWithAI = regenerateWithAI;
  w.insertAIGenResult = insertAIGenResult;
  w.replaceAIGenResult = replaceAIGenResult;
}
