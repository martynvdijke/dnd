import { expose } from '../lib/expose';
import { esc, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';
import { renderError } from '../lib/errors';
import {
  campaignTelegramSaveErrorText,
  campaignTelegramStatusText,
  formatAutoPostLastAt,
  formatHourLabel,
  isValidCadence,
  isValidQuietPair,
  MAX_CADENCE_MINUTES,
  MAX_QUIET_HOUR,
  parseCadenceMinutes,
  parseChatId,
  parseQuietHour,
} from '../lib/telegram';
import type { CampaignTelegramSettings } from '../lib/telegram';
import { currentCampaign, currentUser } from '../lib/state';

// Window shown when quiet hours are switched on for the first time. The DM
// sees both hours filled in rather than an empty select.
const DEFAULT_QUIET_START_HOUR = 22;
const DEFAULT_QUIET_END_HOUR = 7;

export async function getCampaignTelegram(campaignId: number): Promise<CampaignTelegramSettings> {
  return api('GET', `/api/campaigns/${campaignId}/telegram`);
}

export async function saveCampaignTelegram(
  campaignId: number,
  payload: {
    chat_id: number | null;
    is_enabled: boolean;
    auto_post_enabled: boolean;
    auto_save_media?: boolean;
    auto_post_cadence_minutes?: number;
    quiet_start_hour?: number | null;
    quiet_end_hour?: number | null;
  },
): Promise<{ ok: true }> {
  return api('PUT', `/api/campaigns/${campaignId}/telegram`, payload);
}

/** Quiet-hour selects are hidden while the toggle is off, so they stay untouched. */
export function toggleTgQuietHours(enabled: boolean): void {
  const wrap = document.getElementById('tgQuietHoursFields');
  if (!wrap) return;
  wrap.classList.toggle('d-none', !enabled);
}

function hourOptions(selected: number): string {
  let out = '';
  for (let h = 0; h <= MAX_QUIET_HOUR; h++) {
    out += `<option value="${h}"${selected === h ? ' selected' : ''}>${formatHourLabel(h)}</option>`;
  }
  return out;
}

function isCurrentUserDm(campaignId: number): boolean {
  // Admin can always manage. Otherwise check currentCampaign's role.
  if (currentUser?.role === 'admin') return true;
  const cc: any = currentCampaign;
  if (cc && cc.id === campaignId && cc.my_role === 'dm') return true;
  // If campaign mismatch, be permissive and let server enforce — show the form anyway
  // but the PUT will 403 if not DM. This avoids false negatives during navigation.
  if (!cc || cc.id !== campaignId) return true;
  return cc.my_role === 'dm';
}

export async function showCampaignTelegram(campaignId: number): Promise<void> {
  const cid = campaignId || (currentCampaign as any)?.id;
  if (!cid) {
    toast('Select a campaign first', true);
    return;
  }
  if (!isCurrentUserDm(cid)) {
    toast('Only the campaign DM can manage Telegram settings', true);
    return;
  }
  let settings: CampaignTelegramSettings | null = null;
  try {
    settings = await getCampaignTelegram(cid);
  } catch (e: unknown) {
    // Allow opening the form even if fetch failed (e.g. 404 before backend merge)
    const msg = e instanceof Error ? e.message : String(e);
    if (msg.includes('404') || msg.toLowerCase().includes('not found')) {
      settings = { chat_id: null, chat_type: '', title_cache: '', is_enabled: false, auto_post_enabled: false, auto_save_media: false, bound_at: null, auto_post_cadence_minutes: 0, quiet_start_hour: null, quiet_end_hour: null, auto_post_last_at: null };
    } else {
      renderError(e);
      return;
    }
  }
  const chatVal = settings?.chat_id != null ? String(settings.chat_id) : '';
  const enabled = !!settings?.is_enabled;
  const autoPost = !!settings?.auto_post_enabled;
  const autoSave = !!settings?.auto_save_media;
  const status = campaignTelegramStatusText(settings);
  const cadence = settings?.auto_post_cadence_minutes ?? 0;
  // Quiet hours are both-or-neither server-side, so only a stored pair counts
  // as switched on; saving with the toggle off clears whatever is there.
  const quietStart = settings?.quiet_start_hour ?? null;
  const quietEnd = settings?.quiet_end_hour ?? null;
  const quietEnabled = quietStart !== null && quietEnd !== null;
  const quietStartVal = quietStart ?? DEFAULT_QUIET_START_HOUR;
  const quietEndVal = quietEnd ?? DEFAULT_QUIET_END_HOUR;
  const lastAutoPostAt = formatAutoPostLastAt(settings?.auto_post_last_at);
  const lastAutoPostText = lastAutoPostAt ? `Last automatic post: ${lastAutoPostAt}` : 'No automatic post has been sent yet.';
  const nowUtcHour = formatHourLabel(new Date().getUTCHours());

  showModal('Telegram — Campaign', `
    <div class="mb-3">
      <label class="form-label" for="telegramCampaignChatId">Telegram chat ID</label>
      <input type="text" id="telegramCampaignChatId" class="form-control" value="${esc(chatVal)}" placeholder="-1001234567890" data-testid="telegram-campaign-chat-id">
      <div class="form-text">Add the bot to the group, then forward a message from that group to @userinfobot to get the chat ID.</div>
    </div>
    <div class="form-check form-switch mb-3">
      <input class="form-check-input" type="checkbox" id="telegramCampaignEnabled" ${enabled ? 'checked' : ''} data-testid="telegram-enable-toggle">
      <label class="form-check-label" for="telegramCampaignEnabled">Enable Telegram delivery</label>
    </div>
    <div class="form-check form-switch mb-2">
      <input class="form-check-input" type="checkbox" id="telegramCampaignAutoPost" ${autoPost ? 'checked' : ''} data-testid="telegram-auto-post-toggle">
      <label class="form-check-label" for="telegramCampaignAutoPost">Auto-post recap after session</label>
    </div>
    <div class="form-text mb-2">Auto-post fires after the recap's session end date once the grace window passes.</div>
    <div class="mb-3">
      <label class="form-label" for="tgAutopostCadence">Auto-post cadence (minutes)</label>
      <input type="number" id="tgAutopostCadence" class="form-control" min="0" max="${MAX_CADENCE_MINUTES}" step="1" value="${esc(String(cadence))}" data-testid="tg-autopost-cadence">
      <div class="form-text">0 posts as soon as there is something to post. A higher value waits at least that many minutes after the last automatic post, so only one automatic post is sent per window. Manual and command posts are not paced.</div>
      <div class="form-text">${esc(lastAutoPostText)}</div>
    </div>
    <div class="form-check form-switch mb-2">
      <input class="form-check-input" type="checkbox" id="tgQuietHoursToggle" ${quietEnabled ? 'checked' : ''} onchange="toggleTgQuietHours(this.checked)" data-testid="tg-quiet-hours-toggle">
      <label class="form-check-label" for="tgQuietHoursToggle">Quiet hours (UTC)</label>
    </div>
    <div id="tgQuietHoursFields" class="mb-3${quietEnabled ? '' : ' d-none'}">
      <div class="row g-2">
        <div class="col-6">
          <label class="form-label" for="tgQuietStart">Quiet from (UTC)</label>
          <select id="tgQuietStart" class="form-select" data-testid="tg-quiet-start">${hourOptions(quietStartVal)}</select>
        </div>
        <div class="col-6">
          <label class="form-label" for="tgQuietEnd">Quiet until (UTC)</label>
          <select id="tgQuietEnd" class="form-select" data-testid="tg-quiet-end">${hourOptions(quietEndVal)}</select>
        </div>
      </div>
      <div class="form-text">No automatic posts are sent between these UTC hours: the from hour is inside the window and the until hour is not. It is currently ${nowUtcHour} UTC. Manual and command posts are not affected.</div>
    </div>
    <div class="form-check form-switch mb-2">
      <input class="form-check-input" type="checkbox" id="telegramCampaignAutoSaveMedia" ${autoSave ? 'checked' : ''} data-testid="telegram-auto-save-media-toggle">
      <label class="form-check-label" for="telegramCampaignAutoSaveMedia">Auto-save media from Telegram</label>
    </div>
    <div class="form-text mb-3">When enabled, every photo, audio, video or document posted in the bound chat is saved to the campaign gallery without needing /save.</div>
    <div class="small text-muted mb-3" id="telegramCampaignStatus" data-testid="telegram-campaign-status">${esc(status)}</div>
    <button class="btn btn-primary w-100" onclick="saveCampaignTelegramForm(${cid})" data-testid="telegram-save-campaign"><i class="fa-solid fa-floppy-disk me-1"></i>Save Telegram settings</button>
  `);
}

export async function saveCampaignTelegramForm(campaignId: number): Promise<void> {
  const raw = (document.getElementById('telegramCampaignChatId') as HTMLInputElement | null)?.value ?? '';
  const trimmed = raw.trim();
  let chatId: number | null = null;
  if (trimmed !== '') {
    chatId = parseChatId(trimmed);
    if (chatId === null) {
      toast('Chat ID must be a number, for example -1001234567890', true);
      return;
    }
  }
  const isEnabled = !!(document.getElementById('telegramCampaignEnabled') as HTMLInputElement | null)?.checked;
  const autoPost = !!(document.getElementById('telegramCampaignAutoPost') as HTMLInputElement | null)?.checked;
  const autoSave = !!(document.getElementById('telegramCampaignAutoSaveMedia') as HTMLInputElement | null)?.checked;

  const cadence = parseCadenceMinutes((document.getElementById('tgAutopostCadence') as HTMLInputElement | null)?.value ?? '');
  if (cadence === null || !isValidCadence(cadence)) {
    toast(`Cadence must be a whole number of minutes from 0 to ${MAX_CADENCE_MINUTES}`, true);
    return;
  }

  // Both hours are sent as null when the toggle is off, which clears the window.
  const quietEnabled = !!(document.getElementById('tgQuietHoursToggle') as HTMLInputElement | null)?.checked;
  let quietStart: number | null = null;
  let quietEnd: number | null = null;
  if (quietEnabled) {
    quietStart = parseQuietHour((document.getElementById('tgQuietStart') as HTMLSelectElement | null)?.value ?? '');
    quietEnd = parseQuietHour((document.getElementById('tgQuietEnd') as HTMLSelectElement | null)?.value ?? '');
    if (!isValidQuietPair(quietStart, quietEnd)) {
      toast(`Quiet hours need a start hour and an end hour from 0 to ${MAX_QUIET_HOUR}`, true);
      return;
    }
  }

  try {
    await saveCampaignTelegram(campaignId, {
      chat_id: chatId,
      is_enabled: isEnabled,
      auto_post_enabled: autoPost,
      auto_save_media: autoSave,
      auto_post_cadence_minutes: cadence,
      quiet_start_hour: quietStart,
      quiet_end_hour: quietEnd,
    });
    toast('Telegram settings saved');
    hideModal();
  } catch (e: unknown) {
    // A 400 names a field the server refused; show the rule instead of the name.
    if ((e as { status?: number } | null)?.status === 400) {
      toast(campaignTelegramSaveErrorText(e instanceof Error ? e.message : String(e)), true);
      return;
    }
    renderError(e);
  }
}

expose('toggleTgQuietHours', toggleTgQuietHours);
expose('showCampaignTelegram', showCampaignTelegram);
expose('saveCampaignTelegramForm', saveCampaignTelegramForm);
expose('getCampaignTelegram', getCampaignTelegram);
