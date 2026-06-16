import { test, expect } from './fixtures.js';
import { login, waitLoadingDone } from './helpers.js';

const uniqueName = () => `Resp-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function ensureNavOpen(page) {
  const toggler = page.locator('.navbar-toggler');
  if (await toggler.isVisible()) {
    await toggler.click();
    await page.waitForTimeout(300);
  }
}

async function waitModalClosed(page) {
  await page.waitForFunction(() => {
    const modal = document.getElementById('genericModal');
    return !modal || !modal.classList.contains('show');
  }, { timeout: 10000 }).catch(() => {});
}

test.describe('Responsive design', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('desktop layout works at 1280x720', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await expect(page.locator('.navbar')).toBeVisible();
    await expect(page.locator('.container').first()).toBeVisible();
  });

  test('mobile layout works at 767x1024', async ({ page }) => {
    await page.setViewportSize({ width: 767, height: 1024 });
    await waitLoadingDone(page);
    await expect(page.locator('.bottom-tab-bar')).toBeVisible();
    await expect(page.locator('#charGrid')).toBeVisible();
  });

  test('mobile layout works at 390x844 (iPhone 14)', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });

    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
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
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await expect(page.locator('#sheetName')).toBeVisible();
    await page.click('#tabBar button:has-text("Combat")');
    await expect(page.locator('#combatSection')).toBeVisible();
  });

  test('dice roller is usable on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.click('#bottomTabBar button[data-nav="dice"]');
    await expect(page.locator('#diceExpr')).toBeVisible({ timeout: 5000 });

    await page.fill('#diceExpr', '1d20+5');
    await page.click('text=Roll the Bones');
    const result = page.locator('#diceResult');
    await expect(result).toBeVisible({ timeout: 10000 });
  });

  test('admin panel is responsive', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(300);
    await page.goto('/admin', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('#adminUsers .card-header')).toContainText('Users');

    await page.click('#adminTabs button:has-text("Compendium")');
    await expect(page.locator('#adminUnifiedCompendium .card-header')).toContainText('Compendium');

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
      await waitModalClosed(page);
    }

    await page.setViewportSize({ width: 1280, height: 720 });
    await expect(page.locator('#charGrid')).toContainText(`${prefix}-0`);

    await page.setViewportSize({ width: 390, height: 844 });
    await expect(page.locator('#charGrid')).toContainText(`${prefix}-0`);
  });
});
