import { test, expect } from '@playwright/test';
import { waitLoadingDone, clickSecondaryNavItem } from './helpers.js';

const uniqueName = () => `OSC-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

/** Load HTMX content into oneshotSection */
async function loadHtmx(page, url) {
  await page.evaluate(async (u) => {
    const resp = await fetch(u, { credentials: 'same-origin' });
    const el = document.getElementById('oneshotSection')!;
    el.innerHTML = await resp.text();
    (window as any).htmx?.process(el);
  }, url);
}

/** Click the One-Shots nav link, handling mobile hamburger menu */
async function navigateToOneShots(page) {
  await clickSecondaryNavItem(page, 'One-Shots', 'moreNavOneshot');
  await page.waitForSelector('#oneshotSection', { state: 'visible', timeout: 5000 });
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
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded', timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
    await page.waitForTimeout(300);
  });

  // ─── One-Shot Items ───

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
      await page.waitForTimeout(500);
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
      await expect(page.locator('#oneshotSection')).toContainText(shopName, { timeout: 5000 });
    });
  });

  // ─── One-Shot Monsters ───

  test.describe('Monsters', () => {
    test('Create monster library entry', async ({ page }) => {
      const name = uniqueName();
      const result = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/monster-library', {
          name: n, ac: 15, hp: 50, str: 16, dex: 12, con: 14, int_: 8, wis: 10, cha: 8,
          cr: '3', source: 'Monster Manual', is_full: true,
          saves: 'Str+5,Con+4', skills: 'Athletics+5,Perception+2',
          damage_resistances: 'fire', senses: 'darkvision 60ft', languages: 'Common',
          special_abilities: 'Keen Senses', actions: 'Multiattack',
          description: 'A powerful creature',
        });
      }, name);
      expect(result.id).toBeGreaterThan(0);
    });

    test('List monster library returns created entries', async ({ page }) => {
      const name = uniqueName();
      await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/monster-library', {
          name: n, ac: 12, hp: 30, str: 14, dex: 10, con: 12, int_: 6, wis: 8, cha: 6,
          cr: '1', source: 'Homebrew', is_full: false,
        });
      }, name);

      const entries = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/monster-library');
      });
      expect(Array.isArray(entries)).toBe(true);
      expect(entries.some((e: any) => e.name === name)).toBe(true);
    });

    test('Update monster library entry', async ({ page }) => {
      const name = uniqueName();
      const created = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/monster-library', {
          name: n, ac: 10, hp: 20, str: 10, dex: 10, con: 10, int_: 10, wis: 10, cha: 10,
          cr: '0', source: 'Test', is_full: false,
        });
      }, name);

      await page.evaluate(async (entry) => {
        return (window as any).api('PUT', `/api/monster-library/${entry.id}`, {
          name: entry.name + '-updated', ac: 18, hp: 100, str: 20, dex: 14, con: 18, int_: 10, wis: 12, cha: 10,
          cr: '5', source: 'Updated', is_full: true,
        });
      }, created);

      const entries = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/monster-library');
      });
      const updated = entries.find((e: any) => e.id === created.id);
      expect(updated).toBeDefined();
      expect(updated.name).toBe(name + '-updated');
      expect(updated.ac).toBe(18);
      expect(updated.hp).toBe(100);
    });

    test('Delete monster library entry', async ({ page }) => {
      const name = uniqueName();
      const created = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/monster-library', {
          name: n, ac: 10, hp: 10, str: 10, dex: 10, con: 10, int_: 10, wis: 10, cha: 10,
          cr: '0', source: 'Delete test', is_full: false,
        });
      }, name);

      await page.evaluate(async (id) => {
        return (window as any).api('DELETE', `/api/monster-library/${id}`);
      }, created.id);

      const entries = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/monster-library');
      });
      expect(entries.some((e: any) => e.id === created.id)).toBe(false);
    });

    test('Create and list act monsters', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      expect(adv.id).toBeGreaterThan(0);

      // Get the one-shot detail to find act IDs
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThan(0);
      const actId = detail.acts[0].id;

      const monsterName = uniqueName();
      const created = await page.evaluate(async ({ actId, name }) => {
        return (window as any).api('POST', `/api/oneshot-acts/${actId}/monsters`, {
          name, ac: 14, hp: 30, str: 14, dex: 10, con: 12, int_: 6, wis: 8, cha: 6,
          cr: '2', source: 'Homebrew', is_full: false,
        });
      }, { actId, name: monsterName });
      expect(created.id).toBeGreaterThan(0);

      // List monsters for this act
      const monsters = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-acts/${id}/monsters`);
      }, actId);
      expect(monsters.some((m: any) => m.name === monsterName)).toBe(true);
    });

    test('Update and delete act monster', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      const actId = detail.acts[0].id;

      const monsterName = uniqueName();
      const created = await page.evaluate(async ({ actId, name }) => {
        return (window as any).api('POST', `/api/oneshot-acts/${actId}/monsters`, {
          name, ac: 10, hp: 10, str: 10, dex: 10, con: 10, int_: 10, wis: 10, cha: 10,
          cr: '0', source: 'Test', is_full: false,
        });
      }, { actId, name: monsterName });

      // Update
      await page.evaluate(async (monster) => {
        return (window as any).api('PUT', `/api/oneshot-monsters/${monster.id}`, {
          name: monster.name + '-updated', ac: 20, hp: 80, str: 18, dex: 14, con: 16, int_: 10, wis: 12, cha: 10,
          cr: '4', source: 'Updated', is_full: true,
        });
      }, created);

      const monsters = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-acts/${id}/monsters`);
      }, actId);
      const updated = monsters.find((m: any) => m.id === created.id);
      expect(updated).toBeDefined();
      expect(updated.name).toBe(created.name + '-updated');
      expect(updated.ac).toBe(20);

      // Delete
      await page.evaluate(async (id) => {
        return (window as any).api('DELETE', `/api/oneshot-monsters/${id}`);
      }, created.id);

      const monstersAfter = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-acts/${id}/monsters`);
      }, actId);
      expect(monstersAfter.some((m: any) => m.id === created.id)).toBe(false);
    });

    test('Monsters HTMX section loads in one-shot detail view', async ({ page }) => {
      const title = uniqueName();
      const adv = await createOneShot(page, title);

      await navigateToOneShots(page);
      await loadHtmx(page, `/htmx/oneshot-adventures/${adv.id}`);
      await page.waitForTimeout(1000);
      await expect(page.locator('#oneshotSection')).toContainText('Monsters', { timeout: 5000 });
    });
  });

  // ─── Linked Player Characters ───

  test.describe('Player Characters', () => {
    test('Create a character for linking', async ({ page }) => {
      const name = uniqueName();
      const result = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/characters', {
          name: n, race: 'Human', class: 'Fighter', level: 3,
          str: 16, dex: 14, con: 15, int_: 10, wis: 12, cha: 8,
          hp_max: 30, hp_current: 30, ac: 18, speed: 30,
        });
      }, name);
      expect(result.id).toBeGreaterThan(0);
    });

    test('Link character to a one-shot', async ({ page }) => {
      const adv = await createOneShot(page, uniqueName());

      // Create a character first
      const charName = uniqueName();
      const character = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/characters', {
          name: n, race: 'Elf', class: 'Wizard', level: 5,
          str: 8, dex: 14, con: 12, int_: 18, wis: 14, cha: 10,
          hp_max: 28, hp_current: 28, ac: 12, speed: 30,
        });
      }, charName);

      // Link character
      const result = await page.evaluate(async ({ advId, charId }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${advId}/characters`, {
          character_id: charId, role: 'party_member', notes: 'The party wizard',
        });
      }, { advId: adv.id, charId: character.id });
      expect(result.id).toBeGreaterThan(0);
    });

    test('List linked characters for a one-shot', async ({ page }) => {
      const adv = await createOneShot(page, uniqueName());
      const charName = uniqueName();
      const character = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/characters', {
          name: n, race: 'Dwarf', class: 'Cleric', level: 4,
          str: 14, dex: 8, con: 16, int_: 10, wis: 18, cha: 12,
          hp_max: 40, hp_current: 40, ac: 18, speed: 25,
        });
      }, charName);

      await page.evaluate(async ({ advId, charId }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${advId}/characters`, {
          character_id: charId, role: 'party_member', notes: '',
        });
      }, { advId: adv.id, charId: character.id });

      const linked = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}/characters`);
      }, adv.id);
      expect(linked.length).toBeGreaterThan(0);
      expect(linked.some((pc: any) => pc.char_name === charName)).toBe(true);
    });

    test('Unlink character from a one-shot', async ({ page }) => {
      const adv = await createOneShot(page, uniqueName());
      const charName = uniqueName();
      const character = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/characters', {
          name: n, race: 'Halfling', class: 'Rogue', level: 2,
          str: 8, dex: 18, con: 12, int_: 14, wis: 10, cha: 16,
          hp_max: 18, hp_current: 18, ac: 15, speed: 25,
        });
      }, charName);

      await page.evaluate(async ({ advId, charId }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${advId}/characters`, {
          character_id: charId, role: 'party_member', notes: '',
        });
      }, { advId: adv.id, charId: character.id });

      // Unlink
      await page.evaluate(async ({ advId, charId }) => {
        return (window as any).api('DELETE', `/api/oneshot-adventures/${advId}/characters/${charId}`);
      }, { advId: adv.id, charId: character.id });

      const linked = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}/characters`);
      }, adv.id);
      expect(linked.some((pc: any) => pc.character_id === character.id)).toBe(false);
    });

    test('PCs HTMX section loads in one-shot detail view', async ({ page }) => {
      const title = uniqueName();
      const adv = await createOneShot(page, title);

      await navigateToOneShots(page);
      await loadHtmx(page, `/htmx/oneshot-adventures/${adv.id}`);
      await page.waitForTimeout(1000);
      await expect(page.locator('#oneshotSection')).toContainText('Player Characters', { timeout: 5000 });
    });
  });

  // ─── NPC ↔ Item Links ───

  test.describe('NPC-Item Links', () => {
    test('Create NPC-item link', async ({ page }) => {
      const adv = await createOneShot(page, uniqueName());

      // Create an NPC
      const npcName = uniqueName();
      const npc = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/npcs', {
          name: n, race: 'Human', class: 'Merchant', description: 'A shady dealer',
          str: 10, dex: 10, con: 10, int_: 12, wis: 14, cha: 16,
          hp_max: 10, hp_current: 10,
        });
      }, npcName);

      // Create an item
      const itemName = uniqueName();
      const item = await page.evaluate(async ({ id, name }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/items`, {
          name, description: 'Trade good', category: 'treasure',
          quantity: 1, weight: 0.1, price_gp: 100, is_magical: false, attunement: false,
        });
      }, { id: adv.id, name: itemName });

      // Link NPC to item
      const link = await page.evaluate(async ({ advId, npcId, itemId }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${advId}/npc-item-links`, {
          npc_id: npcId, item_id: itemId, adventure_id: advId,
          relationship_type: 'owns', notes: 'Carries the item on his belt',
        });
      }, { advId: adv.id, npcId: npc.id, itemId: item.id });
      expect(link.id).toBeGreaterThan(0);
    });

    test('List items for NPC', async ({ page }) => {
      const adv = await createOneShot(page, uniqueName());
      const npcName = uniqueName();
      const npc = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/npcs', {
          name: n, race: 'Orc', class: 'Warrior', description: 'A brute',
          str: 16, dex: 10, con: 14, int_: 6, wis: 8, cha: 6,
          hp_max: 20, hp_current: 20,
        });
      }, npcName);
      const itemName = uniqueName();
      const item = await page.evaluate(async ({ id, name }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/items`, {
          name, description: 'Battleaxe', category: 'weapon',
          quantity: 1, weight: 5, price_gp: 30, is_magical: false, attunement: false,
        });
      }, { id: adv.id, name: itemName });

      // Create link
      await page.evaluate(async ({ advId, npcId, itemId }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${advId}/npc-item-links`, {
          npc_id: npcId, item_id: itemId, adventure_id: advId,
          relationship_type: 'wields', notes: '',
        });
      }, { advId: adv.id, npcId: npc.id, itemId: item.id });

      // List items for NPC
      const items = await page.evaluate(async ({ advId, npcId }) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${advId}/npcs/${npcId}/items`);
      }, { advId: adv.id, npcId: npc.id });
      expect(items.length).toBeGreaterThan(0);
      expect(items.some((l: any) => l.item_name === itemName)).toBe(true);
    });

    test('List NPCs for item', async ({ page }) => {
      const adv = await createOneShot(page, uniqueName());
      const npcName = uniqueName();
      const npc = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/npcs', {
          name: n, race: 'Goblin', class: 'Scout', description: 'A sneaky goblin',
          str: 8, dex: 16, con: 10, int_: 10, wis: 12, cha: 8,
          hp_max: 10, hp_current: 10,
        });
      }, npcName);
      const itemName = uniqueName();
      const item = await page.evaluate(async ({ id, name }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/items`, {
          name, description: 'Shortbow', category: 'weapon',
          quantity: 1, weight: 2, price_gp: 25, is_magical: false, attunement: false,
        });
      }, { id: adv.id, name: itemName });

      await page.evaluate(async ({ advId, npcId, itemId }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${advId}/npc-item-links`, {
          npc_id: npcId, item_id: itemId, adventure_id: advId,
          relationship_type: 'wields', notes: '',
        });
      }, { advId: adv.id, npcId: npc.id, itemId: item.id });

      // List NPCs for item
      const npcs = await page.evaluate(async (itemId) => {
        return (window as any).api('GET', `/api/oneshot-items/${itemId}/npcs`);
      }, item.id);
      expect(npcs.length).toBeGreaterThan(0);
      expect(npcs.some((l: any) => l.npc_name === npcName)).toBe(true);
    });

    test('Delete NPC-item link', async ({ page }) => {
      const adv = await createOneShot(page, uniqueName());
      const npcName = uniqueName();
      const npc = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/npcs', {
          name: n, race: 'Tiefling', class: 'Warlock', description: 'A pact-bound fiend',
          str: 8, dex: 12, con: 14, int_: 10, wis: 10, cha: 18,
          hp_max: 24, hp_current: 24,
        });
      }, npcName);
      const itemName = uniqueName();
      const item = await page.evaluate(async ({ id, name }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/items`, {
          name, description: 'Rod of the Pact Keeper', category: 'wand',
          quantity: 1, weight: 1, price_gp: 500, is_magical: true, attunement: true,
        });
      }, { id: adv.id, name: itemName });

      const link = await page.evaluate(async ({ advId, npcId, itemId }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${advId}/npc-item-links`, {
          npc_id: npcId, item_id: itemId, adventure_id: advId,
          relationship_type: 'owns', notes: 'Arcane focus',
        });
      }, { advId: adv.id, npcId: npc.id, itemId: item.id });

      // Delete the link
      await page.evaluate(async (linkId) => {
        return (window as any).api('DELETE', `/api/npc-item-links/${linkId}`);
      }, link.id);

      // Verify it's gone from NPC's items
      const items = await page.evaluate(async ({ advId, npcId }) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${advId}/npcs/${npcId}/items`);
      }, { advId: adv.id, npcId: npc.id });
      expect(items.some((l: any) => l.id === link.id)).toBe(false);
    });
  });

  // ─── Inline Editing ───

  test.describe('Inline Editing', () => {
    test('Update act duration', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThan(0);
      const actId = detail.acts[0].id;

      const result = await page.evaluate(async ({ id, minutes }) => {
        return (window as any).api('PATCH', `/api/oneshot-acts/${id}/duration`, {
          estimated_minutes: minutes,
        });
      }, { id: actId, minutes: 45 });
      expect(result.ok).toBe(true);
      expect(result.estimated_minutes).toBe(45);
    });

    test('Update scene duration', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThan(0);
      expect(detail.acts[0].scenes.length).toBeGreaterThan(0);
      const sceneId = detail.acts[0].scenes[0].id;

      const result = await page.evaluate(async ({ id, minutes }) => {
        return (window as any).api('PATCH', `/api/oneshot-scenes/${id}/duration`, {
          estimated_minutes: minutes,
        });
      }, { id: sceneId, minutes: 20 });
      expect(result.ok).toBe(true);
      expect(result.estimated_minutes).toBe(20);
    });

    test('Reorder acts', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThanOrEqual(2);

      // Reverse the order
      const reversed = [...detail.acts].reverse().map((a: any) => a.id);
      const result = await page.evaluate(async ({ id, order }) => {
        return (window as any).api('PUT', `/api/oneshot-adventures/${id}/acts/reorder`, { order });
      }, { id: adv.id, order: reversed });
      expect(result.ok).toBe(true);

      // Verify the order changed
      const updated = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(updated.acts[0].id).toBe(reversed[0]);
    });

    test('Reorder scenes within an act', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThan(0);

      const actId = detail.acts[0].id;
      // Ensure at least 2 scenes for reorder
      if (detail.acts[0].scenes.length < 2) {
        await page.evaluate(async ({ actId, title }) => {
          return (window as any).api('POST', `/api/oneshot-acts/${actId}/scenes`, {
            title, scene_type: 'exploration', estimated_minutes: 15,
          });
        }, { actId, title: 'Extra Scene' });
        const updated = await page.evaluate(async (id) => {
          return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
        }, adv.id);
        detail.acts = updated.acts;
      }
      expect(detail.acts[0].scenes.length).toBeGreaterThanOrEqual(2);

      const reversed = [...detail.acts[0].scenes].reverse().map((s: any) => s.id);
      const result = await page.evaluate(async ({ id, order }) => {
        return (window as any).api('PUT', `/api/oneshot-acts/${id}/scenes/reorder`, { order });
      }, { id: actId, order: reversed });
      expect(result.ok).toBe(true);

      // Verify the order changed
      const updated = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      const updatedAct = updated.acts.find((a: any) => a.id === actId);
      expect(updatedAct).toBeDefined();
      expect(updatedAct.scenes[0].id).toBe(reversed[0]);
    });
  });

  // ─── Act & Scene Editing ───

  test.describe('Act & Scene Editing', () => {
    test('Edit act title via HTMX', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThan(0);
      const actId = detail.acts[0].id;
      const newTitle = 'Edited ' + uniqueName();

      await navigateToOneShots(page);
      await loadHtmx(page, `/htmx/oneshot-adventures/${adv.id}`);
      await page.waitForTimeout(300);

      // Click edit button on first act
      const editBtn = page.locator(`.sortable-act[data-id="${actId}"] .btn-outline-secondary`).first();
      await editBtn.click();
      await page.waitForTimeout(300);

      // Fill modal and submit
      const titleInput = page.locator('#genericModalBody input[name="title"]');
      await titleInput.fill(newTitle);
      const submitBtn = page.locator('#genericModalBody button[class*="btn-primary"]');
      await submitBtn.click();
      await page.waitForTimeout(500);

      // Verify updated title in detail
      await expect(page.locator('#oneshotSection')).toContainText(newTitle, { timeout: 5000 });
    });

    test('Edit scene details via HTMX', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThan(0);
      expect(detail.acts[0].scenes.length).toBeGreaterThan(0);
      const sceneId = detail.acts[0].scenes[0].id;
      const newTitle = 'Edited Scene ' + uniqueName();

      await navigateToOneShots(page);
      await loadHtmx(page, `/htmx/oneshot-adventures/${adv.id}`);
      await page.waitForTimeout(300);

      // Click edit button on first scene
      const editBtn = page.locator(`.sortable-scene[data-id="${sceneId}"] .btn-outline-secondary`).first();
      await editBtn.click();
      await page.waitForTimeout(300);

      // Fill modal
      const titleInput = page.locator('#genericModalBody input[name="title"]');
      await titleInput.fill(newTitle);
      const submitBtn = page.locator('#genericModalBody button[class*="btn-primary"]');
      await submitBtn.click();
      await page.waitForTimeout(500);

      // Verify updated title
      await expect(page.locator('#oneshotSection')).toContainText(newTitle, { timeout: 5000 });
    });

    test('Edit act parent via API', async ({ page }) => {
      const title = uniqueName();
      const adv = await createOneShot(page, title);

      // Create 2 acts
      const act1 = await page.evaluate(async ({ id }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/acts`, {
          title: 'Root Act', number: 1, sort_order: 1,
        });
      }, { id: adv.id });

      const act2 = await page.evaluate(async ({ id }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/acts`, {
          title: 'Child Act', number: 1, sort_order: 1,
        });
      }, { id: adv.id });

      // Move act2 under act1
      await page.evaluate(async ({ id, parentId }) => {
        return (window as any).api('PUT', `/api/oneshot-acts/${id}`, {
          title: 'Child Act', number: 1, sort_order: 1, parent_act_id: parentId,
        });
      }, { id: act2.id, parentId: act1.id });

      // Verify tree
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBe(1);
      expect(detail.acts[0].children.length).toBe(1);
      expect(detail.acts[0].children[0].title).toBe('Child Act');
    });
  });

  // ─── Scene Dialogs ───

  test.describe('Scene Dialogs', () => {
    async function createDialog(page: any, sceneId: number, speaker: string, text: string) {
      return page.evaluate(async ({ sceneId, speaker, text }) => {
        const formData = new URLSearchParams();
        formData.append('speaker', speaker);
        formData.append('dialog_text', text);
        const resp = await fetch(`/htmx/oneshot-scenes/${sceneId}/dialogs`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: formData.toString(),
        });
        return resp.ok;
      }, { sceneId, speaker, text });
    }

    test('Create and view scene dialogs', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThan(0);
      const sceneId = detail.acts[0].scenes[0].id;

      await navigateToOneShots(page);
      await loadHtmx(page, `/htmx/oneshot-adventures/${adv.id}`);
      await page.waitForTimeout(300);

      // Open dialog modal for first scene
      const dialogBtn = page.locator(`.sortable-scene[data-id="${sceneId}"] .btn-outline-warning`).first();
      await dialogBtn.click();
      await page.waitForTimeout(300);

      // Create a dialog entry
      const speaker = 'Guard ' + uniqueName();
      const dtext = 'Halt! Who goes there?';
      const created = await createDialog(page, sceneId, speaker, dtext);
      expect(created).toBe(true);

      // Reload dialogs in modal
      await loadHtmx(page, `/htmx/oneshot-scenes/${sceneId}/dialogs`);
      await page.waitForTimeout(300);
      await expect(page.locator('#genericModalBody')).toContainText(speaker, { timeout: 5000 });
      await expect(page.locator('#genericModalBody')).toContainText(dtext, { timeout: 5000 });
    });

    test('Edit scene dialog via HTMX', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      const sceneId = detail.acts[0].scenes[0].id;
      const newText = 'Password? ' + uniqueName();

      // Create dialog via API
      const created = await createDialog(page, sceneId, 'Captain', 'Old text');
      expect(created).toBe(true);

      // Get dialog ID from DB
      const dialogId = await page.evaluate(async (sceneId) => {
        const resp = await fetch(`/htmx/oneshot-scenes/${sceneId}/dialogs`);
        const html = await resp.text();
        const match = html.match(/data-id="(\d+)"/);
        return match ? parseInt(match[1]) : null;
      }, sceneId);

      // Open dialog edit form
      await loadHtmx(page, `/htmx/oneshot-scene-dialogs/${dialogId}/edit`);
      await page.waitForTimeout(200);

      // Change text and submit
      const textarea = page.locator('#genericModalBody textarea[name="dialog_text"]');
      await textarea.fill(newText);
      const submitBtn = page.locator('#genericModalBody button[class*="btn-primary"]');
      await submitBtn.click();
      await page.waitForTimeout(300);

      // Verify updated text
      await expect(page.locator('#genericModalBody')).toContainText(newText, { timeout: 5000 });
    });

    test('Delete scene dialog', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      const sceneId = detail.acts[0].scenes[0].id;
      const speaker = 'DeleteMe ' + uniqueName();

      // Create dialog
      await createDialog(page, sceneId, speaker, 'Delete this');
      const dialogId = await page.evaluate(async (sceneId) => {
        const resp = await fetch(`/htmx/oneshot-scenes/${sceneId}/dialogs`);
        const html = await resp.text();
        const match = html.match(/data-id="(\d+)"/);
        return match ? parseInt(match[1]) : null;
      }, sceneId);
      expect(dialogId).toBeGreaterThan(0);

      // Load dialogs, then delete
      await loadHtmx(page, `/htmx/oneshot-scenes/${sceneId}/dialogs`);
      await page.waitForTimeout(200);

      await expect(page.locator('#genericModalBody')).toContainText(speaker, { timeout: 5000 });

      // Click delete button
      const delBtn = page.locator(`.dialog-card[data-id="${dialogId}"] .btn-outline-danger`).first();
      await delBtn.click();
      await page.waitForTimeout(300);

      // Verify deleted - speaker should no longer be visible
      // The dialogs container gets replaced, so we check the new content
      await expect(page.locator('#genericModalBody')).not.toContainText(speaker, { timeout: 5000 });
    });
  });
});
