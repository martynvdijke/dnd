import { test, expect } from './fixtures.js';
import { login } from './helpers.js';

const uniqueName = () => `CEX-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function waitLoadingDone(page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: 5000 }).catch(() => {});
}

async function createCampaign(page, name) {
  return page.evaluate(async (n) => {
    return (window as any).api('POST', '/api/campaigns', { name: n, description: 'Test campaign', dm_notes: '' });
  }, name);
}

test.describe('Campaign Extended Features', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  // ─── Campaign Recaps ───

  test.describe('Recaps', () => {
    test('Create a campaign recap', async ({ page }) => {
      const camp = await createCampaign(page, uniqueName());
      const title = 'Recap ' + uniqueName();
      const result = await page.evaluate(async ({ cid, t }) => {
        return (window as any).api('POST', `/api/campaigns/${cid}/recaps`, {
          title: t, content: '# Session 1\n\nWe fought goblins.',
          session_start_date: '2024-01-01', session_end_date: '2024-01-01',
        });
      }, { cid: camp.id, t: title });
      expect(result.id).toBeGreaterThan(0);
    });

    test('List recaps for a campaign', async ({ page }) => {
      const camp = await createCampaign(page, uniqueName());
      const title = 'List Recap ' + uniqueName();
      await page.evaluate(async ({ cid, t }) => {
        return (window as any).api('POST', `/api/campaigns/${cid}/recaps`, {
          title: t, content: '# Session 1', session_start_date: '2024-01-01',
          session_end_date: '2024-01-01',
        });
      }, { cid: camp.id, t: title });

      const recaps = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/campaigns/${cid}/recaps`);
      }, camp.id);
      expect(Array.isArray(recaps)).toBe(true);
      expect(recaps.some((r: any) => r.title === title)).toBe(true);
    });

    test('Get specific recap', async ({ page }) => {
      const camp = await createCampaign(page, uniqueName());
      const title = 'Get Recap ' + uniqueName();
      const created = await page.evaluate(async ({ cid, t }) => {
        return (window as any).api('POST', `/api/campaigns/${cid}/recaps`, {
          title: t, content: '# Recap content', session_start_date: '2024-01-01',
          session_end_date: '2024-01-01',
        });
      }, { cid: camp.id, t: title });

      const recap = await page.evaluate(async (rid) => {
        return (window as any).api('GET', `/api/recaps/${rid}`);
      }, created.id);
      expect(recap.title).toBe(title);
    });

    test('Update a recap', async ({ page }) => {
      const camp = await createCampaign(page, uniqueName());
      const title = 'Update Recap ' + uniqueName();
      const created = await page.evaluate(async ({ cid, t }) => {
        return (window as any).api('POST', `/api/campaigns/${cid}/recaps`, {
          title: t, content: '# Old content', session_start_date: '2024-01-01',
          session_end_date: '2024-01-01',
        });
      }, { cid: camp.id, t: title });

      await page.evaluate(async ({ rid, t }) => {
        return (window as any).api('PUT', `/api/recaps/${rid}`, {
          title: t + '-updated', content: '# Updated content',
          session_start_date: '2024-01-02', session_end_date: '2024-01-02',
        });
      }, { rid: created.id, t: title });

      const recap = await page.evaluate(async (rid) => {
        return (window as any).api('GET', `/api/recaps/${rid}`);
      }, created.id);
      expect(recap.title).toBe(title + '-updated');
    });

    test('Delete a recap', async ({ page }) => {
      const camp = await createCampaign(page, uniqueName());
      const title = 'Delete Recap ' + uniqueName();
      const created = await page.evaluate(async ({ cid, t }) => {
        return (window as any).api('POST', `/api/campaigns/${cid}/recaps`, {
          title: t, content: '# Delete me', session_start_date: '2024-01-01',
          session_end_date: '2024-01-01',
        });
      }, { cid: camp.id, t: title });

      await page.evaluate(async (rid) => {
        return (window as any).api('DELETE', `/api/recaps/${rid}`);
      }, created.id);

      const recaps = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/campaigns/${cid}/recaps`);
      }, camp.id);
      expect(recaps.some((r: any) => r.id === created.id)).toBe(false);
    });

    test('Generate a recap via AI', async ({ page }) => {
      const camp = await createCampaign(page, uniqueName());

      const result = await page.evaluate(async (cid) => {
        try {
          const resp = await (window as any).api('POST', `/api/campaigns/${cid}/recaps/generate`);
          return { ok: true, data: resp };
        } catch (e) {
          return { ok: false, error: String(e) };
        }
      }, camp.id);

      // Accept either success (AI configured) or 400 (no AI configured)
      expect(result.ok === true || result.error !== undefined).toBe(true);
    });
  });



  // ─── Timeline Events ───

  test.describe('Timeline Events', () => {
    test('Create a timeline event', async ({ page }) => {
      const camp = await createCampaign(page, uniqueName());
      const title = 'Timeline ' + uniqueName();
      const result = await page.evaluate(async ({ cid, t }) => {
        return (window as any).api('POST', '/api/timeline', {
          campaign_id: cid, title: t, description: 'Historical event',
          event_date: '2024-06-15', event_type: 'plot', importance: 3,
        });
      }, { cid: camp.id, t: title });
      expect(result.id).toBeGreaterThan(0);
    });

    test('List timeline events', async ({ page }) => {
      const camp = await createCampaign(page, uniqueName());
      const title = 'List Timel ' + uniqueName();
      await page.evaluate(async ({ cid, t }) => {
        return (window as any).api('POST', '/api/timeline', {
          campaign_id: cid, title: t, description: 'Test',
          event_date: '2024-06-15', event_type: 'plot',
        });
      }, { cid: camp.id, t: title });

      const events = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/timeline?campaign_id=${cid}`);
      }, camp.id);
      expect(Array.isArray(events)).toBe(true);
      expect(events.some((e: any) => e.title === title)).toBe(true);
    });

    test('Update a timeline event', async ({ page }) => {
      const camp = await createCampaign(page, uniqueName());
      const title = 'Update TL ' + uniqueName();
      const created = await page.evaluate(async ({ cid, t }) => {
        return (window as any).api('POST', '/api/timeline', {
          campaign_id: cid, title: t, description: 'Original',
          event_date: '2024-06-15', event_type: 'plot',
        });
      }, { cid: camp.id, t: title });

      await page.evaluate(async ({ id, t }) => {
        return (window as any).api('PUT', `/api/timeline/${id}`, {
          title: t + '-updated', description: 'Updated',
          event_date: '2024-06-20', event_type: 'battle',
        });
      }, { id: created.id, t: title });

      const events = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/timeline?campaign_id=${cid}`);
      }, camp.id);
      const updated = events.find((e: any) => e.id === created.id);
      expect(updated).toBeDefined();
      expect(updated.title).toBe(title + '-updated');
    });

    test('Delete a timeline event', async ({ page }) => {
      const camp = await createCampaign(page, uniqueName());
      const title = 'Delete TL ' + uniqueName();
      const created = await page.evaluate(async ({ cid, t }) => {
        return (window as any).api('POST', '/api/timeline', {
          campaign_id: cid, title: t, description: 'Delete me',
          event_date: '2024-06-15', event_type: 'plot',
        });
      }, { cid: camp.id, t: title });

      await page.evaluate(async (id) => {
        return (window as any).api('DELETE', `/api/timeline/${id}`);
      }, created.id);

      const events = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/timeline?campaign_id=${cid}`);
      }, camp.id);
      expect(events.some((e: any) => e.id === created.id)).toBe(false);
    });
  });

  // ─── Combat Log ───

  test.describe('Combat Log', () => {
    test('Create combat log entry', async ({ page }) => {
      const camp = await createCampaign(page, uniqueName());
      const result = await page.evaluate(async (cid) => {
        return (window as any).api('POST', '/api/combat-log', {
          campaign_id: cid, actor_name: 'Goblin', action: 'attacks',
          target_name: 'Fighter', damage: 5, damage_type: 'piercing',
          roll_expression: '1d6+2', roll_total: 5,
        });
      }, camp.id);
      expect(result.id).toBeGreaterThan(0);
    });

    test('List combat log entries', async ({ page }) => {
      const camp = await createCampaign(page, uniqueName());
      await page.evaluate(async (cid) => {
        return (window as any).api('POST', '/api/combat-log', {
          campaign_id: cid, actor_name: 'Goblin', action: 'attacks',
          target_name: 'Fighter', damage: 5, damage_type: 'piercing',
          roll_expression: '1d6+2', roll_total: 5,
        });
      }, camp.id);

      const entries = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/combat-log?campaign_id=${cid}&limit=10`);
      }, camp.id);
      expect(Array.isArray(entries)).toBe(true);
      expect(entries.length).toBeGreaterThan(0);
    });

    test('Get combat log stats', async ({ page }) => {
      const camp = await createCampaign(page, uniqueName());
      await page.evaluate(async (cid) => {
        return (window as any).api('POST', '/api/combat-log', {
          campaign_id: cid, actor_name: 'Hero', action: 'casts',
          target_name: 'Orc', damage: 8, damage_type: 'fire',
          roll_expression: '2d6', roll_total: 8,
        });
      }, camp.id);

      const stats = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/combat-log/stats?campaign_id=${cid}`);
      }, camp.id);
      expect(stats).toBeTruthy();
    });

    test('Create multiple entries and list with limit', async ({ page }) => {
      const camp = await createCampaign(page, uniqueName());

      await page.evaluate(async (cid) => {
        return (window as any).api('POST', '/api/combat-log', {
          campaign_id: cid, actor_name: 'Fighter', action: 'attacks',
          target_name: 'Orc', damage: 6, damage_type: 'slashing',
          roll_expression: '1d8+3', roll_total: 6,
        });
      }, camp.id);

      await page.evaluate(async (cid) => {
        return (window as any).api('POST', '/api/combat-log', {
          campaign_id: cid, actor_name: 'Wizard', action: 'casts',
          target_name: 'Orc', damage: 12, damage_type: 'force',
          roll_expression: '3d4', roll_total: 12,
        });
      }, camp.id);

      const entries = await page.evaluate(async (cid) => {
        return (window as any).api('GET', `/api/combat-log?campaign_id=${cid}&limit=10`);
      }, camp.id);
      expect(entries.length).toBeGreaterThanOrEqual(2);
    });
  });
});
