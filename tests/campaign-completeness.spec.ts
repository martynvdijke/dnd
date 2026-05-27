import { test, expect } from '@playwright/test';

const uniqueName = () => `CampaignTest-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function waitLoadingDone(page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: 5000 }).catch(() => {});
}

test.describe('Campaign Completeness', () => {
  let campaignId: number;

  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded', timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);

    // Create a campaign and character for testing
    const name = uniqueName();
    const camp = await page.evaluate(async (n) => {
      return (window as any).api('POST', '/api/campaigns', {
        name: n,
        description: 'E2E test campaign',
      });
    }, name);
    campaignId = camp.id;

    // Create a character in this campaign
    await page.evaluate(async (cid) => {
      await (window as any).api('POST', '/api/characters', {
        name: 'E2E Hero',
        race: 'Human',
        class: 'Fighter',
        level: 5,
        str: 15, dex: 14, con: 13, int: 12, wis: 10, cha: 8,
        hp_max: 50, hp_current: 50,
        ac: 17, speed: 30,
        campaign_id: cid,
      });
    }, campaignId);
  });

  // ─── Party Items API ───

  test('Party items: empty list, create, list, delete', async ({ page }) => {
    // List empty
    let items = await page.evaluate(async (cid) => {
      return (window as any).api('GET', `/api/campaigns/${cid}/party-items`);
    }, campaignId);
    expect(Array.isArray(items)).toBe(true);

    // Create item
    const created = await page.evaluate(async (cid) => {
      return (window as any).api('POST', `/api/campaigns/${cid}/party-items`, {
        name: 'Dragon Hoard Coins',
        quantity: 5000,
        notes: '5000 gp in mixed coins',
      });
    }, campaignId);
    expect(created.id).toBeGreaterThan(0);

    // List shows created item
    items = await page.evaluate(async (cid) => {
      return (window as any).api('GET', `/api/campaigns/${cid}/party-items`);
    }, campaignId);
    expect(items.length).toBe(1);
    expect(items[0].name).toBe('Dragon Hoard Coins');
    expect(items[0].quantity).toBe(5000);

    // Delete item
    await page.evaluate(async (iid) => {
      return (window as any).api('DELETE', `/api/party-items/${iid}`);
    }, created.id);

    // List is empty again
    items = await page.evaluate(async (cid) => {
      return (window as any).api('GET', `/api/campaigns/${cid}/party-items`);
    }, campaignId);
    expect(items.length).toBe(0);
  });

  // ─── Session Plans API ───

  test('Session plans: CRUD cycle', async ({ page }) => {
    // List empty
    let plans = await page.evaluate(async (cid) => {
      return (window as any).api('GET', `/api/campaigns/${cid}/session-plans`);
    }, campaignId);
    expect(Array.isArray(plans)).toBe(true);

    // Create
    const created = await page.evaluate(async (cid) => {
      return (window as any).api('POST', `/api/campaigns/${cid}/session-plans`, {
        title: 'Into the Dungeon',
        session_date: '2025-06-15',
        status: 'planned',
        dm_notes: 'Prepare goblin ambush',
        planned_encounters: JSON.stringify(['Goblin Ambush', 'Trap Room']),
        npc_ids: JSON.stringify([]),
        player_goals: JSON.stringify(['Find the treasure', 'Rescue the prisoner']),
        expected_duration: 180,
      });
    }, campaignId);
    expect(created.id).toBeGreaterThan(0);

    // List shows created
    plans = await page.evaluate(async (cid) => {
      return (window as any).api('GET', `/api/campaigns/${cid}/session-plans`);
    }, campaignId);
    expect(plans.length).toBe(1);
    expect(plans[0].title).toBe('Into the Dungeon');
    expect(plans[0].status).toBe('planned');

    // Update
    const sid = created.id;
    await page.evaluate(async ({ sid }) => {
      return (window as any).api('PUT', `/api/session-plans/${sid}`, {
        title: 'Into the Dungeon - Updated',
        session_date: '2025-06-16',
        status: 'ready',
        dm_notes: 'Now with a red dragon!',
        planned_encounters: JSON.stringify(['Goblin Ambush', 'Trap Room', 'Red Dragon']),
        npc_ids: JSON.stringify([]),
        player_goals: JSON.stringify(['Survive']),
        expected_duration: 240,
      });
    }, { sid });

    // Verify update
    plans = await page.evaluate(async (cid) => {
      return (window as any).api('GET', `/api/campaigns/${cid}/session-plans`);
    }, campaignId);
    expect(plans[0].title).toBe('Into the Dungeon - Updated');
    expect(plans[0].status).toBe('ready');
    expect(plans[0].expected_duration).toBe(240);

    // Delete
    await page.evaluate(async (sid) => {
      return (window as any).api('DELETE', `/api/session-plans/${sid}`);
    }, created.id);

    plans = await page.evaluate(async (cid) => {
      return (window as any).api('GET', `/api/campaigns/${cid}/session-plans`);
    }, campaignId);
    expect(plans.length).toBe(0);
  });

  // ─── Dashboard API ───

  test('Dashboard returns campaign overview data', async ({ page }) => {
    const dash = await page.evaluate(async (cid) => {
      return (window as any).api('GET', `/api/campaigns/${cid}/dashboard`);
    }, campaignId);

    expect(dash).toBeDefined();
    expect(dash.name).toBeDefined();
    expect(dash.characters).toBeDefined();
    expect(Array.isArray(dash.characters)).toBe(true);
    expect(dash.characters.length).toBeGreaterThanOrEqual(1);

    // Character summary fields
    const char = dash.characters[0];
    expect(char.name).toBe('E2E Hero');
    expect(char.level).toBe(5);
    expect(char.hp_current).toBe(50);
  });

  // ─── Exhaustion API ───

  test('Exhaustion: set level via API', async ({ page }) => {
    // Find the character
    const chars = await page.evaluate(async () => {
      return (window as any).api('GET', '/api/characters');
    });
    const hero = chars.find((c: any) => c.name === 'E2E Hero');
    expect(hero).toBeDefined();

    // Set exhaustion to 2
    const result = await page.evaluate(async (charId) => {
      return (window as any).api('PATCH', `/api/characters/${charId}/exhaustion`, { level: 2 });
    }, hero.id);
    expect(result.ok).toBe(true);

    // Invalid level returns error
    try {
      await page.evaluate(async (charId) => {
        return (window as any).api('PATCH', `/api/characters/${charId}/exhaustion`, { level: 7 });
      }, hero.id);
      expect(false).toBe(true); // Should have thrown
    } catch (e: any) {
      expect(e.message).toContain('exhaustion level must be 0-6');
    }

    // Reset to 0
    const reset = await page.evaluate(async (charId) => {
      return (window as any).api('PATCH', `/api/characters/${charId}/exhaustion`, { level: 0 });
    }, hero.id);
    expect(reset.ok).toBe(true);
  });

  // ─── Spell Preparation API ───

  test('Spell preparation: batch prepare/unprepare', async ({ page }) => {
    // Find character
    const chars = await page.evaluate(async () => {
      return (window as any).api('GET', '/api/characters');
    });
    const hero = chars.find((c: any) => c.name === 'E2E Hero');
    expect(hero).toBeDefined();

    // Add spells to the character
    const spell1 = await page.evaluate(async (charId) => {
      return (window as any).api('POST', `/api/characters/${charId}/spells`, {
        name: 'Magic Missile',
        level: 1,
        school: 'evocation',
        prepared: true,
      });
    }, hero.id);
    expect(spell1.id).toBeGreaterThan(0);

    const spell2 = await page.evaluate(async (charId) => {
      return (window as any).api('POST', `/api/characters/${charId}/spells`, {
        name: 'Shield',
        level: 1,
        school: 'abjuration',
        prepared: true,
      });
    }, hero.id);
    expect(spell2.id).toBeGreaterThan(0);

    // Unprepare all
    const unprepResult = await page.evaluate(async (charId) => {
      return (window as any).api('PUT', `/api/characters/${charId}/spells/prepare`, { spell_ids: [] });
    }, hero.id);
    expect(unprepResult.ok).toBe(true);

    // Prepare only Magic Missile
    const prepResult = await page.evaluate(async ({ charId, spell1Id }) => {
      return (window as any).api('PUT', `/api/characters/${charId}/spells/prepare`, { spell_ids: [spell1Id] });
    }, { charId: hero.id, spell1Id: spell1.id });
    expect(prepResult.ok).toBe(true);

    // Verify by fetching the character
    const updated = await page.evaluate(async (charId) => {
      return (window as any).api('GET', `/api/characters/${charId}`);
    }, hero.id);

    const ms = (updated.spells || []).find((s: any) => s.name === 'Magic Missile');
    expect(ms).toBeDefined();
    expect(ms.prepared).toBe(true);

    const shield = (updated.spells || []).find((s: any) => s.name === 'Shield');
    expect(shield).toBeDefined();
    expect(shield.prepared).toBe(false);
  });
});
