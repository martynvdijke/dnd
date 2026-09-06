import { test, expect } from './fixtures.js';
import { waitLoadingDone, clickSecondaryNavItem, login, NAV_TIMEOUT } from './helpers.js';

const uniqueName = () => `OSC-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

/** Load HTMX content into oneshotSection */
async function loadHtmx(page, url, target?: string) {
  await page.evaluate(async ({ u, t }) => {
    const resp = await fetch(u, { credentials: 'same-origin' });
    const el = document.getElementById(t || 'oneshotSection')!;
    el.innerHTML = await resp.text();
    (window as any).htmx?.process(el);
  }, { u: url, t: target || null });
}

/** Click the One-Shots nav link, handling mobile hamburger menu */
async function navigateToOneShots(page) {
  await clickSecondaryNavItem(page, 'oneshots', 'moreNavOneshot', 'One-Shots');
  await page.waitForSelector('#oneshotSection', { state: 'visible', timeout: NAV_TIMEOUT });
}

/** Create a one-shot adventure via API and return its ID */
async function createOneShot(page, title: string) {
  return page.evaluate(async (t) => {
    return (window as any).api('POST', '/api/oneshot-adventures', {
      title: t, template: 'custom', difficulty: 'medium', estimated_minutes: 120,
    });
  }, title);
}

/** Create a one-shot adventure with acts/scenes via template */
async function createGeneratedOneShot(page, title: string) {
  return page.evaluate(async (t) => {
    return (window as any).api('POST', '/api/oneshot-adventures/generate', {
      title: t, template: 'five_room_dungeon', difficulty: 'easy', estimated_minutes: 60,
    });
  }, title);
}

test.describe('One-Shot Content Features', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test.describe('Items', () => {
    test('Create an item for a one-shot', async ({ page }) => {
      const title = uniqueName();
      const adv = await createOneShot(page, title);
      expect(adv.id).toBeGreaterThan(0);

      const item = await page.evaluate(async ({ id, name }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/items`, {
          name, description: 'A magical longsword', category: 'weapon',
          quantity: 1, weight: 3, price_gp: 200, is_magical: true, attunement: false, notes: 'Flame tongue',
        });
      }, { id: adv.id, name: uniqueName() });
      expect(item.id).toBeGreaterThan(0);
    });

    test('List items for a one-shot', async ({ page }) => {
      const title = uniqueName();
      const adv = await createOneShot(page, title);
      const itemName = uniqueName();
      await page.evaluate(async ({ id, name }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/items`, {
          name, description: 'Test item', category: 'weapon',
          quantity: 1, weight: 1, price_gp: 10, is_magical: false, attunement: false,
        });
      }, { id: adv.id, name: itemName });

      const items = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}/items`);
      }, adv.id);
      expect(Array.isArray(items)).toBe(true);
      expect(items.some((i: any) => i.name === itemName)).toBe(true);
    });

    test('Update item details', async ({ page }) => {
      const title = uniqueName();
      const adv = await createOneShot(page, title);
      const itemName = uniqueName();
      const created = await page.evaluate(async ({ id, name }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/items`, {
          name, description: 'Old desc', category: 'weapon',
          quantity: 1, weight: 1, price_gp: 10, is_magical: false, attunement: false,
        });
      }, { id: adv.id, name: itemName });

      await page.evaluate(async (item) => {
        return (window as any).api('PUT', `/api/oneshot-items/${item.id}`, {
          name: item.name + '-updated', description: 'Updated description', category: 'armor',
          quantity: 2, weight: 5, price_gp: 50, is_magical: true, attunement: true,
        });
      }, created);

      const items = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}/items`);
      }, adv.id);
      const updated = items.find((i: any) => i.id === created.id);
      expect(updated).toBeDefined();
      expect(updated.name).toBe(itemName + '-updated');
      expect(updated.is_magical).toBe(true);
      expect(updated.attunement).toBe(true);
    });

    test('Delete item removes it from list', async ({ page }) => {
      const title = uniqueName();
      const adv = await createOneShot(page, title);
      const itemName = uniqueName();
      const created = await page.evaluate(async ({ id, name }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/items`, {
          name, description: 'Delete me', category: 'potion',
          quantity: 1, weight: 0.5, price_gp: 50, is_magical: true, attunement: false,
        });
      }, { id: adv.id, name: itemName });

      await page.evaluate(async (itemId) => {
        return (window as any).api('DELETE', `/api/oneshot-items/${itemId}`);
      }, created.id);

      const items = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}/items`);
      }, adv.id);
      expect(items.some((i: any) => i.id === created.id)).toBe(false);
    });

    test('List item uploads returns empty array', async ({ page }) => {
      const title = uniqueName();
      const adv = await createOneShot(page, title);
      const itemName = uniqueName();
      const created = await page.evaluate(async ({ id, name }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/items`, {
          name, description: 'Upload test', category: 'weapon',
          quantity: 1, weight: 1, price_gp: 10, is_magical: false, attunement: false,
        });
      }, { id: adv.id, name: itemName });

      const uploads = await page.evaluate(async (itemId) => {
        return (window as any).api('GET', `/api/oneshot-items/${itemId}/uploads`);
      }, created.id);
      expect(Array.isArray(uploads)).toBe(true);
      expect(uploads.length).toBe(0);
    });
  });

  // ─── One-Shot Shops ───

  test.describe('Shops', () => {
    test('Create a shop within a one-shot', async ({ page }) => {
      const title = uniqueName();
      const adv = await createOneShot(page, title);

      const result = await page.evaluate(async ({ id, name }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/shops`, {
          name, description: 'A dark alchemy shop',
          markup_percent: 150, markup_buy_percent: 50,
        });
      }, { id: adv.id, name: uniqueName() });
      expect(result.id).toBeGreaterThan(0);
    });

    test('List shops for a one-shot', async ({ page }) => {
      const title = uniqueName();
      const adv = await createOneShot(page, title);
      const shopName = uniqueName();
      await page.evaluate(async ({ id, name }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/shops`, {
          name, description: 'Potion shop',
          markup_percent: 100, markup_buy_percent: 50,
        });
      }, { id: adv.id, name: shopName });

      const shops = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}/shops`);
      }, adv.id);
      expect(Array.isArray(shops)).toBe(true);
      expect(shops.some((s: any) => s.name === shopName)).toBe(true);
    });

    test('Create shop item within a one-shot shop', async ({ page }) => {
      const title = uniqueName();
      const adv = await createOneShot(page, title);
      const shopName = uniqueName();
      await page.evaluate(async ({ id, name }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/shops`, {
          name, description: 'General store',
          markup_percent: 120, markup_buy_percent: 40,
        });
      }, { id: adv.id, name: shopName });

      // Get shop ID
      const shops = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}/shops`);
      }, adv.id);
      const shop = shops.find((s: any) => s.name === shopName);
      expect(shop).toBeDefined();

      // Add item to shop
      const result = await page.evaluate(async ({ shopId, advId }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${advId}/shops/${shopId}/items`, {
          item_name: 'Potion of Healing', category: 'potion',
          price_gp: 50, quantity_available: 5,
          description: 'Restores 2d4+2 HP', is_magical: true, attunement_required: false,
        });
      }, { shopId: shop.id, advId: adv.id });
      expect(result.ok).toBe(true);
    });

    test('Delete shop within a one-shot', async ({ page }) => {
      const title = uniqueName();
      const adv = await createOneShot(page, title);
      const shopName = uniqueName();
      await page.evaluate(async ({ id, name }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/shops`, {
          name, description: 'Delete me',
          markup_percent: 100, markup_buy_percent: 50,
        });
      }, { id: adv.id, name: shopName });

      const shops = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}/shops`);
      }, adv.id);
      const shop = shops.find((s: any) => s.name === shopName);
      expect(shop).toBeDefined();

      await page.evaluate(async ({ id, shopId }) => {
        return (window as any).api('DELETE', `/api/oneshot-adventures/${id}/shops/${shopId}`);
      }, { id: adv.id, shopId: shop.id });

      const shopsAfter = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}/shops`);
      }, adv.id);
      expect(shopsAfter.some((s: any) => s.id === shop.id)).toBe(false);
    });

    test('Shop items HTMX section loads in one-shot detail view', async ({ page }) => {
      // Create a one-shot with a shop
      const title = uniqueName();
      const adv = await createOneShot(page, title);
      const shopName = 'Test Shop ' + uniqueName();
      await page.evaluate(async ({ id, name }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/shops`, {
          name, description: 'Shop for HTMX test',
          markup_percent: 100, markup_buy_percent: 50,
        });
      }, { id: adv.id, name: shopName });

      // Navigate to one-shots
      await navigateToOneShots(page);
      // Load detail view
      await loadHtmx(page, `/htmx/oneshot-adventures/${adv.id}`);
      // Manually load shops section into its container
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
      const shopsHtml = await page.evaluate(async (advId) => {
        const resp = await fetch(`/htmx/oneshot-adventures/${advId}/shops`, { credentials: 'same-origin' });
        if (!resp.ok) return '';
        return resp.text();
      }, adv.id);
      if (shopsHtml) {
        await page.evaluate((html) => {
          const card = document.querySelector('[hx-get*="/shops"]');
          if (card) card.innerHTML = html;
        }, shopsHtml);
      }
      await expect(page.locator('#oneshotSection')).toContainText(shopName, { timeout: NAV_TIMEOUT });
    });
  });

  // ─── One-Shot Monsters ───


});
