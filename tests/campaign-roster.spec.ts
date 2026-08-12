import { test, expect } from './fixtures.js';
import { clickNavItem, login, NAV_TIMEOUT } from './helpers.js';

const uniqueName = () => `Roster-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function logoutAndLoginAs(page, username, password) {
  await page.evaluate(async () => { await window.api('POST', '/api/logout'); });
  await page.goto('/login', { waitUntil: 'domcontentloaded' });
  await page.fill('#username', username);
  await page.fill('#password', password);
  await Promise.all([
    page.waitForURL('/', { timeout: NAV_TIMEOUT, waitUntil: 'domcontentloaded' }),
    page.getByTestId('login-submit').click(),
  ]);
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: NAV_TIMEOUT }).catch(() => {});
}

/**
 * Campaign roster flows: creating a campaign lets the user attach their own
 * (player-controlled) characters; the manage view shows the current roster,
 * lets members add external characters (owned by other campaign members) and
 * remove characters, and exposes DM notes on create/edit.
 *
 * Note on "create with external characters": candidates are scoped to own
 * characters + this campaign's members' characters (design D3), and a fresh
 * campaign has only the creator as member, so external characters can only be
 * added after members exist — covered via the manage flow below.
 */
test.describe('Campaign roster management', () => {
  test.beforeEach(async ({ page }) => {
    await login(page); // admin
    await page.waitForTimeout(300);
  });

  test('create campaign with roster selection persists own characters and DM notes', async ({ page }) => {
    const name = uniqueName();
    const charName = name + '-Hero';

    // Admin creates an unassigned character to attach during creation.
    const ch = await page.evaluate(async (cn) => {
      return window.api('POST', '/api/characters', {
        name: cn, race: 'Human', class: 'Fighter',
      });
    }, charName);
    expect(ch.id).toBeGreaterThan(0);

    // Open the create campaign modal from the party view.
    await clickNavItem(page, 'party', 'party');
    await expect(page.locator('#partyContent')).toContainText('Party View', { timeout: NAV_TIMEOUT });
    await page.getByRole('button', { name: 'New Campaign' }).click();
    await page.fill('#newCampaignName', name);
    await page.fill('#newPartyName', 'Dawnbringers');
    await page.fill('#newCampaignDesc', 'Roster test campaign');
    await page.fill('#newCampaignDmNotes', 'Secret dragon lair plan');
    await page.getByRole('button', { name: 'Create' }).click();

    // Step 2: roster picker — the own character is offered player-controlled.
    const picker = page.getByTestId('roster-picker-confirm');
    await expect(picker).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#genericModalBody')).toContainText('Your Characters (player-controlled)');
    await page.locator(`[data-testid="roster-candidate-${ch.id}"]`).getByRole('checkbox').check();
    await picker.click();

    // Campaign + character appear in the party view; DM notes persisted.
    await expect(page.locator('#partyContent')).toContainText(name, { timeout: NAV_TIMEOUT });
    await expect(page.locator('#partyContent')).toContainText(charName, { timeout: NAV_TIMEOUT });
    const saved = await page.evaluate(async () => {
      return window.api('GET', '/api/campaigns');
    });
    const camp = saved.find((c: any) => c.name === name);
    expect(camp).toBeTruthy();
    expect(camp.dm_notes).toBe('Secret dragon lair plan');

    // Character is attached to the campaign (roster endpoint reflects it).
    const roster = await page.evaluate(async (cid) => {
      return window.api('GET', `/api/campaigns/${cid}/characters`);
    }, camp.id);
    expect(roster.some((r: any) => r.id === ch.id)).toBe(true);
  });

  test('create campaign with empty roster is allowed', async ({ page }) => {
    const name = uniqueName();
    await clickNavItem(page, 'party', 'party');
    await expect(page.locator('#partyContent')).toContainText('Party View', { timeout: NAV_TIMEOUT });
    await page.getByRole('button', { name: 'New Campaign' }).click();
    await page.fill('#newCampaignName', name);
    await page.getByRole('button', { name: 'Create' }).click();

    const picker = page.getByTestId('roster-picker-confirm');
    await expect(picker).toBeVisible({ timeout: NAV_TIMEOUT });
    await picker.click(); // no selection — campaign created with empty roster

    // Campaign exists with an empty roster (party view only lists campaigns
    // that have characters, so nothing to assert there).
    const saved = await page.evaluate(async () => {
      return window.api('GET', '/api/campaigns');
    });
    const camp = saved.find((c: any) => c.name === name);
    expect(camp).toBeTruthy();
    const roster = await page.evaluate(async (cid) => {
      return window.api('GET', `/api/campaigns/${cid}/characters`);
    }, camp.id);
    expect(roster.length).toBe(0);
  });

  test('manage view shows roster, adds external member character, removes it, edits DM notes', async ({ page }) => {
    const name = uniqueName();
    const member = 'Member-' + Math.random().toString(36).slice(2, 8);
    const adminChar = name + '-Admin';
    const externalChar = name + '-External';

    // Seed: member user, campaign owned by admin with member added, admin's
    // character already in the roster. Return the ids (the window would be
    // wiped by the logout/reload below).
    const seed = await page.evaluate(async ({ name, member, adminChar }) => {
      await window.api('POST', '/api/admin/users', {
        username: member, password: 'testpassword123', role: 'user',
      });
      const camp = await window.api('POST', '/api/campaigns', { name, party_name: 'Shared Party' });
      await window.api('POST', `/api/campaigns/${camp.id}/members`, { username: member });
      const adminCharRes = await window.api('POST', '/api/characters', {
        name: adminChar, race: 'Dwarf', class: 'Cleric', campaign_id: camp.id,
      });
      return { campaignId: camp.id, adminCharId: adminCharRes.id };
    }, { name, member, adminChar });
    const { campaignId } = seed;

    // Member creates their own (external) character.
    await logoutAndLoginAs(page, member, 'testpassword123');
    const externalCharId = await page.evaluate(async (cn) => {
      const res = await window.api('POST', '/api/characters', {
        name: cn, race: 'Elf', class: 'Wizard',
      });
      return res.id;
    }, externalChar);

    // Back to admin: open the manage modal for the campaign card.
    await logoutAndLoginAs(page, 'admin', 'testpassword123');
    await clickNavItem(page, 'party', 'party');
    await expect(page.locator('#partyContent')).toContainText(name, { timeout: NAV_TIMEOUT });
    const card = page.locator('#partyContent .card').filter({ hasText: name });
    await card.getByRole('button', { name: 'Manage' }).click();

    // Roster section shows the pre-existing admin character (player-controlled).
    let ch = { id: seed.adminCharId };
    await expect(page.getByTestId(`roster-member-${ch.id}`)).toContainText(adminChar, { timeout: NAV_TIMEOUT });
    await expect(page.getByTestId(`roster-member-${ch.id}`)).toContainText('Player-controlled');

    // DM notes field is present (empty for this campaign).
    await expect(page.locator('#editCampaignDmNotes')).toBeVisible({ timeout: NAV_TIMEOUT });

    // Open the picker: the member's character shows under external with the
    // owner's username; the admin character is pre-selected (in roster).
    await page.getByTestId('roster-add-open').click();
    const pickerConfirm = page.getByTestId('roster-picker-confirm');
    await expect(pickerConfirm).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#genericModalBody')).toContainText("Campaign Members' Characters (external)");
    const adminCand = page.locator('[data-testid^="roster-candidate-"]').filter({ hasText: adminChar });
    await expect(adminCand.getByRole('checkbox')).toBeChecked();
    await expect(adminCand).toContainText('In roster');

    const extCand = page.locator('[data-testid^="roster-candidate-"]').filter({ hasText: externalChar });
    await expect(extCand).toBeVisible();
    await expect(extCand).toContainText('External');
    await expect(extCand).toContainText(member);

    // Add the external character and verify it lands in the roster list.
    await extCand.getByRole('checkbox').check();
    await pickerConfirm.click();
    ch = { id: externalCharId };
    await expect(page.getByTestId(`roster-member-${ch.id}`)).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId(`roster-member-${ch.id}`)).toContainText('External');
    await expect(page.getByTestId(`roster-member-${ch.id}`)).toContainText(member);

    // Edit DM notes via the manage form and confirm persistence.
    await page.fill('#editCampaignDmNotes', 'Updated lair notes');
    await page.getByRole('button', { name: 'Save Settings' }).click();
    const saved = await page.evaluate(async () => {
      return window.api('GET', '/api/campaigns');
    });
    const camp = saved.find((c: any) => c.id === campaignId);
    expect(camp.dm_notes).toBe('Updated lair notes');

    // Remove the external character from the roster.
    await page.locator('#partyContent .card').filter({ hasText: name }).getByRole('button', { name: 'Manage' }).click();
    await expect(page.getByTestId(`roster-member-${ch.id}`)).toBeVisible({ timeout: NAV_TIMEOUT });
    page.once('dialog', (d) => d.accept());
    await page.getByTestId(`roster-remove-${ch.id}`).click();
    await expect(page.getByTestId(`roster-member-${ch.id}`)).toHaveCount(0, { timeout: NAV_TIMEOUT });

    const roster = await page.evaluate(async (cid) => {
      return window.api('GET', `/api/campaigns/${cid}/characters`);
    }, campaignId);
    expect(roster.some((r: any) => r.name === externalChar)).toBe(false);
  });
});
