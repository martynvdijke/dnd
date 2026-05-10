import { test, expect } from '@playwright/test';

const uniqueName = () => `Resp-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('Responsive design', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
  });

  test('desktop layout works at 1280x720', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await expect(page.locator('.navbar')).toBeVisible();
    await expect(page.locator('.container').first()).toBeVisible();
  });

  test('tablet layout works at 768x1024', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await expect(page.locator('.navbar')).toBeVisible();
    await expect(page.locator('.container').first()).toBeVisible();
  });

  test('mobile layout works at 390x844 (iPhone 14)', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });

    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await expect(page.locator('#sheetName')).toBeVisible();

    await expect(page.locator('#statsSection .ability-box').first()).toBeVisible();
  });

  test('small mobile layout works at 320x568 (iPhone SE)', async ({ page }) => {
    await page.setViewportSize({ width: 320, height: 568 });

    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Cleric');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.locator('.character-card').filter({ hasText: name }).click();

    await expect(page.locator('#sheetName')).toBeVisible();
    await page.click('text=Combat');
    await expect(page.locator('#combatSection')).toBeVisible();
  });

  test('dice roller is usable on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.click('a:has-text("Dice")');
    await page.waitForTimeout(200);
    await expect(page.locator('#diceExpr')).toBeVisible();

    await page.fill('#diceExpr', '1d20+5');
    await page.click('text=Roll the Bones');
    const result = page.locator('#diceResult');
    await expect(result).toBeVisible();
  });

  test('admin panel is responsive', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto('/admin', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('#adminUsers .card-header')).toContainText('Users');

    await page.click('#adminTabs button:has-text("Compendium")');
    await expect(page.locator('#adminCompendium .card-header')).toContainText('Compendium Management');

    await page.click('#adminTabs button:has-text("Backup")');
    await expect(page.locator('#adminBackup .card-header').first()).toContainText('Backup Settings');
  });

  test('character grid adapts to viewport', async ({ page }) => {
    const prefix = uniqueName();
    for (let i = 0; i < 3; i++) {
      await page.click('text=New Character');
      await page.fill('#newName', `${prefix}-${i}`);
      await page.fill('#newRace', 'Human');
      await page.fill('#newClass', 'Fighter');
      await page.click('text=Create');
      await page.waitForTimeout(300);
    }

    await page.setViewportSize({ width: 1280, height: 720 });
    await expect(page.locator('#charGrid')).toContainText(`${prefix}-0`);

    await page.setViewportSize({ width: 390, height: 844 });
    await expect(page.locator('#charGrid')).toContainText(`${prefix}-0`);
  });
});
