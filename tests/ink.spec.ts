import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT, getApiToken, waitLoadingDone } from './helpers.js';

const uniqueName = () => `Ink-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function getCSRFToken(page: any): Promise<string> {
  const resp = await page.request.get('/api/csrf-token');
  const data = await resp.json();
  return data.token;
}

async function createCampaign(page: any, name: string) {
  return page.evaluate(async (n: string) => (window as any).api('POST', '/api/campaigns', { name: n, description: 'ink e2e', dm_notes: '' }), name);
}

async function openRecaps(page: any) {
  const nav = page.locator('[data-testid="nav-recaps"]');
  if (await nav.first().isVisible().catch(() => false)) {
    await nav.first().click();
  } else {
    await page.evaluate(() => (window as any).showRecaps?.());
  }
  await expect(page.locator('#recapsView')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ink-inbox')).toBeVisible({ timeout: NAV_TIMEOUT });
}

const ONE_PIXEL_PNG = Buffer.from(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=',
  'base64',
);

async function uploadCapture(page: any, campaignId: number, title: string, text: string) {
  const csrf = await getCSRFToken(page);
  const apiToken = await getApiToken(page);
  const resp = await page.request.post('/api/ink/captures', {
    multipart: {
      campaign_id: String(campaignId),
      title,
      source: 'share',
      text,
      pages: { name: 'page.png', mimeType: 'image/png', buffer: ONE_PIXEL_PNG },
    },
    headers: { 'X-CSRF-Token': csrf, Authorization: `Bearer ${apiToken}` },
  });
  expect(resp.status()).toBe(201);
  return resp.json();
}

test.describe('Ink inbox', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('shows an empty state for a campaign without pages', async ({ page }) => {
    const camp = await createCampaign(page, `InkEmpty ${uniqueName()}`);
    await page.evaluate((id: number) => (window as any).selectCampaign(id), camp.id as number);
    await waitLoadingDone(page);
    await openRecaps(page);
    await expect(page.getByTestId('ink-empty')).toBeVisible({ timeout: NAV_TIMEOUT });
  });

  test('reviews a shared page and promotes it to an unpublished recap', async ({ page }) => {
    const name = uniqueName();
    const camp = await createCampaign(page, `InkCamp ${name}`);
    const cid = camp.id as number;
    const title = `Boox notes ${name}`;
    const text = 'The party found a sealed door beneath the mill.';
    await page.evaluate((id: number) => (window as any).selectCampaign(id), cid);
    await waitLoadingDone(page);

    const first = await uploadCapture(page, cid, title, text);
    const captureId = first.capture.id as number;
    await openRecaps(page);

    const capture = page.getByTestId('ink-capture').first();
    await expect(capture).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(capture).toContainText(title);
    await expect(capture).toContainText('Transcribed');
    await expect(capture.locator('[data-testid="ink-page-text"]').first()).toHaveValue(text);

    const promoted = page.waitForResponse(
      (r) => r.url().includes('/promote') && r.request().method() === 'POST',
      { timeout: NAV_TIMEOUT },
    );
    await capture.getByRole('button', { name: /Recap/ }).click();
    expect((await promoted).status()).toBe(201);

    await expect(page.locator('#recapsContent')).toContainText(title, { timeout: NAV_TIMEOUT });
    await expect(page.locator('#inkInbox')).toContainText('Promoted', { timeout: NAV_TIMEOUT });

    // A second promote is rejected, so the button path cannot duplicate recaps.
    const duplicate = await page.evaluate(async (id: number) => {
      try {
        await (window as any).api('POST', `/api/ink/captures/${id}/promote`, { target: 'recap' });
        return 0;
      } catch (e: any) {
        return e.status;
      }
    }, captureId);
    expect(duplicate).toBe(409);
  });
});
