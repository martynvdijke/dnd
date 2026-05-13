import { test, expect } from '@playwright/test';

test.describe('Shops & Trading', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded', timeout: 10000 });
    await page.waitForTimeout(800);
  });

  test('shops nav is visible for admin', async ({ page }) => {
    await expect(page.locator('#shopsNavItem')).toBeVisible();
  });

  test('create a shop via API', async ({ page }) => {
    await page.evaluate(async () => {
      await window.api('POST', '/api/admin/shops', {
        name: 'Arcane Emporium',
        description: 'Magical goods and potions',
        markup_percent: 120,
        markup_buy_percent: 40,
      });
    });
    await page.click('#shopsNavItem a');
    await page.waitForTimeout(500);
    await expect(page.locator('#shopSelect')).toContainText('Arcane Emporium');
  });

  test('adds items to shop and views them', async ({ page }) => {
    await page.evaluate(async () => {
      await window.api('POST', '/api/admin/shops', { name: 'Blacksmith', description: 'Weapons and armor', markup_percent: 100, markup_buy_percent: 50 });
      const shops = await window.api('GET', '/api/shops');
      const shop = shops.find((s: any) => s.name === 'Blacksmith');
      if (shop) {
        await window.api('POST', '/api/admin/shops/' + shop.id + '/items', {
          item_name: 'Longsword', category: 'weapon', price_gp: 15, quantity_available: 5, description: 'A sharp blade',
        });
      }
    });

    await page.click('#shopsNavItem a');
    await page.waitForTimeout(500);
    await expect(page.locator('#shopSelect')).toContainText('Blacksmith');
  });
});
