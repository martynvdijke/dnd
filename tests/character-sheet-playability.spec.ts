import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT, clickNavItem, waitLoadingDone } from './helpers.js';

async function openSheet(page: import('@playwright/test').Page, name: string): Promise<void> {
  await clickNavItem(page, 'characters', 'characters');
  await waitLoadingDone(page);
  await page.locator('.character-card, .char-card, [onclick*="showSheet"]', { hasText: name }).first().click();
  await waitLoadingDone(page);
}

test.describe('Character sheet playability', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('roll mode toggle applies to the next skill check', async ({ page }) => {
    const name = `RollMode-${Date.now()}`;
    await page.evaluate(async (charName) => {
      await window.api('POST', '/api/characters', {
        name: charName, race: 'Human', class: 'Fighter', level: 1,
        hp_max: 20, hp_current: 20, ac: 15, proficiency_bonus: 2,
        str_mod: 3, dex_mod: 1, con_mod: 2, int_mod: -1, wis_mod: 0, cha_mod: 0,
      });
    }, name);
    await openSheet(page, name);

    const advantage = page.locator('[data-testid="roll-mode-advantage"]');
    await expect(advantage).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(advantage).toHaveAttribute('aria-pressed', 'false');
    await advantage.click();
    await expect(advantage).toHaveAttribute('aria-pressed', 'true');

    const reqPromise = page.waitForRequest((r) => r.url().includes('/api/roll/check'));
    await page.locator('.skill-row', { hasText: 'Stealth' }).first().click();
    const req = await reqPromise;
    expect(req.postDataJSON()?.advantage).toBe('advantage');

    // Clicking the active mode toggles back to normal.
    await advantage.click();
    await expect(advantage).toHaveAttribute('aria-pressed', 'false');
    await expect(page.locator('[data-testid="roll-mode-normal"]')).toHaveAttribute('aria-pressed', 'true');
    await expect(page.locator('[data-testid="roll-mode-disadvantage"]')).toHaveAttribute('aria-pressed', 'false');
  });

  test('combat tab shows proficient saves, weapon attacks and configurable hit dice', async ({ page }) => {
    const name = `CombatTab-${Date.now()}`;
    const weapon = `Longsword-${Date.now()}`;
    await page.evaluate(async (opts) => {
      const c = await window.api('POST', '/api/characters', {
        name: opts.name, race: 'Human', class: 'Fighter', level: 1,
        hp_max: 20, hp_current: 20, ac: 15, proficiency_bonus: 2,
        str: 16, dex: 12, con: 14, int: 8, wis: 10, cha: 10,
        hit_dice: '3d8', hit_dice_current: 3,
      });
      await window.api('POST', `/api/characters/${c.id}/inventory`, {
        name: opts.weapon, category: 'weapon', damage_dice: '1d8', damage_type: 'slashing', is_equipped: true, quantity: 1,
      });
    }, { name, weapon });
    await openSheet(page, name);
    await page.locator('#tab-combat').click();

    // Fighter proficiency: STR/CON get the bonus, DEX does not.
    await expect(page.locator('[data-testid="saving-throws"]')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('[data-save="str"]')).toContainText('+5');
    await expect(page.locator('[data-save="con"]')).toContainText('+4');
    await expect(page.locator('[data-save="dex"]')).toContainText('+1');

    // Equipped weapon appears with one-click attack/damage.
    const quick = page.locator('[data-testid="combat-quick-attacks"]');
    await expect(quick).toContainText(weapon);
    await expect(quick.locator('[data-testid="weapon-attack-btn"]')).toBeVisible();
    await expect(quick.locator('[data-testid="weapon-damage-btn"]')).toBeVisible();
    await quick.locator('[data-testid="weapon-attack-btn"]').click();
    await expect(page.locator('#toastContainer')).toContainText('attack', { timeout: NAV_TIMEOUT });

    // Short rest submits the chosen hit dice count.
    const diceInput = page.locator('[data-testid="short-rest-hit-dice-input"]');
    await expect(diceInput).toBeVisible();
    await diceInput.fill('2');
    const restPromise = page.waitForRequest((r) => r.url().includes('/rest') && r.method() === 'POST');
    await page.locator('#combatSection button', { hasText: 'Short Rest' }).click();
    const restReq = await restPromise;
    expect(restReq.postDataJSON()?.hit_dice_count).toBe(2);
  });

  test('sheet tab filters hide non-play tabs', async ({ page }) => {
    const name = `TabFilter-${Date.now()}`;
    await page.evaluate(async (charName) => {
      await window.api('POST', '/api/characters', { name: charName, race: 'Human', class: 'Fighter', level: 1, hp_max: 10, hp_current: 10, ac: 15 });
    }, name);
    await openSheet(page, name);

    await expect(page.locator('#tab-crafting')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.locator('[data-filter="play"]').click();
    await expect(page.locator('#tab-crafting')).toBeHidden();
    await expect(page.locator('#tab-combat')).toBeVisible();
    await page.locator('[data-filter="all"]').click();
    await expect(page.locator('#tab-crafting')).toBeVisible();
  });
});
