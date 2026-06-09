import { test, expect } from '@playwright/test';
import { isMobile } from './helpers.js';

const uniqueName = () => `Search-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function ensureNavOpen(page) {
  const toggler = page.locator('.navbar-toggler');
  if (await toggler.isVisible()) {
    await toggler.click();
    await page.waitForTimeout(300);
  }
}

async function waitModalClosed(page) {
  await page.waitForFunction(() => {
    const modal = document.getElementById('genericModal');
    return !modal || !modal.classList.contains('show');
  }, { timeout: 10000 }).catch(() => {});
}

test.describe('Advanced Search', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded', timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);
    await page.waitForFunction(() => {
      const o = document.getElementById('loadingOverlay');
      return o && o.classList.contains('d-none');
    }, { timeout: 5000 }).catch(() => {});
    await page.waitForTimeout(300);
  });

  test('search bar is visible in navbar', async ({ page }) => {
    if (await isMobile(page)) return;
    await ensureNavOpen(page);
    await expect(page.locator('#searchInput')).toBeVisible();
    await expect(page.locator('#searchBtn')).toBeVisible();
  });

  test('doSearch function exists on window', async ({ page }) => {
    const exists = await page.evaluate(() => typeof window.doSearch);
    expect(exists).toBe('function');
  });

  test('search finds fireball in compendium', async ({ page }) => {
    if (await isMobile(page)) {
      await page.evaluate(() => { (document.getElementById('searchInput') as HTMLInputElement).value = 'fireball'; });
    } else {
      await ensureNavOpen(page);
      await page.fill('#searchInput', 'fireball');
    }
    await page.evaluate(() => window.doSearch());
    await page.waitForTimeout(1000);
    await expect(page.locator('#searchPanel')).toContainText('Spells');
  });

  test('search shows no results for nonsense query', async ({ page }) => {
    if (await isMobile(page)) {
      await page.evaluate(() => { (document.getElementById('searchInput') as HTMLInputElement).value = 'xyznonexistent12345'; });
    } else {
      await ensureNavOpen(page);
      await page.fill('#searchInput', 'xyznonexistent12345');
    }
    await page.evaluate(() => window.doSearch());
    await page.waitForTimeout(1000);
    await expect(page.locator('#searchPanel')).toContainText('No Results');
  });

  test('search finds character by name', async ({ page }) => {
    const name = uniqueName();
    await page.click('button:has-text("New Character")');
    await page.locator('#newName').waitFor({ state: 'visible', timeout: 5000 });
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Ranger');
    await page.click('.modal button:has-text("Create")');
    await waitModalClosed(page);

    const searchTerm = name.slice(0, 10);
    if (await isMobile(page)) {
      await page.evaluate((term) => { (document.getElementById('searchInput') as HTMLInputElement).value = term; }, searchTerm);
    } else {
      await ensureNavOpen(page);
      await page.fill('#searchInput', searchTerm);
    }
    await page.evaluate(() => window.doSearch());
    await page.waitForTimeout(1000);
    await expect(page.locator('#searchPanel')).toContainText(name);
  });

  test('search result navigates to character sheet', async ({ page }) => {
    const name = uniqueName();
    await page.click('button:has-text("New Character")');
    await page.locator('#newName').waitFor({ state: 'visible', timeout: 5000 });
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Fighter');
    await page.click('.modal button:has-text("Create")');
    await waitModalClosed(page);

    if (await isMobile(page)) {
      await page.evaluate((n) => { (document.getElementById('searchInput') as HTMLInputElement).value = n; }, name);
    } else {
      await ensureNavOpen(page);
      await page.fill('#searchInput', name);
    }
    await page.evaluate(() => window.doSearch());
    await page.waitForTimeout(1000);

    const result = page.locator('.search-result-item').first();
    await expect(result).toBeVisible();
    await result.click();
    await page.waitForTimeout(500);

    await expect(page.locator('#sheetName')).toContainText(name);
  });
});
