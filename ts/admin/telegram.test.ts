import { describe, it, expect, beforeEach, vi } from 'vitest';

const mockApi = vi.fn();
vi.mock('./state', () => ({
  api: (...args: any[]) => mockApi(...args),
}));

vi.mock('../lib/dom', () => ({
  esc: (s: string) => s || '',
  toast: vi.fn(),
}));

vi.mock('../lib/errors', () => ({
  renderError: vi.fn(),
}));

beforeEach(() => {
  mockApi.mockReset();
  document.body.innerHTML = `
    <input id="telegramToken" value="">
    <div id="telegramTokenMasked"></div>
    <select id="telegramMode"><option value="off">Off</option><option value="polling">Polling</option><option value="webhook">Webhook</option><option value="auto">Auto</option></select>
    <input id="telegramGraceMinutes" value="">
    <input id="telegramWebhookUrl" value="">
    <div id="telegramStatus"></div>
    <div id="telegramWebhookField" style="display:none"></div>
    <button id="telegramTestBtn"></button>
  `;
});

describe('admin telegram fetch wrappers', () => {
  it('getAdminTelegramSettings calls GET', async () => {
    mockApi.mockResolvedValue({ has_token: false, token_masked: '', mode: 'off', webhook_url: '', grace_minutes: 30, status: { mode: 'off' } });
    const m = await import('./telegram');
    const res = await m.getAdminTelegramSettings();
    expect(mockApi).toHaveBeenCalledWith('GET', '/api/admin/telegram-settings');
    expect(res.mode).toBe('off');
  });

  it('saveAdminTelegramSettings calls POST', async () => {
    mockApi.mockResolvedValue({ has_token: true, token_masked: '***123', mode: 'polling', webhook_url: '', grace_minutes: 45, status: { mode: 'polling' } });
    const m = await import('./telegram');
    const res = await m.saveAdminTelegramSettings({ mode: 'polling', grace_minutes: 45 });
    expect(mockApi).toHaveBeenCalledWith('POST', '/api/admin/telegram-settings', { mode: 'polling', grace_minutes: 45 });
    expect(res.grace_minutes).toBe(45);
  });

  it('testAdminTelegram calls POST test endpoint', async () => {
    mockApi.mockResolvedValue({ sent: true });
    const m = await import('./telegram');
    const res = await m.testAdminTelegram();
    expect(mockApi).toHaveBeenCalledWith('POST', '/api/admin/telegram-test');
    expect(res.sent).toBe(true);
  });

  it('loadTelegramSettings populates DOM', async () => {
    mockApi.mockResolvedValue({ has_token: true, token_masked: 'bot****', mode: 'polling', webhook_url: '', grace_minutes: 30, status: { mode: 'polling' } });
    const m = await import('./telegram');
    // call exposed function via window
    await (window as any).loadTelegramSettings();
    expect((document.getElementById('telegramMode') as HTMLSelectElement).value).toBe('polling');
    expect((document.getElementById('telegramGraceMinutes') as HTMLInputElement).value).toBe('30');
    expect(document.getElementById('telegramStatus')?.textContent).toContain('polling');
  });

  it('saveTelegramSettings builds payload from DOM', async () => {
    mockApi.mockResolvedValue({ has_token: true, token_masked: '***', mode: 'auto', webhook_url: '', grace_minutes: 60, status: { mode: 'auto' } });
    (document.getElementById('telegramToken') as HTMLInputElement).value = ' 123:ABC ';
    (document.getElementById('telegramMode') as HTMLSelectElement).value = 'auto';
    (document.getElementById('telegramGraceMinutes') as HTMLInputElement).value = '60';
    const m = await import('./telegram');
    await (window as any).saveTelegramSettings();
    expect(mockApi).toHaveBeenCalledWith('POST', '/api/admin/telegram-settings', expect.objectContaining({ token: '123:ABC', mode: 'auto', grace_minutes: 60 }));
  });
});
