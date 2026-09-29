import { expose } from '../lib/expose';
import { esc, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';
import { renderError } from '../lib/errors';
import { formatLinkCodeDisplay } from '../lib/telegram';
import type { TelegramStatus, TelegramLinkCode } from '../lib/telegram';

export async function getTelegramStatus(): Promise<TelegramStatus> {
  return api('GET', '/api/telegram/status');
}

export async function generateLinkCode(): Promise<TelegramLinkCode> {
  return api('POST', '/api/telegram/link-code');
}

export async function unlinkTelegram(): Promise<{ ok: true }> {
  return api('DELETE', '/api/telegram/unlink');
}

export async function updateTelegramPrefs(dmEnabled: boolean): Promise<{ ok: true }> {
  return api('PUT', '/api/telegram/prefs', { dm_enabled: !!dmEnabled });
}

async function renderTelegramAccountContent(containerId: string) {
  const container = document.getElementById(containerId);
  if (!container) return;
  container.innerHTML = '<div class="text-center py-3"><i class="fa-solid fa-spinner fa-spin me-2"></i>Loading...</div>';
  let status: TelegramStatus;
  try {
    status = await getTelegramStatus();
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : String(e);
    if (msg.includes('404')) {
      // No status endpoint yet — treat as unlinked
      status = { linked: false, telegram_username: '', dm_enabled: false };
    } else {
      container.innerHTML = `<div class="alert alert-danger">${esc(msg)}</div>`;
      return;
    }
  }

  if (status.linked) {
    const username = status.telegram_username ? `@${esc(status.telegram_username)}` : 'linked';
    container.innerHTML = `
      <div class="mb-3">
        <span class="badge bg-success" data-testid="telegram-linked-badge"><i class="fa-solid fa-link me-1"></i>Linked ${username}</span>
      </div>
      <div class="form-check form-switch mb-3">
        <input class="form-check-input" type="checkbox" id="telegramDmEnabled" ${status.dm_enabled ? 'checked' : ''} data-testid="telegram-dm-toggle" onchange="toggleTelegramDm(this.checked)">
        <label class="form-check-label" for="telegramDmEnabled">DM me recaps</label>
      </div>
      <div class="d-flex gap-2">
        <button class="btn btn-outline-danger btn-sm" onclick="unlinkTelegramAccount()" data-testid="telegram-unlink"><i class="fa-solid fa-unlink me-1"></i>Unlink Telegram</button>
        <button class="btn btn-outline-secondary btn-sm" onclick="hideModal()">Close</button>
      </div>
    `;
  } else {
    container.innerHTML = `
      <p class="small text-muted mb-3">Link your Telegram account to receive recap notifications and use the bot's <code>/recap</code> commands.</p>
      <button class="btn btn-primary w-100 mb-3" onclick="generateTelegramLinkCode()" data-testid="telegram-generate-code"><i class="fa-brands fa-telegram me-1"></i>Generate link code</button>
      <div id="telegramLinkCodeArea" style="display:none" class="border rounded p-3 mb-3 bg-light">
        <label class="form-label small mb-1">Your code</label>
        <div class="input-group mb-2">
          <input class="form-control font-monospace" id="telegramLinkCode" readonly data-testid="telegram-code">
          <button class="btn btn-outline-primary" onclick="copyTelegramCode()" data-testid="telegram-copy-code" title="Copy code"><i class="fa-solid fa-copy"></i></button>
        </div>
        <label class="form-label small mb-1">Deep link</label>
        <div class="input-group mb-2">
          <input class="form-control font-monospace" id="telegramDeepLink" readonly data-testid="telegram-deep-link">
          <button class="btn btn-outline-primary" onclick="copyTelegramLink()" data-testid="telegram-copy-link" title="Copy link"><i class="fa-solid fa-link"></i></button>
        </div>
        <div class="small text-muted mb-2">This code expires in 15 minutes and can be used once. Send <code>/start &lt;code&gt;</code> to the bot.</div>
        <a id="telegramDeepLinkAnchor" href="#" target="_blank" rel="noopener" class="btn btn-outline-secondary btn-sm w-100" data-testid="telegram-open-bot"><i class="fa-brands fa-telegram me-1"></i>Open in Telegram</a>
      </div>
      <div class="form-check form-switch mb-3">
        <input class="form-check-input" type="checkbox" id="telegramDmEnabledUnlinked" data-testid="telegram-dm-toggle" onchange="toggleTelegramDm(this.checked)">
        <label class="form-check-label" for="telegramDmEnabledUnlinked">DM me recaps (applies after linking)</label>
      </div>
      <div class="text-center"><button class="btn btn-outline-secondary btn-sm" onclick="hideModal()">Close</button></div>
    `;
  }
}

export async function showTelegramLink(): Promise<void> {
  showModal('Link Telegram', `<div id="telegramAccountContent"><div class="text-center py-3"><i class="fa-solid fa-spinner fa-spin"></i> Loading...</div></div>`);
  await renderTelegramAccountContent('telegramAccountContent');
}

export async function generateTelegramLinkCode(): Promise<void> {
  const btn = document.querySelector<HTMLButtonElement>('[data-testid="telegram-generate-code"]');
  if (btn) {
    btn.disabled = true;
    btn.innerHTML = '<i class="fa-solid fa-spinner fa-spin me-1"></i>Generating...';
  }
  try {
    const res = await generateLinkCode();
    const area = document.getElementById('telegramLinkCodeArea');
    if (area) area.style.display = 'block';
    const codeInput = document.getElementById('telegramLinkCode') as HTMLInputElement | null;
    if (codeInput) codeInput.value = formatLinkCodeDisplay(res.code);
    const linkInput = document.getElementById('telegramDeepLink') as HTMLInputElement | null;
    if (linkInput) linkInput.value = res.url || '';
    const anchor = document.getElementById('telegramDeepLinkAnchor') as HTMLAnchorElement | null;
    if (anchor) anchor.href = res.url || '#';
    toast('Code generated — open the link in Telegram');
  } catch (e: unknown) {
    renderError(e);
  } finally {
    if (btn) {
      btn.disabled = false;
      btn.innerHTML = '<i class="fa-brands fa-telegram me-1"></i>Generate link code';
    }
  }
}

export async function unlinkTelegramAccount(): Promise<void> {
  if (!confirm('Unlink your Telegram account? You will stop receiving Telegram recaps.')) return;
  try {
    await unlinkTelegram();
    toast('Telegram account unlinked');
    // Re-render the same modal content
    await renderTelegramAccountContent('telegramAccountContent');
  } catch (e: unknown) {
    renderError(e);
  }
}

export async function toggleTelegramDm(enabled: boolean): Promise<void> {
  try {
    await updateTelegramPrefs(enabled);
    toast(enabled ? 'You will receive recaps by DM' : 'DM recaps disabled');
  } catch (e: unknown) {
    renderError(e);
  }
}

function copyToClipboard(value: string, label: string) {
  if (!value) return;
  navigator.clipboard.writeText(value).then(() => toast(`${label} copied`)).catch(() => {
    // Fallback: select the input
    toast(label + ': ' + value);
  });
}

export function copyTelegramCode(): void {
  const el = document.getElementById('telegramLinkCode') as HTMLInputElement | null;
  if (el) copyToClipboard(el.value, 'Code');
}

export function copyTelegramLink(): void {
  const el = document.getElementById('telegramDeepLink') as HTMLInputElement | null;
  if (el) copyToClipboard(el.value, 'Link');
}

expose('showTelegramLink', showTelegramLink);
expose('generateTelegramLinkCode', generateTelegramLinkCode);
expose('unlinkTelegramAccount', unlinkTelegramAccount);
expose('toggleTelegramDm', toggleTelegramDm);
expose('copyTelegramCode', copyTelegramCode);
expose('copyTelegramLink', copyTelegramLink);
expose('getTelegramStatus', getTelegramStatus);
