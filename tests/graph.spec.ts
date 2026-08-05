import { test, expect } from './fixtures.js';
import { login, waitLoadingDone, waitModalClosed, NAV_TIMEOUT } from './helpers.js';

const uniqueName = () => `G-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('D3 Graph Visualization', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
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
    await expect(svg).toBeVisible({ timeout: NAV_TIMEOUT });
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

    await page.evaluate(async (cid) => {
      await (window as any).showCampaignGraph(cid);
    }, result);

    const modal = page.locator('#genericModal');
    await expect(modal).toHaveClass(/show/);

    await page.waitForFunction(() => {
      const el = document.getElementById('campaignGraphStats');
      return el && !el.textContent?.includes('Loading');
    }, { timeout: 15000 });

    const svg = modal.locator('svg');
    await expect(svg).toBeVisible({ timeout: NAV_TIMEOUT });
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
