import { test, expect } from './fixtures.js';
import { NAV_TIMEOUT, clickNavItem, login, waitLoadingDone, waitModalClosed } from './helpers.js';

const uniqueName = () => `Sel-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;




/**
 * Create a member user + campaign + character via the admin API, then log in
 * as the member. Returns { username, password, campaignId, charId }.
 */
async function setupCampaignForMember(page, memberName) {
  const password = 'testpassword123';
  const user = await page.evaluate(async ({ name, pwd }) => {
    const res = await window.api('POST', '/api/admin/users', {
      username: name, password: pwd, role: 'user',
    });
    return res;
  }, { name: memberName, pwd: password });

  const camp = await page.evaluate(async (n) => {
    return window.api('POST', '/api/campaigns', { name: n, party_name: 'Test Party' });
  }, `Campaign ${memberName}`);

  // Add the member (player role) to the campaign
  await page.evaluate(async ({ cid, username }) => {
    await window.api('POST', `/api/campaigns/${cid}/members`, { username });
  }, { cid: camp.id, username: memberName });

  // Admin creates a character in the campaign (owned by admin, shared for the member)
  const ch = await page.evaluate(async ({ cid }) => {
    return window.api('POST', '/api/characters', {
      name: 'Shared Hero', race: 'Human', class: 'Fighter', campaign_ids: [cid],
    });
  }, { cid: camp.id });

  return { username: memberName, password, campaignId: camp.id, charId: ch.id };
}

async function logoutAndLoginAs(page, username, password) {
  await page.evaluate(async () => { await window.api('POST', '/api/logout'); });
  await page.goto('/login', { waitUntil: 'domcontentloaded' });
  await page.fill('#username', username);
  await page.fill('#password', password);
  await Promise.all([
    page.waitForURL('/', { timeout: NAV_TIMEOUT, waitUntil: 'domcontentloaded' }),
    page.getByTestId('login-submit').click(),
  ]);
  await waitLoadingDone(page);
}

test.describe('Campaign-first character selection', () => {
  test.beforeEach(async ({ page }) => {
    await login(page); // admin
  });

  test('member sees campaign picker, then character picker, then sheet', async ({ page }) => {
    const member = uniqueName();
    const setup = await setupCampaignForMember(page, member);
    await logoutAndLoginAs(page, setup.username, setup.password);

    // Campaign picker is shown first
    await expect(page.getByTestId('campaign-picker-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('[data-testid="campaign-picker-card"]').first()).toContainText('Campaign');

    // Select the campaign → character picker
    await page.locator('[data-testid="campaign-picker-card"]').first().click();
    await expect(page.getByTestId('character-picker-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    const sharedCard = page.locator('[data-testid="character-picker-card"]').filter({ hasText: 'Shared Hero' });
    await expect(sharedCard).toBeVisible();
    await expect(sharedCard).toContainText('Shared');

    // Select the character → sheet opens in read-only mode
    await sharedCard.click();
    await expect(page.getByTestId('sheet-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#sheetView.readonly')).toBeVisible({ timeout: NAV_TIMEOUT });
  });

  test('shared character sheet is read-only, owned character is editable', async ({ page }) => {
    const member = uniqueName();
    const setup = await setupCampaignForMember(page, member);
    await logoutAndLoginAs(page, setup.username, setup.password);

    // Select the campaign through the picker first
    await expect(page.getByTestId('campaign-picker-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.locator('[data-testid="campaign-picker-card"]').first().click();
    await expect(page.getByTestId('character-picker-view')).toBeVisible({ timeout: NAV_TIMEOUT });

    // Member creates their own character in the campaign (owned)
    const own = await page.evaluate(async ({ cid }) => {
      return window.api('POST', '/api/characters', {
        name: 'Owned Hero', race: 'Elf', class: 'Wizard', campaign_ids: [cid],
      });
    }, { cid: setup.campaignId });

    // Open the character picker again and select the owned character
    await page.evaluate(() => (window as any).switchCharacter?.());
    await expect(page.getByTestId('character-picker-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    const ownedCard = page.locator('[data-testid="character-picker-card"]').filter({ hasText: 'Owned Hero' });
    await expect(ownedCard).toContainText('Owned');
    await ownedCard.click();
    await expect(page.getByTestId('sheet-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#sheetView.readonly')).toHaveCount(0, { timeout: NAV_TIMEOUT });

    // Character folio shows ownership badges
    await clickNavItem(page, 'characters', 'characters');
    await expect(page.locator('[data-testid="character-card"]').filter({ hasText: 'Owned Hero' })).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('[data-testid="character-card"]').filter({ hasText: 'Shared Hero' })).toContainText('Shared');

    // Campaign context button shows the current campaign
    await expect(page.getByTestId('campaign-context-btn')).toContainText('Campaign');
  });

  test('member cannot write to a shared character via API (403)', async ({ page }) => {
    const member = uniqueName();
    const setup = await setupCampaignForMember(page, member);
    await logoutAndLoginAs(page, setup.username, setup.password);

    const status = await page.evaluate(async ({ cid, username }) => {
      // The member's API token is stored by the SPA under a user-scoped key.
      const apiToken = localStorage.getItem(`villum-api-token-${username}`) || '';
      const res = await fetch(`/api/characters/${cid}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || '',
          ...(apiToken ? { 'Authorization': `Bearer ${apiToken}` } : {}),
        },
        credentials: 'include',
        body: JSON.stringify({ name: 'Hijacked' }),
      });
      return res.status;
    }, { cid: setup.charId, username: setup.username });
    expect(status).toBe(403);
  });

  test('own campaign-less characters appear under Unassigned in the picker', async ({ page }) => {
    const member = uniqueName();
    const setup = await setupCampaignForMember(page, member);
    await logoutAndLoginAs(page, setup.username, setup.password);

    // Member creates a character NOT assigned to any campaign
    await page.evaluate(async () => {
      return window.api('POST', '/api/characters', {
        name: 'Lone Wolf', race: 'Halfling', class: 'Rogue',
      });
    });

    // Select the campaign → picker should show the Unassigned group with the character
    await expect(page.getByTestId('campaign-picker-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.locator('[data-testid="campaign-picker-card"]').first().click();
    await expect(page.getByTestId('character-picker-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('[data-testid="character-picker-card"]').filter({ hasText: 'Lone Wolf' })).toBeVisible();
    await expect(page.locator('#characterPickerList')).toContainText('Unassigned');

    // Selecting it opens the sheet (owned)
    await page.locator('[data-testid="character-picker-card"]').filter({ hasText: 'Lone Wolf' }).click();
    await expect(page.getByTestId('sheet-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#sheetView.readonly')).toHaveCount(0, { timeout: NAV_TIMEOUT });
  });

  test('user with no campaigns creates one from the empty picker and lands in Party View', async ({ page }) => {
    const fresh = uniqueName();
    await page.evaluate(async ({ name }) => {
      await window.api('POST', '/api/admin/users', { username: name, password: 'testpassword123', role: 'user' });
    }, { name: fresh });
    await logoutAndLoginAs(page, fresh, 'testpassword123');

    await expect(page.getByTestId('campaign-picker-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.getByTestId('create-campaign-from-picker').click();
    await page.fill('#newCampaignName', `Fresh ${fresh}`);
    await page.locator('.modal button:has-text("Create")').click();
    await expect(page.getByTestId('roster-picker-confirm')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.getByTestId('roster-picker-confirm').click();

    // The new campaign is adopted as DM and the Party View opens (the roster
    // is empty, so the Party View shows its empty state).
    await expect(page.locator('#partyView')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('character-picker-view')).toBeHidden();
    const stored = await page.evaluate(() => JSON.parse(localStorage.getItem('villum_campaign') || 'null'));
    expect(stored?.name).toBe(`Fresh ${fresh}`);
  });

  test('DM selection skips the character picker; Continue as DM returns to Party View', async ({ page }) => {
    const dm = uniqueName();
    await page.evaluate(async ({ name }) => {
      await window.api('POST', '/api/admin/users', { username: name, password: 'testpassword123', role: 'user' });
    }, { name: dm });
    await logoutAndLoginAs(page, dm, 'testpassword123');

    // The fresh DM creates a campaign they run, then re-opens the picker.
    await page.evaluate(async () => {
      await window.api('POST', '/api/campaigns', { name: 'DM Campaign', party_name: 'DM Party' });
    });
    await page.evaluate(() => (window as any).loadCampaignPicker());
    await expect(page.getByTestId('campaign-picker-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.locator('[data-testid="campaign-picker-card"]').first().click();

    // DM entry goes straight to Party View — no character step.
    await expect(page.locator('#partyView')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('character-picker-view')).toBeHidden();

    // Switching character for a DM offers "Continue as DM" and goes back to Party View.
    await page.evaluate(() => (window as any).switchCharacter());
    await expect(page.getByTestId('character-picker-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('continue-as-dm')).toBeVisible();
    await page.getByTestId('continue-as-dm').click();
    await expect(page.locator('#partyView')).toBeVisible({ timeout: NAV_TIMEOUT });
  });

  test('multi-campaign character appears in both pickers; detaching one keeps the other', async ({ page }) => {
    const member = uniqueName();
    const setup = await setupCampaignForMember(page, member);

    // Second campaign run by the same admin, same member, same character attached.
    await page.evaluate(async ({ username, charId }) => {
      const camp = await window.api('POST', '/api/campaigns', { name: 'Second Campaign', party_name: 'Party B' });
      await window.api('POST', `/api/campaigns/${camp.id}/members`, { username });
      await window.api('POST', `/api/campaigns/${camp.id}/characters`, { character_id: charId });
    }, { username: member, charId: setup.charId });

    await logoutAndLoginAs(page, member, 'testpassword123');

    // Both campaigns show the same shared character.
    await expect(page.getByTestId('campaign-picker-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.locator('[data-testid="campaign-picker-card"]').filter({ hasText: `Campaign ${member}` }).click();
    await expect(page.locator('[data-testid="character-picker-card"]').filter({ hasText: 'Shared Hero' })).toBeVisible();
    await page.evaluate(() => (window as any).loadCampaignPicker());
    await expect(page.getByTestId('campaign-picker-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.locator('[data-testid="campaign-picker-card"]').filter({ hasText: 'Second Campaign' }).click();
    await expect(page.locator('[data-testid="character-picker-card"]').filter({ hasText: 'Shared Hero' })).toBeVisible();

    // Detach from the first campaign only.
    await page.evaluate(async ({ cid, charId }) => {
      await window.api('DELETE', `/api/campaigns/${cid}/characters/${charId}`);
    }, { cid: setup.campaignId, charId: setup.charId });

    await page.evaluate(() => (window as any).loadCampaignPicker());
    await page.locator('[data-testid="campaign-picker-card"]').filter({ hasText: `Campaign ${member}` }).click();
    await expect(page.locator('[data-testid="character-picker-card"]').filter({ hasText: 'Shared Hero' })).toHaveCount(0);
    await page.evaluate(() => (window as any).loadCampaignPicker());
    await page.locator('[data-testid="campaign-picker-card"]').filter({ hasText: 'Second Campaign' }).click();
    await expect(page.locator('[data-testid="character-picker-card"]').filter({ hasText: 'Shared Hero' })).toBeVisible();
  });

  test('wizard and sheet expose campaign membership controls', async ({ page }) => {
    const tag = uniqueName();
    const camp = await page.evaluate(async ({ n }) => {
      return window.api('POST', '/api/campaigns', { name: n, party_name: 'Membership Party' });
    }, { n: `Membership ${tag}` });
    const ch = await page.evaluate(async ({ cid, n }) => {
      return window.api('POST', '/api/characters', {
        name: n, race: 'Human', class: 'Fighter', campaign_ids: [cid],
      });
    }, { cid: camp.id, n: `Member Hero ${tag}` });

    // Sheet shows the editable campaigns section for the owner.
    await page.evaluate((id: number) => (window as any).openChar(id), ch.id);
    await expect(page.getByTestId('sheet-view')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('character-campaigns-section')).toBeVisible();
    await expect(page.getByTestId('membership-campaign-option').first()).toBeVisible();

    // Character folio renders the campaign badge.
    await clickNavItem(page, 'characters', 'characters');
    await expect(page.locator('[data-testid="character-card"]').filter({ hasText: `Member Hero ${tag}` })).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('character-campaign-badge').first()).toBeVisible();

    // Wizard review step lists campaigns to join.
    await page.getByTestId('new-character').click();
    await page.getByTestId('guided-builder').click();
    await expect(page.getByTestId('character-wizard')).toBeVisible();
    await page.fill('#wizName', `Wiz ${tag}`);
    await page.getByTestId('wizard-next').click();
    await page.getByTestId('wizard-next').click();
    await page.getByTestId('wizard-next').click();
    await expect(page.getByTestId('wizard-campaigns')).toBeVisible();
    await expect(page.getByTestId('wizard-campaign-option').first()).toBeVisible();
  });
});

test.describe('FAB cleanup', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('FAB menu no longer contains Generate with AI', async ({ page }) => {
    await page.evaluate(() => (window as any).toggleFabMenu?.());
    await expect(page.locator('#fabMenu')).toContainText('New Character');
    await expect(page.locator('#fabMenu')).not.toContainText('Generate with AI');
  });
});
