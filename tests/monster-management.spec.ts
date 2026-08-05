import { test, expect } from './fixtures.js';
import { login } from './helpers.js';

const uniqueName = () => `MNG-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('Monster Management', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.waitForTimeout(200);
  });

  // ─── Monster Library CRUD ───

  test.describe('Monster Library CRUD', () => {
    test('Create monster library entry with full stats', async ({ page }) => {
    test.slow();
      const name = uniqueName();
      const result = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/monster-library', {
          name: n, ac: 18, hp: 120, str: 18, dex: 14, con: 16, int_: 10, wis: 14, cha: 12,
          cr: '5', source: 'Monster Manual', is_full: true,
          saves: 'Str+7,Con+6', skills: 'Athletics+7,Perception+5',
          damage_resistances: 'fire,cold', senses: 'darkvision 60ft',
          languages: 'Common,Draconic', special_abilities: 'Keen Senses',
          actions: 'Multiattack', description: 'A fearsome dragon',
        });
      }, name);
      expect(result.id).toBeGreaterThan(0);
    });

    test('Create monster library entry with minimal fields', async ({ page }) => {
    test.slow();
      const name = uniqueName();
      const result = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/monster-library', {
          name: n, ac: 12, hp: 30, str: 14, dex: 10, con: 12, int_: 6, wis: 8, cha: 6,
          cr: '1', source: 'Homebrew', is_full: false,
        });
      }, name);
      expect(result.id).toBeGreaterThan(0);
    });

    test('List monster library returns created entries', async ({ page }) => {
    test.slow();
      const name = uniqueName();
      await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/monster-library', {
          name: n, ac: 10, hp: 20, str: 10, dex: 10, con: 10, int_: 10, wis: 10, cha: 10,
          cr: '0', source: 'Test', is_full: false,
        });
      }, name);

      const entries = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/monster-library');
      });
      expect(Array.isArray(entries)).toBe(true);
      expect(entries.some((e: any) => e.name === name)).toBe(true);
    });

    test('Update monster library entry', async ({ page }) => {
    test.slow();
      const name = uniqueName();
      const created = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/monster-library', {
          name: n, ac: 10, hp: 20, str: 10, dex: 10, con: 10, int_: 10, wis: 10, cha: 10,
          cr: '0', source: 'Test', is_full: false,
        });
      }, name);

      await page.evaluate(async (entry) => {
        return (window as any).api('PUT', `/api/monster-library/${entry.id}`, {
          name: entry.name + '-updated', ac: 18, hp: 100, str: 20, dex: 14, con: 18,
          int_: 10, wis: 12, cha: 10, cr: '5', source: 'Updated', is_full: true,
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
    test.slow();
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
  });


});
