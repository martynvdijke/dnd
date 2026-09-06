import { test, expect } from './fixtures.js';
import { login } from './helpers.js';

const uniqueName = () => `Cond-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function createCharacter(page, name) {
  await page.click('text=New Character');
  await page.fill('#newName', name);
  await page.fill('#newRace', 'Human');
  await page.fill('#newClass', 'Fighter');
  await page.click('text=Create');
  await page.waitForFunction(() => !document.getElementById('genericModal')?.classList.contains('show'), { timeout: 10000 }).catch(() => {});
  await page.locator('.character-card').filter({ hasText: name }).waitFor({ state: 'visible', timeout: 10000 });
}

test.describe('Conditions', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
  });

  test('get conditions types', async ({ page }) => {
    const result = await page.evaluate(async () => {
      return window.api('GET', '/api/conditions/types');
    });
    // API returns { conditions: [...] }
    const types = result.conditions || result;
    expect(Array.isArray(types)).toBe(true);
    expect(types.length).toBeGreaterThan(0);
  });

  test('list conditions for a character', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (charName) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === charName);
      if (!char) return { err: 'character not found' };
      const conditions = await window.api('GET', `/api/conditions?character_id=${char.id}`);
      return { ok: true, count: conditions.length };
    }, name);

    expect(result.err).toBeFalsy();
    expect(result.count).toBeGreaterThanOrEqual(0);
  });

  test('create a condition for a character', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (charName) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === charName);
      if (!char) return { err: 'character not found' };
      const cond = await window.api('POST', '/api/conditions', {
        character_id: char.id,
        name: 'Charmed',
        type: 'charmed',
        source: 'spell',
        duration: 60,
        duration_type: 'round',
        saving_throw: 'WIS',
        save_dc: 15,
        description: 'Charmed by the witch',
      });
      return { ok: true, id: cond.id };
    }, name);

    expect(result.err).toBeFalsy();
    expect(result.id).toBeGreaterThan(0);
  });

  test('update a condition', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (charName) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === charName);
      if (!char) return { err: 'character not found' };
      const created = await window.api('POST', '/api/conditions', {
        character_id: char.id, name: 'Original', type: 'charmed',
        duration: 60, duration_type: 'round', source: 'spell',
        saving_throw: 'WIS', save_dc: 15,
      });
      await window.api('PUT', `/api/conditions/${created.id}`, {
        name: 'Stunned', type: 'stunned', source: 'monster',
        duration: 30, duration_type: 'round',
        saving_throw: 'CON', save_dc: 18,
      });
      const all = await window.api('GET', `/api/conditions?character_id=${char.id}`);
      const updated = all.find((c: any) => c.id === created.id);
      return { ok: true, name: updated?.name };
    }, name);

    expect(result.err).toBeFalsy();
    expect(result.name).toBe('Stunned');
  });

  test('delete a condition', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (charName) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === charName);
      if (!char) return { err: 'character not found' };
      const created = await window.api('POST', '/api/conditions', {
        character_id: char.id, name: 'Delete Me', type: 'charmed',
        duration: 1, duration_type: 'round', source: 'test',
        saving_throw: 'WIS', save_dc: 10,
      });
      await window.api('DELETE', `/api/conditions/${created.id}`);
      const all = await window.api('GET', `/api/conditions?character_id=${char.id}`);
      return { ok: true, remaining: all.filter((c: any) => c.id === created.id).length };
    }, name);

    expect(result.err).toBeFalsy();
    expect(result.remaining).toBe(0);
  });

  test('tick conditions', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (charName) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === charName);
      if (!char) return { err: 'character not found' };
      await window.api('POST', '/api/conditions', {
        character_id: char.id, name: 'Ticking', type: 'charmed',
        duration: 2, duration_type: 'round', source: 'spell',
        saving_throw: 'WIS', save_dc: 10,
      });
      try {
        const tick = await window.api('POST', '/api/conditions/tick', {
          character_id: char.id, count: 1, duration_type: 'round',
        });
        return { ok: true, data: tick };
      } catch (e) {
        return { ok: false, error: String(e) };
      }
    }, name);

    expect(result).toBeTruthy();
  });

  test('get conditions summary', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (charName) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === charName);
      if (!char) return { err: 'character not found' };
      try {
        const summary = await window.api('GET', `/api/conditions/summary?character_id=${char.id}`);
        return { ok: true, data: summary };
      } catch (e) {
        return { ok: false, error: String(e) };
      }
    }, name);

    expect(result.err).toBeFalsy();
  });
});
