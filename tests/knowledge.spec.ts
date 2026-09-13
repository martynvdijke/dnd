import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT } from './helpers.js';

const uniqueName = () => `Know-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function logoutAndLoginAs(page: any, username: string, password: string) {
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

test.describe('Party Knowledge', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('create campaign and knowledge entry via API then list renders', async ({ page }) => {
    const campName = uniqueName();
    const title = 'Ancient Secret ' + uniqueName();

    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Knowledge test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'campaign not found' } as any;
      const k = await window.api('POST', `/api/campaigns/${c.id}/knowledge`, {
        title: opts.title,
        content: '<p>The dragon sleeps beneath the mountain</p>',
        source: 'Old tome',
        status: 'rumor',
      });
      return { ok: true, campId: c.id, kid: k.id };
    }, { campName, title });

    expect(result.err).toBeFalsy();

    await page.evaluate((cid) => window.showKnowledge(cid), result.campId);

    await expect(page.locator('#knowledgeContent')).toContainText(title, { timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('knowledge-list')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('knowledge-card').first()).toBeVisible({ timeout: NAV_TIMEOUT });
    // also assert status filter select is rendered
    await expect(page.getByTestId('knowledge-status-filter')).toBeVisible({ timeout: NAV_TIMEOUT });
  });

  test('detail rendering shows manage controls for owner', async ({ page }) => {
    const campName = uniqueName();
    const title = 'Detail Entry ' + uniqueName();
    const content = 'Hidden lore content ' + uniqueName();

    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Detail test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'not found' } as any;
      const k = await window.api('POST', `/api/campaigns/${c.id}/knowledge`, {
        title: opts.title,
        content: opts.content,
        source: 'Scout report',
        status: 'confirmed',
      });
      return { campId: c.id, kid: k.id };
    }, { campName, title, content });

    await page.evaluate((cid) => window.showKnowledge(cid), result.campId);
    await expect(page.locator('#knowledgeContent')).toContainText(title, { timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('knowledge-card').first()).toBeVisible({ timeout: NAV_TIMEOUT });

    // click card via onclick -> loadKnowledge
    await page.evaluate((kid) => window.loadKnowledge(kid), result.kid);

    await expect(page.getByTestId('knowledge-detail')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('knowledge-detail')).toContainText(content, { timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('share-toggle')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('bulk-reveal-btn')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('known-by-picker')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('entity-link-picker')).toBeVisible({ timeout: NAV_TIMEOUT });
  });

  test('status filter hides non-matching cards', async ({ page }) => {
    const campName = uniqueName();
    const titleRumor = 'Rumor Entry ' + uniqueName();
    const titleFalse = 'False Entry ' + uniqueName();

    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Filter test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'not found' } as any;
      await window.api('POST', `/api/campaigns/${c.id}/knowledge`, {
        title: opts.titleRumor, content: 'rumor content', source: '', status: 'rumor',
      });
      await window.api('POST', `/api/campaigns/${c.id}/knowledge`, {
        title: opts.titleFalse, content: 'false content', source: '', status: 'false',
      });
      await window.api('POST', `/api/campaigns/${c.id}/knowledge`, {
        title: 'Confirmed One', content: 'confirmed', source: '', status: 'confirmed',
      });
      return { campId: c.id };
    }, { campName, titleRumor, titleFalse });

    await page.evaluate((cid) => window.showKnowledge(cid), result.campId);
    await expect(page.locator('#knowledgeContent')).toContainText(titleRumor, { timeout: NAV_TIMEOUT });
    await expect(page.locator('#knowledgeContent')).toContainText(titleFalse, { timeout: NAV_TIMEOUT });

    // select "false" status
    await page.getByTestId('knowledge-status-filter').selectOption('false');

    // client-side filtering: only false card visible
    await expect(page.locator('[data-testid="knowledge-card"]').filter({ hasText: titleFalse })).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('[data-testid="knowledge-card"]').filter({ hasText: titleRumor })).toHaveCount(0, { timeout: NAV_TIMEOUT });

    // reset to All
    await page.getByTestId('knowledge-status-filter').selectOption('');
    await expect(page.locator('[data-testid="knowledge-card"]').filter({ hasText: titleRumor })).toBeVisible({ timeout: NAV_TIMEOUT });
  });

  test('visibility: member sees only shared knowledge', async ({ page }) => {
    const campName = uniqueName();
    const member = 'KnowMember-' + Math.random().toString(36).slice(2, 8);
    const unsharedTitle = 'Unshared Secret ' + uniqueName();
    const sharedTitle = 'Shared Secret ' + uniqueName();

    // create member user and campaign via admin
    const seed = await page.evaluate(async (opts) => {
      await window.api('POST', '/api/admin/users', { username: opts.member, password: 'testpassword123', role: 'user' });
      const camp = await window.api('POST', '/api/campaigns', { name: opts.campName, description: 'Visibility test', dm_notes: '' });
      await window.api('POST', `/api/campaigns/${camp.id}/members`, { username: opts.member });
      const unshared = await window.api('POST', `/api/campaigns/${camp.id}/knowledge`, {
        title: opts.unsharedTitle,
        content: 'dm only content',
        source: '',
        status: 'rumor',
      });
      const shared = await window.api('POST', `/api/campaigns/${camp.id}/knowledge`, {
        title: opts.sharedTitle,
        content: 'party content',
        source: '',
        status: 'rumor',
      });
      // mark second as shared
      await window.api('PUT', `/api/knowledge/${shared.id}`, { shared: true });
      return { campId: camp.id, unsharedId: unshared.id, sharedId: shared.id };
    }, { campName, member, unsharedTitle, sharedTitle });

    // login as member
    await logoutAndLoginAs(page, member, 'testpassword123');

    await page.evaluate((cid) => window.showKnowledge(cid), seed.campId);
    await expect(page.getByTestId('knowledge-list')).toBeVisible({ timeout: NAV_TIMEOUT });
    // shared visible, unshared not visible
    await expect(page.locator('#knowledgeContent')).toContainText(sharedTitle, { timeout: NAV_TIMEOUT });
    await expect(page.locator('#knowledgeContent')).not.toContainText(unsharedTitle, { timeout: NAV_TIMEOUT });

    // verify via API as member: list only returns shared
    const listCheck = await page.evaluate(async (campId) => {
      const list = await window.api('GET', `/api/campaigns/${campId}/knowledge`);
      return list.map((k: any) => k.title);
    }, seed.campId);
    expect(listCheck).toContain(sharedTitle);
    expect(listCheck).not.toContain(unsharedTitle);

    // back to admin for cleanup
    await logoutAndLoginAs(page, 'admin', 'testpassword123');
  });

  test('bulk reveal marks shared and known-by', async ({ page }) => {
    const campName = uniqueName();
    const title = 'Reveal Me ' + uniqueName();

    await page.evaluate(async (name) => {
      await window.api('POST', '/api/campaigns', { name, description: 'Bulk reveal test', dm_notes: '' });
    }, campName);

    const result = await page.evaluate(async (opts) => {
      const camps = await window.api('GET', '/api/campaigns');
      const c = camps.find((x: any) => x.name === opts.campName);
      if (!c) return { err: 'not found' } as any;
      // ensure at least one character exists for known-by assertion
      const ch = await window.api('POST', '/api/characters', {
        name: 'RevealChar-' + Math.random().toString(36).slice(2, 6),
        race: 'Human', class: 'Fighter', campaign_id: c.id,
      });
      const k = await window.api('POST', `/api/campaigns/${c.id}/knowledge`, {
        title: opts.title, content: 'bulk reveal content', source: '', status: 'rumor',
      });
      return { campId: c.id, kid: k.id, charId: ch.id };
    }, { campName, title });

    await page.evaluate((cid) => window.showKnowledge(cid), result.campId);
    await expect(page.locator('#knowledgeContent')).toContainText(title, { timeout: NAV_TIMEOUT });

    await page.evaluate((kid) => window.loadKnowledge(kid), result.kid);
    await expect(page.getByTestId('knowledge-detail')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('bulk-reveal-btn')).toBeVisible({ timeout: NAV_TIMEOUT });

    await page.evaluate((kid) => { (window as any).__bulkKid = kid; }, result.kid);

    // handle confirm dialog then click bulk reveal
    page.once('dialog', (d) => d.accept());
    await page.getByTestId('bulk-reveal-btn').click();

    // wait for API side-effect: shared flag
    await page.waitForFunction(async () => {
      try {
        const kid = (window as any).__bulkKid;
        if (!kid) return false;
        const k = await (window as any).api('GET', `/api/knowledge/${kid}`);
        return k.shared === true;
      } catch { return false; }
    }, { timeout: NAV_TIMEOUT }).catch(() => {});

    const verify = await page.evaluate(async (kid) => {
      const k = await window.api('GET', `/api/knowledge/${kid}`);
      const known = await window.api('GET', `/api/knowledge/${kid}/known-by`).catch(() => []);
      return { shared: k.shared, knownCount: Array.isArray(known) ? known.length : 0 };
    }, result.kid);

    expect(verify.shared).toBe(true);
    // if character creation succeeded, known-by should be non-empty; otherwise at least shared flag suffices
    expect(verify.knownCount).toBeGreaterThanOrEqual(1);
  });
});
