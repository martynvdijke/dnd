import { test, expect } from './fixtures.js';
import { NAV_TIMEOUT, login, waitLoadingDone } from './helpers.js';

test.describe('VTT vision', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await waitLoadingDone(page);
  });

  test('draws walls, saves them, and shows a vision mask', async ({ page }) => {
    const ids = await page.evaluate(async () => {
      const c = await (window as any).api('POST', '/api/campaigns', {
        name: 'Vision Camp', party_name: 'Testers', description: '', dm_notes: '',
      });
      await (window as any).api('POST', `/api/campaigns/${c.id}/maps`, {
        name: 'Crypt', width: 1000, height: 800, grid_size: 50,
      });
      await (window as any).api('POST', `/api/campaigns/${c.id}/battlemap/tokens`, {
        name: 'Scout', aura_radius: 2, vision_radius: 6,
      });
      return { cid: c.id as number };
    });

    await page.evaluate((cid) => (window as any).showBattlemap(cid), ids.cid);
    await expect(page.locator('[data-testid="battlemap-board"]')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('[data-testid="bm-overlay"]')).toBeVisible();

    // Draw a wall: toggle wall mode, click two points, save.
    await page.locator('[data-testid="bm-walls-toggle"]').click();
    await expect(page.locator('[data-testid="bm-walls-save"]')).toBeVisible();
    await expect(page.locator('[data-testid="bm-walls-clear"]')).toBeVisible();
    const board = page.locator('[data-testid="battlemap-board"]');
    await board.click({ position: { x: 100, y: 100 } });
    await board.click({ position: { x: 100, y: 300 } });
    await page.locator('[data-testid="bm-walls-save"]').click();

    const walls = await page.evaluate(async (cid) => {
      const maps = await (window as any).api('GET', `/api/campaigns/${cid}/maps`);
      const map = Array.isArray(maps) ? maps[0] : maps.maps?.[0];
      const res = await (window as any).api('GET', `/api/maps/${map.id}/walls`);
      return res.walls as unknown[];
    }, ids.cid);
    expect(walls.length).toBeGreaterThan(0);

    // Toggle vision and pick the token as the viewer.
    await page.locator('[data-testid="bm-vision-toggle"]').click();
    const viewer = page.locator('[data-testid="bm-vision-token"]');
    await expect(viewer).toBeVisible();
    await viewer.selectOption({ index: 1 });
    await expect(page.locator('.bm-vision-mask')).toBeVisible();
  });
});
