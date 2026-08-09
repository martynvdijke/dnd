import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT, clickNavItem } from './helpers.js';

const uniqueName = () => `Sel-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function waitLoadingDone(page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: NAV_TIMEOUT }).catch(() => {});
}

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
      name: 'Shared Hero', race: 'Human', class: 'Fighter', campaign_id: cid,
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
        name: 'Owned Hero', race: 'Elf', class: 'Wizard', campaign_id: cid,
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

    const status = await page.evaluate(async ({ cid }) => {
      const res = await fetch(`/api/characters/${cid}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || '',
        },
        credentials: 'include',
        body: JSON.stringify({ name: 'Hijacked' }),
      });
      return res.status;
    }, { cid: setup.charId });
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
