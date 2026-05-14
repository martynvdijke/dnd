import { test, expect } from '@playwright/test';

const uniqueName = () => `Wiki-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function waitLoadingDone(page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: 5000 }).catch(() => {});
}

test.describe('Campaign Wiki', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded', timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
    await page.waitForTimeout(200);
  });

  test('create campaign and wiki page via API', async ({ page }) => {
    const campName = uniqueName();
    const pageTitle = 'World Lore';
    const pageContent = '# The World\n\nThis is the **lore** of our world.';

    // Create campaign
    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Test campaign', dm_notes: '' });
    }, campName);

    // Get campaign ID and create wiki page in one evaluate
    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'campaign not found' };
      await window.api('POST', `/api/campaigns/${c.id}/wiki`, {
        campaign_id: c.id,
        title: opts.pageTitle,
        content: opts.pageContent,
        visibility: 'public',
        tags: '[]',
        sort_order: 0,
      });
      return { ok: true, campId: c.id };
    }, { campName, pageTitle, pageContent });

    expect(result.err).toBeFalsy();

    // View wiki
    await page.evaluate((cid) => window.showWiki(cid), result.campId);
    await page.waitForTimeout(500);

    await expect(page.locator('#wikiContent')).toContainText(pageTitle);
  });
});
