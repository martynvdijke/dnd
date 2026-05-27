import { test, expect } from '@playwright/test';

const uniqueName = () => `Wiki-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function waitLoadingDone(page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: 5000 }).catch(() => {});
}

async function login(page) {
  await page.goto('/login', { waitUntil: 'domcontentloaded' });
  await page.fill('#username', 'admin');
  await page.fill('#password', 'testpassword123');
  await Promise.all([
    page.waitForURL('/', { waitUntil: 'domcontentloaded', timeout: 10000 }),
    page.click('button[type="submit"]'),
  ]);
  await waitLoadingDone(page);
  await page.waitForTimeout(200);
}

test.describe('Campaign Wiki', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('create campaign and wiki page via API', async ({ page }) => {
    const campName = uniqueName();
    const pageTitle = 'World Lore';
    const pageContent = '# The World\n\nThis is the **lore** of our world.';

    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Test campaign', dm_notes: '' });
    }, campName);

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

    await page.evaluate((cid) => window.showWiki(cid), result.campId);
    await page.waitForTimeout(500);

    await expect(page.locator('#wikiContent')).toContainText(pageTitle);
  });

  test('wiki page renders markdown content', async ({ page }) => {
    const campName = uniqueName();
    const pageTitle = 'Markdown Test';
    const pageContent = '# Heading\n\n**Bold text** *italic*';

    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'not found' };
      const p = await window.api('POST', `/api/campaigns/${c.id}/wiki`, {
        campaign_id: c.id, title: opts.pageTitle, content: opts.pageContent,
        visibility: 'public', tags: '[]', sort_order: 0,
      });
      return { campId: c.id, pageId: p.id };
    }, { campName, pageTitle, pageContent });

    // Pass campaign ID so showWiki renders its sidebar (otherwise picks campaigns[0])
    await page.evaluate((cid) => window.showWiki(cid), result.campId);
    await page.waitForTimeout(500);

    await page.evaluate((pid) => window.loadWikiPage(pid), result.pageId);
    await page.waitForTimeout(500);

    const wikiContent = page.locator('#wikiPageContent .wiki-content');
    await expect(wikiContent).toContainText('Heading', { timeout: 10000 });
    await expect(wikiContent).toContainText('Bold text', { timeout: 5000 });
  });

  test('wiki offcanvas works on mobile viewport', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });

    const campName = uniqueName();
    const pageTitle = 'Mobile Page';

    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'not found' };
      await window.api('POST', `/api/campaigns/${c.id}/wiki`, {
        campaign_id: c.id, title: opts.pageTitle, content: '# Test',
        visibility: 'public', tags: '[]', sort_order: 0,
      });
      return { campId: c.id };
    }, { campName, pageTitle });

    await page.evaluate((cid) => window.showWiki(cid), result.campId);
    await page.waitForTimeout(500);

    const sidebarDesktop = page.locator('.col-md-3.d-none.d-md-block');
    await expect(sidebarDesktop).toBeHidden();

    const toggleBtn = page.locator('[onclick^="toggleWikiSidebar()"]');
    if (await toggleBtn.isVisible()) {
      await toggleBtn.click();
      await page.waitForTimeout(400);
      const offcanvas = page.locator('#wikiOffcanvas');
      await expect(offcanvas).toHaveClass(/show/);
    }
  });
});
