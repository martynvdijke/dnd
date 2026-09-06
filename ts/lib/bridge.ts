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

  // Navigation — centralized wrappers so window.show* never depends on
  // fragile side-effect imports. showView is the single source of truth.
  w.showView = showView;
  w.showCompendium = () => showView('compendium');
  w.showEncounterBuilder = () => showView('encounter');
  w.showFactions = () => showView('factions');
  w.showParty = () => showView('party');
  w.showTimeline = () => showView('timeline');
  w.showCombatTracker = () => showView('combatTracker');
  w.showDice = () => showView('dice');
  w.showWiki = () => showView('wiki');
  w.showShops = () => showView('shops');
  w.showOneShots = () => showView('oneshot');

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
