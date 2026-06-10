import { test, expect } from '@playwright/test';
import { ensureNavOpen, waitLoadingDone, clickNavItem } from './helpers.js';

test.describe('Dice rolling', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('dice roller works', async ({ page }) => {
    await clickNavItem(page, 'Dice', 'dice');
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
    await clickNavItem(page, 'Dice', 'dice');
    const input = page.locator('#diceExpr');
    await input.fill('1d20');
    await page.click('text=Roll the Bones');
    await expect(page.locator('#diceResult')).toBeVisible({ timeout: 10000 });

    const history = page.locator('.dice-history-item');
    await expect(history.first()).toBeVisible({ timeout: 10000 });
  });

  test('rpg notation: keep highest 3 of 4d6', async ({ page }) => {
    await clickNavItem(page, 'Dice', 'dice');
    const input = page.locator('#diceExpr');
    await input.fill('4d6kh3');
    await page.click('text=Roll the Bones');
    const result = page.locator('#diceResult');
    await expect(result).toBeVisible({ timeout: 10000 });
    await expect(result).toContainText('4d6kh3');
    // The 3D dice container should have dice elements
    const diceContainer = page.locator('#dice3dContainer');
    await expect(diceContainer.locator('.dice-3d-die')).toHaveCount(4, { timeout: 5000 });
  });

  test('rpg notation: exploding dice', async ({ page }) => {
    await clickNavItem(page, 'Dice', 'dice');
    const input = page.locator('#diceExpr');
    await input.fill('1d6!');
    await page.click('text=Roll the Bones');
    const result = page.locator('#diceResult');
    await expect(result).toBeVisible({ timeout: 10000 });
    await expect(result).toContainText('1d6!');
  });

  test('rpg notation: percentile d100', async ({ page }) => {
    await clickNavItem(page, 'Dice', 'dice');
    const input = page.locator('#diceExpr');
    await input.fill('1d100');
    await page.click('text=Roll the Bones');
    const result = page.locator('#diceResult');
    await expect(result).toBeVisible({ timeout: 10000 });
    await expect(result).toContainText('1d100');
    // d100 should render as a 3D die (not 2D fallback)
    const diceContainer = page.locator('#dice3dContainer');
    await expect(diceContainer.locator('.dice-3d-die.d100').first()).toBeVisible({ timeout: 5000 });
  });

  test('3D dice renders for all standard polyhedral types', async ({ page }) => {
    await clickNavItem(page, 'Dice', 'dice');
    for (const die of ['d4', 'd6', 'd8', 'd10', 'd12', 'd20']) {
      const input = page.locator('#diceExpr');
      await input.fill('1' + die);
      await page.click('text=Roll the Bones');
      const result = page.locator('#diceResult');
      await expect(result).toBeVisible({ timeout: 10000 });

      // Verify the 3D die element rendered for this type
      // 3D dice render async and may not be visible immediately
      await page.waitForFunction((d) => {
        const el = document.querySelector(`#dice3dContainer .dice-3d-die.${d}`);
        return el && el.getBoundingClientRect().width > 0 && el.getBoundingClientRect().height > 0;
      }, die, { timeout: 10000 });
    }
  });

  test('advantage quick-preset works', async ({ page }) => {
    await clickNavItem(page, 'Dice', 'dice');
    const input = page.locator('#diceExpr');
    await input.fill('1d20');
    // Click the "Advantage" preset button
    await page.click('button:has-text("Advantage")');
    const result = page.locator('#diceResult');
    await expect(result).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#diceExpr')).toHaveValue('2d20kh1');
  });

  test('invalid expression shows error', async ({ page }) => {
    await clickNavItem(page, 'Dice', 'dice');
    const input = page.locator('#diceExpr');
    await input.fill('not-a-dice-roll');
    await page.click('text=Roll the Bones');
    // Should show a toast error
    const toast = page.locator('.toast, .toast-container, .alert-danger').first();
    await expect(toast).toBeVisible({ timeout: 5000 });
  });

  test('placeholder text indicates RPG notation', async ({ page }) => {
    await clickNavItem(page, 'Dice', 'dice');
    const input = page.locator('#diceExpr');
    await expect(input).toHaveAttribute('placeholder', /4d6kh3|rpg|notation|kh|!/i);
  });
});
