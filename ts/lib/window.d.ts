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
    htmx: { process: (el: Element) => void; trigger: (el: Element, event: string) => void; ajax: (method: string, path: string, opts: unknown) => void } & Record<string, unknown>;
    showCombatTracker: () => Promise<void>;
    showEncounterBuilder: () => Promise<void>;
    showEncounterDetail: (id: number) => Promise<void>;
    showParty: () => Promise<void>;
    showManageCampaign: (id: number, name: string, partyName?: string) => Promise<void>;
    showFactions: () => void;
    showTimeline: () => void;
    loadCompendiumTab: (tab: string) => void;
    renameParty: (id: number, name: string, desc: string) => void;
    showCampaignDashboard: (id: number, name: string) => void;
    showCharStatsModal: (id: number) => Promise<void>;
    showCharNotes: (id: number) => void;
    renderStepper: (field: string, value: number, delta: number, min?: number, max?: number, label?: string, size?: string) => string;
    renderStats: () => void;
    renderCombat: () => void;
    renderResources: () => Promise<void> | void;
    renderInventory: () => void;
    renderSheet: () => void;
    renderSpells: () => void;
    updateField: (field: string, value: unknown) => void;
    saveCharacter: () => Promise<void>;
    autoSaveField: (field: string, el: HTMLElement) => void;
    stepperField: (field: string, delta: number, min?: number, max?: number) => void;
    editStepperValue: (field: string, el: HTMLElement) => void;
    coinStepper: (coin: string, delta: number) => Promise<void>;
    updateSaveBtnState: () => void;
    canEditCharacter: boolean;
    renderCrafting: () => void;
    renderDetails: () => void;
    updateXPBar: () => void;
    // allow any other expose names without error
    [key: string]: unknown;
  }
}
