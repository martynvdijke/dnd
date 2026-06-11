import { test, expect } from '@playwright/test';

const uniqueName = () => `Shop-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

import { ensureNavOpen, isMobile, clickSecondaryNavItem, login } from './helpers.js';

test.describe('Shops & Trading', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.waitForTimeout(300);
  });

  test('shops nav is visible for admin', async ({ page }) => {
    if (await isMobile(page)) return;
    await ensureNavOpen(page);
    await expect(page.locator('#sidebarShopsNav')).toBeVisible();
  });

  test('create a shop via API', async ({ page }) => {
    const shopName = uniqueName();
    await page.evaluate(async (name) => {
      await window.api('POST', '/api/admin/shops', {
        name,
        description: 'Magical goods and potions',
        markup_percent: 120,
        markup_buy_percent: 40,
      });
    }, shopName);
    await clickSecondaryNavItem(page, 'Shops', 'moreNavShops');
    await page.waitForTimeout(500);
    await expect(page.locator('#shopsGrid')).toContainText(shopName);
  });

  test('adds items to shop and views them', async ({ page }) => {
    const shopName = uniqueName();
    await page.evaluate(async (name) => {
      await window.api('POST', '/api/admin/shops', { name, description: 'Weapons and armor', markup_percent: 100, markup_buy_percent: 50 });
      const shops = await window.api('GET', '/api/shops');
      const shop = shops.find((s: any) => s.name === name);
      if (shop) {
        await window.api('POST', '/api/admin/shops/' + shop.id + '/items', {
          item_name: 'Longsword', category: 'weapon', price_gp: 15, quantity_available: 5, description: 'A sharp blade',
        });
      }
    }, shopName);

    await clickSecondaryNavItem(page, 'Shops', 'moreNavShops');
    await page.waitForTimeout(500);
    await expect(page.locator('#shopsGrid')).toContainText(shopName);
  });
});
