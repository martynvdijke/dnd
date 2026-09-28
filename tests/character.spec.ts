import { test, expect } from './fixtures.js';
import { NAV_TIMEOUT, login, waitLoadingDone, waitModalClosed } from './helpers.js';

const uniqueName = () => `Test-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;







test.describe('Character management', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('shows character list', async ({ page }) => {
    await expect(page.locator('.navbar-brand')).toContainText('villum');
  });

  test('creates a new character', async ({ page }) => {
    const name = uniqueName();
    await page.getByTestId('new-character').click();
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Wizard');
    await page.click('.modal button:has-text("Create")');
    await waitModalClosed(page);

    await expect(page.getByText(name).first()).toBeVisible();
  });

  test('opens character sheet', async ({ page }) => {
    const name = uniqueName();
    await page.getByTestId('new-character').click();
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('.modal button:has-text("Create")');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
    await expect(page.getByTestId('sheet-view')).toBeVisible();
    await expect(page.locator('#sheetName')).toContainText(name);
    await page.waitForFunction((n) => {
      const el = document.getElementById('sheetSubtitle');
      return el && el.textContent?.includes(n);
    }, name, { timeout: 10000 }).catch(() => {});
    await expect(page.locator('#sheetSubtitle')).toContainText('Human Fighter', { timeout: 10000 });
  });

  test('shows ability scores', async ({ page }) => {
    const name = uniqueName();
    await page.getByTestId('new-character').click();
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Barbarian');
    await page.click('.modal button:has-text("Create")');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
    const abilityValues = await page.locator('.ability-box .stepper-value').allTextContents();
    expect(abilityValues.length).toBeGreaterThanOrEqual(6);
  });

  test('exports a printable PDF sheet', async ({ page }) => {
    const name = uniqueName();
    await page.getByTestId('new-character').click();
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Halfling');
    await page.fill('#newClass', 'Rogue');
    await page.click('.modal button:has-text("Create")');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
    await expect(page.getByTestId('export-pdf')).toBeVisible();

    const id = await page.evaluate(async (n) => {
      const res = await fetch('/api/characters', { credentials: 'include' });
      const data = await res.json();
      const list = Array.isArray(data) ? data : (data.characters || []);
      const c = list.find((x: any) => x.name === n);
      return c ? c.id : 0;
    }, name);
    expect(id).toBeGreaterThan(0);

    const resp = await page.request.get(`/api/characters/${id}/print?format=html`);
    expect(resp.status()).toBe(200);
    expect(resp.headers()['content-type']).toContain('text/html');
    expect(await resp.text()).toContain(name);
  });
});
