import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT } from './helpers.js';

const uniqueName = () => `Invite-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('Campaign invitations', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('create, view, and revoke an email invitation', async ({ page }) => {
    const campName = uniqueName();
    const email = `invitee-${Date.now()}@example.com`;

    const campId = await page.evaluate(async (name: string) => {
      const camp = await (window as any).api('POST', '/api/campaigns', { name, description: 'invite e2e', dm_notes: '' });
      return camp.id as number;
    }, campName);

    const inv = await page.evaluate(async (opts: any) => {
      return (window as any).api('POST', `/api/campaigns/${opts.campId}/invitations`, { email: opts.email, role: 'player' });
    }, { campId, email });
    expect(inv.token).toBeTruthy();
    expect(inv.url).toContain('/invite/' + inv.token);

    // Pending list contains the invitation.
    const pending = await page.evaluate(async (cid: number) => (window as any).api('GET', `/api/campaigns/${cid}/invitations`), campId);
    expect(pending.some((i: any) => i.email === email)).toBeTruthy();

    // Public invite page renders (logged-in campaign owner is already a member).
    await page.goto('/invite/' + inv.token);
    await expect(page.locator('body')).toContainText(campName, { timeout: NAV_TIMEOUT });
    await expect(page.locator('body')).toContainText("You're invited", { timeout: NAV_TIMEOUT });

    // Return to the SPA (window.api is only defined there) to revoke.
    await page.goto('/');
    await page.waitForFunction(() => typeof (window as any).api === 'function');
    await page.evaluate(async (opts: any) => {
      await (window as any).api('DELETE', `/api/campaigns/${opts.campId}/invitations/${opts.id}`);
    }, { campId, id: inv.id });
    const after = await page.evaluate(async (cid: number) => (window as any).api('GET', `/api/campaigns/${cid}/invitations`), campId);
    expect(after.some((i: any) => i.id === inv.id)).toBeFalsy();
  });
});
