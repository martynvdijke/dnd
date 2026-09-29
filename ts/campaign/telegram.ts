import { expose } from '../lib/expose';
import { esc, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';
import { renderError } from '../lib/errors';
import { campaignTelegramStatusText, parseChatId } from '../lib/telegram';
import type { CampaignTelegramSettings } from '../lib/telegram';
import { currentCampaign, currentUser } from '../lib/state';

export async function getCampaignTelegram(campaignId: number): Promise<CampaignTelegramSettings> {
  return api('GET', `/api/campaigns/${campaignId}/telegram`);
}

export async function saveCampaignTelegram(
  campaignId: number,
  payload: { chat_id: number | null; is_enabled: boolean; auto_post_enabled: boolean },
): Promise<{ ok: true }> {
  return api('PUT', `/api/campaigns/${campaignId}/telegram`, payload);
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
      settings = { chat_id: null, chat_type: '', title_cache: '', is_enabled: false, auto_post_enabled: false, bound_at: null };
    } else {
      renderError(e);
      return;
    }
  }
  const chatVal = settings?.chat_id != null ? String(settings.chat_id) : '';
  const enabled = !!settings?.is_enabled;
  const autoPost = !!settings?.auto_post_enabled;
  const status = campaignTelegramStatusText(settings);

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
    <div class="form-text mb-3">Auto-post fires after the recap's session end date once the grace window passes.</div>
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
  try {
    await saveCampaignTelegram(campaignId, { chat_id: chatId, is_enabled: isEnabled, auto_post_enabled: autoPost });
    toast('Telegram settings saved');
    hideModal();
  } catch (e: unknown) {
    renderError(e);
  }
}

expose('showCampaignTelegram', showCampaignTelegram);
expose('saveCampaignTelegramForm', saveCampaignTelegramForm);
expose('getCampaignTelegram', getCampaignTelegram);
