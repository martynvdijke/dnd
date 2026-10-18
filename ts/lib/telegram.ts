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
  auto_save_media?: boolean;
  bound_at: string | null;
  /** Automatic post pacing, whole minutes. 0 means no pacing. */
  auto_post_cadence_minutes?: number | null;
  /** First UTC hour of the quiet window, or null when quiet hours are off. */
  quiet_start_hour?: number | null;
  /** End of the quiet window, exclusive, or null when quiet hours are off. */
  quiet_end_hour?: number | null;
  /** UTC time of the last automatic post, or null when none has been sent. */
  auto_post_last_at?: string | null;
}

/**
 * Auto-post scheduling bounds. The server rejects anything outside these, and
 * it compares UTC hours, so nothing in this module converts to local time.
 */
export const MAX_CADENCE_MINUTES = 1440;
export const MAX_QUIET_HOUR = 23;

/** Parse a whole number from a form value. Empty, fractional and junk are null. */
function parseWholeNumber(raw: unknown): number | null {
  const s = String(raw ?? '').trim();
  if (s === '') return null;
  const n = Number(s);
  // Reject "60.5" rather than rounding: the server takes whole minutes, and a
  // silently rounded cadence would pace differently from what was typed.
  if (!Number.isFinite(n) || !Number.isInteger(n)) return null;
  return n;
}

/** Cadence form value in whole minutes, or null when it is not a whole number. */
export function parseCadenceMinutes(raw: unknown): number | null {
  return parseWholeNumber(raw);
}

export function isValidCadence(n: number): boolean {
  return Number.isInteger(n) && n >= 0 && n <= MAX_CADENCE_MINUTES;
}

/** Quiet-hour select value in whole hours, or null when it is not a whole number. */
export function parseQuietHour(raw: unknown): number | null {
  return parseWholeNumber(raw);
}

export function isValidQuietHour(n: number | null | undefined): boolean {
  if (n === null || n === undefined) return true;
  return Number.isInteger(n) && n >= 0 && n <= MAX_QUIET_HOUR;
}

/**
 * Quiet hours are both-or-neither: a single set hour is not a window, and the
 * server rejects a half-set pair, so the form must never send one.
 */
export function isValidQuietPair(start: number | null | undefined, end: number | null | undefined): boolean {
  const startSet = start !== null && start !== undefined;
  const endSet = end !== null && end !== undefined;
  if (startSet !== endSet) return false;
  return isValidQuietHour(start) && isValidQuietHour(end);
}

/** Label for an hour select option, e.g. 7 -> "07:00". */
export function formatHourLabel(hour: number): string {
  return `${String(hour).padStart(2, '0')}:00`;
}

/**
 * Readable form of the stored last automatic post time, or '' when there is
 * none. The server stores UTC; a value without a zone marker is read as UTC
 * rather than as browser-local, because showing local time here would
 * contradict the (UTC) labels on the quiet hours.
 */
export function formatAutoPostLastAt(raw: string | null | undefined): string {
  const s = String(raw ?? '').trim();
  if (s === '') return '';
  const withZone = /[zZ]|[+-]\d{2}:?\d{2}$/.test(s) ? s : `${s}Z`;
  const d = new Date(withZone);
  if (Number.isNaN(d.getTime())) return '';
  const p = (n: number) => String(n).padStart(2, '0');
  return `${d.getUTCFullYear()}-${p(d.getUTCMonth() + 1)}-${p(d.getUTCDate())} ${p(d.getUTCHours())}:${p(d.getUTCMinutes())} UTC`;
}

/**
 * A 400 from the settings endpoint names the field it rejected; that name alone
 * is not something a DM can act on, so map it to the rule that applies.
 */
export function campaignTelegramSaveErrorText(serverMessage: string): string {
  const m = String(serverMessage ?? '').toLowerCase();
  if (m.includes('cadence')) {
    return `Cadence must be a whole number of minutes from 0 to ${MAX_CADENCE_MINUTES}`;
  }
  if (m.includes('quiet')) {
    return `Quiet hours must be set as a start hour and an end hour from 0 to ${MAX_QUIET_HOUR}, or both turned off`;
  }
  const raw = String(serverMessage ?? '').trim();
  return raw ? `Telegram settings were not saved: ${raw}` : 'Telegram settings were not saved';
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
  const save = (s as any).auto_save_media ? 'auto-save on' : 'auto-save off';
  const title = s.title_cache ? ` — ${s.title_cache}` : '';
  return `${enabled}, ${auto}, ${save}${title}`;
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
