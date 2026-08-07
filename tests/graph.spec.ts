import { test, expect } from './fixtures.js';
import { login, waitLoadingDone, waitModalClosed, NAV_TIMEOUT, clickNavItem } from './helpers.js';

const uniqueName = () => `G-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('D3 Graph Visualization', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('party graph renders SVG', async ({ page }) => {
    const name = uniqueName();

    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    const charId = await page.evaluate(async (charName) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === charName);
      return char ? char.id : null;
    }, name);
    expect(charId).toBeTruthy();

    await page.evaluate(async (cid) => {
      const loc = await window.api('POST', '/api/locations', { name: 'Waterdeep', type: 'city', description: 'City' });
      await window.api('POST', `/api/characters/${cid}/locations`, { location_id: loc.id, relationship: 'home', notes: '' });
    }, charId);

    await clickNavItem(page, 'party', 'party');
    await page.locator('#partySubTabBar').waitFor({ state: 'visible', timeout: NAV_TIMEOUT });
    await page.locator('#partySubTabBar button:has-text("Graph")').click();
    await waitLoadingDone(page);

    await expect(page.locator('#partyContent h5').first()).toContainText('Campaign Graph');
    const svg = page.locator('#partyGraphSvg svg');
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

  test('party graph nodes have text labels', async ({ page }) => {
    const name = uniqueName();

    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Rogue');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    const charId = await page.evaluate(async (charName) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === charName);
      return char ? char.id : null;
    }, name);
    expect(charId).toBeTruthy();

    await page.evaluate(async (cid) => {
      const npc = await window.api('POST', '/api/npcs', { name: 'Luthien', race: 'Elf', class: 'Scout', description: 'Companion' });
      await window.api('POST', `/api/characters/${cid}/npcs`, { npc_id: npc.id, relationship: 'ally', notes: '' });
    }, charId);

    await clickNavItem(page, 'party', 'party');
    await page.locator('#partySubTabBar').waitFor({ state: 'visible', timeout: NAV_TIMEOUT });
    await page.locator('#partySubTabBar button:has-text("Graph")').click();
    await waitLoadingDone(page);

    const svgTexts = page.locator('#partyGraphSvg svg text');
    await expect(svgTexts.first()).toBeVisible({ timeout: NAV_TIMEOUT });
    const count = await svgTexts.count();
    expect(count).toBeGreaterThanOrEqual(1);
  });
});
