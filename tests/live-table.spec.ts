import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT, clickNavItem, waitLoadingDone } from './helpers.js';

const uniqueName = () => `LiveTable-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('Live Table', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('table view renders when enabled', async ({ page }) => {
    const campName = uniqueName();
    const charName = uniqueName();
    const campId = await page.evaluate(async (opts) => {
      const c = await window.api('POST', '/api/campaigns', { name: opts.campName, description: 'lt', dm_notes: '' });
      await window.api('POST', '/api/characters', { name: opts.charName, race: 'Human', class: 'Fighter', level: 1, campaign_ids: [c.id], hp_max: 10, hp_current: 10, ac: 15 });
      return c.id;
    }, { campName, charName });

    await page.evaluate((cid) => window.showTable(cid), campId);
    await expect(page.locator('[data-testid="table-view"]')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('[data-testid="table-roll-btn"]')).toBeVisible();
    await expect(page.locator('[data-testid="table-dice-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="table-roll-feed"]')).toBeAttached();
    await expect(page.locator('[data-testid="table-initiative"]')).toBeAttached();
    await expect(page.locator('[data-testid="table-handouts"]')).toBeAttached();
    await page.evaluate(() => window.hideTable());
    await expect(page.locator('[data-testid="table-view"]')).toBeHidden();
  });

  test('roll shows in the feed', async ({ page }) => {
    const campName = uniqueName();
    const charName = uniqueName();
    const campId = await page.evaluate(async (opts) => {
      const c = await window.api('POST', '/api/campaigns', { name: opts.campName, description: 'lt', dm_notes: '' });
      await window.api('POST', '/api/characters', { name: opts.charName, race: 'Human', class: 'Fighter', level: 1, campaign_ids: [c.id], hp_max: 10, hp_current: 10, ac: 15 });
      return c.id;
    }, { campName, charName });

    await page.evaluate((cid) => window.showTable(cid), campId);
    await expect(page.locator('[data-testid="table-view"]')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.locator('[data-testid="table-dice-input"]').fill('1d20');
    await page.locator('[data-testid="table-roll-btn"]').click();
    await expect(page.locator('[data-testid="table-roll-feed"]')).toContainText('1d20', { timeout: NAV_TIMEOUT });
  });

  test('shared handout appears but unshared does not', async ({ page }) => {
    const campName = uniqueName();
    const charName = uniqueName();
    const sharedTitle = `Shared-${uniqueName()}`;
    const hiddenTitle = `Hidden-${uniqueName()}`;
    const campId = await page.evaluate(async (opts) => {
      const c = await window.api('POST', '/api/campaigns', { name: opts.campName, description: 'lt', dm_notes: '' });
      await window.api('POST', '/api/characters', { name: opts.charName, race: 'Human', class: 'Fighter', level: 1, campaign_ids: [c.id], hp_max: 10, hp_current: 10, ac: 15 });
      await window.api('POST', `/api/campaigns/${c.id}/knowledge`, { title: opts.sharedTitle, content: 'visible', source: '', status: 'revealed', shared: true });
      await window.api('POST', `/api/campaigns/${c.id}/knowledge`, { title: opts.hiddenTitle, content: 'secret', source: '', status: 'rumor', shared: false });
      return c.id;
    }, { campName, charName, sharedTitle, hiddenTitle });

    await page.evaluate((cid) => window.showTable(cid), campId);
    await expect(page.locator('[data-testid="table-view"]')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.evaluate(() => window.refreshTableHandouts());
    await expect(page.locator('[data-testid="table-handouts"]')).toContainText(sharedTitle, { timeout: NAV_TIMEOUT });
    await expect(page.locator('[data-testid="table-handouts"]')).not.toContainText(hiddenTitle);
  });

  test('docked HUD keeps the character sheet usable in session mode', async ({ page }) => {
    const campName = uniqueName();
    const charName = uniqueName();
    await page.evaluate(async (opts) => {
      const c = await window.api('POST', '/api/campaigns', { name: opts.campName, description: 'lt', dm_notes: '' });
      await window.api('POST', '/api/characters', { name: opts.charName, race: 'Human', class: 'Fighter', level: 1, campaign_ids: [c.id], hp_max: 10, hp_current: 10, ac: 15 });
    }, { campName, charName });

    await clickNavItem(page, 'characters', 'characters');
    await waitLoadingDone(page);
    await page.locator('.character-card, .char-card, [onclick*="showSheet"]', { hasText: charName }).first().click();
    await waitLoadingDone(page);

    await page.evaluate(() => (window as unknown as { activateSessionMode: () => void }).activateSessionMode());

    const table = page.locator('[data-testid="table-view"]');
    await expect(table).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(table).toHaveClass(/table-view-docked/);

    // The sheet's session quick-action bar is no longer covered by the table view.
    await expect(page.locator('#sheetQuickActions')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#sheetQuickActions').getByText('Damage')).toBeVisible();

    // The sheet stays interactive while the HUD is docked.
    await page.locator('#tab-combat').click();
    await expect(page.locator('#combatSection')).toBeVisible({ timeout: NAV_TIMEOUT });

    // Docked <-> expanded focus toggle.
    await page.locator('[data-testid="table-view-dock-toggle"]').click();
    await expect(table).toHaveClass(/table-view-expanded/);
    await page.locator('[data-testid="table-view-dock-toggle"]').click();
    await expect(table).not.toHaveClass(/table-view-expanded/);

    await page.evaluate(() => (window as unknown as { deactivateSessionMode: () => void }).deactivateSessionMode());
    await expect(table).toBeHidden();
  });

  test('dice preset buttons roll from the table HUD', async ({ page }) => {
    const campName = uniqueName();
    const charName = uniqueName();
    const campId = await page.evaluate(async (opts) => {
      const c = await window.api('POST', '/api/campaigns', { name: opts.campName, description: 'lt', dm_notes: '' });
      await window.api('POST', '/api/characters', { name: opts.charName, race: 'Human', class: 'Fighter', level: 1, campaign_ids: [c.id], hp_max: 10, hp_current: 10, ac: 15 });
      return c.id;
    }, { campName, charName });

    await page.evaluate((cid) => window.showTable(cid), campId);
    await expect(page.locator('[data-testid="table-view"]')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.locator('[data-testid="table-dice-presets"] button', { hasText: 'd20' }).click();
    await expect(page.locator('[data-testid="table-roll-feed"]')).toContainText('1d20', { timeout: NAV_TIMEOUT });
  });

  test('DM reveal reaches a member table live', async ({ browser }) => {
    const campName = uniqueName();
    const charName = uniqueName();
    const member = `lt${Date.now().toString(36)}`.slice(0, 12);
    const password = 'testpassword123';

    const adminCtx = await browser.newContext();
    const adminPage = await adminCtx.newPage();
    await login(adminPage);
    const campId = await adminPage.evaluate(async (opts) => {
      const c = await window.api('POST', '/api/campaigns', { name: opts.campName, description: 'lt', dm_notes: '' });
      await window.api('POST', '/api/admin/users', { username: opts.member, password: opts.password, role: 'user' });
      await window.api('POST', `/api/campaigns/${c.id}/members`, { username: opts.member });
      return c.id;
    }, { campName, charName, member, password });

    const memberCtx = await browser.newContext();
    const memberPage = await memberCtx.newPage();
    await memberPage.goto('/login', { waitUntil: 'domcontentloaded' });
    await memberPage.fill('#username', member);
    await memberPage.fill('#password', password);
    await Promise.all([
      memberPage.waitForURL('/', { timeout: 60000, waitUntil: 'domcontentloaded' }),
      memberPage.getByTestId('login-submit').click(),
    ]);
    await memberPage.waitForFunction(() => (window as any).__wsReady === true, { timeout: 60000 });
    await memberPage.evaluate(async (opts) => {
      await window.api('POST', '/api/characters', { name: opts.charName, race: 'Human', class: 'Fighter', level: 1, campaign_ids: [opts.cid], hp_max: 10, hp_current: 10, ac: 15 });
    }, { charName, cid: campId });
    await memberPage.evaluate((cid) => window.showTable(cid), campId);
    await expect(memberPage.locator('[data-testid="table-view"]')).toBeVisible({ timeout: NAV_TIMEOUT });

    const title = `Reveal-${uniqueName()}`;
    await adminPage.evaluate(async (opts) => {
      await window.api('POST', `/api/campaigns/${opts.cid}/knowledge`, { title: opts.title, content: 'secret', source: '', status: 'rumor', shared: false });
      const list: any[] = await window.api('GET', `/api/campaigns/${opts.cid}/knowledge`);
      const k = list.find((e: any) => e.title === opts.title);
      await window.api('POST', `/api/knowledge/${k.id}/reveal`, {});
    }, { cid: campId, title });

    await expect(memberPage.locator('[data-testid="table-handouts"]')).toContainText(title, { timeout: NAV_TIMEOUT });

    await adminCtx.close();
    await memberCtx.close();
  });
});
