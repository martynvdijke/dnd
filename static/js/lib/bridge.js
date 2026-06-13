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
import { toggleTheme, } from './theme';
import { showModal, hideModal, toast, } from './dom';
import { setDiceExpr, rollWithAdvantage, renderDiceTab, doRoll, loadDiceHistory, } from '../dice';
import { showSearchOverlay, hideSearchOverlay, doSearch, } from '../search';
import { openPdfViewer, pdfViewerPrevPage, pdfViewerNextPage, pdfViewerZoomIn, pdfViewerZoomOut, pdfViewerFitToWidth, pdfViewerDownload, } from '../pdf-viewer';
import { openAIGenModal, generateWithAI, regenerateWithAI, insertAIGenResult, replaceAIGenResult, } from '../ai';
export function initBridge() {
    const w = window;
    // Navigation
    w.showView = showView;
    // Theme
    w.toggleTheme = toggleTheme;
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
