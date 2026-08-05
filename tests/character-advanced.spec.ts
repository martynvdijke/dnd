import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT } from './helpers.js';

const uniqueName = () => `CHA-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function waitLoadingDone(page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: NAV_TIMEOUT }).catch(() => {});
}

async function createCharacter(page, name) {
  return page.evaluate(async (n) => {
    return (window as any).api('POST', '/api/characters', {
      name: n, race: 'Human', class: 'Fighter', level: 3,
      str: 16, dex: 14, con: 15, int_: 10, wis: 12, cha: 8,
      hp_max: 30, hp_current: 30, ac: 18, speed: 30,
    });
  }, name);
}

test.describe('Character Advanced Features', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  // ─── Downtime Activities ───

  test.describe('Downtime Activities', () => {
    test('Get downtime types', async ({ page }) => {
      const types = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/downtime/types');
      });
      expect(Array.isArray(types)).toBe(true);
      expect(types.length).toBeGreaterThan(0);
    });

    test('Create a downtime activity', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      const name = 'Research ' + uniqueName();
      const result = await page.evaluate(async ({ cid, n }) => {
        return (window as any).api('POST', `/api/characters/${cid}/downtime`, {
          activity_type: 'research', name: n, description: 'Studying ancient texts',
          dc: 15, days_required: 10, cost_per_day: 5, total_cost: 50,
          reward: 'Knowledge', status: 'in-progress',
        });
      }, { cid: char.id, n: name });
      expect(result.id).toBeGreaterThan(0);
    });

    test('List downtime activities', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      const name = 'List Downtime ' + uniqueName();
      await page.evaluate(async ({ cid, n }) => {
        return (window as any).api('POST', `/api/characters/${cid}/downtime`, {
          activity_type: 'research', name: n, description: 'Test',
          dc: 10, days_required: 5, cost_per_day: 2, total_cost: 10,
          reward: 'Info', status: 'in-progress',
        });
      }, { cid: char.id, n: name });

      const activities = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/characters/${cid}/downtime`);
      }, char.id);
      expect(Array.isArray(activities)).toBe(true);
      expect(activities.some((a: any) => a.name === name)).toBe(true);
    });

    test('Update a downtime activity', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      const name = 'Update DT ' + uniqueName();
      const created = await page.evaluate(async ({ cid, n }) => {
        return (window as any).api('POST', `/api/characters/${cid}/downtime`, {
          activity_type: 'crafting', name: n, description: 'Original',
          dc: 12, days_required: 7, cost_per_day: 10, total_cost: 70,
          reward: 'Item', status: 'in-progress',
        });
      }, { cid: char.id, n: name });

      await page.evaluate(async ({ id, n }) => {
        return (window as any).api('PUT', `/api/downtime/${id}`, {
          name: n + '-updated', description: 'Updated', status: 'in-progress',
        });
      }, { id: created.id, n: name });

      const activities = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/characters/${cid}/downtime`);
      }, char.id);
      const updated = activities.find((a: any) => a.id === created.id);
      expect(updated).toBeDefined();
      expect(updated.name).toBe(name + '-updated');
    });

    test('Advance downtime day', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      const name = 'Advance DT ' + uniqueName();
      const created = await page.evaluate(async ({ cid, n }) => {
        return (window as any).api('POST', `/api/characters/${cid}/downtime`, {
          activity_type: 'training', name: n, description: 'Advance test',
          dc: 10, days_required: 5, cost_per_day: 2, total_cost: 10,
          reward: 'Skill', status: 'in-progress',
        });
      }, { cid: char.id, n: name });

      const result = await page.evaluate(async (id) => {
        return (window as any).api('POST', `/api/downtime/${id}/advance`);
      }, created.id);
      expect(result).toBeTruthy();
    });

    test('Delete a downtime activity', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      const name = 'Delete DT ' + uniqueName();
      const created = await page.evaluate(async ({ cid, n }) => {
        return (window as any).api('POST', `/api/characters/${cid}/downtime`, {
          activity_type: 'other', name: n, description: 'Delete me',
          dc: 10, days_required: 1, cost_per_day: 0, total_cost: 0,
          reward: '', status: 'in-progress',
        });
      }, { cid: char.id, n: name });

      await page.evaluate(async (id) => {
        return (window as any).api('DELETE', `/api/downtime/${id}`);
      }, created.id);

      const activities = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/characters/${cid}/downtime`);
      }, char.id);
      expect(activities.some((a: any) => a.id === created.id)).toBe(false);
    });
  });

  // ─── Level Planner ───

  test.describe('Level Planner', () => {
    test('Save a level up plan', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      const result = await page.evaluate(async (cid) => {
        return (window as any).api('POST', `/api/characters/${cid}/level-plan`, {
          target_level: 5,
          plan_data: [{ class: 'Fighter', feats: ['Tough'], asi: 'Str+2' }],
          notes: 'Plan to level up',
        });
      }, char.id);
      expect(result).toBeTruthy();
    });

    test('Get level up plan', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      await page.evaluate(async (cid) => {
        return (window as any).api('POST', `/api/characters/${cid}/level-plan`, {
          target_level: 5,
          plan_data: [{ class: 'Fighter' }],
          notes: 'Test plan',
        });
      }, char.id);

      const plan = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/characters/${cid}/level-plan`);
      }, char.id);
      expect(plan).toBeTruthy();
      expect(plan.target_level || plan.id).toBeTruthy();
    });

    test('Update level up plan', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      await page.evaluate(async (cid) => {
        return (window as any).api('POST', `/api/characters/${cid}/level-plan`, {
          target_level: 3,
          plan_data: [{ class: 'Fighter' }],
          notes: 'Original',
        });
      }, char.id);

      await page.evaluate(async (cid) => {
        return (window as any).api('POST', `/api/characters/${cid}/level-plan`, {
          target_level: 6,
          plan_data: [{ class: 'Fighter', subclass: 'Champion' }],
          notes: 'Updated plan',
        });
      }, char.id);

      const plan = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/characters/${cid}/level-plan`);
      }, char.id);
      expect(plan.target_level || plan.plan_data).toBeTruthy();
    });

    test('Delete level up plan', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      await page.evaluate(async (cid) => {
        return (window as any).api('POST', `/api/characters/${cid}/level-plan`, {
          target_level: 4,
          plan_data: [{ class: 'Fighter' }],
          notes: 'Delete me',
        });
      }, char.id);

      await page.evaluate(async (cid) => {
        return (window as any).api('DELETE', `/api/characters/${cid}/level-plan`);
      }, char.id);

      const plan = await page.evaluate(async (cid) => {
        try {
          return await (window as any).api('GET', `/api/characters/${cid}/level-plan`);
        } catch (e) {
          return null;
        }
      }, char.id);
      // Plan should be gone - either null or empty object
      expect(plan === null || plan.id === undefined || plan.target_level === undefined || plan.id === 0).toBe(true);
    });

    test('Get level up suggestions', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      const suggestions = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/characters/${cid}/level-suggestions`);
      }, char.id);
      expect(Array.isArray(suggestions) || typeof suggestions === 'object').toBe(true);
    });
  });

  // ─── Companions ───

  test.describe('Companions', () => {
    test('Create a companion', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      const name = 'Buddy ' + uniqueName();
      const result = await page.evaluate(async ({ cid, n }) => {
        return (window as any).api('POST', '/api/companions', {
          character_id: cid, name: n, type: 'animal', race: 'Wolf',
          hp_max: 15, hp_current: 15, ac: 13,
          str: 14, dex: 15, con: 12, int: 2, wis: 12, cha: 6,
        });
      }, { cid: char.id, n: name });
      expect(result.id).toBeGreaterThan(0);
    });

    test('List companions for a character', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      const name = 'List Comp ' + uniqueName();
      await page.evaluate(async ({ cid, n }) => {
        return (window as any).api('POST', '/api/companions', {
          character_id: cid, name: n, type: 'animal', race: 'Dog',
          hp_max: 10, hp_current: 10, ac: 12,
          str: 12, dex: 14, con: 10, int: 2, wis: 10, cha: 6,
        });
      }, { cid: char.id, n: name });

      const companions = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/companions?character_id=${cid}`);
      }, char.id);
      expect(Array.isArray(companions)).toBe(true);
      expect(companions.some((c: any) => c.name === name)).toBe(true);
    });

    test('Update a companion', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      const name = 'Update Comp ' + uniqueName();
      const created = await page.evaluate(async ({ cid, n }) => {
        return (window as any).api('POST', '/api/companions', {
          character_id: cid, name: n, type: 'animal', race: 'Wolf',
          hp_max: 15, hp_current: 15, ac: 13,
          str: 14, dex: 15, con: 12, int: 2, wis: 12, cha: 6,
        });
      }, { cid: char.id, n: name });

      await page.evaluate(async ({ id, n }) => {
        return (window as any).api('PUT', `/api/companions/${id}`, {
          name: n + '-updated', race: 'Dire Wolf',
          hp_max: 30, hp_current: 30, ac: 15,
          str: 18, dex: 14, con: 16, int: 2, wis: 12, cha: 6,
        });
      }, { id: created.id, n: name });

      const companions = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/companions?character_id=${cid}`);
      }, char.id);
      const updated = companions.find((c: any) => c.id === created.id);
      expect(updated).toBeDefined();
      expect(updated.name).toBe(name + '-updated');
    });

    test('Delete a companion', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      const name = 'Delete Comp ' + uniqueName();
      const created = await page.evaluate(async ({ cid, n }) => {
        return (window as any).api('POST', '/api/companions', {
          character_id: cid, name: n, type: 'animal', race: 'Cat',
          hp_max: 5, hp_current: 5, ac: 10,
          str: 4, dex: 16, con: 8, int: 2, wis: 12, cha: 6,
        });
      }, { cid: char.id, n: name });

      await page.evaluate(async (id) => {
        return (window as any).api('DELETE', `/api/companions/${id}`);
      }, created.id);

      const companions = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/companions?character_id=${cid}`);
      }, char.id);
      expect(companions.some((c: any) => c.id === created.id)).toBe(false);
    });

    test('Create companion with full stats', async ({ page }) => {
      const char = await createCharacter(page, uniqueName());
      const name = 'Full Comp ' + uniqueName();
      const result = await page.evaluate(async ({ cid, n }) => {
        return (window as any).api('POST', '/api/companions', {
          character_id: cid, name: n, type: 'mount', race: 'Warhorse',
          hp_max: 30, hp_current: 30, ac: 14,
          str: 18, dex: 12, con: 16, int: 3, wis: 12, cha: 8,
          speed: 60, abilities: 'Trampling Charge', notes: 'A loyal steed',
          is_alive: true,
        });
      }, { cid: char.id, n: name });
      expect(result.id).toBeGreaterThan(0);

      const companions = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/companions?character_id=${cid}`);
      }, char.id);
      const found = companions.find((c: any) => c.id === result.id);
      expect(found).toBeDefined();
      expect(found.type).toBe('mount');
    });
  });
});
