import { expose } from '../lib/expose';
import { esc, toast } from '../lib/dom';
import { api } from './state';
import { renderError } from '../lib/errors';
import {
  adminTelegramStatusText,
  maskedTokenDisplay,
} from '../lib/telegram';
import type { AdminTelegramSettings } from '../lib/telegram';

function setStatusText(s: AdminTelegramSettings | null) {
  const el = document.getElementById('telegramStatus');
  if (!el) return;
  if (!s) {
    el.textContent = 'Not configured';
    return;
  }
  el.textContent = adminTelegramStatusText(s);
}

function setTokenHint(s: AdminTelegramSettings | null) {
  const el = document.getElementById('telegramTokenMasked');
  if (!el) return;
  if (!s) {
    el.textContent = 'No token configured';
    return;
  }
  el.textContent = maskedTokenDisplay(s.token_masked || '', !!s.has_token);
}

function updateWebhookVisibility(mode: string) {
  const field = document.getElementById('telegramWebhookField');
  if (!field) return;
  field.style.display = (mode === 'webhook' || mode === 'auto') ? 'block' : 'none';
}

// Bind mode change to webhook visibility when DOM is ready
if (typeof document !== 'undefined') {
  document.addEventListener('DOMContentLoaded', () => {
    const sel = document.getElementById('telegramMode') as HTMLSelectElement | null;
    if (sel) sel.addEventListener('change', () => updateWebhookVisibility(sel.value));
  });
  // Also bind immediately if element already exists (admin page may not have DOMContentLoaded after module load)
  const immediate = document.getElementById('telegramMode') as HTMLSelectElement | null;
  if (immediate) immediate.addEventListener('change', () => updateWebhookVisibility(immediate.value));
}

export async function getAdminTelegramSettings(): Promise<AdminTelegramSettings> {
  return api('GET', '/api/admin/telegram-settings');
}

export async function saveAdminTelegramSettings(payload: {
  token?: string;
  mode?: string;
  grace_minutes?: number;
  clear_token?: boolean;
}): Promise<AdminTelegramSettings> {
  return api('POST', '/api/admin/telegram-settings', payload);
}

export async function testAdminTelegram(): Promise<{ sent: boolean }> {
  return api('POST', '/api/admin/telegram-test');
}

async function loadTelegramSettings() {
  try {
    const s = await getAdminTelegramSettings();
    const tokenInput = document.getElementById('telegramToken') as HTMLInputElement | null;
    if (tokenInput) {
      tokenInput.value = '';
      tokenInput.placeholder = s.has_token ? (s.token_masked || 'Token is set — leave blank to keep') : 'Not set — paste bot token from @BotFather';
    }
    const modeSel = document.getElementById('telegramMode') as HTMLSelectElement | null;
    if (modeSel) {
      modeSel.value = s.mode || 'off';
      updateWebhookVisibility(modeSel.value);
    }
    const grace = document.getElementById('telegramGraceMinutes') as HTMLInputElement | null;
    if (grace) grace.value = String(s.grace_minutes ?? 30);
    const webhookEl = document.getElementById('telegramWebhookUrl') as HTMLInputElement | null;
    if (webhookEl) webhookEl.value = s.webhook_url || '';
    setStatusText(s);
    setTokenHint(s);
  } catch (e: unknown) {
    setStatusText(null);
    // Avoid toast spam on initial load when endpoint may not yet exist (parallel backend build)
    const msg = e instanceof Error ? e.message : String(e);
    if (!msg.includes('404') && !msg.includes('Not found')) {
      // silent — admin will retry on interaction
    }
  }
}

async function saveTelegramSettings() {
  const tokenRaw = (document.getElementById('telegramToken') as HTMLInputElement | null)?.value.trim() || '';
  const mode = (document.getElementById('telegramMode') as HTMLSelectElement | null)?.value || 'off';
  const graceRaw = (document.getElementById('telegramGraceMinutes') as HTMLInputElement | null)?.value || '';
  const grace = graceRaw ? Number(graceRaw) : undefined;
  const payload: Record<string, unknown> = {};
  if (tokenRaw) payload.token = tokenRaw;
  if (mode) payload.mode = mode;
  if (grace !== undefined && Number.isFinite(grace)) payload.grace_minutes = Math.floor(grace);
  try {
    const s = await saveAdminTelegramSettings(payload as any);
    const tokenInput = document.getElementById('telegramToken') as HTMLInputElement | null;
    if (tokenInput) tokenInput.value = '';
    if (s) {
      const modeSel = document.getElementById('telegramMode') as HTMLSelectElement | null;
      if (modeSel) modeSel.value = s.mode || mode;
      const graceEl = document.getElementById('telegramGraceMinutes') as HTMLInputElement | null;
      if (graceEl) graceEl.value = String(s.grace_minutes ?? graceRaw ?? 30);
      setStatusText(s);
      setTokenHint(s);
    }
    toast('Telegram settings saved');
  } catch (e: unknown) {
    renderError(e);
  }
}

async function testTelegramSettings() {
  const btn = document.getElementById('telegramTestBtn') as HTMLButtonElement | null;
  if (btn) {
    btn.disabled = true;
    btn.innerHTML = '<i class="fa-solid fa-spinner fa-spin me-1"></i>Sending...';
  }
  try {
    const r = await testAdminTelegram();
    toast(r.sent ? 'Test message sent' : 'Test did not send — check bot token and chat setup');
  } catch (e: unknown) {
    renderError(e);
  } finally {
    if (btn) {
      btn.disabled = false;
      btn.innerHTML = '<i class="fa-solid fa-paper-plane me-1"></i>Send test message';
    }
  }
}

async function clearTelegramToken() {
  if (!confirm('Clear the bot token? Telegram delivery will stop until a new token is saved.')) return;
  try {
    const s = await saveAdminTelegramSettings({ clear_token: true } as any);
    const tokenInput = document.getElementById('telegramToken') as HTMLInputElement | null;
    if (tokenInput) tokenInput.value = '';
    setStatusText(s);
    setTokenHint(s);
    toast('Token cleared');
  } catch (e: unknown) {
    renderError(e);
  }
}

expose('loadTelegramSettings', loadTelegramSettings);
expose('saveTelegramSettings', saveTelegramSettings);
expose('testTelegramSettings', testTelegramSettings);
expose('clearTelegramToken', clearTelegramToken);
