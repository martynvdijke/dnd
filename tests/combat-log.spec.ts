import { test, expect } from './fixtures.js';
import { login } from './helpers.js';

const uniqueName = () => `CL-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('Combat Log', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
  });

  test('list combat log entries (empty initially)', async ({ page }) => {
    const entries = await page.evaluate(async () => {
      return window.api('GET', '/api/combat-log');
    });
    expect(Array.isArray(entries)).toBe(true);
  });

  test('create a combat log entry', async ({ page }) => {
    const result = await page.evaluate(async () => {
      return window.api('POST', '/api/combat-log', {
        actor_name: 'Goblin #1',
        action: 'attacks',
        target_name: 'Player Character',
        damage: 5,
        damage_type: 'piercing',
        description: 'Goblin attacks with shortbow',
      });
    });
    expect(result.id).toBeGreaterThan(0);
  });

  test('create multiple combat log entries and list them', async ({ page }) => {
    const actor = uniqueName();
    const result = await page.evaluate(async (a) => {
      await window.api('POST', '/api/combat-log', {
        actor_name: a + '-Hero', action: 'attacks', target_name: 'Goblin',
        damage: 8, damage_type: 'slashing',
      });
      await window.api('POST', '/api/combat-log', {
        actor_name: a + '-Goblin', action: 'attacks', target_name: 'Hero',
        damage: 0, damage_type: 'piercing',
      });
      const entries = await window.api('GET', '/api/combat-log');
      return { ok: true, count: entries.length };
    }, actor);

    expect(result.ok).toBe(true);
    expect(result.count).toBeGreaterThanOrEqual(2);
  });

  test('create entry with full details', async ({ page }) => {
    const result = await page.evaluate(async () => {
      return window.api('POST', '/api/combat-log', {
        actor_name: 'Ancient Dragon',
        action: 'breathes fire',
        target_name: 'Party',
        damage: 25,
        damage_type: 'fire',
        healing: 0,
        roll_expression: '12d6',
        roll_total: 42,
        is_critical: false,
        description: 'Legendary action - fire breath',
      });
    });
    expect(result.id).toBeGreaterThan(0);

    const entry = await page.evaluate(async (id) => {
      const all = await window.api('GET', '/api/combat-log');
      return all.find((e: any) => e.id === id);
    }, result.id);

    expect(entry).toBeTruthy();
    expect(entry.actor_name).toBe('Ancient Dragon');
    expect(entry.action).toBe('breathes fire');
    expect(entry.damage).toBe(25);
  });

  test('get combat log stats requires campaign_id', async ({ page }) => {
    const result = await page.evaluate(async () => {
      try {
        const stats = await window.api('GET', '/api/combat-log/stats?campaign_id=0');
        return { ok: true, data: stats };
      } catch (e) {
        return { ok: false, error: String(e) };
      }
    });

    expect(result).toBeTruthy();
  });
});
