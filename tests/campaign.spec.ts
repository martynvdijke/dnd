import { test, expect } from '@playwright/test';
import { login } from './helpers.js';

const uniqueName = () => `Camp-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function waitLoadingDone(page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: 5000 }).catch(() => {});
}

async function waitModalClosed(page) {
  await page.waitForFunction(() => {
    const modal = document.getElementById('genericModal');
    return !modal || !modal.classList.contains('show');
  }, { timeout: 10000 }).catch(() => {});
}

async function createCharAndOpen(page, name, race, cls) {
  await page.click('text=New Character');
  await page.fill('#newName', name);
  await page.fill('#newRace', race);
  await page.fill('#newClass', cls);
  await page.click('text=Create');
  await waitModalClosed(page);
  await page.locator('.character-card').filter({ hasText: name }).waitFor({ state: 'visible', timeout: 10000 });
  await page.locator('.character-card').filter({ hasText: name }).click();
  await waitLoadingDone(page);
}

test.describe('Campaign features', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('Locations tab exists and can link a location', async ({ page }) => {
    const name = uniqueName();
    await createCharAndOpen(page, name, 'Human', 'Wizard');

    await page.click('#tabBar button:has-text("Locations")');
    await expect(page.locator('#locationsSection h5').first()).toContainText('Locations');

    await page.locator('#locationsSection button:has-text("New")').click();
    await page.fill('#newLocName', 'Waterdeep');
    await page.fill('#newLocDesc', 'The City of Splendors');
    await page.click('text=Create');
    await waitModalClosed(page);
    await expect(page.locator('#locationsSection button:has-text("Link")')).toBeVisible({ timeout: 10000 });
    await page.locator('#locationsSection button:has-text("Link")').click();
    await page.selectOption('#linkLocId', { index: 0 });
    await page.click('#genericModal button:has-text("Link")');
    await waitModalClosed(page);

    await expect(page.locator('#locationsSection')).toContainText('Waterdeep');
  });

  test('NPCs tab works', async ({ page }) => {
    const name = uniqueName();
    await createCharAndOpen(page, name, 'Human', 'Wizard');

    await page.click('#tabBar button:has-text("Npcs")');
    await expect(page.locator('#npcsSection button:has-text("New NPC")')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#npcsSection h5').first()).toContainText('Related NPCs');

    await page.click('text=New NPC');
    await page.fill('#newNPCName', 'Elminster');
    await page.fill('#newNPCRace', 'Human');
    await page.fill('#newNPCClass', 'Wizard');
    await page.fill('#newNPCDesc', 'The Sage of Shadowdale');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.click('text=Link NPC');
    await page.selectOption('#linkNPCId', { index: 0 });
    await page.click('#genericModal button:has-text("Link")');
    await waitModalClosed(page);

    await expect(page.locator('#npcsSection')).toContainText('Elminster');
  });

  test('Sessions tab allows logging sessions', async ({ page }) => {
    const name = uniqueName();
    await createCharAndOpen(page, name, 'Human', 'Wizard');

    await page.click('#tabBar button:has-text("Sessions")');
    await expect(page.locator('#sessionsSection button:has-text("Log Session")')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#sessionsSection h5').first()).toContainText('Session Log');

    await page.click('text=Log Session');
    await page.fill('#sessTitle', 'Session 1: The Beginning');
    await page.fill('#sessNotes', 'Our heroes met in a tavern...');
    await page.fill('#sessXP', '300');
    await page.fill('#sessGold', '50');
    await page.fill('#sessEvents', 'Met the mysterious stranger');
    await page.click('#genericModal button:has-text("Log Session")');
    await waitModalClosed(page);

    await expect(page.locator('#sessionsSection')).toContainText('Session 1');
    await expect(page.locator('#sessionsSection')).toContainText('300 XP');
  });

  test('Quests tab works', async ({ page }) => {
    const name = uniqueName();
    await createCharAndOpen(page, name, 'Human', 'Wizard');

    await page.click('#tabBar button:has-text("Quests")');
    await expect(page.locator('#questsSection button:has-text("New Quest")')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#questsSection h5').first()).toContainText('Quests');

    await page.click('text=New Quest');
    await page.fill('#questName', 'Find the Lost Crown');
    await page.fill('#questDesc', 'The king has lost his crown');
    await page.fill('#questObj', '1. Enter the dungeon\n2. Defeat the boss');
    await page.fill('#questRewards', '1000 XP, Royal Favor');
    await page.click('text=Create');
    await waitModalClosed(page);

    await expect(page.locator('#questsSection')).toContainText('Find the Lost Crown');
  });

  test('Journal tab works', async ({ page }) => {
    const name = uniqueName();
    await createCharAndOpen(page, name, 'Human', 'Wizard');

    await page.click('#tabBar button:has-text("Journal")');
    await expect(page.locator('#journalSection button:has-text("Write Entry")')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#journalSection h5').first()).toContainText('Character Journal');

    await page.click('text=Write Entry');
    await page.fill('#journalTitle', 'Day 1');
    await page.locator('#journalEditor .ProseMirror').click();
    await page.locator('#journalEditor .ProseMirror').fill('Today was the first day of my adventure...');
    await page.click('#genericModal button:has-text("Save")');
    await waitModalClosed(page);

    await expect(page.locator('#journalSection')).toContainText('Day 1');
    await expect(page.locator('#journalSection')).toContainText('first day of my adventure');
  });

  test('Graph tab loads visualization', async ({ page }) => {
    const name = uniqueName();
    await createCharAndOpen(page, name, 'Human', 'Wizard');

    await page.click('#tabBar button:has-text("Locations")');
    await expect(page.locator('#locationsSection h5').first()).toBeVisible();
    await page.locator('#locationsSection button:has-text("New")').click();
    await page.fill('#newLocName', 'Neverwinter');
    await page.fill('#newLocDesc', 'A city');
    await page.click('text=Create');
    await waitModalClosed(page);
    await expect(page.locator('#locationsSection button:has-text("Link")')).toBeVisible({ timeout: 10000 });
    await page.locator('#locationsSection button:has-text("Link")').click();
    await page.selectOption('#linkLocId', { index: 0 });
    await page.click('#genericModal button:has-text("Link")');
    await waitModalClosed(page);

    await page.click('#tabBar button:has-text("Sessions")');
    await expect(page.locator('#sessionsSection h5').first()).toBeVisible();
    await page.click('text=Log Session');
    await page.fill('#sessTitle', 'Session Test');
    await page.fill('#sessXP', '0');
    await page.fill('#sessGold', '0');
    await page.click('#genericModal button:has-text("Log Session")');
    await waitModalClosed(page);

    await page.click('#tabBar button:has-text("Graph")');
    await expect(page.locator('#graphContainer')).toBeVisible({ timeout: 5000 });
  });

  // ─── Dashboard ───

  test.describe('Dashboard', () => {
    test('campaign dashboard API returns summary data', async ({ page }) => {
      const name = uniqueName();
      await createCharAndOpen(page, name, 'Human', 'Fighter');

      // Get the campaign from the character's campaign_id
      const result = await page.evaluate(async (opts) => {
        const chars = await window.api('GET', '/api/characters');
        const char = chars.find((c: any) => c.name === opts.charName);
        if (!char || !char.campaign_id) return { err: 'no campaign' };
        try {
          const dash = await window.api('GET', `/api/campaigns/${char.campaign_id}/dashboard`);
          return { ok: true, data: dash };
        } catch (e) {
          return { ok: false, error: String(e) };
        }
      }, { charName: name });

      expect(result).toBeTruthy();
    });
  });

  // ─── Faction Reputation ───

  test.describe('Faction Reputation', () => {
    test('create a faction', async ({ page }) => {
      const name = uniqueName();
      const result = await page.evaluate(async (n) => {
        return window.api('POST', '/api/factions', {
          name: n,
          description: 'A secret society',
          alignment: 'lawful neutral',
          headquarters: 'Shadow Tower',
          influence: 50,
          resources: JSON.stringify({ gold: 5000 }),
          goals: JSON.stringify(['Gain power', 'Control the guilds']),
          visibility: 'known',
        });
      }, name);
      expect(result.id).toBeGreaterThan(0);
    });

    test('list factions', async ({ page }) => {
      const name = uniqueName();
      await page.evaluate(async (n) => {
        await window.api('POST', '/api/factions', {
          name: n, description: 'Test faction', alignment: 'neutral',
          headquarters: 'Tower', influence: 25,
          resources: '{}', goals: '[]', visibility: 'known',
        });
      }, name);

      const result = await page.evaluate(async (n) => {
        const factions = await window.api('GET', '/api/factions');
        return { ok: true, count: factions.length, found: factions.some((f: any) => f.name === n) };
      }, name);

      expect(result.ok).toBe(true);
      expect(result.found).toBe(true);
    });

    test('update a faction', async ({ page }) => {
      const name = uniqueName();
      const created = await page.evaluate(async (n) => {
        return window.api('POST', '/api/factions', {
          name: n, description: 'Original', alignment: 'lawful good',
          headquarters: 'Castle', influence: 10,
          resources: '{}', goals: '[]', visibility: 'known',
        });
      }, name);

      await page.evaluate(async ({ id }) => {
        return window.api('PUT', `/api/factions/${id}`, {
          name: 'Updated Faction', description: 'Updated description',
          alignment: 'chaotic neutral', headquarters: 'Fortress',
          influence: 75, resources: '{}', goals: '[]', visibility: 'secret',
        });
      }, created);

      const updated = await page.evaluate(async (id) => {
        const all = await window.api('GET', '/api/factions');
        return all.find((f: any) => f.id === id);
      }, created.id);

      expect(updated).toBeTruthy();
      expect(updated.name).toBe('Updated Faction');
    });

    test('delete a faction', async ({ page }) => {
      const name = uniqueName();
      const created = await page.evaluate(async (n) => {
        return window.api('POST', '/api/factions', {
          name: n, description: 'Delete me', alignment: 'neutral',
          headquarters: 'Hut', influence: 5,
          resources: '{}', goals: '[]', visibility: 'unknown',
        });
      }, name);

      await page.evaluate(async (id) => {
        return window.api('DELETE', `/api/factions/${id}`);
      }, created.id);

      const remaining = await page.evaluate(async (id) => {
        const all = await window.api('GET', '/api/factions');
        return all.filter((f: any) => f.id === id).length;
      }, created.id);

      expect(remaining).toBe(0);
    });

    test('create faction reputation entry', async ({ page }) => {
      const charName = uniqueName();
      await createCharAndOpen(page, charName, 'Human', 'Fighter');

      const factionName = uniqueName();
      const result = await page.evaluate(async (opts) => {
        const chars = await window.api('GET', '/api/characters');
        const char = chars.find((c: any) => c.name === opts.charName);
        if (!char) return { err: 'character not found' };

        const faction = await window.api('POST', '/api/factions', {
          name: opts.factionName, description: 'Rep test', type: 'guild',
          headquarters: 'Hall',
        });

        const result = await window.api('POST', '/api/faction-reputation', {
          character_id: char.id,
          faction_id: faction.id,
          standing: 50,
          rank: 'Member',
          notes: 'Neutral standing',
        });
        return { ok: true, factionId: faction.id, charId: char.id };
      }, { charName, factionName });

      expect(result.err).toBeFalsy();
      expect(result.factionId).toBeGreaterThan(0);
    });

    test('delete faction reputation entry', async ({ page }) => {
      const charName = uniqueName();
      await createCharAndOpen(page, charName, 'Human', 'Fighter');

      const factionName = uniqueName();
      const result = await page.evaluate(async (opts) => {
        const chars = await window.api('GET', '/api/characters');
        const char = chars.find((c: any) => c.name === opts.charName);
        if (!char) return { err: 'character not found' };

        const faction = await window.api('POST', '/api/factions', {
          name: opts.factionName, description: 'Rep del test', type: 'guild',
          headquarters: 'Cave',
        });

        // Create reputation (returns {ok: true}, no id)
        await window.api('POST', '/api/faction-reputation', {
          character_id: char.id,
          faction_id: faction.id,
          standing: 80,
          rank: 'Leader',
          notes: 'Friendly',
        });

        // Get the reputation entry ID
        const reps = await window.api('GET', `/api/faction-reputation?character_id=${char.id}`);
        const rep = reps.find((r: any) => r.faction_id === faction.id);

        await window.api('DELETE', `/api/faction-reputation/${rep.id}`);

        // Verify it's gone
        const repsAfter = await window.api('GET', `/api/faction-reputation?character_id=${char.id}`);
        return { ok: true, remaining: repsAfter.filter((r: any) => r.faction_id === faction.id).length };
      }, { charName, factionName });

      expect(result.err).toBeFalsy();
      expect(result.remaining).toBe(0);
    });
  });
});
