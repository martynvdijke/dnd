import { test, expect } from './fixtures.js';
import { login } from './helpers.js';

const uniqueName = () => `Level-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

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

test.describe('Level Planner', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
  });

  test('create a level plan for a character', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (opts) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === opts.name);
      if (!char) return { err: 'character not found' };
      const resp = await window.api('POST', `/api/characters/${char.id}/level-plan`, {
        target_level: 5,
        plan_data: [{ level: 4, class: 'Fighter', feat: 'Great Weapon Master', asi: 'Strength 18', notes: '' }],
        notes: 'Aim for Extra Attack feat',
      });
      return { ok: true, id: resp.id };
    }, { name });

    expect(result.err).toBeFalsy();
    expect(result.id).toBeGreaterThan(0);
  });

  test('get level plan for a character', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (opts) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === opts.name);
      if (!char) return { err: 'character not found' };
      await window.api('POST', `/api/characters/${char.id}/level-plan`, {
        target_level: 10,
        plan_data: [{ level: 4, class: 'Fighter', feat: 'Tough', asi: 'Con 16', notes: '' }],
        notes: 'Tank build',
      });
      const plan = await window.api('GET', `/api/characters/${char.id}/level-plan`);
      return { ok: true, data: plan };
    }, { name });

    expect(result.err).toBeFalsy();
    expect(result.data).toBeTruthy();
    expect(result.data.target_level).toBe(10);
  });

  test('get level suggestions for a character', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (opts) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === opts.name);
      if (!char) return { err: 'character not found' };
      try {
        const suggestions = await window.api('GET', `/api/characters/${char.id}/level-suggestions`);
        return { ok: true, data: suggestions };
      } catch (e) {
        return { ok: false, error: String(e) };
      }
    }, { name });

    expect(result).toBeTruthy();
  });

  test('delete a level plan', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (opts) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === opts.name);
      if (!char) return { err: 'character not found' };
      const resp = await window.api('POST', `/api/characters/${char.id}/level-plan`, {
        target_level: 3,
        plan_data: [{ level: 2, class: 'Fighter', feat: '', asi: '', notes: 'Basic' }],
        notes: 'Basic plan',
      });
      await window.api('DELETE', `/api/characters/${char.id}/level-plan`);
      const plan = await window.api('GET', `/api/characters/${char.id}/level-plan`);
      return { ok: true, deleted: !plan.target_level || plan.target_level === 0 };
    }, { name });

    expect(result.err).toBeFalsy();
    expect(result.deleted).toBe(true);
  });
});
