import { describe, it, expect } from 'vitest';
import {
  isValidTelegramMode,
  formatGraceMinutes,
  buildDeepLink,
  formatLinkCodeDisplay,
  parseChatId,
  isValidChatId,
  campaignTelegramStatusText,
  adminTelegramStatusText,
  maskedTokenDisplay,
  telegramPrefsPayload,
  campaignTelegramSaveErrorText,
  formatAutoPostLastAt,
  formatHourLabel,
  isValidCadence,
  isValidQuietHour,
  isValidQuietPair,
  parseCadenceMinutes,
  parseQuietHour,
  MAX_CADENCE_MINUTES,
  MAX_QUIET_HOUR,
} from './telegram';

describe('isValidTelegramMode', () => {
  it('accepts known modes', () => {
    expect(isValidTelegramMode('off')).toBe(true);
    expect(isValidTelegramMode('polling')).toBe(true);
    expect(isValidTelegramMode('webhook')).toBe(true);
    expect(isValidTelegramMode('auto')).toBe(true);
  });
  it('rejects unknown', () => {
    expect(isValidTelegramMode('bogus')).toBe(false);
    expect(isValidTelegramMode('')).toBe(false);
  });
});

describe('formatGraceMinutes', () => {
  it('defaults to 30 on invalid', () => {
    expect(formatGraceMinutes(undefined)).toBe(30);
    expect(formatGraceMinutes(null)).toBe(30);
    expect(formatGraceMinutes(NaN)).toBe(30);
    expect(formatGraceMinutes(0)).toBe(30);
    expect(formatGraceMinutes(-5)).toBe(30);
  });
  it('floors valid values', () => {
    expect(formatGraceMinutes(30.9)).toBe(30);
    expect(formatGraceMinutes('45')).toBe(45);
  });
});

describe('buildDeepLink', () => {
  it('builds link with code', () => {
    expect(buildDeepLink('ABC123', 'mybot')).toBe('https://t.me/mybot?start=ABC123');
  });
  it('returns empty on empty code', () => {
    expect(buildDeepLink('')).toBe('');
  });
});

describe('formatLinkCodeDisplay', () => {
  it('trims whitespace', () => {
    expect(formatLinkCodeDisplay('  ABC ')).toBe('ABC');
  });
  it('returns empty for empty', () => {
    expect(formatLinkCodeDisplay('')).toBe('');
  });
});

describe('parseChatId', () => {
  it('parses numeric ids', () => {
    expect(parseChatId('-100123')).toBe(-100123);
    expect(parseChatId('123')).toBe(123);
    expect(parseChatId('  -100123  ')).toBe(-100123);
  });
  it('returns null for blank or non-numeric', () => {
    expect(parseChatId('')).toBeNull();
    expect(parseChatId('   ')).toBeNull();
    expect(parseChatId('abc')).toBeNull();
  });
  it('truncates decimals', () => {
    expect(parseChatId('123.9')).toBe(123);
  });
});

describe('isValidChatId', () => {
  it('validates', () => {
    expect(isValidChatId('123')).toBe(true);
    expect(isValidChatId('')).toBe(false);
    expect(isValidChatId('abc')).toBe(false);
  });
});

describe('campaignTelegramStatusText', () => {
  it('handles null', () => {
    expect(campaignTelegramStatusText(null)).toBe('Not configured');
  });
  it('handles no chat', () => {
    expect(campaignTelegramStatusText({ chat_id: null, chat_type: '', title_cache: '', is_enabled: false, auto_post_enabled: false, bound_at: null })).toBe('No chat linked');
  });
  it('formats enabled with title', () => {
    const s = { chat_id: -100123, chat_type: 'group', title_cache: 'My Group', is_enabled: true, auto_post_enabled: true, bound_at: null };
    expect(campaignTelegramStatusText(s)).toContain('Enabled');
    expect(campaignTelegramStatusText(s)).toContain('auto-post on');
    expect(campaignTelegramStatusText(s)).toContain('My Group');
  });
  it('formats disabled', () => {
    const s = { chat_id: -100123, chat_type: 'group', title_cache: '', is_enabled: false, auto_post_enabled: false, bound_at: null };
    expect(campaignTelegramStatusText(s)).toContain('Disabled');
  });
  it('keeps reporting auto-save media', () => {
    const base = { chat_id: -100123, chat_type: 'group', title_cache: '', is_enabled: true, auto_post_enabled: true, bound_at: null };
    expect(campaignTelegramStatusText({ ...base, auto_save_media: true })).toContain('auto-save on');
    expect(campaignTelegramStatusText({ ...base, auto_save_media: false })).toContain('auto-save off');
  });
});

describe('adminTelegramStatusText', () => {
  it('handles null', () => {
    expect(adminTelegramStatusText(null)).toBe('Not configured');
  });
  it('off mode', () => {
    expect(adminTelegramStatusText({ has_token: false, token_masked: '', mode: 'off', webhook_url: '', grace_minutes: 30, status: { mode: 'off' } })).toBe('Telegram is off');
  });
  it('webhook with url', () => {
    const s = { has_token: true, token_masked: '****', mode: 'webhook' as const, webhook_url: 'https://example.com/hook', grace_minutes: 30, status: { mode: 'webhook', webhook_url: 'https://example.com/hook' } };
    expect(adminTelegramStatusText(s)).toContain('https://example.com/hook');
  });
  it('polling mode', () => {
    const s = { has_token: true, token_masked: '****', mode: 'polling' as const, webhook_url: '', grace_minutes: 30, status: { mode: 'polling' } };
    expect(adminTelegramStatusText(s)).toBe('Mode: polling');
  });
});

describe('maskedTokenDisplay', () => {
  it('no token', () => {
    expect(maskedTokenDisplay('', false)).toBe('No token configured');
  });
  it('with masked', () => {
    expect(maskedTokenDisplay('bot****', true)).toContain('bot****');
  });
});

describe('telegramPrefsPayload', () => {
  it('maps boolean', () => {
    expect(telegramPrefsPayload(true)).toEqual({ dm_enabled: true });
    expect(telegramPrefsPayload(false)).toEqual({ dm_enabled: false });
  });
});

describe('parseCadenceMinutes', () => {
  it('parses whole minutes', () => {
    expect(parseCadenceMinutes('60')).toBe(60);
    expect(parseCadenceMinutes(' 0 ')).toBe(0);
    expect(parseCadenceMinutes('1440')).toBe(1440);
    expect(parseCadenceMinutes(30)).toBe(30);
  });
  it('returns null for blank, junk and fractional values', () => {
    expect(parseCadenceMinutes('')).toBeNull();
    expect(parseCadenceMinutes('   ')).toBeNull();
    expect(parseCadenceMinutes(null)).toBeNull();
    expect(parseCadenceMinutes(undefined)).toBeNull();
    expect(parseCadenceMinutes('abc')).toBeNull();
    // Rounding 60.5 to 60 would pace differently from what was typed.
    expect(parseCadenceMinutes('60.5')).toBeNull();
  });
});

describe('isValidCadence', () => {
  it('accepts the whole 0 to 1440 range', () => {
    expect(isValidCadence(0)).toBe(true);
    expect(isValidCadence(1)).toBe(true);
    expect(isValidCadence(MAX_CADENCE_MINUTES)).toBe(true);
  });
  it('rejects out of range and non-integer values', () => {
    expect(isValidCadence(-1)).toBe(false);
    expect(isValidCadence(1441)).toBe(false);
    expect(isValidCadence(60.5)).toBe(false);
    expect(isValidCadence(NaN)).toBe(false);
  });
});

describe('parseQuietHour', () => {
  it('parses whole hours and leaves range checking to isValidQuietHour', () => {
    expect(parseQuietHour('0')).toBe(0);
    expect(parseQuietHour('7')).toBe(7);
    expect(parseQuietHour(23)).toBe(23);
    expect(parseQuietHour('24')).toBe(24);
  });
  it('returns null for blank and non-hour values', () => {
    expect(parseQuietHour('')).toBeNull();
    expect(parseQuietHour(null)).toBeNull();
    expect(parseQuietHour('22:00')).toBeNull();
    expect(parseQuietHour('evening')).toBeNull();
  });
});

describe('isValidQuietHour', () => {
  it('accepts 0 to 23 and treats unset as valid', () => {
    expect(isValidQuietHour(0)).toBe(true);
    expect(isValidQuietHour(7)).toBe(true);
    expect(isValidQuietHour(MAX_QUIET_HOUR)).toBe(true);
    expect(isValidQuietHour(null)).toBe(true);
    expect(isValidQuietHour(undefined)).toBe(true);
  });
  it('rejects out of range and fractional hours', () => {
    expect(isValidQuietHour(-1)).toBe(false);
    expect(isValidQuietHour(24)).toBe(false);
    expect(isValidQuietHour(7.5)).toBe(false);
  });
});

describe('isValidQuietPair', () => {
  it('accepts both set or both cleared', () => {
    expect(isValidQuietPair(null, null)).toBe(true);
    expect(isValidQuietPair(undefined, undefined)).toBe(true);
    expect(isValidQuietPair(22, 7)).toBe(true);
    expect(isValidQuietPair(0, 0)).toBe(true);
  });
  it('rejects a half-set pair', () => {
    expect(isValidQuietPair(22, null)).toBe(false);
    expect(isValidQuietPair(null, 7)).toBe(false);
    expect(isValidQuietPair(22, undefined)).toBe(false);
  });
  it('rejects an out of range hour on either side', () => {
    expect(isValidQuietPair(24, 7)).toBe(false);
    expect(isValidQuietPair(22, 24)).toBe(false);
  });
});

describe('formatHourLabel', () => {
  it('pads to a two digit hour with :00', () => {
    expect(formatHourLabel(0)).toBe('00:00');
    expect(formatHourLabel(7)).toBe('07:00');
    expect(formatHourLabel(22)).toBe('22:00');
    expect(formatHourLabel(23)).toBe('23:00');
  });
});

describe('formatAutoPostLastAt', () => {
  it('renders a stored UTC time', () => {
    expect(formatAutoPostLastAt('2026-10-17T21:04:00Z')).toBe('2026-10-17 21:04 UTC');
  });
  it('converts an offset timestamp to UTC', () => {
    expect(formatAutoPostLastAt('2026-10-17T23:04:00+02:00')).toBe('2026-10-17 21:04 UTC');
  });
  it('reads an unzoned value as UTC rather than local time', () => {
    expect(formatAutoPostLastAt('2026-10-17 21:04:00')).toBe('2026-10-17 21:04 UTC');
  });
  it('returns empty when there is no usable time', () => {
    expect(formatAutoPostLastAt(null)).toBe('');
    expect(formatAutoPostLastAt('')).toBe('');
    expect(formatAutoPostLastAt('   ')).toBe('');
    expect(formatAutoPostLastAt('not a date')).toBe('');
  });
});

describe('campaignTelegramSaveErrorText', () => {
  it('explains a rejected cadence instead of naming the field', () => {
    const msg = campaignTelegramSaveErrorText('auto_post_cadence_minutes must be between 0 and 1440');
    expect(msg).toContain(`0 to ${MAX_CADENCE_MINUTES}`);
    expect(msg).not.toContain('auto_post_cadence_minutes');
  });
  it('explains a rejected quiet-hour pair', () => {
    const msg = campaignTelegramSaveErrorText('quiet_start_hour and quiet_end_hour must both be set');
    expect(msg).toContain(`0 to ${MAX_QUIET_HOUR}`);
    expect(msg).toContain('both turned off');
  });
  it('keeps the server text for anything else', () => {
    expect(campaignTelegramSaveErrorText('unauthorized')).toBe('Telegram settings were not saved: unauthorized');
  });
  it('has a fallback when the server sends nothing', () => {
    expect(campaignTelegramSaveErrorText('')).toBe('Telegram settings were not saved');
  });
});
