import { test, expect, type Page } from './fixtures.js';
import { ensureNavOpen, waitModalClosed, isMobile, login } from './helpers.js';

async function openCombat(page: Page) {
  if (await isMobile(page)) {
    await page.click('#moreTabBtn');
    await page.waitForTimeout(300);
    // Bottom sheet buttons are dynamically created without IDs, click by text
    await page.locator('#bottom-sheet-more-nav button').filter({ hasText: 'Combat' }).click();
  } else {
    await ensureNavOpen(page);
    await page.locator('#appSidebar button[data-nav="combat"]').click();
  }
  await page.waitForTimeout(500);
}

test.describe.serial('Combat Tracker', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.waitForTimeout(300);
  });

  test('combat nav item visible for admin', async ({ page }) => {
    await openCombat(page);
    await expect(page.locator('#combatTrackerView')).toBeVisible();
  });

  test('opens combat tracker view', async ({ page }) => {
    await openCombat(page);
    await expect(page.locator('#combatTrackerView')).toBeVisible();
    await expect(page.locator('#combatTrackerContent')).toBeVisible();
  });

  test('adds a combatant and shows in tracker', async ({ page }) => {
    await openCombat(page);

    const uniqueEnemy = `Goblin-${Date.now()}`;
    // Hide FAB on mobile to avoid interception
    await page.evaluate(() => {
      const fab = document.getElementById('fabBtn');
      if (fab) fab.style.display = 'none';
    });
    await page.locator('#combatTrackerContent button:has-text("Add")').first().click();
    await page.fill('#ceName', uniqueEnemy);
    await page.selectOption('#ceType', 'monster');
    await page.fill('#ceAC', '15');
    await page.fill('#ceHPMax', '27');
    await page.click('.modal button:has-text("Add")');
    await waitModalClosed(page);

    await expect(page.locator('#combatTrackerContent')).toContainText(uniqueEnemy);
  });

  test('rolls initiative for combatants', async ({ page }) => {
    await openCombat(page);

    const uniqueEnemy = `Goblin-${Date.now()}`;
    await page.evaluate(() => {
      const fab = document.getElementById('fabBtn');
      if (fab) fab.style.display = 'none';
    });
    await page.locator('#combatTrackerContent button:has-text("Add")').first().click();
    await page.fill('#ceName', uniqueEnemy);
    await page.selectOption('#ceType', 'monster');
    await page.fill('#ceAC', '15');
    await page.fill('#ceHPMax', '27');
    await page.click('.modal button:has-text("Add")');
    await waitModalClosed(page);

    await page.evaluate(() => (window as any).rollAllInitiative());
    await page.waitForTimeout(500);

    const initCells = await page.locator('#combatTrackerTable td:nth-child(3)').allTextContents();
    const validRolls = initCells.filter(v => v.trim() !== '-' && !isNaN(parseInt(v)));
    expect(validRolls.length).toBeGreaterThanOrEqual(1);
  });

  test('applies damage to combatant', async ({ page }) => {
    await openCombat(page);

    const uniqueEnemy = `Goblin-${Date.now()}`;
    await page.evaluate(() => {
      const fab = document.getElementById('fabBtn');
      if (fab) fab.style.display = 'none';
    });
    await page.locator('#combatTrackerContent button:has-text("Add")').first().click();
    await page.fill('#ceName', uniqueEnemy);
    await page.selectOption('#ceType', 'monster');
    await page.fill('#ceAC', '15');
    await page.fill('#ceHPMax', '27');
    await page.click('.modal button:has-text("Add")');
    await waitModalClosed(page);

    // Find the damage input for this specific combatant by locating their row
    const row = page.locator('#combatTrackerContent tr', { hasText: uniqueEnemy });
    const dmgInput = row.locator('input[type="number"]');
    await dmgInput.fill('10');
    const dmgBtn = row.locator('button[onclick*="combatTrackerDamage"]');
    await dmgBtn.click();
    await page.waitForTimeout(500);

    await expect(page.locator('#combatTrackerContent')).toContainText('17/27');
  });
});
