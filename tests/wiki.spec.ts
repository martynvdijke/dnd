import { test, expect } from '@playwright/test';
import { login } from './helpers.js';

const uniqueName = () => `Wiki-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

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

  test('update a wiki page', async ({ page }) => {
    const campName = uniqueName();
    const pageTitle = 'Update Test';
    const updatedTitle = 'Updated Page';

    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Update wiki test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'campaign not found' };
      const p = await window.api('POST', `/api/campaigns/${c.id}/wiki`, {
        campaign_id: c.id, title: opts.pageTitle, content: '# Original',
        visibility: 'public', tags: '[]', sort_order: 0,
      });
      return { campId: c.id, pageId: p.id };
    }, { campName, pageTitle });

    await page.evaluate(async (opts) => {
      // Get wiki pages list to find the page's update endpoint
      const pages = await window.api('GET', `/api/campaigns/${opts.campId}/wiki`);
      const wp = pages.find((p: any) => p.id === opts.pageId);
      if (wp) {
        await window.api('PUT', `/api/wiki/${opts.pageId}`, {
          title: opts.updatedTitle, content: '# Updated Content',
          visibility: 'public', tags: '[]', sort_order: 0,
        });
      }
      return { ok: true };
    }, { campId: result.campId, pageId: result.pageId, updatedTitle });

    // Verify update by loading the page
    await page.evaluate((cid) => window.showWiki(cid), result.campId);
    await page.waitForTimeout(500);
    await page.evaluate((pid) => window.loadWikiPage(pid), result.pageId);
    await page.waitForTimeout(500);

    await expect(page.locator('#wikiPageContent .wiki-content')).toContainText('Updated Content', { timeout: 5000 });
  });

  test('delete a wiki page', async ({ page }) => {
    const campName = uniqueName();
    const pageTitle = 'Delete Test';

    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Delete wiki test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'campaign not found' };
      const p = await window.api('POST', `/api/campaigns/${c.id}/wiki`, {
        campaign_id: c.id, title: opts.pageTitle, content: '# Delete Me',
        visibility: 'public', tags: '[]', sort_order: 0,
      });
      return { campId: c.id, pageId: p.id };
    }, { campName, pageTitle });

    await page.evaluate(async (opts) => {
      try {
        await window.api('DELETE', `/api/wiki/${opts.pageId}`);
        return { ok: true };
      } catch (e) {
        return { ok: false, error: String(e) };
      }
    }, { campId: result.campId, pageId: result.pageId });

    // Verify deletion by trying to get the specific page (should fail)
    const verify = await page.evaluate(async (pageId) => {
      try {
        await window.api('GET', `/api/wiki/${pageId}`);
        return { exists: true };
      } catch (e) {
        return { exists: false, error: String(e) };
      }
    }, result.pageId);

    expect(verify.exists).toBe(false);
  });

  test('list wiki pages for a campaign', async ({ page }) => {
    const campName = uniqueName();

    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'List wiki test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'campaign not found' };
      await window.api('POST', `/api/campaigns/${c.id}/wiki`, {
        campaign_id: c.id, title: 'Page 1', content: '# One',
        visibility: 'public', tags: '[]', sort_order: 0,
      });
      await window.api('POST', `/api/campaigns/${c.id}/wiki`, {
        campaign_id: c.id, title: 'Page 2', content: '# Two',
        visibility: 'public', tags: '[]', sort_order: 0,
      });
      const pages = await window.api('GET', `/api/campaigns/${c.id}/wiki`);
      return { ok: true, count: pages.length, titles: pages.map((p: any) => p.title) };
    }, { campName });

    expect(result.err).toBeFalsy();
    expect(result.count).toBeGreaterThanOrEqual(2);
    expect(result.titles).toContain('Page 1');
    expect(result.titles).toContain('Page 2');
  });

  test('wiki page with special characters in content', async ({ page }) => {
    const campName = uniqueName();
    const pageTitle = 'Special Char';

    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Special chars test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'not found' };
      const p = await window.api('POST', `/api/campaigns/${c.id}/wiki`, {
        campaign_id: c.id, title: opts.pageTitle,
        content: '# Special\n\nD&D 5e characters: áéíóú üñ ©®™\n```\ncode block\n```',
        visibility: 'public', tags: '[]', sort_order: 0,
      });
      return { campId: c.id, pageId: p.id };
    }, { campName, pageTitle });

    await page.evaluate((cid) => window.showWiki(cid), result.campId);
    await page.waitForTimeout(500);
    await page.evaluate((pid) => window.loadWikiPage(pid), result.pageId);

    // Wait for wiki content to be rendered (may take longer on mobile with DB contention)
    await page.waitForFunction(() => {
      const el = document.querySelector('#wikiPageContent .wiki-content');
      return el && el.textContent && el.textContent.length > 0;
    }, { timeout: 15000 });

    await expect(page.locator('#wikiPageContent .wiki-content')).toContainText('Special', { timeout: 5000 });
    await expect(page.locator('#wikiPageContent .wiki-content')).toContainText('D&D 5e', { timeout: 5000 });
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
