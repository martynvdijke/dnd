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
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
      await expect(page.locator('#oneshotSection')).toContainText('Monsters', { timeout: NAV_TIMEOUT });
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
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
      await expect(page.locator('#oneshotSection')).toContainText('Player Characters', { timeout: NAV_TIMEOUT });
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


});
