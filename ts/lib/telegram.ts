/**
 * Telegram helpers — pure, testable utilities for the Telegram integration.
 *
 * No DOM or fetch side-effects here; see ts/admin/telegram.ts and
 * ts/campaign/telegram.ts for the API wrappers that use these.
 */

export type TelegramMode = 'off' | 'polling' | 'webhook' | 'auto';

export interface TelegramStatus {
  linked: boolean;
  telegram_username: string;
  dm_enabled: boolean;
}

export interface TelegramLinkCode {
  code: string;
  url: string;
  expires_at: string;
}

export interface AdminTelegramSettings {
  has_token: boolean;
  token_masked: string;
  mode: TelegramMode;
  webhook_url: string;
  grace_minutes: number;
  status: {
    mode: string;
    webhook_url?: string;
    last_error?: string;
    [k: string]: unknown;
  };
}

export interface CampaignTelegramSettings {
  chat_id: number | null;
  chat_type: string;
  title_cache: string;
  is_enabled: boolean;
  auto_post_enabled: boolean;
  bound_at: string | null;
}

export const TELEGRAM_MODES: TelegramMode[] = ['off', 'polling', 'webhook', 'auto'];

export function isValidTelegramMode(v: string): v is TelegramMode {
  return (TELEGRAM_MODES as string[]).includes(v);
}

export function formatTelegramMode(mode: string): string {
  if (!mode) return 'off';
  return mode;
}

export function formatGraceMinutes(n: unknown): number {
  const v = Number(n);
  if (!Number.isFinite(v) || v < 1) return 30;
  return Math.floor(v);
}

/**
 * Build a t.me deep link for a start code. If botUsername is not known,
 * callers may pass an empty string and the URL will be `https://t.me/<bot>?start=CODE`
 * — backend already returns the full url, this helper is for display fallback.
 */
export function buildDeepLink(code: string, botUsername?: string): string {
  if (!code) return '';
  if (botUsername) return `https://t.me/${botUsername}?start=${encodeURIComponent(code)}`;
  return `https://t.me/villumbot?start=${encodeURIComponent(code)}`;
}

export function formatLinkCodeDisplay(code: string): string {
  if (!code) return '';
  // Codes are short (8 chars) — display as-is, trimmed.
  return code.trim();
}

export function parseChatId(raw: string): number | null {
  const s = String(raw || '').trim();
  if (!s) return null;
  const n = Number(s);
  if (!Number.isFinite(n)) return null;
  // Telegram chat ids are 64-bit signed; JS can hold them exactly up to 2^53.
  // Just return the numeric value; backend validates range.
  return Math.trunc(n);
}

export function isValidChatId(raw: string): boolean {
  if (raw.trim() === '') return false;
  return parseChatId(raw) !== null;
}

export function campaignTelegramStatusText(s: CampaignTelegramSettings | null): string {
  if (!s) return 'Not configured';
  if (!s.chat_id) return 'No chat linked';
  const enabled = s.is_enabled ? 'Enabled' : 'Disabled';
  const auto = s.auto_post_enabled ? 'auto-post on' : 'auto-post off';
  const title = s.title_cache ? ` — ${s.title_cache}` : '';
  return `${enabled}, ${auto}${title}`;
}

export function adminTelegramStatusText(s: AdminTelegramSettings | null): string {
  if (!s) return 'Not configured';
  const mode = s.mode || 'off';
  if (mode === 'off') return 'Telegram is off';
  const url = s.webhook_url || (s.status && (s.status.webhook_url as string)) || '';
  if ((mode === 'webhook' || mode === 'auto') && url) return `Mode: ${mode} — ${url}`;
  return `Mode: ${mode}`;
}

export function maskedTokenDisplay(masked: string, hasToken: boolean): string {
  if (!hasToken) return 'No token configured';
  if (masked) return `Token: ${masked}`;
  return 'Token is set';
}

export function telegramPrefsPayload(dmEnabled: boolean): { dm_enabled: boolean } {
  return { dm_enabled: !!dmEnabled };
}
