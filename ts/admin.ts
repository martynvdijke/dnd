import { expose } from './lib/expose';
import { api, setCsrfToken, setApiToken, setCurrentUser, currentUser } from './admin/state';
import './admin/site-settings';
import './admin/api-tokens';
import './admin/compendium';
import './admin/entries';
import './admin/logs';
import './admin/users';
import './admin/schemas';
import './admin/integrations';
import './admin/import-wizard';
import './admin/events';
import './admin/utils';
import './admin/pdf';

(() => {

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

async function ensureApiToken(): Promise<void> {
  try {
    const cu = currentUser;
    const key = `villum-api-token-${cu?.username || 'admin'}`;
    const stored = localStorage.getItem(key);
    if (stored) {
      setApiToken(stored);
      return;
    }
    const tokens = await api('GET', '/api/tokens');
    const active = Array.isArray(tokens)
      ? tokens.find((t: any) => !t.revoked_at && (!t.expires_at || new Date(t.expires_at) > new Date()))
      : null;
    if (active) {
      const rotated = await api('POST', `/api/tokens/${active.id}/rotate`);
      localStorage.setItem(key, rotated.token);
      setApiToken(rotated.token);
      return;
    }
    const created = await api('POST', '/api/tokens', { name: 'admin-panel' });
    localStorage.setItem(key, created.token);
    setApiToken(created.token);
  } catch {
  }
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
    await ensureApiToken();
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
  const allTabs = ['users', 'unified-compendium', 'backup', 'email', 'push', 'ai-endpoints', 'analytics', 'telemetry', 'events', 'import', 'e-ink', 'settings', 'logs'];
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
