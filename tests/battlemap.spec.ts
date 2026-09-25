import { test, expect } from './fixtures.js';
import { NAV_TIMEOUT, login, waitLoadingDone } from './helpers.js';

test.describe('Battlemap', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await waitLoadingDone(page);
  });

  test('syncs combatants onto the battlemap as tokens', async ({ page }) => {
    const ids = await page.evaluate(async () => {
      const c = await (window as any).api('POST', '/api/campaigns', {
        name: 'Map Camp', party_name: 'Testers', description: '', dm_notes: '',
      });
      const e = await (window as any).api('POST', '/api/combat', {
        campaign_id: c.id, name: 'Goblin', type: 'monster',
        initiative_roll: 12, initiative_mod: 2, hp_max: 30, hp_current: 30, ac: 14,
      });
      return { cid: c.id as number, eid: e.id as number };
    });

    await page.evaluate((cid) => (window as any).showBattlemap(cid), ids.cid);
    await expect(page.locator('[data-testid="battlemap-board"]')).toBeVisible({ timeout: NAV_TIMEOUT });

    await page.evaluate((cid) => (window as any).battlemapSync(cid), ids.cid);
    await expect(page.locator('.bm-token[data-name="Goblin"]')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#battlemapView')).toContainText('Goblin');
  });
});
