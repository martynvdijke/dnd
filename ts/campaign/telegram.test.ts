import { describe, it, expect, beforeEach, vi } from 'vitest';

const mockApi = vi.fn();
vi.mock('../lib/api', () => ({
  api: (...args: any[]) => mockApi(...args),
}));

vi.mock('../lib/dom', () => ({
  esc: (s: string | null | undefined) => s || '',
  showModal: vi.fn((title: string, body: string) => {
    document.getElementById('genericModalTitle')!.textContent = title;
    document.getElementById('genericModalBody')!.innerHTML = body;
  }),
  hideModal: vi.fn(),
  toast: vi.fn(),
}));

vi.mock('../lib/errors', () => ({
  renderError: vi.fn(),
}));

vi.mock('../lib/state', () => ({
  currentCampaign: { id: 42, my_role: 'dm' },
  currentUser: { id: 1, role: 'dm' },
}));

beforeEach(() => {
  mockApi.mockReset();
  document.body.innerHTML = `
    <div id="genericModal"><div id="genericModalTitle"></div><div id="genericModalBody"></div></div>
    <div id="toastContainer"></div>
  `;
});

describe('campaign telegram fetch wrappers', () => {
  it('getCampaignTelegram calls GET with campaign id', async () => {
    mockApi.mockResolvedValue({ chat_id: -100123, chat_type: 'group', title_cache: '', is_enabled: true, auto_post_enabled: false, bound_at: null });
    const m = await import('./telegram');
    const res = await m.getCampaignTelegram(42);
    expect(mockApi).toHaveBeenCalledWith('GET', '/api/campaigns/42/telegram');
    expect(res.chat_id).toBe(-100123);
  });

  it('saveCampaignTelegram calls PUT', async () => {
    mockApi.mockResolvedValue({ ok: true });
    const m = await import('./telegram');
    await m.saveCampaignTelegram(42, { chat_id: -100123, is_enabled: true, auto_post_enabled: true });
    expect(mockApi).toHaveBeenCalledWith('PUT', '/api/campaigns/42/telegram', { chat_id: -100123, is_enabled: true, auto_post_enabled: true });
  });

  it('showCampaignTelegram renders modal with inputs and data-testids', async () => {
    mockApi.mockResolvedValue({ chat_id: -100123, chat_type: 'group', title_cache: 'My Group', is_enabled: true, auto_post_enabled: false, bound_at: null });
    const m = await import('./telegram');
    await m.showCampaignTelegram(42);
    const body = document.getElementById('genericModalBody')!.innerHTML;
    expect(body).toContain('data-testid="telegram-campaign-chat-id"');
    expect(body).toContain('data-testid="telegram-enable-toggle"');
    expect(body).toContain('data-testid="telegram-auto-post-toggle"');
    expect(body).toContain('data-testid="telegram-save-campaign"');
    expect(body).toContain('-100123');
  });

  it('saveCampaignTelegramForm validates chat_id and calls API', async () => {
    document.body.innerHTML += `
      <input id="telegramCampaignChatId" value="-100999">
      <input type="checkbox" id="telegramCampaignEnabled" checked>
      <input type="checkbox" id="telegramCampaignAutoPost">
    `;
    mockApi.mockResolvedValue({ ok: true });
    const m = await import('./telegram');
    await m.saveCampaignTelegramForm(42);
    expect(mockApi).toHaveBeenCalledWith('PUT', '/api/campaigns/42/telegram', expect.objectContaining({ chat_id: -100999, is_enabled: true }));
  });

  it('saveCampaignTelegramForm allows empty chat_id as null', async () => {
    document.body.innerHTML += `
      <input id="telegramCampaignChatId" value="">
      <input type="checkbox" id="telegramCampaignEnabled">
      <input type="checkbox" id="telegramCampaignAutoPost">
    `;
    mockApi.mockResolvedValue({ ok: true });
    const m = await import('./telegram');
    // Need to reset body to have only these inputs visible after previous?
    document.getElementById('telegramCampaignChatId')!.setAttribute('value', '');
    (document.getElementById('telegramCampaignChatId') as HTMLInputElement).value = '';
    await m.saveCampaignTelegramForm(99);
    expect(mockApi).toHaveBeenCalledWith('PUT', '/api/campaigns/99/telegram', expect.objectContaining({ chat_id: null }));
  });
});
