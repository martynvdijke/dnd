import { test, expect } from './fixtures.js';

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
    await clickSecondaryNavItem(page, 'shops', 'moreNavShops', 'Shops');
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

    await clickSecondaryNavItem(page, 'shops', 'moreNavShops', 'Shops');
    await page.waitForTimeout(500);
    await expect(page.locator('#shopsGrid')).toContainText(shopName);
  });

  test('shop item linked from compendium snapshots data and can be unlinked', async ({ page }) => {
    const shopName = uniqueName();
    await page.evaluate(async (name) => {
      const comp = await window.api('GET', '/api/compendium/equipment');
      const entry = comp[0];
      if (!entry) throw new Error('no compendium equipment seeded');
      await window.api('POST', '/api/admin/shops', { name, description: 'Compendium stock', markup_percent: 100, markup_buy_percent: 50 });
      const shops = await window.api('GET', '/api/shops');
      const shop = shops.find((s: any) => s.name === name);
      if (!shop) throw new Error('shop not created');
      await window.api('POST', '/api/admin/shops/' + shop.id + '/items', {
        item_name: 'Wrong Name', category: 'gear', price_gp: 9, quantity_available: 2, compendium_equipment_id: entry.id,
      });
      const items = await window.api('GET', '/api/shops/' + shop.id + '/items');
      return { items, shopId: shop.id };
    }, shopName).then(async (result: any) => {
      // Name is snapshotted from the compendium entry and the link is kept.
      const item = result.items[0];
      expect(item.compendium_equipment_id).toBeTruthy();
      expect(item.item_name).not.toBe('Wrong Name');
      // Unlink preserves the data but drops the reference.
      await page.evaluate(async (args) => {
        await window.api('DELETE', '/api/shop-items/' + args.item.id + '/link');
      }, { item });
      const after = await page.evaluate(async (shopId: number) => {
        const items = await window.api('GET', '/api/shops/' + shopId + '/items');
        return items[0];
      }, result.shopId);
      expect(after.compendium_equipment_id).toBeFalsy();
      expect(after.item_name).toBe(item.item_name);
    });
  });
});
