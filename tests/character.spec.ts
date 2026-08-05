import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT } from './helpers.js';

const uniqueName = () => `Test-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function waitLoadingDone(page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: NAV_TIMEOUT }).catch(() => {});
}

async function waitModalClosed(page) {
  await page.waitForFunction(() => {
    const modal = document.getElementById('genericModal');
    return !modal || !modal.classList.contains('show');
  }, { timeout: 10000 }).catch(() => {});
}

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
});
