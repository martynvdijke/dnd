import { test, expect } from '@playwright/test';

test.describe('Full application smoke test', () => {
  test.beforeEach(async ({ page }) => {
    const resp = await page.goto('/login', { waitUntil: 'domcontentloaded' });
    expect(resp?.status()).toBe(200);
  });

  test('all static assets load correctly', async ({ page }) => {
    const assets = [
      '/static/style.css',
      '/static/js/app.js',
      '/static/js/login.js',
    ];
    for (const asset of assets) {
      const resp = await page.goto(asset, { waitUntil: 'domcontentloaded' });
      expect(resp?.status()).toBe(200);
    }
  });

  test('login flow works', async ({ page }) => {
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);
    await expect(page.locator('#userName')).toContainText('admin', { timeout: 5000 });
    await expect(page.locator('.navbar-brand')).toBeVisible();
  });

  test('character list loads', async ({ page }) => {
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);
    await expect(page.locator('#charGrid')).toBeVisible({ timeout: 5000 });
  });

  test('navigation between views works', async ({ page }) => {
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);

    const views = [
      { link: 'Compendium', heading: 'Compendium' },
      { link: 'Dice', heading: 'Dice Roller' },
      { link: 'Encounters', heading: 'Encounter Builder' },
      { link: 'Factions', heading: 'Factions' },
    ];
    for (const { link, heading } of views) {
      await page.locator(`nav a:has-text("${link}")`).click();
      await expect(page.locator('h1')).toContainText(heading, { timeout: 5000 });
    }
  });

  test('character create and sheet open', async ({ page }) => {
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);

    const name = `Smoke-${Date.now()}`;
    await page.click('button:has-text("New Character")');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('.modal button:has-text("Create")');
    await page.waitForFunction(() => {
      const modal = document.getElementById('genericModal');
      return !modal || !modal.classList.contains('show');
    }, { timeout: 10000 }).catch(() => {});

    await expect(page.locator('.character-card').filter({ hasText: name })).toBeVisible({ timeout: 5000 });
    await page.locator('.character-card').filter({ hasText: name }).click();
    await expect(page.locator('#sheetName')).toContainText(name, { timeout: 10000 });
  });

  test('logout works', async ({ page }) => {
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);

    await page.locator('a:has-text("Logout")').click();
    await expect(page).toHaveURL(/\/login/, { timeout: 5000 });
  });

  test('app version is displayed', async ({ page }) => {
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);

    await expect(page.locator('footer')).toContainText('v1.15.0', { timeout: 5000 });
  });
});
