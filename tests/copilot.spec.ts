import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT } from './helpers.js';
import http from 'node:http';

let server: http.Server;
let mockPort = 0;

test.beforeAll(async () => {
  server = http.createServer((req, res) => {
    if (req.method === 'POST' && req.url?.endsWith('/chat/completions')) {
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ choices: [{ message: { content: 'Mock copilot answer.' }, finish_reason: 'stop' }] }));
    } else { res.writeHead(404); res.end(); }
  });
  await new Promise<void>(r => server.listen(0, '127.0.0.1', () => { mockPort = (server.address() as any).port; r(); }));
});

test.afterAll(async () => { await new Promise<void>(r => server.close(() => r())); });

const uniqueName = () => `Copilot-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('Campaign Copilot', () => {
  test('chat, conversations, transcript ingest and summarize', async ({ page }) => {
    await login(page);

    const campName = uniqueName();
    const wikiTitle = 'Wiki-' + uniqueName();

    // Seed campaign and wiki page + AI endpoint
    const seed = await page.evaluate(async (opts) => {
      const camp: any = await (window as any).api('POST', '/api/campaigns', { name: opts.campName, description: 'copilot test', dm_notes: '' });
      // create wiki page for citation
      await (window as any).api('POST', `/api/campaigns/${camp.id}/wiki`, { campaign_id: camp.id, title: opts.wikiTitle, content: 'Ancient lore about dragons', visibility: 'public' });
      const ep: any = await (window as any).api('POST', '/api/admin/ai-endpoints', {
        name: 'mock-' + Date.now() + '-' + Math.random().toString(36).slice(2, 5),
        type: 'text',
        base_url: 'http://127.0.0.1:' + opts.mockPort + '/v1',
        api_key: 'x',
        model: 'mock',
        enabled: true,
      });
      return { campId: camp.id, epId: ep.id, wikiTitle: opts.wikiTitle };
    }, { campName, wikiTitle, mockPort });

    // Open copilot view
    await page.evaluate((cid) => (window as any).showCopilot(cid), seed.campId);

    await expect(page.getByTestId('copilot-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('copilot-content')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('copilot-query')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('copilot-ask')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('copilot-answer')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('copilot-sources')).toBeAttached({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('copilot-conversations')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('copilot-new')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('copilot-transcript')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('copilot-ingest')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('copilot-summarize')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('copilot-prep')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('copilot-history')).toBeAttached({ timeout: NAV_TIMEOUT });

    // Chat
    const chatResp = page.waitForResponse((r) => r.url().includes(`/api/campaigns/${seed.campId}/copilot/chat`) && r.request().method() === 'POST', { timeout: NAV_TIMEOUT });
    await page.getByTestId('copilot-query').fill('dragons lore');
    await page.getByTestId('copilot-ask').click();
    await chatResp;

    await expect(page.getByTestId('copilot-answer')).toContainText('Mock copilot answer.', { timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('copilot-sources')).toBeVisible({ timeout: NAV_TIMEOUT });
    // at least one source link whose text matches seeded wiki title
    await expect(page.getByTestId('copilot-source-link').first()).toBeVisible({ timeout: NAV_TIMEOUT });
    const sourceTexts = await page.getByTestId('copilot-source-link').allTextContents();
    expect(sourceTexts.join(' ')).toContain(wikiTitle);

    // Conversations list eventually shows conversation
    await expect(page.getByTestId('copilot-conversation').first()).toBeVisible({ timeout: NAV_TIMEOUT });

    // Load first conversation
    const convId = await page.getByTestId('copilot-conversation').first().getAttribute('data-id');
    if (convId) {
      const convResp = page.waitForResponse((r) => r.url().includes(`/api/campaigns/${seed.campId}/copilot/conversations/${convId}`) && r.request().method() === 'GET', { timeout: NAV_TIMEOUT });
      await page.evaluate((id: any) => (window as any).copilotLoadConversation(id), Number(convId) || convId);
      await convResp.catch(() => {});
      await expect(page.getByTestId('copilot-history')).toBeVisible({ timeout: NAV_TIMEOUT });
    }

    // Transcript ingest
    const ingestResp = page.waitForResponse((r) => r.url().includes(`/api/campaigns/${seed.campId}/copilot/transcript`) && r.request().method() === 'POST', { timeout: NAV_TIMEOUT });
    await page.getByTestId('copilot-transcript').fill('Session transcript: the party fought a dragon near the mountain.');
    await page.getByTestId('copilot-ingest').click();
    await ingestResp;

    // Summarize
    const sumResp = page.waitForResponse((r) => r.url().includes('/copilot/transcript/summarize') && r.request().method() === 'POST', { timeout: NAV_TIMEOUT });
    await page.getByTestId('copilot-summarize').click();
    await sumResp;
    await expect(page.getByTestId('copilot-answer')).toBeVisible({ timeout: NAV_TIMEOUT });

    // New conversation button
    await page.getByTestId('copilot-new').click();
    await expect(page.getByTestId('copilot-history')).toBeAttached({ timeout: NAV_TIMEOUT });

    // Prep
    const prepResp = page.waitForResponse((r) => r.url().includes(`/api/campaigns/${seed.campId}/copilot/prep`) && r.request().method() === 'POST', { timeout: NAV_TIMEOUT });
    await page.getByTestId('copilot-prep').click();
    await prepResp;
    await expect(page.getByTestId('copilot-answer')).toBeVisible({ timeout: NAV_TIMEOUT });
  });
});
