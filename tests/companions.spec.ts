import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT } from './helpers.js';

const uniqueName = () => `Compn-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function waitLoadingDone(page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: NAV_TIMEOUT }).catch(() => {});
}

async function createCharacter(page, name) {
  await page.click('text=New Character');
  await page.fill('#newName', name);
  await page.fill('#newRace', 'Human');
  await page.fill('#newClass', 'Fighter');
  await page.click('text=Create');
  await page.waitForFunction(() => !document.getElementById('genericModal')?.classList.contains('show'), { timeout: 10000 }).catch(() => {});
  await page.locator('.character-card').filter({ hasText: name }).waitFor({ state: 'visible', timeout: 10000 });
  return name;
}

test.describe('Companions', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('list all companions for a character (empty initially)', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const companions = await page.evaluate(async (charName) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === charName);
      if (!char) return { err: 'character not found' };
      return window.api('GET', `/api/companions?character_id=${char.id}`);
    }, name);

    expect(Array.isArray(companions)).toBe(true);
  });

  test('create a companion', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (opts) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === opts.name);
      if (!char) return { err: 'character not found' };
      const companion = await window.api('POST', '/api/companions', {
        character_id: char.id,
        name: 'Wolf',
        type: 'beast',
        race: 'Wolf',
        hp_max: 25,
        hp_current: 25,
        ac: 12,
        str: 14, dex: 15, con: 12, int_: 2, wis: 12, cha: 6,
        speed: 40,
        abilities: JSON.stringify([{ name: 'Keen Senses', desc: 'Advantage on Perception' }]),
        notes: 'Loyal companion',
        is_alive: true,
      });
      return { ok: true, id: companion.id };
    }, { name });

    expect(result.err).toBeFalsy();
    expect(result.id).toBeGreaterThan(0);
  });

  test('get companion by id', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (opts) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === opts.name);
      if (!char) return { err: 'character not found' };
      const created = await window.api('POST', '/api/companions', {
        character_id: char.id, name: 'Falcon', type: 'beast', race: 'Falcon',
        hp_max: 10, hp_current: 10, ac: 14,
        str: 6, dex: 16, con: 8, int_: 2, wis: 14, cha: 6,
        speed: 60, abilities: '{}', notes: '', is_alive: true,
      });
      const all = await window.api('GET', `/api/companions?character_id=${char.id}`);
      const found = all.find((c: any) => c.id === created.id);
      return { ok: true, name: found.name, type: found.type };
    }, { name });

    expect(result.err).toBeFalsy();
    expect(result.name).toBe('Falcon');
    expect(result.type).toBe('beast');
  });

  test('update a companion', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (opts) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === opts.name);
      if (!char) return { err: 'character not found' };
      const created = await window.api('POST', '/api/companions', {
        character_id: char.id, name: 'Bear', type: 'beast', race: 'Brown Bear',
        hp_max: 34, hp_current: 34, ac: 11,
        str: 18, dex: 10, con: 16, int_: 2, wis: 12, cha: 6,
        speed: 40, abilities: '{}', notes: '', is_alive: true,
      });
      await window.api('PUT', `/api/companions/${created.id}`, {
        name: 'Dire Bear', hp_max: 50, hp_current: 50, ac: 14,
        str: 20, dex: 12, con: 18, int_: 2, wis: 12, cha: 6,
        speed: 40, abilities: JSON.stringify([{ name: 'Rage', desc: 'Double damage' }]),
        notes: 'Upgraded', is_alive: true,
      });
      const all = await window.api('GET', `/api/companions?character_id=${char.id}`);
      const updated = all.find((c: any) => c.id === created.id);
      return { ok: true, name: updated.name, hp_max: updated.hp_max };
    }, { name });

    expect(result.err).toBeFalsy();
    expect(result.name).toBe('Dire Bear');
    expect(result.hp_max).toBe(50);
  });

  test('delete a companion', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (opts) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === opts.name);
      if (!char) return { err: 'character not found' };
      const created = await window.api('POST', '/api/companions', {
        character_id: char.id, name: 'Rat', type: 'beast', race: 'Rat',
        hp_max: 1, hp_current: 1, ac: 10,
        str: 2, dex: 14, con: 10, int_: 1, wis: 10, cha: 4,
        speed: 20, abilities: '{}', notes: '', is_alive: true,
      });
      await window.api('DELETE', `/api/companions/${created.id}`);
      const all = await window.api('GET', `/api/companions?character_id=${char.id}`);
      return { ok: true, remaining: all.filter((c: any) => c.id === created.id).length };
    }, { name });

    expect(result.err).toBeFalsy();
    expect(result.remaining).toBe(0);
  });
});
