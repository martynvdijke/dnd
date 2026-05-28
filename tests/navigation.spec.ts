import { test, expect } from '@playwright/test';
import { clickBottomTabForce } from './helpers.js';

async function waitLoadingDone(page: import('@playwright/test').Page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: 5000 }).catch(() => {});
}

test.describe('Navigation: Bottom Tab Bar & Sidebar', () => {
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

  test('desktop sidebar is visible at 1280x720', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    const sidebar = page.locator('#appSidebar');
    await expect(sidebar).toBeVisible({ timeout: 5000 });
  });

  test('bottom tab bar is visible on mobile at 390x844', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    const bottomBar = page.locator('.bottom-tab-bar');
    await expect(bottomBar).toBeVisible({ timeout: 5000 });
  });

  test('bottom tabs change views', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });

    await clickBottomTabForce(page, 'dice');
    await expect(page.locator('#diceView')).toBeVisible({ timeout: 5000 });

    await clickBottomTabForce(page, 'characters');
    await expect(page.locator('#charactersView')).toBeVisible({ timeout: 5000 });

    await clickBottomTabForce(page, 'party');
    await expect(page.locator('#partyView')).toBeVisible({ timeout: 5000 });

    await clickBottomTabForce(page, 'compendium');
    await expect(page.locator('#compendiumView')).toBeVisible({ timeout: 5000 });
  });

  test('sidebar nav items change views on desktop', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });

    await page.locator('.sidebar-nav-item[data-nav="dice"]').click();
    await expect(page.locator('#diceView')).toBeVisible({ timeout: 5000 });

    await page.locator('.sidebar-nav-item[data-nav="characters"]').click();
    await expect(page.locator('#charactersView')).toBeVisible({ timeout: 5000 });
  });

  test('More menu button exists on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await expect(page.locator('#moreTabBtn')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#moreTabBtn')).toContainText('More');
  });

  test('session mode topbar exists in DOM', async ({ page }) => {
    await expect(page.locator('#sessionModeTopbar')).toBeAttached({ timeout: 5000 });
  });
});
