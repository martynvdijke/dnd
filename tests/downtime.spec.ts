import { test, expect } from '@playwright/test';

const uniqueName = () => `DT-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function waitLoadingDone(page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: 5000 }).catch(() => {});
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

test.describe('Downtime Activities', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded', timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
    await page.waitForTimeout(200);
  });

  test('get downtime types', async ({ page }) => {
    const types = await page.evaluate(async () => {
      return window.api('GET', '/api/downtime/types');
    });
    expect(Array.isArray(types)).toBe(true);
  });

  test('create downtime activity for a character', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (opts) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === opts.name);
      if (!char) return { err: 'character not found' };
      const dt = await window.api('POST', `/api/characters/${char.id}/downtime`, {
        activity_type: 'training',
        name: 'Training',
        description: 'Practicing swordsmanship',
        days_required: 30,
        days_completed: 0,
        dc: 10,
        cost_per_day: 0,
        total_cost: 0,
        reward: 'Weapon proficiency',
      });
      return { ok: true, id: dt.id };
    }, { name });

    expect(result.err).toBeFalsy();
    expect(result.id).toBeGreaterThan(0);
  });

  test('list downtime activities for a character', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (opts) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === opts.name);
      if (!char) return { err: 'character not found' };
      await window.api('POST', `/api/characters/${char.id}/downtime`, {
        activity_type: 'training', name: 'Training', description: 'Sword practice',
        days_required: 30, days_completed: 0, dc: 10, cost_per_day: 0, total_cost: 0, reward: '+1 to hit',
      });
      await window.api('POST', `/api/characters/${char.id}/downtime`, {
        activity_type: 'research', name: 'Research', description: 'Library study',
        days_required: 14, days_completed: 5, dc: 12, cost_per_day: 0, total_cost: 0, reward: 'Lore knowledge',
      });
      const activities = await window.api('GET', `/api/characters/${char.id}/downtime`);
      return { ok: true, count: activities.length, names: activities.map((a: any) => a.name) };
    }, { name });

    expect(result.err).toBeFalsy();
    expect(result.count).toBeGreaterThanOrEqual(2);
    expect(result.names).toContain('Training');
    expect(result.names).toContain('Research');
  });

  test('update a downtime activity', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (opts) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === opts.name);
      if (!char) return { err: 'character not found' };
      const created = await window.api('POST', `/api/characters/${char.id}/downtime`, {
        activity_type: 'training', name: 'Original', description: 'Original desc',
        days_required: 30, days_completed: 0, dc: 10, cost_per_day: 0, total_cost: 0, reward: 'None',
      });
      await window.api('PUT', `/api/downtime/${created.id}`, {
        activity_type: 'training', name: 'Updated Activity', description: 'Updated desc',
        days_required: 20, days_completed: 10, dc: 12, cost_per_day: 0, total_cost: 0, reward: '+1 proficiency',
      });
      const all = await window.api('GET', `/api/characters/${char.id}/downtime`);
      const updated = all.find((a: any) => a.id === created.id);
      return { ok: true, name: updated?.name, days_completed: updated?.days_completed };
    }, { name });

    expect(result.err).toBeFalsy();
    expect(result.name).toBe('Updated Activity');
    expect(result.days_completed).toBe(10);
  });

  test('advance a downtime activity', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (opts) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === opts.name);
      if (!char) return { err: 'character not found' };
      const created = await window.api('POST', `/api/characters/${char.id}/downtime`, {
        activity_type: 'research', name: 'Research', description: 'Library',
        days_required: 10, days_completed: 0, dc: 10, cost_per_day: 0, total_cost: 0, reward: 'Lore',
      });
      const advanced = await window.api('POST', `/api/downtime/${created.id}/advance`, { days: 1 });
      return { ok: true, data: advanced };
    }, { name });

    expect(result.err).toBeFalsy();
    expect(result.data).toBeTruthy();
  });

  test('delete a downtime activity', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);

    const result = await page.evaluate(async (opts) => {
      const chars = await window.api('GET', '/api/characters');
      const char = chars.find((c: any) => c.name === opts.name);
      if (!char) return { err: 'character not found' };
      const created = await window.api('POST', `/api/characters/${char.id}/downtime`, {
        activity_type: 'training', name: 'Delete Me', description: 'To delete',
        days_required: 5, days_completed: 0, dc: 10, cost_per_day: 0, total_cost: 0, reward: 'None',
      });
      await window.api('DELETE', `/api/downtime/${created.id}`);
      const all = await window.api('GET', `/api/characters/${char.id}/downtime`);
      return { ok: true, remaining: all.filter((a: any) => a.id === created.id).length };
    }, { name });

    expect(result.err).toBeFalsy();
    expect(result.remaining).toBe(0);
  });
});
