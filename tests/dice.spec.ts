import { test, expect } from '@playwright/test';

test.describe('Dice rolling', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
  });

  test('dice roller works', async ({ page }) => {
    await page.click('text=Dice');
    await expect(page.locator('h1')).toContainText('Dice Roller');
    await expect(page.locator('#diceExpr')).toBeVisible();

    const input = page.locator('#diceExpr');
    await input.fill('2d6+3');
    await page.click('text=Roll the Bones');

    const result = page.locator('#diceResult');
    await expect(result).toBeVisible();
    const text = await result.textContent();
    expect(text).toContain('2d6+3');
  });

  test('saves dice roll history', async ({ page }) => {
    await page.click('text=Dice');
    const input = page.locator('#diceExpr');
    await input.fill('1d20');
    await page.click('text=Roll the Bones');
    await page.waitForTimeout(500);

    const history = page.locator('.dice-history-item');
    await expect(history.first()).toBeVisible();
  });
});
