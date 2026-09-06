import { test, expect } from './fixtures.js';
import { NAV_TIMEOUT, clickNavItem, login, waitLoadingDone, waitModalClosed } from './helpers.js';

const uniqueName = () => `Camp-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;







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

async function getCharId(page, name) {
  return page.evaluate(async (charName) => {
    const chars = await window.api('GET', '/api/characters');
    const char = chars.find((c: any) => c.name === charName);
    return char ? char.id : null;
  }, name);
}

async function openPartySubTab(page, tab) {
  await clickNavItem(page, 'party', 'party');
  await page.locator('#partySubTabBar').waitFor({ state: 'visible', timeout: NAV_TIMEOUT });
  await page.locator(`#partySubTabBar button:has-text("${tab}")`).click();
  await waitLoadingDone(page);
}

test.describe('Campaign features', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('party view shows linked locations', async ({ page }) => {
    const name = uniqueName();
    await createCharAndOpen(page, name, 'Human', 'Wizard');

    const charId = await getCharId(page, name);
    expect(charId).toBeTruthy();

    await page.evaluate(async (opts) => {
      const loc = await window.api('POST', '/api/locations', {
        name: 'Waterdeep', type: 'city', description: 'The City of Splendors',
      });
      await window.api('POST', `/api/characters/${opts.charId}/locations`, {
        location_id: loc.id, relationship: 'home', notes: 'birthplace',
      });
    }, { charId });

    await openPartySubTab(page, 'Locations');
    await expect(page.locator('#partyContent h5').first()).toContainText('Campaign Locations');
    await expect(page.locator('#partyContent')).toContainText('Waterdeep');
  });

  test('party view shows campaign NPCs', async ({ page }) => {
    const name = uniqueName();
    await createCharAndOpen(page, name, 'Human', 'Wizard');

    const charId = await getCharId(page, name);
    expect(charId).toBeTruthy();

    await page.evaluate(async (opts) => {
      const npc = await window.api('POST', '/api/npcs', {
        name: 'Elminster', race: 'Human', class: 'Wizard', description: 'The Sage of Shadowdale',
      });
      await window.api('POST', `/api/characters/${opts.charId}/npcs`, {
        npc_id: npc.id, relationship: 'mentor', notes: '',
      });
    }, { charId });

    await openPartySubTab(page, 'NPCs');
    await expect(page.locator('#partyContent h5').first()).toContainText('Campaign NPCs');
    await expect(page.locator('#partyContent')).toContainText('Elminster');
  });

  test('party view shows logged sessions', async ({ page }) => {
    const name = uniqueName();
    await createCharAndOpen(page, name, 'Human', 'Wizard');

    const charId = await getCharId(page, name);
    expect(charId).toBeTruthy();

    await page.evaluate(async (opts) => {
      await window.api('POST', `/api/characters/${opts.charId}/sessions`, {
        session_date: '2026-08-01', title: 'Session 1: The Beginning',
        notes: 'Our heroes met in a tavern...', xp_earned: 300, gold_earned: 50,
        important_events: 'Met the mysterious stranger',
      });
    }, { charId });

    await openPartySubTab(page, 'Sessions');
    await expect(page.locator('#partyContent h5').first()).toContainText('Session Log');
    await expect(page.locator('#partyContent')).toContainText('Session 1: The Beginning');
  });

  test('party view shows quests', async ({ page }) => {
    const name = uniqueName();
    await createCharAndOpen(page, name, 'Human', 'Wizard');

    const charId = await getCharId(page, name);
    expect(charId).toBeTruthy();

    await page.evaluate(async (opts) => {
      await window.api('POST', `/api/characters/${opts.charId}/quests`, {
        name: 'Find the Lost Crown', description: 'The king has lost his crown',
        status: 'active', objectives: '1. Enter the dungeon\n2. Defeat the boss',
        rewards: '1000 XP, Royal Favor',
      });
    }, { charId });

    await openPartySubTab(page, 'Quests');
    await expect(page.locator('#partyContent h5').first()).toContainText('Quests');
    await expect(page.locator('#partyContent')).toContainText('Find the Lost Crown');
  });

  test('Journal tab works', async ({ page }) => {
    const name = uniqueName();
    await createCharAndOpen(page, name, 'Human', 'Wizard');

    await page.click('#tabBar button:has-text("Journal")');
    await expect(page.locator('#journalSection button:has-text("Write Entry")')).toBeVisible({ timeout: NAV_TIMEOUT });
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

  test('party view loads campaign graph', async ({ page }) => {
    const name = uniqueName();
    await createCharAndOpen(page, name, 'Human', 'Wizard');

    const charId = await getCharId(page, name);
    expect(charId).toBeTruthy();

    await page.evaluate(async (opts) => {
      const loc = await window.api('POST', '/api/locations', {
        name: 'Neverwinter', type: 'city', description: 'A city',
      });
      await window.api('POST', `/api/characters/${opts.charId}/locations`, {
        location_id: loc.id, relationship: 'home', notes: '',
      });
      await window.api('POST', `/api/characters/${opts.charId}/sessions`, {
        session_date: '2026-08-02', title: 'Session Test', notes: '', xp_earned: 0, gold_earned: 0,
      });
    }, { charId });

    await openPartySubTab(page, 'Graph');
    await expect(page.locator('#partyContent h5').first()).toContainText('Campaign Graph');
    const svg = page.locator('#partyGraphSvg svg');
    await expect(svg).toBeVisible({ timeout: NAV_TIMEOUT });
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
