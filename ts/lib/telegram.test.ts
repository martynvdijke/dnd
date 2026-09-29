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
