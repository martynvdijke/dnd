import { test, expect } from './fixtures.js';
import { login, waitLoadingDone } from './helpers.js';

test.describe('Ambience Soundboard', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await waitLoadingDone(page);
  });

  test('DM broadcasts a soundscape and the client plays it', async ({ page }) => {
    const cid = await page.evaluate(async () => {
      const c = await (window as any).api('POST', '/api/campaigns', {
        name: 'Amb Camp', party_name: 'Testers', description: '', dm_notes: '',
      });
      return c.id as number;
    });

    await page.evaluate((id) => (window as any).ambiencePlay(id, 'rain'), cid);
    await expect.poll(async () => page.evaluate(() => document.body.dataset.ambience)).toBe('rain');

    await page.evaluate((id) => (window as any).ambienceStop(id), cid);
    await expect.poll(async () => page.evaluate(() => document.body.dataset.ambience ?? null)).toBe(null);
  });
});
