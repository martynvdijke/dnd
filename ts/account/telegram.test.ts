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

import * as telegramMod from './telegram';

function stubClipboard() {
  const writeMock = vi.fn().mockResolvedValue(undefined);
  Object.defineProperty(globalThis.navigator, 'clipboard', {
    value: { writeText: writeMock },
    configurable: true,
    writable: true,
  });
  return writeMock;
}

beforeEach(() => {
  mockApi.mockReset();
  stubClipboard();
  document.body.innerHTML = `
    <div id="genericModal"><div id="genericModalTitle"></div><div id="genericModalBody"></div></div>
    <div id="toastContainer"></div>
  `;
});

describe('account telegram fetch wrappers', () => {
  it('getTelegramStatus calls GET', async () => {
    mockApi.mockResolvedValue({ linked: true, telegram_username: 'testuser', dm_enabled: true });
    const res = await telegramMod.getTelegramStatus();
    expect(mockApi).toHaveBeenCalledWith('GET', '/api/telegram/status');
    expect(res.linked).toBe(true);
  });

  it('generateLinkCode calls POST', async () => {
    mockApi.mockResolvedValue({ code: 'ABCD1234', url: 'https://t.me/bot?start=ABCD1234', expires_at: '2026-01-01T00:15:00Z' });
    const res = await telegramMod.generateLinkCode();
    expect(mockApi).toHaveBeenCalledWith('POST', '/api/telegram/link-code');
    expect(res.code).toBe('ABCD1234');
  });

  it('unlinkTelegram calls DELETE', async () => {
    mockApi.mockResolvedValue({ ok: true });
    await telegramMod.unlinkTelegram();
    expect(mockApi).toHaveBeenCalledWith('DELETE', '/api/telegram/unlink');
  });

  it('updateTelegramPrefs calls PUT with dm_enabled', async () => {
    mockApi.mockResolvedValue({ ok: true });
    await telegramMod.updateTelegramPrefs(true);
    expect(mockApi).toHaveBeenCalledWith('PUT', '/api/telegram/prefs', { dm_enabled: true });
  });

  it('showTelegramLink renders linked state', async () => {
    mockApi.mockResolvedValue({ linked: true, telegram_username: 'alice', dm_enabled: false });
    document.body.innerHTML = `
      <div id="genericModal"><div id="genericModalTitle"></div><div id="genericModalBody"><div id="telegramAccountContent"></div></div></div>
    `;
    // showTelegramLink creates modal then fills telegramAccountContent, so we need to set up showModal mock to preserve container
    // Our mock replaces body innerHTML, so after call, body will contain rendered content
    await telegramMod.showTelegramLink();
    // After showTelegramLink, the genericModalBody should contain linked badge
    const body = document.getElementById('genericModalBody')?.innerHTML || document.body.innerHTML;
    expect(body).toContain('data-testid="telegram-dm-toggle"');
    expect(body).toContain('data-testid="telegram-unlink"');
  });

  it('showTelegramLink renders unlinked state with generate button', async () => {
    mockApi.mockResolvedValue({ linked: false, telegram_username: '', dm_enabled: false });
    document.body.innerHTML = `
      <div id="genericModal"><div id="genericModalTitle"></div><div id="genericModalBody"><div id="telegramAccountContent"></div></div></div>
    `;
    await telegramMod.showTelegramLink();
    const body = document.getElementById('genericModalBody')?.innerHTML || document.body.innerHTML;
    expect(body).toContain('data-testid="telegram-generate-code"');
    expect(body).toContain('data-testid="telegram-dm-toggle"');
  });

  it('generateTelegramLinkCode fills code and link inputs', async () => {
    mockApi.mockResolvedValue({ code: 'CODE1234', url: 'https://t.me/mybot?start=CODE1234', expires_at: '2026-01-01T00:15:00Z' });
    document.body.innerHTML = `
      <div id="telegramLinkCodeArea" style="display:none">
        <input id="telegramLinkCode" value="">
        <input id="telegramDeepLink" value="">
        <a id="telegramDeepLinkAnchor" href="#"></a>
      </div>
      <button data-testid="telegram-generate-code"></button>
    `;
    await telegramMod.generateTelegramLinkCode();
    expect((document.getElementById('telegramLinkCode') as HTMLInputElement).value).toBe('CODE1234');
    expect((document.getElementById('telegramDeepLink') as HTMLInputElement).value).toBe('https://t.me/mybot?start=CODE1234');
    expect((document.getElementById('telegramDeepLinkAnchor') as HTMLAnchorElement).href).toContain('CODE1234');
    expect(document.getElementById('telegramLinkCodeArea')?.style.display).toBe('block');
  });

  it('copy helpers use clipboard', async () => {
    const writeMock = stubClipboard();
    document.body.innerHTML = `
      <input id="telegramLinkCode" value="MYCODE">
      <input id="telegramDeepLink" value="https://t.me/bot?start=MYCODE">
    `;
    telegramMod.copyTelegramCode();
    expect(writeMock).toHaveBeenCalledWith('MYCODE');
    telegramMod.copyTelegramLink();
    expect(writeMock).toHaveBeenCalledWith('https://t.me/bot?start=MYCODE');
  });
});
