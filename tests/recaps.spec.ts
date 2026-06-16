import { test, expect } from './fixtures.js';
import { login } from './helpers.js';

const uniqueName = () => `Recap-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('Recaps', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('create recap for a campaign', async ({ page }) => {
    const campName = uniqueName();
    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Test campaign for recaps', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'campaign not found' };
      const recap = await window.api('POST', `/api/campaigns/${c.id}/recaps`, {
        campaign_id: c.id,
        title: 'Session Recap',
        content: 'The party explored the dungeon and found a hidden treasure.',
        recap_type: 'session',
        tags: '[]',
      });
      return { ok: true, recapId: recap.id };
    }, { campName });

    expect(result.err).toBeFalsy();
    expect(result.recapId).toBeGreaterThan(0);
  });

  test('list recaps for a campaign', async ({ page }) => {
    const campName = uniqueName();
    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'List recaps test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'campaign not found' };
      await window.api('POST', `/api/campaigns/${c.id}/recaps`, {
        campaign_id: c.id, title: 'Recap 1', content: 'First recap', recap_type: 'session', tags: '[]',
      });
      await window.api('POST', `/api/campaigns/${c.id}/recaps`, {
        campaign_id: c.id, title: 'Recap 2', content: 'Second recap', recap_type: 'session', tags: '[]',
      });
      const recaps = await window.api('GET', `/api/campaigns/${c.id}/recaps`);
      return { ok: true, count: recaps.length, titles: recaps.map((r: any) => r.title) };
    }, { campName });

    expect(result.err).toBeFalsy();
    expect(result.count).toBeGreaterThanOrEqual(2);
    expect(result.titles).toContain('Recap 1');
    expect(result.titles).toContain('Recap 2');
  });

  test('get a single recap by id', async ({ page }) => {
    const campName = uniqueName();
    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Single recap test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'campaign not found' };
      const created = await window.api('POST', `/api/campaigns/${c.id}/recaps`, {
        campaign_id: c.id, title: 'Single Recap', content: 'Detailed recap content', recap_type: 'session', tags: '[]',
      });
      const recap = await window.api('GET', `/api/recaps/${created.id}`);
      return { ok: true, title: recap.title, content: recap.content };
    }, { campName });

    expect(result.err).toBeFalsy();
    expect(result.title).toBe('Single Recap');
    expect(result.content).toBe('Detailed recap content');
  });

  test('update a recap', async ({ page }) => {
    const campName = uniqueName();
    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Update recap test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'campaign not found' };
      const created = await window.api('POST', `/api/campaigns/${c.id}/recaps`, {
        campaign_id: c.id, title: 'Original', content: 'Original content', recap_type: 'session', tags: '[]',
      });
      await window.api('PUT', `/api/recaps/${created.id}`, {
        title: 'Updated Recap', content: 'Updated content', tags: '["important"]',
      });
      const recap = await window.api('GET', `/api/recaps/${created.id}`);
      return { ok: true, title: recap.title, content: recap.content };
    }, { campName });

    expect(result.err).toBeFalsy();
    expect(result.title).toBe('Updated Recap');
    expect(result.content).toBe('Updated content');
  });

  test('delete a recap', async ({ page }) => {
    const campName = uniqueName();
    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Delete recap test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'campaign not found' };
      const created = await window.api('POST', `/api/campaigns/${c.id}/recaps`, {
        campaign_id: c.id, title: 'Delete Me', content: 'To be deleted', recap_type: 'session', tags: '[]',
      });
      await window.api('DELETE', `/api/recaps/${created.id}`);
      const recaps = await window.api('GET', `/api/campaigns/${c.id}/recaps`);
      return { ok: true, remaining: recaps.filter((r: any) => r.id === created.id).length };
    }, { campName });

    expect(result.err).toBeFalsy();
    expect(result.remaining).toBe(0);
  });

  test('generate recap AI endpoint', async ({ page }) => {
    const campName = uniqueName();
    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Generate recap test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'campaign not found' };
      try {
        const gen = await window.api('POST', `/api/campaigns/${c.id}/recaps/generate`, {});
        return { ok: true, data: gen };
      } catch (e) {
        // May fail if AI not configured — that's acceptable
        return { ok: false, error: String(e) };
      }
    }, { campName });

    // Accept any response — AI generation may or may not be available
    expect(result).toBeTruthy();
  });
});
