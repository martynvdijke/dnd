import type { ViewState } from '../types';

export {};

declare global {
  interface Window {
    showView: (view: ViewState) => void;
    toggleTheme: () => void;
    toggleEink: () => void;
    showModal: (title: string, bodyHtml: string) => void;
    hideModal: () => void;
    toast: (msg: string, opts?: unknown, duration?: number) => void;
    bootstrap: unknown;
    setDiceExpr: (expr: string) => void;
    rollWithAdvantage: (isAdv: boolean) => Promise<void>;
    renderDiceTab: () => void;
    doRoll: () => Promise<void>;
    loadDiceHistory: () => Promise<void>;
    toggleDiceSound: () => void;
    showSearchOverlay: () => void;
    hideSearchOverlay: () => void;
    doSearch: (query?: string) => Promise<void>;
    initSearch: () => void;
    openPdfViewer: (url: string, title?: string) => void;
    pdfViewerPrevPage: () => void;
    pdfViewerNextPage: () => void;
    pdfViewerZoomIn: () => void;
    pdfViewerZoomOut: () => void;
    pdfViewerFitToWidth: () => void;
    pdfViewerDownload: () => void;
    openAIGenModal: (mode: string, targetId: string, promptHint?: string, title?: string) => void;
    generateWithAI: () => Promise<void>;
    regenerateWithAI: () => void;
    insertAIGenResult: () => void;
    replaceAIGenResult: () => void;
    htmx: { process: (el: Element) => void } & Record<string, unknown>;
    // allow any other expose names without error
    [key: string]: unknown;
  }
}
