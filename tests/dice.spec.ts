import { test, expect } from './fixtures.js';
import { ensureNavOpen, waitLoadingDone, clickNavItem, login, NAV_TIMEOUT } from './helpers.js';

test.describe('Dice rolling', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('dice roller works', async ({ page }) => {
    await clickNavItem(page, 'dice', 'dice');
    await expect(page.locator('#diceView h1')).toContainText('Dice Roller');
    await expect(page.locator('#diceExpr')).toBeVisible();

    const input = page.locator('#diceExpr');
    await input.fill('2d6+3');
    await page.click('text=Roll the Bones');

    const result = page.locator('#diceResult');
    await expect(result).toBeVisible({ timeout: 10000 });
    await expect(result).toContainText('2d6+3');
  });

  test('dice roller renders on direct #/dice deep link and reload', async ({ page }) => {
    // Router navigation (deep link / reload / back-forward) must render the
    // roller too — the dice view has no static content in app.html.
    await page.goto('/#/dice');
    await expect(page.locator('#diceView h1')).toContainText('Dice Roller');
    await expect(page.locator('#diceExpr')).toBeVisible();
    await expect(page.locator('#dice3dContainer')).toBeVisible();

    // Reloading while on #/dice (initial-hash navigation) must re-render.
    await page.reload();
    await expect(page.locator('#diceExpr')).toBeVisible();

    const input = page.locator('#diceExpr');
    await input.fill('1d20');
    await page.click('text=Roll the Bones');
    await expect(page.locator('#diceResult')).toBeVisible({ timeout: 10000 });
  });

  test('saves dice roll history', async ({ page }) => {
    await clickNavItem(page, 'dice', 'dice');
    const input = page.locator('#diceExpr');
    await input.fill('1d20');
    await page.click('text=Roll the Bones');
    await expect(page.locator('#diceResult')).toBeVisible({ timeout: 10000 });

    const history = page.locator('.dice-history-item');
    await expect(history.first()).toBeVisible({ timeout: 10000 });
  });

  test('rpg notation: keep highest 3 of 4d6', async ({ page }) => {
    await clickNavItem(page, 'dice', 'dice');
    const input = page.locator('#diceExpr');
    await input.fill('4d6kh3');
    await page.click('text=Roll the Bones');
    const result = page.locator('#diceResult');
    await expect(result).toBeVisible({ timeout: 10000 });
    await expect(result).toContainText('4d6kh3');
    // The dice container should have die elements
    const diceContainer = page.locator('#dice3dContainer');
    await expect(diceContainer.locator('.die')).toHaveCount(4, { timeout: NAV_TIMEOUT });
  });

  test('rpg notation: exploding dice', async ({ page }) => {
    await clickNavItem(page, 'dice', 'dice');
    const input = page.locator('#diceExpr');
    await input.fill('1d6!');
    await page.click('text=Roll the Bones');
    const result = page.locator('#diceResult');
    await expect(result).toBeVisible({ timeout: 10000 });
    await expect(result).toContainText('1d6!');
  });

  test('rpg notation: percentile d100', async ({ page }) => {
    await clickNavItem(page, 'dice', 'dice');
    const input = page.locator('#diceExpr');
    await input.fill('1d100');
    await page.click('text=Roll the Bones');
    const result = page.locator('#diceResult');
    await expect(result).toBeVisible({ timeout: 10000 });
    await expect(result).toContainText('1d100');
    // d100 should render as a die
    const diceContainer = page.locator('#dice3dContainer');
    await expect(diceContainer.locator('.die.d100').first()).toBeVisible({ timeout: NAV_TIMEOUT });
  });

  test('dice renders for all standard polyhedral types', async ({ page }) => {
    await clickNavItem(page, 'dice', 'dice');
    for (const die of ['d4', 'd6', 'd8', 'd10', 'd12', 'd20']) {
      const input = page.locator('#diceExpr');
      await input.fill('1' + die);
      await page.click('text=Roll the Bones');
      const result = page.locator('#diceResult');
      await expect(result).toBeVisible({ timeout: 10000 });

      // Verify the die element rendered for this type
      await expect(page.locator(`#dice3dContainer .die.${die}`)).toBeAttached({ timeout: 10000 });
    }
  });

  test('advantage quick-preset works', async ({ page }) => {
    await clickNavItem(page, 'dice', 'dice');
    const input = page.locator('#diceExpr');
    await input.fill('1d20');
    // Click the "Advantage" preset button
    await page.click('button:has-text("Advantage")');
    const result = page.locator('#diceResult');
    await expect(result).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#diceExpr')).toHaveValue('2d20kh1');
  });

  test('invalid expression shows error', async ({ page }) => {
    await clickNavItem(page, 'dice', 'dice');
    const input = page.locator('#diceExpr');
    await input.fill('not-a-dice-roll');
    await page.click('text=Roll the Bones');
    // Should show a toast error
    const toast = page.locator('.toast, .toast-container, .alert-danger').first();
    await expect(toast).toBeVisible({ timeout: NAV_TIMEOUT });
  });

  test('placeholder text indicates RPG notation', async ({ page }) => {
    await clickNavItem(page, 'dice', 'dice');
    const input = page.locator('#diceExpr');
    await expect(input).toHaveAttribute('placeholder', /4d6kh3|rpg|notation|kh|!/i);
  });

  test('crit celebration appears on nat 20', async ({ page }) => {
    await clickNavItem(page, 'dice', 'dice');
    // Mock the roll API to always return a nat 20 — deterministic, no RNG
    await page.route('**/api/roll', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          total: 20,
          expression: '1d20',
          text: '1d20 (20) = 20',
          breakdown: [{ die: 'd20', rolls: [{ value: 20, useInTotal: true, modifierFlags: '' }], total: 20 }],
        }),
      });
    });
    const input = page.locator('#diceExpr');
    await input.fill('1d20');
    await page.click('text=Roll the Bones');
    // Settle animation takes 900ms, crit overlay appears right after
    const critOverlay = page.locator('.crit-overlay');
    await expect(critOverlay).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('.crit-text')).toContainText(/CRITICAL/i);
  });
});
