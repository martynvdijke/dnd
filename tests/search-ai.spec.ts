import { test, expect } from './fixtures.js';
import { LOGIN_TIMEOUT, NAV_TIMEOUT, login } from './helpers.js';

async function waitForSearchOverlay(page) {
  await page.waitForFunction(() => {
    const overlay = document.getElementById('searchOverlay');
    return overlay && overlay.style.display !== 'none';
  }, { timeout: NAV_TIMEOUT });
}

async function openPalette(page) {
  await page.evaluate(() => {
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true, bubbles: true }));
  });
  await waitForSearchOverlay(page);
}

async function fillQueryAndSearch(page, query) {
  await page.evaluate((q) => {
    const input = document.getElementById('cpSearchInput') as HTMLInputElement | null;
    if (input) input.value = q;
    if (input) input.dispatchEvent(new Event('input', { bubbles: true }));
  }, query);
  await page.evaluate((q) => window.doSearch(q), query);
}

test.describe('Search AI panel', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
  });

  test('toggle is disabled until query exists, enabled after query', async ({ page }) => {
    await openPalette(page);
    const toggle = page.getByTestId('search-ai-toggle');
    await expect(toggle).toBeVisible();
    await expect(toggle).toBeDisabled();
    await fillQueryAndSearch(page, 'fireball');
    await expect(page.locator('#cpResults')).toContainText(/Fireball|No Results/i, { timeout: NAV_TIMEOUT });
    await expect(toggle).toBeEnabled();
  });

  test('clicking Ask AI shows loading then renders 200 answer and sources', async ({ page }) => {
    let capturedBody: { query?: string } | undefined;
    await page.route('**/api/search/ai', async (route) => {
      const req = route.request();
      expect(req.method()).toBe('POST');
      const body = JSON.parse(req.postData() || '{}');
      expect(body.query).toBeTruthy();
      capturedBody = body;
      await new Promise((r) => setTimeout(r, 150));
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          answer: 'Fireball is a level 3 evocation spell.',
          sources: [{ kind: 'spell', id: 1, title: 'Fireball', subtitle: 'Level 3 Evocation', url: '/compendium/spell/1', snippet: 'A bright streak...' }],
          scope: 'compendium',
        }),
      });
    });
    await openPalette(page);
    await fillQueryAndSearch(page, 'what is fireball');
    const toggle = page.getByTestId('search-ai-toggle');
    await expect(toggle).toBeEnabled({ timeout: NAV_TIMEOUT });
    await toggle.click();
    await expect(page.getByTestId('search-ai-loading')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('search-ai-loading')).toHaveAttribute('role', 'status');
    await expect(page.getByTestId('search-ai-answer')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('search-ai-answer')).toContainText('Fireball is a level 3 evocation spell');
    await expect(page.getByTestId('search-ai-sources')).toBeVisible();
    await expect(page.getByTestId('search-ai-source-link').first()).toBeVisible();
    await expect(page.getByTestId('search-ai-source-link').first()).toContainText('Fireball');
    expect(capturedBody?.query).toContain('fireball');
    await page.unroute('**/api/search/ai');
  });

  test('Cmd+Enter triggers AI search (keyboard)', async ({ page }) => {
    await page.route('**/api/search/ai', async (route) => {
      expect(route.request().method()).toBe('POST');
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          answer: 'Keyboard triggered answer.',
          sources: [{ kind: 'spell', id: 2, title: 'Magic Missile', subtitle: 'Level 1' }],
          scope: 'compendium',
        }),
      });
    });
    await openPalette(page);
    await page.evaluate((q) => {
      const input = document.getElementById('cpSearchInput') as HTMLInputElement | null;
      if (input) { input.value = q; input.dispatchEvent(new Event('input', { bubbles: true })); }
    }, 'magic missile');
    await page.evaluate((q) => window.doSearch(q), 'magic missile');
    await expect(page.locator('#cpResults')).toContainText(/Magic|No Results/i, { timeout: NAV_TIMEOUT });
    await page.evaluate(() => {
      const input = document.getElementById('cpSearchInput') as HTMLInputElement | null;
      if (input) {
        input.focus();
        input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', metaKey: true, bubbles: true }));
      }
    });
    await expect(page.getByTestId('search-ai-answer')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('search-ai-answer')).toContainText('Keyboard triggered answer');
    await page.unroute('**/api/search/ai');
  });

  test('back button returns to lexical results', async ({ page }) => {
    await page.route('**/api/search/ai', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ answer: 'Answer for back test', sources: [{ kind: 'spell', id: 1, title: 'Fireball' }], scope: 'compendium' }),
      });
    });
    await openPalette(page);
    await fillQueryAndSearch(page, 'fireball');
    await expect(page.locator('#cpResults')).toContainText(/Fireball|No Results/i, { timeout: NAV_TIMEOUT });
    await page.getByTestId('search-ai-toggle').click();
    await expect(page.getByTestId('search-ai-answer')).toBeVisible({ timeout: NAV_TIMEOUT });
    const backBtn = page.getByTestId('search-ai-back').first();
    await expect(backBtn).toBeVisible();
    await backBtn.click();
    await expect(page.getByTestId('search-ai-answer')).toBeHidden({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('search-ai-loading')).toBeHidden({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#cpResults')).toBeVisible();
    await page.unroute('**/api/search/ai');
  });

  test('503 renders error message and still shows direct-match source', async ({ page }) => {
    await page.route('**/api/search/ai', async (route) => {
      expect(route.request().method()).toBe('POST');
      const body = JSON.parse(route.request().postData() || '{}');
      expect(body.query).toBeTruthy();
      await route.fulfill({
        status: 503,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'AI is not configured; showing direct matches instead.', sources: [{ kind: 'spell', id: 1, title: 'Fireball' }], scope: 'compendium' }),
      });
    });
    await openPalette(page);
    await fillQueryAndSearch(page, 'fireball');
    await page.getByTestId('search-ai-toggle').click();
    await expect(page.locator('#cpResults')).toContainText('AI is not configured; showing direct matches instead.', { timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('search-ai-sources')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('search-ai-source-link').first()).toContainText('Fireball');
    await page.unroute('**/api/search/ai');
  });

  test('Escape in AI mode goes back to lexical results, second Esc closes palette', async ({ page }) => {
    await page.route('**/api/search/ai', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ answer: 'Escape test answer', sources: [{ kind: 'spell', id: 1, title: 'Fireball' }], scope: 'compendium' }),
      });
    });
    await openPalette(page);
    await fillQueryAndSearch(page, 'fireball');
    await page.getByTestId('search-ai-toggle').click();
    await expect(page.getByTestId('search-ai-answer')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.keyboard.press('Escape');
    await expect(page.getByTestId('search-ai-answer')).toBeHidden({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#cpResults')).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(page.locator('#searchOverlay')).toBeHidden({ timeout: NAV_TIMEOUT });
    await page.unroute('**/api/search/ai');
  });
});
