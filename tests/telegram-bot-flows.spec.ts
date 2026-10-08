import { test, expect } from './fixtures.js';
import { login } from './helpers.js';

async function getApiToken(page: any): Promise<string> {
  const csrfResp = await page.request.get('/api/csrf-token');
  const j = await csrfResp.json();
  const csrf = j.token;
  const createResp = await page.request.post('/api/tokens', {
    data: { name: 'e2e-bot-flows' },
    headers: { 'X-CSRF-Token': csrf },
  });
  if (createResp.status() !== 201) throw new Error('token create failed ' + createResp.status());
  const d = await createResp.json();
  return d.token;
}

async function apiWithToken(page: any, method: string, path: string, body?: any) {
  const token = await getApiToken(page);
  const csrfResp = await page.request.get('/api/csrf-token');
  const csrf = (await csrfResp.json()).token;
  const headers: Record<string, string> = { 'X-CSRF-Token': csrf, Authorization: `Bearer ${token}` };
  const opts: any = { headers };
  if (body !== undefined) opts.data = body;
  if (method === 'GET') return page.request.get(path, opts);
  if (method === 'POST') return page.request.post(path, opts);
  if (method === 'PUT') return page.request.put(path, opts);
  if (method === 'DELETE') return page.request.delete(path, opts);
  throw new Error('unknown method');
}

test.describe('Telegram bot flows', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('login: email sign-in + magic link + status linked', async ({ page, workerData }) => {
    // slow: supervised client must start for /login + email round-trip
    test.slow();
    const secret = 'login-flow-secret-' + Date.now();
    const saveBot = await apiWithToken(page, 'POST', '/api/admin/telegram-settings', {
      token: workerData.telegramMock.botToken,
      mode: 'polling',
      webhook_secret: secret,
    });
    expect(saveBot.ok()).toBeTruthy();
    await expect.poll(() => workerData.telegramMock.commands.length, { timeout: 20000 }).toBeGreaterThan(0);

    const emailSave = await apiWithToken(page, 'POST', '/api/admin/email-settings', {
      smtp_host: '127.0.0.1',
      smtp_port: (workerData as any).smtpMock.port,
      username: '',
      password: '',
      from_addr: 'bot@villum.test',
      enabled: true,
    });
    expect(emailSave.ok()).toBeTruthy();

    const tgUser: any = { id: 700100001, is_bot: false, first_name: 'Login', username: 'e2e_login' };
    const chat: any = { id: 700100001, type: 'private' };
    let updateID = 710000;

    const send = async (text: string): Promise<string> => {
      workerData.telegramMock.clear();
      const resp = await page.request.post('/api/telegram/webhook', {
        data: { update_id: ++updateID, message: { message_id: 1, from: tgUser, chat, text } },
        headers: { 'X-Telegram-Bot-Api-Secret-Token': secret },
      });
      expect(resp.status()).toBe(200);
      await expect.poll(() => workerData.telegramMock.sentMessages.length, { timeout: 15000 }).toBeGreaterThan(0);
      return workerData.telegramMock.sentMessages.map((m) => m.text).join('\n---\n');
    };

    // Clear SMTP mock
    (workerData as any).smtpMock.clear();

    const loginPrompt = await send('/login');
    expect(loginPrompt).toMatch(/Email sign-in/i);

    const inboxMsg = await send('e2e-login@example.test');
    expect(inboxMsg).toMatch(/Check your inbox/i);

    await expect.poll(() => (workerData as any).smtpMock.messages.length, { timeout: 15000 }).toBeGreaterThan(0);
    const rawEmail: string = (workerData as any).smtpMock.messages[0].raw ?? (workerData as any).smtpMock.messages[0].data;
    // Extract /telegram/auth?token=... URL — email body contains full link with BASE_URL
    const linkMatch = rawEmail.match(/https?:\/\/[^\s]+?\/telegram\/auth\?token=[^\s"']+/);
    expect(linkMatch).not.toBeNull();
    const fullLink = linkMatch![0].trim();

    // Navigate on same browser context so session cookie is set
    const resp = await page.goto(fullLink, { waitUntil: 'domcontentloaded' });
    expect(resp?.ok() ?? true).toBeTruthy();
    await expect(page.locator('body')).toContainText("You're signed in");
    // Verify session cookie exists
    const cookies = await page.context().cookies();
    expect(cookies.some((c) => c.name === 'session')).toBeTruthy();

    const statusReply = await send('/status');
    expect(statusReply).toMatch(/Linked/i);
    expect(statusReply).toContain('e2e_login');
  });

  test('claim: inline-button claim, re-claim, and unclaim', async ({ page, workerData }) => {
    test.slow();
    const secret = 'claim-flow-secret-' + Date.now();
    const saveBot = await apiWithToken(page, 'POST', '/api/admin/telegram-settings', {
      token: workerData.telegramMock.botToken,
      mode: 'polling',
      webhook_secret: secret,
    });
    expect(saveBot.ok()).toBeTruthy();
    await expect.poll(() => workerData.telegramMock.commands.length, { timeout: 20000 }).toBeGreaterThan(0);

    // Link a fresh tg user via link code (admin bypass => can claim any)
    const codeResp = await apiWithToken(page, 'POST', '/api/telegram/link-code');
    expect(codeResp.ok()).toBeTruthy();
    const code = (await codeResp.json()).code as string;
    const tgUser: any = { id: 700100002, is_bot: false, first_name: 'Claim', username: 'e2e_claim' };
    const chat: any = { id: 700100002, type: 'private' };
    let updateID = 720000;

    const linkResp = await page.request.post('/api/telegram/webhook', {
      data: { update_id: ++updateID, message: { message_id: 1, from: tgUser, chat, text: `/start ${code}` } },
      headers: { 'X-Telegram-Bot-Api-Secret-Token': secret },
    });
    expect(linkResp.status()).toBe(200);

    // Seed a player character
    const charName = `ClaimHero-${Date.now()}`;
    const charResp = await apiWithToken(page, 'POST', '/api/characters', {
      name: charName,
      race: 'Human',
      class: 'Fighter',
      level: 3,
      campaign_ids: [],
    });
    expect(charResp.ok()).toBeTruthy();
    const charId = (await charResp.json()).id as number;

    const send = async (text: string): Promise<string> => {
      workerData.telegramMock.clear();
      const resp = await page.request.post('/api/telegram/webhook', {
        data: { update_id: ++updateID, message: { message_id: 1, from: tgUser, chat, text } },
        headers: { 'X-Telegram-Bot-Api-Secret-Token': secret },
      });
      expect(resp.status()).toBe(200);
      await expect.poll(() => workerData.telegramMock.sentMessages.length, { timeout: 15000 }).toBeGreaterThan(0);
      return workerData.telegramMock.sentMessages.map((m) => m.text).join('\n---\n');
    };

    const claimList = await send('/claim');
    expect(claimList).toMatch(/Tap a character to claim it:/i);

    // Character names live in the inline-keyboard button labels, not the message text.
    const lastMsg: any = workerData.telegramMock.sentMessages.at(-1);
    const rawMarkup: string = lastMsg?.body?.reply_markup ?? '';
    expect(rawMarkup).toBeTruthy();
    let callbackData = '';
    let buttonLabel = '';
    const markup = JSON.parse(rawMarkup);
    const kb: any[][] = markup.inline_keyboard ?? [];
    for (const row of kb) {
      for (const btn of row) {
        if (btn.callback_data === `claim:${charId}`) {
          callbackData = btn.callback_data;
          buttonLabel = String(btn.text ?? '');
        }
      }
    }
    expect(buttonLabel).toContain(charName);
    expect(callbackData).toBe(`claim:${charId}`);

    // Tap the inline button via callback_query
    workerData.telegramMock.clear();
    const cbResp = await page.request.post('/api/telegram/webhook', {
      data: {
        update_id: ++updateID,
        callback_query: {
          id: 'cb1',
          from: tgUser,
          message: { message_id: 1, date: Math.floor(Date.now() / 1000), chat },
          data: callbackData,
        },
      },
      headers: { 'X-Telegram-Bot-Api-Secret-Token': secret },
    });
    expect(cbResp.status()).toBe(200);
    await expect.poll(() => workerData.telegramMock.sentMessages.length, { timeout: 15000 }).toBeGreaterThan(0);
    const claimReply = workerData.telegramMock.sentMessages.map((m) => m.text).join('\n---\n');
    expect(claimReply).toContain(charName);

    const claimedAgain = await send('/claim');
    expect(claimedAgain).toMatch(/Currently claimed/i);
    expect(claimedAgain).toContain(charName);

    const unclaimReply = await send('/unclaim');
    expect(unclaimReply).toMatch(/released|unclaim/i);
  });

  test('status responds for a linked user', async ({ page, workerData }) => {
    test.slow();
    const secret = 'status-secret-' + Date.now();
    const saveBot = await apiWithToken(page, 'POST', '/api/admin/telegram-settings', {
      token: workerData.telegramMock.botToken,
      mode: 'polling',
      webhook_secret: secret,
    });
    expect(saveBot.ok()).toBeTruthy();
    await expect.poll(() => workerData.telegramMock.commands.length, { timeout: 20000 }).toBeGreaterThan(0);

    const codeResp = await apiWithToken(page, 'POST', '/api/telegram/link-code');
    expect(codeResp.ok()).toBeTruthy();
    const code = (await codeResp.json()).code as string;
    const tgUser: any = { id: 700100003, is_bot: false, first_name: 'Status', username: 'e2e_status' };
    const chat: any = { id: 700100003, type: 'private' };
    let updateID = 730000;

    const linkResp = await page.request.post('/api/telegram/webhook', {
      data: { update_id: ++updateID, message: { message_id: 1, from: tgUser, chat, text: `/start ${code}` } },
      headers: { 'X-Telegram-Bot-Api-Secret-Token': secret },
    });
    expect(linkResp.status()).toBe(200);

    workerData.telegramMock.clear();
    const r = await page.request.post('/api/telegram/webhook', {
      data: { update_id: ++updateID, message: { message_id: 1, from: tgUser, chat, text: '/status' } },
      headers: { 'X-Telegram-Bot-Api-Secret-Token': secret },
    });
    expect(r.status()).toBe(200);
    await expect.poll(() => workerData.telegramMock.sentMessages.length, { timeout: 15000 }).toBeGreaterThan(0);
    const text = workerData.telegramMock.sentMessages.map((m) => m.text).join('\n---\n');
    expect(text).toMatch(/Linked/i);
    expect(text.length).toBeGreaterThan(0);
  });
});
