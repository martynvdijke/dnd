import { test, expect } from '@playwright/test';

test.describe('Combat Tracker', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded', timeout: 10000 });
    await page.waitForTimeout(800);
  });

  test('combat nav item visible for admin', async ({ page }) => {
    await expect(page.locator('#combatNavItem')).toBeVisible();
  });

  test('opens combat tracker view', async ({ page }) => {
    await page.click('#combatNavItem a');
    await page.waitForTimeout(500);
    await expect(page.locator('#combatTrackerView')).toBeVisible();
    await expect(page.locator('#combatTrackerContent')).toContainText('No Combatants');
  });

  test('adds a combatant and shows in tracker', async ({ page }) => {
    await page.click('#combatNavItem a');
    await page.waitForTimeout(500);

    await page.click('button:has-text("Add")');
    await page.fill('#ceName', 'Goblin Archer');
    await page.selectOption('#ceType', 'monster');
    await page.fill('#ceAC', '15');
    await page.fill('#ceHPMax', '27');
    await page.click('.modal button:has-text("Add")');
    await page.waitForTimeout(500);

    await expect(page.locator('#combatTrackerContent')).toContainText('Goblin Archer');
  });

  test('rolls initiative for combatants', async ({ page }) => {
    await page.click('#combatNavItem a');
    await page.waitForTimeout(500);

    await page.click('button:has-text("Add")');
    await page.fill('#ceName', 'Goblin Archer');
    await page.selectOption('#ceType', 'monster');
    await page.fill('#ceAC', '15');
    await page.fill('#ceHPMax', '27');
    await page.click('.modal button:has-text("Add")');
    await page.waitForTimeout(500);

    await page.click('button:has-text("Roll Init")');
    await page.waitForTimeout(500);

    const initCells = await page.locator('#combatTrackerTable td:nth-child(3)').allTextContents();
    const validRolls = initCells.filter(v => v.trim() !== '-' && !isNaN(parseInt(v)));
    expect(validRolls.length).toBeGreaterThanOrEqual(1);
  });

  test('applies damage to combatant', async ({ page }) => {
    await page.click('#combatNavItem a');
    await page.waitForTimeout(500);

    await page.click('button:has-text("Add")');
    await page.fill('#ceName', 'Goblin Archer');
    await page.selectOption('#ceType', 'monster');
    await page.fill('#ceAC', '15');
    await page.fill('#ceHPMax', '27');
    await page.click('.modal button:has-text("Add")');
    await page.waitForTimeout(500);

    const dmgInput = page.locator('#qdamage-1');
    await dmgInput.fill('10');
    await page.locator('button[onclick*="combatTrackerDamage(1)"]').click();
    await page.waitForTimeout(500);

    await expect(page.locator('#combatTrackerContent')).toContainText('17/27');
  });
});
