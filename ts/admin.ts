import { expose } from './lib/expose';
import { api, setCsrfToken, setApiToken, setCurrentUser, clearApiToken, setUnauthorizedHandler } from './admin/state';
import './admin/site-settings';
import './admin/api-tokens';
import './admin/compendium';
import './admin/entries';
import './admin/logs';
import './admin/users';
import './admin/schemas';
import './admin/integrations';
import './admin/telegram';
import './admin/import-wizard';
import './admin/events';
import './admin/utils';
import './admin/pdf';

(() => {

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

async function provisionApiToken(key: string): Promise<void> {
  try {
    // Mint a fresh, per-device token rather than rotating an existing one, which
    // would invalidate the secret another device already stored.
    const created = await api('POST', '/api/tokens', { name: 'admin-panel' });
    localStorage.setItem(key, created.token);
    setApiToken(created.token);
  } catch {
  }
}

async function ensureApiToken(username: string): Promise<void> {
  const key = `villum-api-token-${username}`;
  const stored = localStorage.getItem(key);
  if (stored) {
    setApiToken(stored);
    return;
  }
  await provisionApiToken(key);
}

function toggleTheme() {
  const html = document.documentElement;
  const isDark = html.getAttribute('data-theme') === 'dark';
  const newTheme = isDark ? 'light' : 'dark';
  html.setAttribute('data-theme', newTheme);
  localStorage.setItem('villum-theme', newTheme);
  const icon = document.getElementById('themeIcon');
  if (icon) icon.className = isDark ? 'fa-solid fa-moon' : 'fa-solid fa-sun';
}
expose('toggleTheme', toggleTheme);

function initTheme() {
  const saved = localStorage.getItem('villum-theme') || 'light';
  document.documentElement.setAttribute('data-theme', saved);
  const icon = document.getElementById('themeIcon');
  if (icon) icon.className = saved === 'dark' ? 'fa-solid fa-sun' : 'fa-solid fa-moon';
}

async function init() {
  initTheme();
  try {
    const cu = await api('GET', '/api/user/me');
    setCurrentUser(cu);
    if (cu.role !== 'admin') {
      window.location.href = '/';
      return;
    }
    const tokenRes = await api('GET', '/api/csrf-token');
    setCsrfToken(tokenRes.token);
    // Self-heal a stale API token on a mutation 401: drop the stored secret,
    // provision a fresh per-device token, and let api() retry once.
    setUnauthorizedHandler(async () => {
      const key = `villum-api-token-${cu.username}`;
      localStorage.removeItem(key);
      clearApiToken();
      await provisionApiToken(key);
    });
    await ensureApiToken(cu.username);
    showAdminTab('users');
    (window as any).loadUsers?.();
  } catch {
    window.location.href = '/login';
  }
}

function showAdminTab(tab: string) {
  document.querySelectorAll('#adminTabs .nav-link').forEach(el => el.classList.remove('active'));
  const tabBtn = document.getElementById('tab' + capitalize(tab) + 'Btn');
  if (tabBtn) tabBtn.classList.add('active');
  const allTabs = ['users', 'unified-compendium', 'backup', 'email', 'push', 'telegram', 'wled', 'ai-endpoints', 'analytics', 'telemetry', 'events', 'import', 'e-ink', 'settings', 'logs'];
  allTabs.forEach(s => {
    const parts = s.split('-').map((p, i) => i === 0 ? capitalize(p) : capitalize(p));
    const id = 'admin' + parts.join('');
    const el = document.getElementById(id);
    if (el) el.style.display = s === tab ? 'block' : 'none';
  });
  const w = window as any;
  if (tab === 'users') w.loadUsers?.();
  if (tab === 'unified-compendium') w.loadUnifiedCompendium?.();
  if (tab === 'backup') { w.loadBackupSettings?.(); w.loadBackupList?.(); }
  if (tab === 'email') w.loadEmailSettings?.();
  if (tab === 'push') w.loadPushSettings?.();
  if (tab === 'telegram') w.loadTelegramSettings?.();
  if (tab === 'wled') w.loadWledSettings?.();
  if (tab === 'ai-endpoints') w.loadAIEndpoints?.();
  if (tab === 'analytics') w.loadUmamiSettings?.();
  if (tab === 'telemetry') w.loadOTelSettings?.();
  if (tab === 'events') { w.loadEventsSettings?.(); w.loadCampaignEventSettings?.(); w.loadEventsPublicLink?.(); }
  if (tab === 'import') { w.loadImportSchemas?.(); w.loadImportLogs?.(); }
  if (tab === 'e-ink') w.loadEinkSetting?.();
  if (tab === 'settings') { w.loadAutoSaveSetting?.(); w.loadApiTokens?.(); w.loadAISetting?.(); }
  if (tab === 'logs') { w.startLogAutoRefresh?.(); }
  else { w.stopLogAutoRefresh?.(); }
}
expose('showAdminTab', showAdminTab);

init();

})();
