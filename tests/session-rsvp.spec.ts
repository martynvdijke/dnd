import { test, expect } from './fixtures.js';
import { NAV_TIMEOUT, login, waitLoadingDone } from './helpers.js';

test.describe('Session RSVP', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await waitLoadingDone(page);
  });

  test('member can RSVP and the planner shows counts', async ({ page }) => {
    const ids = await page.evaluate(async () => {
      const camp = await (window as any).api('POST', '/api/campaigns', {
        name: 'RSVP Camp', party_name: 'The Party', description: '', dm_notes: '',
      });
      const plan = await (window as any).api('POST', `/api/campaigns/${camp.id}/session-plans`, {
        title: 'Session One', session_date: '2026-10-01', status: 'planned',
        dm_notes: '', planned_encounters: '[]', npc_ids: '[]', player_goals: '[]', expected_duration: 180,
      });
      return { campaignId: camp.id as number, planId: plan.id as number };
    });

    const rsvp = await page.evaluate(async (planId) => {
      await (window as any).api('PUT', `/api/session-plans/${planId}/rsvp`, { status: 'yes', note: '' });
      return (window as any).api('GET', `/api/session-plans/${planId}/rsvps`);
    }, ids.planId);
    expect(rsvp.my_status).toBe('yes');
    expect(rsvp.counts.yes).toBeGreaterThanOrEqual(1);

    await page.evaluate((cid) => (window as any).showSessionPlanner(cid), ids.campaignId);
    await expect(page.locator(`#rsvp-${ids.planId}`)).toContainText('Going', { timeout: NAV_TIMEOUT });
  });
});
