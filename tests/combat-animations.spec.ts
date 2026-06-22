import { test, expect, type Page } from './fixtures.js';
import { ensureNavOpen, waitLoadingDone, clickNavItem, login, isMobile } from './helpers.js';

async function createCharacter(page: Page, name: string) {
  return page.evaluate(async (n) => {
    return (window as any).api('POST', '/api/characters', {
      name: n, race: 'Human', class: 'Fighter', level: 3,
      str: 16, dex: 14, con: 15, int_: 10, wis: 12, cha: 8,
      hp_max: 30, hp_current: 30, ac: 18, speed: 30,
    });
  }, name);
}

async function openCombat(page: Page) {
  if (await isMobile(page)) {
    await page.click('#moreTabBtn');
    await page.waitForTimeout(300);
    await page.locator('#bottom-sheet-more-nav button').filter({ hasText: 'Combat' }).click();
  } else {
    await ensureNavOpen(page);
    await page.locator('#appSidebar button[data-nav="combat"]').click();
  }
  await page.waitForTimeout(500);
}

test.describe('Combat animations', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('HP damage shows floating number on character sheet', async ({ page }) => {
    // Create a character with full HP
    const name = `AnimChar-${Date.now()}`;
    await createCharacter(page, name);

    // Navigate to character list and open the sheet
    await clickNavItem(page, 'Characters', 'characters');
    await waitLoadingDone(page);
    await page.locator('.character-card, .char-card, [onclick*="showSheet"]', { hasText: name }).first().click();
    await waitLoadingDone(page);

    // Switch to the Combat tab (damage input is inside #combatSection, hidden by default)
    await page.click('#tabBar button:has-text("Combat")');
    await expect(page.locator('#combatSection')).toBeVisible({ timeout: 5000 });

    // Apply damage via the damage input
    const dmgInput = page.locator('#dmgInput');
    await expect(dmgInput).toBeVisible({ timeout: 5000 });
    await dmgInput.fill('5');
    await page.locator('button[onclick*="applyDamage"]').click();

    // Wait for the floating damage number to appear
    const floatEl = page.locator('.damage-float');
    await expect(floatEl).toBeVisible({ timeout: 5000 });
    await expect(floatEl).toContainText('-5');
  });

  test('combat turn transition marks active row', async ({ page }) => {
    await openCombat(page);
    await expect(page.locator('#combatTrackerView')).toBeVisible({ timeout: 5000 });

    // Add a combatant so there's a row to transition
    const uniqueEnemy = `Goblin-${Date.now()}`;
    await page.evaluate(() => {
      const fab = document.getElementById('fabBtn');
      if (fab) fab.style.display = 'none';
    });
    await page.locator('#combatTrackerContent button:has-text("Add")').first().click();
    await page.fill('#ceName', uniqueEnemy);
    await page.selectOption('#ceType', 'monster');
    await page.fill('#ceHPMax', '27');
    await page.click('.modal button:has-text("Add")');
    // Wait for the combatant to appear
    await expect(page.locator('#combatTrackerContent')).toContainText(uniqueEnemy, { timeout: 5000 });

    // Add a second combatant so Next Turn has somewhere to go
    const secondEnemy = `Orc-${Date.now()}`;
    await page.locator('#combatTrackerContent button:has-text("Add")').first().click();
    await page.fill('#ceName', secondEnemy);
    await page.selectOption('#ceType', 'monster');
    await page.fill('#ceHPMax', '30');
    await page.click('.modal button:has-text("Add")');
    await expect(page.locator('#combatTrackerContent')).toContainText(secondEnemy, { timeout: 5000 });

    // Click "Next Turn"
    const nextTurnBtn = page.locator('button:has-text("Next Turn")');
    await nextTurnBtn.click();
    await waitLoadingDone(page);

    // After turn change, the active row should be marked with combatant-row-active
    // (all rows have table-active, but only the active one gets combatant-row-active from animateTurnChange)
    const activeRow = page.locator('tr.combatant-row-active');
    await expect(activeRow).toBeVisible({ timeout: 5000 });
  });
});
