import { test, expect } from '@playwright/test';

const uniqueName = () => `G-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function waitLoadingDone(page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: 5000 }).catch(() => {});
}

async function waitModalClosed(page) {
  await page.waitForFunction(() => {
    const modal = document.getElementById('genericModal');
    return !modal || !modal.classList.contains('show');
  }, { timeout: 10000 }).catch(() => {});
}

test.describe('D3 Graph Visualization', () => {
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

  test('character graph renders SVG in graph tab', async ({ page }) => {
    const name = uniqueName();

    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Graph")');
    await page.waitForTimeout(1000);

    const graphSection = page.locator('#graphSection');
    await expect(graphSection).toBeVisible();

    const svg = graphSection.locator('svg');
    await expect(svg).toBeVisible({ timeout: 5000 });
  });

  test('campaign graph renders SVG in modal', async ({ page }) => {
    const campName = uniqueName();

    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Graph Test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (name) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === name);
      return c ? c.id : null;
    }, campName);

    expect(result).toBeTruthy();

    await page.evaluate(async (cid) => {
      await window.api('POST', `/api/campaigns/${cid}/wiki`, {
        campaign_id: cid, title: 'Test Page', content: '# Hello',
        visibility: 'public', tags: '[]', sort_order: 0,
      });
    }, result);

    await page.evaluate((cid) => (window as any).showCampaignGraph(cid), result);
    await page.waitForTimeout(1000);

    const modal = page.locator('#genericModal');
    await expect(modal).toHaveClass(/show/);

    const svg = modal.locator('svg');
    await expect(svg).toBeVisible({ timeout: 5000 });

    const stats = page.locator('#campaignGraphStats');
    await expect(stats).not.toContainText('Loading');
  });

  test('graph nodes have text labels', async ({ page }) => {
    const name = uniqueName();

    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Rogue');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Graph")');
    await page.waitForTimeout(1000);

    const svgTexts = page.locator('#graphSection svg text');
    const count = await svgTexts.count();
    expect(count).toBeGreaterThanOrEqual(1);
  });
});
