/**
 * Application initialization and startup orchestration.
 *
 * Manages the init() lifecycle: bootstrap lib modules, authenticate,
 * load initial data, and establish WebSocket connection.
 *
 * Shared state (currentUser, currentChar, allLocations, allNPCs) is
 * managed in app.ts and mutated via window-level references during
 * the transition from the monolithic init() to modular ownership.
 */
import { showView, showViewFromRouter, getCurrentView } from './navigation';
import { initRouter, navigateToInitialHash } from './router';
import { initBridge } from './lib/bridge';
import { initTheme } from './lib/theme';
import { initShortcuts } from './lib/shortcuts';
import { api, setCsrfToken, getCsrfToken, setApiToken, getApiToken } from './lib/api';
import { initSearch } from './search';
import { initAIClickHandler } from './ai';
import { initPdfViewerCleanup } from './pdf-viewer';
import { setCurrentUser, setAllLocations, setAllNPCs } from './lib/state';
import { initSpellCompendium } from './spell-compendium';

// ─── WebSocket ───

let ws: WebSocket | null = null;
let wsReconnectTimer: any = null;

function connectWS() {
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return;
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  ws = new WebSocket(`${proto}//${window.location.host}/api/ws`);
  ws.onmessage = (ev) => {
    try {
      const msg = JSON.parse(ev.data);
      if (msg.type === 'character_update' && getCurrentView() === 'sheet') {
        (window as any).openChar(msg.payload.character_id);
      }
      if (msg.type === 'party_update' && getCurrentView() === 'party') {
        (window as any).showParty();
      }
    } catch {}
  };
  ws.onclose = () => {
    ws = null;
    wsReconnectTimer = setTimeout(connectWS, 5000);
  };
  ws.onerror = () => ws?.close();
}

// ─── API token bootstrap ───

const API_TOKEN_KEY = 'villum-api-token';

// ensureApiToken provisions a user API token for the web app and stores the
// one-time secret in localStorage. Token lifecycle endpoints are session +
// CSRF protected, so this works before any token exists. Failures are
// non-fatal: mutations will surface a 401 that the user can recover from.
async function ensureApiToken(): Promise<void> {
  try {
    const stored = localStorage.getItem(API_TOKEN_KEY);
    if (stored) {
      setApiToken(stored);
      return;
    }
    const tokens = await api('GET', '/api/tokens');
    const active = Array.isArray(tokens)
      ? tokens.find((t: any) => !t.revoked_at && (!t.expires_at || new Date(t.expires_at) > new Date()))
      : null;
    if (active) {
      // An active token exists but its secret was never stored locally
      // (e.g. created from another device) — rotate to obtain a fresh secret.
      const rotated = await api('POST', `/api/tokens/${active.id}/rotate`);
      localStorage.setItem(API_TOKEN_KEY, rotated.token);
      setApiToken(rotated.token);
      return;
    }
    const created = await api('POST', '/api/tokens', { name: 'web-app' });
    localStorage.setItem(API_TOKEN_KEY, created.token);
    setApiToken(created.token);
  } catch {
    // Token bootstrap must not break the app shell.
  }
}

// ─── Init — called from app.ts after imports ───

export async function init() {
  initBridge();
  initTheme();
  initShortcuts();
  initSearch();
  initAIClickHandler();
  initPdfViewerCleanup();
  initSpellCompendium();
  // Initialize hash router — handles back/forward and bookmarks
  initRouter((route) => showViewFromRouter(route.view));
  try {
    const user = await api('GET', '/api/user/me');
    setCurrentUser(user);
    const tokenRes = await api('GET', '/api/csrf-token');
    setCsrfToken(tokenRes.token);
    document.querySelector('meta[name="csrf-token"]')?.setAttribute('content', getCsrfToken());
    document.body.addEventListener('htmx:configRequest', (e: any) => {
      e.detail.headers['X-CSRF-Token'] = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || getCsrfToken();
      const method = (e.detail.verb || 'get').toUpperCase();
      const token = getApiToken();
      if (token && method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS') {
        e.detail.headers['Authorization'] = `Bearer ${token}`;
      }
    });
    await ensureApiToken();
    document.getElementById('userName')!.textContent = user.username;

    // Top navbar visibility
    const show = (id: string) => { const el = document.getElementById(id); if (el) el.style.display = ''; };
    if (user.role === 'admin') {
      show('adminNavItem');
      show('sidebarAdminNav');
    }
    if (user.role === 'admin' || user.role === 'dm') {
      show('combatNavItem');
      show('sidebarCombatNav');
      show('oneshotNavItem');
      show('sidebarOneshotNav');
      show('factionsNavItem');
      show('sidebarFactionsNav');
      show('shopsNavItem');
      show('sidebarShopsNav');
    }
    // If URL has a hash (e.g. #/compendium), navigate to that view
    // instead of default characters view
    const selectionOk = await (window as any).validateSelection();
    if (!selectionOk) {
      // No valid campaign selected — the picker view is already shown.
      connectWS();
      api('GET', '/api/locations').then(setAllLocations).catch(() => {});
      api('GET', '/api/npcs').then(setAllNPCs).catch(() => {});
      return;
    }
    if (location.hash && location.hash.length > 1) {
      navigateToInitialHash((route) => showViewFromRouter(route.view));
    } else {
      showView('characters');
    }
    (window as any).loadCharacters();
    connectWS();
    api('GET', '/api/locations').then(setAllLocations).catch(() => {});
    api('GET', '/api/npcs').then(setAllNPCs).catch(() => {});
  } catch (err) {
    // Offline (service worker serving the cached shell): stay on the app
    // instead of redirecting to /login — cached assets keep the UI usable.
    if (navigator.onLine === false) {
      showView('characters');
    } else {
      window.location.href = '/login';
    }
  }
}
