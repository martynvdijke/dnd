import { test, expect } from './fixtures.js';
import { isMobile, clickNavItem, clickSecondaryNavItem, login, waitModalClosed } from './helpers.js';

test.describe('Full application smoke test', () => {
  test.beforeEach(async ({ page }) => {
    await page.context().clearCookies();
    const resp = await page.goto('/login', { waitUntil: 'domcontentloaded' });
    expect(resp?.status()).toBe(200);
  });

  test('all static assets load correctly', async ({ page }) => {
    const assets = [
      '/static/style.css',
      '/static/js/app.js',
      '/static/js/login.js',
    ];
    for (const asset of assets) {
      const resp = await page.goto(asset, { waitUntil: 'domcontentloaded' });
      expect(resp?.status()).toBe(200);
    }
  });

  test('login flow works', async ({ page }) => {
    await login(page);
    await expect(page.locator('#userName')).toContainText('admin', { timeout: 5000 });
    if (!(await isMobile(page))) {
      await expect(page.locator('.navbar-brand')).toBeVisible();
    }
  });

  test('character list loads', async ({ page }) => {
    await login(page);
    await expect(page.locator('h1:has-text("Character Folio")')).toBeVisible({ timeout: 5000 });
  });

  test('navigation between views works', async ({ page }) => {
    await login(page);

    const views = [
      { link: 'compendium', heading: 'Compendium', bottomNav: 'compendium' },
      { link: 'dice', heading: 'Dice Roller', bottomNav: 'dice' },
      { link: 'encounters', heading: 'Encounter Builder', moreLabel: 'Encounters' },
      { link: 'factions', heading: 'Factions', moreLabel: 'Factions' },
    ];
    for (const { link, heading, bottomNav, moreLabel } of views) {
      if (await isMobile(page)) {
        if (moreLabel) {
          await clickSecondaryNavItem(page, link, 'moreNav', moreLabel);
        } else if (bottomNav) {
          await clickNavItem(page, link, bottomNav);
        }
      } else {
        await page.locator(`#appSidebar button[data-nav="${link}"]`).click();
      }
      await expect(page.locator(`h1:has-text("${heading}")`)).toBeVisible({ timeout: 5000 });
    }
  });

  test('character create and sheet open', async ({ page }) => {
    await login(page);

    const name = `Smoke-${Date.now()}`;
    await page.getByTestId('new-character').click();
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('.modal button:has-text("Create")');
    await waitModalClosed(page);

    await expect(page.locator('.character-card').filter({ hasText: name })).toBeVisible({ timeout: 5000 });
    await page.locator('.character-card').filter({ hasText: name }).click();
    await expect(page.locator('#sheetName')).toContainText(name, { timeout: 10000 });
  });

  test('logout works', async ({ page }) => {
    await login(page);

    if (await isMobile(page)) {
      await page.evaluate(() => (window as any).logout());
    } else {
      const toggler = page.locator('.navbar-toggler');
      if (await toggler.isVisible()) {
        await toggler.click();
        await page.waitForTimeout(300);
      }
      await page.locator('a:has-text("Logout")').click();
    }
    await expect(page).toHaveURL(/\/login/, { timeout: 5000 });
  });

  test('app version is displayed', async ({ page }) => {
    await login(page);

    await expect(page.locator('footer')).toContainText(/v\d+\.\d+\.\d+/, { timeout: 5000 });
  });
});
