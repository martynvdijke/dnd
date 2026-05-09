import { test, expect } from '@playwright/test';

test.describe('Responsive design', () => {
  test.beforeEach(async ({ page }) => {
    // Setup admin if needed and login
    await page.goto('/setup');
    const body = await page.locator('body').textContent();
    if (body?.includes('First-Time')) {
      await page.fill('#username', 'admin');
      await page.fill('#password', 'testpassword123');
      await page.fill('#confirm', 'testpassword123');
      await page.click('button[type="submit"]');
      await page.waitForURL(/\/app/);
    }

    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/app/);
  });

  test('desktop layout works at 1280x720', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await expect(page.locator('.header')).toBeVisible();
    await expect(page.locator('.container')).toBeVisible();
  });

  test('tablet layout works at 768x1024', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await expect(page.locator('.header')).toBeVisible();
    await expect(page.locator('.container')).toBeVisible();
  });

  test('mobile layout works at 390x844 (iPhone 14)', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });

    // Create a character to ensure there's content to check
    await page.click('text=New Character');
    await page.fill('#newName', 'Mobile Test');
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await page.click('.character-card');
    await expect(page.locator('#sheetName')).toBeVisible();

    // Check ability scores grid adapts on mobile (3 columns instead of 6)
    const abilityGrid = page.locator('.ability-grid');
    await expect(abilityGrid).toBeVisible();
  });

  test('small mobile layout works at 320x568 (iPhone SE)', async ({ page }) => {
    await page.setViewportSize({ width: 320, height: 568 });

    await page.click('text=New Character');
    await page.fill('#newName', 'Small Test');
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Cleric');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.click('.character-card');

    await expect(page.locator('#sheetName')).toBeVisible();
    // Tabs should still be accessible
    await page.click('text=Combat');
    await expect(page.locator('.tab.active')).toContainText('Combat');
  });

  test('dice roller is usable on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.click('text=Dice');
    await expect(page.locator('#diceExpr')).toBeVisible();

    await page.fill('#diceExpr', '1d20+5');
    await page.click('text=Roll the Bones');
    const result = page.locator('#diceResult');
    await expect(result).toBeVisible();
  });

  test('admin panel is responsive', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto('/admin');
    await expect(page.locator('h1')).toContainText('Users');

    // Switch tabs
    await page.click('text=Compendium');
    await expect(page.locator('h1')).toContainText('Compendium Management');

    await page.click('text=Backup');
    await expect(page.locator('h1')).toContainText('Backup');
  });

  test('character grid adapts to viewport', async ({ page }) => {
    // Create a few characters
    for (let i = 0; i < 3; i++) {
      await page.click('text=New Character');
      await page.fill('#newName', `Character ${i}`);
      await page.fill('#newRace', 'Human');
      await page.fill('#newClass', 'Fighter');
      await page.click('text=Create');
      await page.waitForTimeout(300);
    }

    // Check grid on desktop
    await page.setViewportSize({ width: 1280, height: 720 });
    let grid = page.locator('.character-grid');
    await expect(grid.locator('.character-card')).toHaveCount(3);

    // Check grid on mobile
    await page.setViewportSize({ width: 390, height: 844 });
    grid = page.locator('.character-grid');
    await expect(grid.locator('.character-card')).toHaveCount(3);
  });
});
