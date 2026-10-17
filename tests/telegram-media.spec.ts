import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT } from './helpers.js';

const PNG_1x1 = Buffer.from(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAAEElEQVR4nGP4z8AARAAEAQIAvOQYVwAAAABJRU5ErkJggg==',
  'base64',
);
const PNG_RED = Buffer.from(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADUlEQVR4nGPw8z8AAwAB/6X9W1oAAAAASUVORK5CYII=',
  'base64',
);
function pngVariant(seed: number): Buffer {
  return seed % 2 === 0 ? PNG_1x1 : PNG_RED;
}
function mp3Bytes(seed = 0): Buffer {
  return Buffer.from(`ID3\x03\x00\x00\x00\x00\x00\x00track-${seed}-` + 'a'.repeat(256));
}

async function getApiToken(page: any): Promise<string> {
  const csrfResp = await page.request.get('/api/csrf-token');
  const j = await csrfResp.json();
  const csrf = j.token;
  const createResp = await page.request.post('/api/tokens', {
    data: { name: 'e2e-media-' + Date.now() },
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
async function uploadForCampaign(page: any, campaignId: number, filename: string, data: Buffer) {
  const token = await getApiToken(page);
  const csrf = (await (await page.request.get('/api/csrf-token')).json()).token;
  const resp = await page.request.post('/api/upload', {
    headers: { 'X-CSRF-Token': csrf, Authorization: `Bearer ${token}` },
    multipart: {
      image: { name: filename, mimeType: 'application/octet-stream', buffer: data },
      owner_type: 'campaign',
      owner_id: String(campaignId),
    },
  });
  if (!resp.ok()) throw new Error(`upload ${filename} failed ${resp.status()} ${await resp.text()}`);
  const j = await resp.json();
  const linkResp = await page.request.post('/api/upload-links', {
    data: { upload_id: j.id, entity_type: 'campaign', entity_id: campaignId, field_name: '' },
    headers: { 'X-CSRF-Token': csrf, Authorization: `Bearer ${token}` },
  });
  if (!linkResp.ok()) throw new Error(`link failed ${await linkResp.text()}`);
  return j as { id: number; url: string; kind: string };
}

let nextUpdateId = Date.now();
function allocUpdateId() {
  return ++nextUpdateId;
}
async function setupBot(page: any, workerData: any, opts?: { autoSave?: boolean; campaignId?: number; chatId?: number }) {
  const secret = 'media-secret-' + Date.now() + '-' + Math.random().toString(36).slice(2, 8);
  const r = await apiWithToken(page, 'POST', '/api/admin/telegram-settings', {
    token: workerData.telegramMock.botToken,
    mode: 'polling',
    webhook_secret: secret,
  });
  expect(r.ok()).toBeTruthy();
  await expect.poll(() => workerData.telegramMock.commands.length, { timeout: 20000 }).toBeGreaterThan(0);
  if (opts?.campaignId && opts?.chatId) {
    const bind = await apiWithToken(page, 'PUT', `/api/campaigns/${opts.campaignId}/telegram`, {
      chat_id: opts.chatId,
      is_enabled: true,
      auto_post_enabled: false,
      auto_save_media: !!opts.autoSave,
    });
    expect(bind.ok()).toBeTruthy();
  }
  return secret;
}

test.describe('Telegram rich media', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('outbound: post gallery image to bound chat asserts sendPhoto', async ({ page, workerData }) => {
    // slow: bot startup + upload + media post round-trip
    test.slow();
    const campResp = await apiWithToken(page, 'POST', '/api/campaigns', { name: `media-camp-${Date.now()}`, description: 'outbound photo' });
    expect(campResp.ok()).toBeTruthy();
    const campId = (await campResp.json()).id as number;
    const chatId = -100400000001 + Math.floor(Math.random() * 1000);
    const secret = await setupBot(page, workerData, { campaignId: campId, chatId });
    void secret;
    const up = await uploadForCampaign(page, campId, `photo-${Date.now()}.png`, PNG_1x1);
    workerData.telegramMock.clear();
    const postResp = await apiWithToken(page, 'POST', '/api/telegram/media', {
      upload_ids: [up.id],
      target: { kind: 'campaign', campaign_id: campId },
      caption: 'e2e photo',
    });
    expect(postResp.ok()).toBeTruthy();
    const body = await postResp.json();
    expect(body.results).toBeDefined();
    expect(body.results[0].status).toBe('sent');
    await expect.poll(() => workerData.telegramMock.sentMessages.length, { timeout: 15000 }).toBeGreaterThan(0);
    const m = workerData.telegramMock.sentMessages.find((x: any) => x.method === 'sendPhoto');
    expect(m).toBeTruthy();
    expect(m.chat_id).toBe(chatId);
    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('body')).toBeVisible({ timeout: NAV_TIMEOUT });
  });

  test('outbound: album posting multi-select photos/videos via sendMediaGroup', async ({ page, workerData }) => {
    test.slow();
    const campResp = await apiWithToken(page, 'POST', '/api/campaigns', { name: `album-camp-${Date.now()}` });
    expect(campResp.ok()).toBeTruthy();
    const campId = (await campResp.json()).id as number;
    const chatId = -100400000002 - Math.floor(Math.random() * 1000);
    await setupBot(page, workerData, { campaignId: campId, chatId });
    const up1 = await uploadForCampaign(page, campId, `a1-${Date.now()}.png`, pngVariant(0));
    const up2 = await uploadForCampaign(page, campId, `a2-${Date.now()}.png`, pngVariant(1));
    workerData.telegramMock.clear();
    const postResp = await apiWithToken(page, 'POST', '/api/telegram/media', {
      upload_ids: [up1.id, up2.id],
      target: { kind: 'campaign', campaign_id: campId },
      caption: 'album',
    });
    expect(postResp.ok()).toBeTruthy();
    const j = await postResp.json();
    expect(j.results.length).toBe(2);
    expect(j.results.filter((r: any) => r.status === 'sent').length).toBe(2);
    await expect.poll(() => workerData.telegramMock.sentMessages.length, { timeout: 15000 }).toBeGreaterThan(0);
    const grps = workerData.telegramMock.sentMessages.filter((x: any) => x.method === 'sendMediaGroup');
    expect(grps.length).toBeGreaterThan(0);
    const totalMedia = grps.reduce((s: number, g: any) => s + (g.body._mediaCount || 0), 0);
    const totalFiles = grps.reduce((s: number, g: any) => s + (g.body._fileCount || 0), 0);
    expect(totalMedia).toBe(2);
    expect(totalFiles).toBe(2);
    expect(grps[0].chat_id).toBe(chatId);
  });

  test('inbound: /save caption in bound chat saves to gallery', async ({ page, workerData }) => {
    test.slow();
    const campResp = await apiWithToken(page, 'POST', '/api/campaigns', { name: `inbound-camp-${Date.now()}` });
    expect(campResp.ok()).toBeTruthy();
    const campId = (await campResp.json()).id as number;
    const chatId = -100400000010 - Math.floor(Math.random() * 10000);
    const secret = await setupBot(page, workerData, { campaignId: campId, chatId });
    const uid = allocUpdateId();
    const fid = `f-in-${uid}`;
    const fuid = `u-in-${uid}`;
    const customPath = `files/${fid}/photo.png`;
    workerData.telegramMock.setFileMeta(fid, customPath, fuid);
    workerData.telegramMock.setFileFixture(customPath, PNG_1x1);
    workerData.telegramMock.setFileFixture(`files/${fid}/file.bin`, PNG_1x1);
    workerData.telegramMock.clear();
    const wh = await page.request.post('/api/telegram/webhook', {
      data: {
        update_id: allocUpdateId(),
        message: {
          message_id: 10,
          from: { id: 900001, is_bot: false, first_name: 'In', username: 'e2e_in' },
          chat: { id: chatId, type: 'supergroup', title: 'E2E' },
          caption: '/save my photo',
          photo: [{ file_id: fid, file_unique_id: fuid, file_size: 1234 }],
        },
      },
      headers: { 'X-Telegram-Bot-Api-Secret-Token': secret },
    });
    expect(wh.status()).toBe(200);
    await expect.poll(() => workerData.telegramMock.sentMessages.length, { timeout: 15000 }).toBeGreaterThan(0);
    expect(workerData.telegramMock.sentMessages.map((m: any) => m.text).join(' ')).toMatch(/Saved/i);
    // Hard assertion on gallery HTML (spec: "stored in the campaign gallery" = /htmx/media-gallery)
    await expect.poll(async () => {
      const r = await page.request.get(`/htmx/media-gallery?owner_type=campaign&owner_id=${campId}`);
      if (!r.ok()) return '';
      return r.text();
    }, { timeout: 15000 }).toContain('media-select-');
  });

  test('inbound: auto_save_media on saves without /save, off does not', async ({ page, workerData }) => {
    test.slow();
    const ca = await (await apiWithToken(page, 'POST', '/api/campaigns', { name: `auto-on-${Date.now()}` })).json();
    const caId = ca.id as number;
    const chatA = -100400001000 - Math.floor(Math.random() * 1000);
    const secret = await setupBot(page, workerData, { campaignId: caId, chatId: chatA });
    await apiWithToken(page, 'PUT', `/api/campaigns/${caId}/telegram`, { chat_id: chatA, is_enabled: true, auto_save_media: true });
    const fuidA = `u-auto-on-${allocUpdateId()}`;
    const fidA = `f-auto-on-${allocUpdateId()}`;
    const pathA = `files/${fidA}/a.png`;
    workerData.telegramMock.setFileMeta(fidA, pathA, fuidA);
    workerData.telegramMock.setFileFixture(pathA, PNG_1x1);
    workerData.telegramMock.clear();
    const whA = await page.request.post('/api/telegram/webhook', {
      data: {
        update_id: allocUpdateId(),
        message: {
          message_id: 11,
          from: { id: 900002, first_name: 'Auto' },
          chat: { id: chatA, type: 'supergroup' },
          photo: [{ file_id: fidA, file_unique_id: fuidA, file_size: 100 }],
        },
      },
      headers: { 'X-Telegram-Bot-Api-Secret-Token': secret },
    });
    expect(whA.status()).toBe(200);
    await expect.poll(() => workerData.telegramMock.sentMessages.length, { timeout: 15000 }).toBeGreaterThan(0);
    expect(workerData.telegramMock.sentMessages.map((m: any) => m.text).join(' ')).toMatch(/Saved/i);
    await expect.poll(async () => {
      const r = await page.request.get(`/htmx/media-gallery?owner_type=campaign&owner_id=${caId}`);
      return r.ok() ? r.text() : '';
    }, { timeout: 15000 }).toContain('media-select-');

    const cb = await (await apiWithToken(page, 'POST', '/api/campaigns', { name: `auto-off-${Date.now()}` })).json();
    const cbId = cb.id as number;
    const chatB = -100400002000 - Math.floor(Math.random() * 1000);
    await apiWithToken(page, 'PUT', `/api/campaigns/${cbId}/telegram`, { chat_id: chatB, is_enabled: true, auto_save_media: false });
    const fuidB = `u-auto-off-${allocUpdateId()}`;
    const fidB = `f-auto-off-${allocUpdateId()}`;
    const pathB = `files/${fidB}/b.png`;
    workerData.telegramMock.setFileMeta(fidB, pathB, fuidB);
    workerData.telegramMock.setFileFixture(pathB, PNG_1x1);
    workerData.telegramMock.clear();
    const whB = await page.request.post('/api/telegram/webhook', {
      data: {
        update_id: allocUpdateId(),
        message: {
          message_id: 12,
          from: { id: 900003, first_name: 'NoAuto' },
          chat: { id: chatB, type: 'supergroup' },
          photo: [{ file_id: fidB, file_unique_id: fuidB, file_size: 100 }],
        },
      },
      headers: { 'X-Telegram-Bot-Api-Secret-Token': secret },
    });
    expect(whB.status()).toBe(200);
    await new Promise((r) => setTimeout(r, 1200));
    expect(workerData.telegramMock.sentMessages.length).toBe(0);
    const htmlB = await (await page.request.get(`/htmx/media-gallery?owner_type=campaign&owner_id=${cbId}`)).text();
    const settingsB = await (await apiWithToken(page, 'GET', `/api/campaigns/${cbId}/telegram`)).json();
    expect(settingsB.auto_save_media).toBe(false);
    expect(htmlB).not.toContain('data-upload-id');
    expect(htmlB).toContain('No Media Yet');
  });

  test('inbound: DM picker flow saves to chosen campaign', async ({ page, workerData }) => {
    test.slow();
    const c1 = await (await apiWithToken(page, 'POST', '/api/campaigns', { name: `dm1-${Date.now()}` })).json();
    const c2 = await (await apiWithToken(page, 'POST', '/api/campaigns', { name: `dm2-${Date.now()}` })).json();
    const secret = await setupBot(page, workerData);
    const codeResp = await apiWithToken(page, 'POST', '/api/telegram/link-code');
    expect(codeResp.ok()).toBeTruthy();
    const code = (await codeResp.json()).code as string;
    const tgUser = { id: 980000001, is_bot: false, first_name: 'DM', username: 'e2e_dm' };
    const dmChat = { id: 980000001, type: 'private' as const };
    const linkWh = await page.request.post('/api/telegram/webhook', {
      data: { update_id: allocUpdateId(), message: { message_id: 1, from: tgUser, chat: dmChat, text: `/start ${code}` } },
      headers: { 'X-Telegram-Bot-Api-Secret-Token': secret },
    });
    expect(linkWh.status()).toBe(200);
    await expect.poll(() => workerData.telegramMock.sentMessages.some((m: any) => /Linked/i.test(m.text)), { timeout: 15000 }).toBeTruthy();
    workerData.telegramMock.clear();
    const fid = `f-dm-${allocUpdateId()}`;
    const fuid = `u-dm-${allocUpdateId()}`;
    const path = `files/${fid}/dm.png`;
    workerData.telegramMock.setFileMeta(fid, path, fuid);
    workerData.telegramMock.setFileFixture(path, pngVariant(0));
    const wh = await page.request.post('/api/telegram/webhook', {
      data: {
        update_id: allocUpdateId(),
        message: { message_id: 20, date: Math.floor(Date.now() / 1000), from: tgUser, chat: dmChat, photo: [{ file_id: fid, file_unique_id: fuid, file_size: 100 }] },
      },
      headers: { 'X-Telegram-Bot-Api-Secret-Token': secret },
    });
    expect(wh.status()).toBe(200);
    await expect.poll(() => workerData.telegramMock.sentMessages.map((m: any) => m.text).join('\n'), { timeout: 15000 }).toMatch(/Choose where to save/i);
    const picker = workerData.telegramMock.sentMessages.map((m: any) => m.text).join('\n');
    expect(picker).toMatch(/1\./);
    expect(picker).toMatch(/2\./);
    workerData.telegramMock.clear();
    const saveWh = await page.request.post('/api/telegram/webhook', {
      data: {
        update_id: allocUpdateId(),
        message: { message_id: 21, from: tgUser, chat: dmChat, text: '/save 2' },
      },
      headers: { 'X-Telegram-Bot-Api-Secret-Token': secret },
    });
    expect(saveWh.status()).toBe(200);
    await expect.poll(() => workerData.telegramMock.sentMessages.map((m: any) => m.text).join(' '), { timeout: 15000 }).toMatch(/Saved/i);
    // Verify saved into second campaign's gallery (hard assertion)
    const c2Html = await (await page.request.get(`/htmx/media-gallery?owner_type=campaign&owner_id=${c2.id}`)).text();
    expect(c2Html).toContain('media-select-');
    // Pending capture consumed: second /save should report no pending
    workerData.telegramMock.clear();
    const saveAgain = await page.request.post('/api/telegram/webhook', {
      data: {
        update_id: allocUpdateId(),
        message: { message_id: 22, from: tgUser, chat: dmChat, text: '/save 1' },
      },
      headers: { 'X-Telegram-Bot-Api-Secret-Token': secret },
    });
    expect(saveAgain.status()).toBe(200);
    await expect.poll(() => workerData.telegramMock.sentMessages.length, { timeout: 15000 }).toBeGreaterThan(0);
    expect(workerData.telegramMock.sentMessages.map((m: any) => m.text).join(' ')).toMatch(/No pending media|expired/i);
    await apiWithToken(page, 'DELETE', '/api/telegram/unlink');
  });

  test('upload: audio file end-to-end renders audio player', async ({ page }) => {
    const campResp = await apiWithToken(page, 'POST', '/api/campaigns', { name: `audio-camp-${Date.now()}` });
    expect(campResp.ok()).toBeTruthy();
    const campId = (await campResp.json()).id as number;
    const up = await uploadForCampaign(page, campId, `track-${Date.now()}.mp3`, mp3Bytes(Date.now()));
    expect(up.kind).toBe('audio');
    const galResp = await page.request.get(`/htmx/media-gallery?owner_type=campaign&owner_id=${campId}`);
    expect(galResp.ok()).toBeTruthy();
    const html = await galResp.text();
    expect(html).toContain(`media-select-${up.id}`);
    expect(html).toContain(`media-post-${up.id}`);
    expect(html).toContain(`media-audio-${up.id}`);
    expect(html).toContain('<audio');
  });
});
