import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT } from './helpers.js';

test.describe('Party radar, group actions and loot', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('party view shows passives and supports group rolls and party rest', async ({ page }) => {
    const seed = await page.evaluate(async () => {
      const suffix = Date.now();
      const camp = await window.api('POST', '/api/campaigns', { name: `PR-${suffix}`, description: 'radar' });
      const names = [`Aria-${suffix}`, `Bram-${suffix}`];
      for (const n of names) {
        await window.api('POST', '/api/characters', {
          name: n, campaign_ids: [camp.id], level: 1, hp_max: 12, hp_current: 6, ac: 13,
          str: 10, dex: 14, con: 12, int: 10, wis: 16, cha: 8, proficiency_bonus: 2,
        });
      }
      return { camp, names };
    });

    await page.evaluate(() => (window as any).showParty());
    await expect(page.locator('[data-testid="party-member-passives"]').first()).toBeVisible({ timeout: NAV_TIMEOUT });

    const select = page.locator('[data-testid="party-group-roll-select"]').first();
    await expect(select).toBeVisible();
    await select.selectOption('skill:Perception');
    await page.locator('[data-testid="party-group-roll-btn"]').first().click();
    await expect(page.locator('#genericModal')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#genericModal')).toContainText(seed.names[0]);
    await page.evaluate(() => (window as any).hideModal());
    await expect(page.locator('#genericModal')).toBeHidden();

    await expect(page.locator('[data-testid="party-short-rest-btn"]').first()).toBeVisible();
    await expect(page.locator('[data-testid="party-long-rest-btn"]').first()).toBeAttached();
    const restPromise = page.waitForRequest((r) => r.url().includes('/rest') && r.method() === 'POST');
    await page.locator('[data-testid="party-short-rest-btn"]').first().click();
    await restPromise;
  });

  test('party loot can be claimed into a character inventory', async ({ page }) => {
    const seed = await page.evaluate(async () => {
      const suffix = Date.now();
      const camp = await window.api('POST', '/api/campaigns', { name: `PL-${suffix}`, description: 'loot' });
      const char = await window.api('POST', '/api/characters', { name: `Looter-${suffix}`, campaign_ids: [camp.id], level: 1, hp_max: 10, hp_current: 10, ac: 12 });
      await window.api('POST', `/api/campaigns/${camp.id}/party-items`, { name: 'Bag of Holding', quantity: 1, notes: 'Claim me' });
      return { camp, char };
    });

    await page.evaluate((cid) => (window as any).showPartyInventory(cid), seed.camp.id);
    await expect(page.locator('[data-testid="party-claim-character"]')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('[data-testid="party-item-claim-btn"]')).toBeVisible();
    await page.locator('[data-testid="party-item-claim-btn"]').first().click();

    await expect(page.locator('#partyInvContent')).toContainText('No party items yet', { timeout: NAV_TIMEOUT });
    const inv = await page.evaluate(async (cid) => {
      const c = await window.api('GET', `/api/characters/${cid}`);
      return c.inventory || [];
    }, seed.char.id);
    expect(inv.some((i: any) => i.name === 'Bag of Holding')).toBe(true);
  });

  test('DM tools can save generated content to the campaign', async ({ page }) => {
    await page.evaluate(async () => {
      (window as any).showDmTools();
      await (window as any).dmGenerateName();
    });

    const save = page.locator('[data-testid="dm-tools-save-btn"]');
    await expect(save).toBeVisible({ timeout: NAV_TIMEOUT });
    const generatedName = ((await page.locator('#dmToolsResult .fs-5').first().textContent()) || '').trim();
    expect(generatedName.length).toBeGreaterThan(0);
    await save.click();

    await expect.poll(async () => {
      return page.evaluate(async (name) => {
        const npcs = await window.api('GET', '/api/npcs');
        return (npcs || []).some((n: any) => n.name === name);
      }, generatedName);
    }, { timeout: NAV_TIMEOUT }).toBe(true);
  });
});
