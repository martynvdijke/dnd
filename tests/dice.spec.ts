import { test, expect } from '@playwright/test';

async function ensureNavOpen(page) {
  const toggler = page.locator('.navbar-toggler');
  if (await toggler.isVisible()) {
    await toggler.click();
    await page.waitForTimeout(300);
  }
}

test.describe('Dice rolling', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
  });

  test('dice roller works', async ({ page }) => {
    await ensureNavOpen(page);
    await page.click('a:has-text("Dice")');
    await expect(page.locator('#diceView h1')).toContainText('Dice Roller');
    await expect(page.locator('#diceExpr')).toBeVisible();

    const input = page.locator('#diceExpr');
    await input.fill('2d6+3');
    await page.click('text=Roll the Bones');

    const result = page.locator('#diceResult');
    await expect(result).toBeVisible({ timeout: 10000 });
    await expect(result).toContainText('2d6+3');
  });

  test('saves dice roll history', async ({ page }) => {
    await ensureNavOpen(page);
    await page.click('a:has-text("Dice")');
    const input = page.locator('#diceExpr');
    await input.fill('1d20');
    await page.click('text=Roll the Bones');
    await expect(page.locator('#diceResult')).toBeVisible({ timeout: 10000 });

    const history = page.locator('.dice-history-item');
    await expect(history.first()).toBeVisible({ timeout: 10000 });
  });
});
