import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT } from './helpers.js';

test.describe('Combat statblocks and encounter rewards', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('monster statblock expands inline in the combat tracker', async ({ page }) => {
    const seed = await page.evaluate(async () => {
      const suffix = Date.now();
      const camp = await window.api('POST', '/api/campaigns', { name: `SB-${suffix}`, description: 'statblock' });
      const entry = await window.api('POST', '/api/combat', {
        name: 'Goblin #1', type: 'monster', ac: 15, hp_max: 7, hp_current: 7, initiative_mod: 2, campaign_id: camp.id,
      });
      return { entry };
    });

    await page.evaluate(() => (window as any).showCombatTracker());
    await expect(page.locator('#combatTrackerContent')).toContainText('Goblin #1', { timeout: NAV_TIMEOUT });

    const row = page.locator('#ce-' + seed.entry.id);
    await row.locator('[data-testid="combat-statblock-toggle"]').click();
    await expect(page.locator('#statblock-' + seed.entry.id)).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#statblock-' + seed.entry.id)).toContainText(/CR|No compendium statblock/);
  });

  test('monster combatants can be selected as attackers', async ({ page }) => {
    const seed = await page.evaluate(async () => {
      const suffix = Date.now();
      const camp = await window.api('POST', '/api/campaigns', { name: `MA-${suffix}`, description: 'monster attacker' });
      const attacker = await window.api('POST', '/api/combat', {
        name: `Goblin Archer-${suffix}`, type: 'monster', ac: 13, hp_max: 7, hp_current: 7, initiative_mod: 2, campaign_id: camp.id,
      });
      const target = await window.api('POST', '/api/combat', {
        name: `Dummy-${suffix}`, type: 'monster', ac: 10, hp_max: 30, hp_current: 30, initiative_mod: 0, campaign_id: camp.id,
      });
      return { target };
    });

    await page.evaluate(() => (window as any).showCombatTracker());
    const row = page.locator('#ce-' + seed.target.id);
    await row.locator('[data-testid="combat-attack-btn"]').click();
    const attacker = page.locator('[data-testid="combat-attack-attacker"]');
    await expect(attacker).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(attacker.locator('option', { hasText: '(monster)' }).first()).toBeAttached();
    await attacker.selectOption({ index: await attacker.locator('option').count() - 1 });
    await page.locator('[data-testid="combat-attack-preview"]').click();
    await expect(page.locator('[data-testid="combat-attack-result"]')).toContainText(/vs AC/, { timeout: NAV_TIMEOUT });
  });

  test('end encounter awards XP and clears the tracker', async ({ page }) => {
    const seed = await page.evaluate(async () => {
      const suffix = Date.now();
      const camp = await window.api('POST', '/api/campaigns', { name: `EE-${suffix}`, description: 'rewards' });
      const char = await window.api('POST', '/api/characters', {
        name: `Reward-${suffix}`, campaign_ids: [camp.id], level: 1, hp_max: 10, hp_current: 10, ac: 12, xp: 0,
      });
      await window.api('POST', '/api/combat', {
        name: 'Goblin', type: 'monster', ac: 15, hp_max: 7, hp_current: 0, initiative_mod: 2, campaign_id: camp.id,
      });
      return { camp, char };
    });

    await page.evaluate(() => (window as any).showCombatTracker());
    await page.locator('[data-testid="combat-end-encounter-btn"]').click();
    await expect(page.locator('[data-testid="combat-end-xp-party"]')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('[data-testid="combat-end-xp-character"]')).toHaveCount(1);
    await expect(page.locator('[data-testid="combat-end-treasure"]')).toBeAttached();
    const xpOverride = page.locator('[data-testid="combat-end-xp-override"]');
    await expect(xpOverride).toBeVisible();
    await xpOverride.fill('200');
    await page.locator('[data-testid="combat-end-confirm"]').click();

    await expect(page.locator('#combatTrackerContent')).toContainText('No Combatants', { timeout: NAV_TIMEOUT });
    const xp = await page.evaluate(async (cid) => {
      const c = await window.api('GET', `/api/characters/${cid}`);
      return c.xp;
    }, seed.char.id);
    expect(xp).toBe(200);
  });
});
