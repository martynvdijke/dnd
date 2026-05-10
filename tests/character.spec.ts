import { test, expect } from '@playwright/test';

const uniqueName = () => `Test-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('Character management', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { timeout: 10000 });
    await page.waitForSelector('.character-card', { timeout: 5000 }).catch(() => {});
  });

  test('shows character list', async ({ page }) => {
    await expect(page.locator('.navbar-brand')).toContainText('villum');
  });

  test('creates a new character', async ({ page }) => {
    const name = uniqueName();
    await page.click('button:has-text("New Character")');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Wizard');
    await page.click('.modal button:has-text("Create")');
    await page.waitForTimeout(1000);

    await expect(page.getByText(name).first()).toBeVisible();
  });

  test('opens character sheet', async ({ page }) => {
    const name = uniqueName();
    await page.click('button:has-text("New Character")');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('.modal button:has-text("Create")');
    await page.waitForTimeout(1000);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await expect(page.locator('#sheetName')).toContainText(name);
    await expect(page.locator('#sheetSubtitle')).toContainText('Human Fighter');
  });

  test('shows ability scores', async ({ page }) => {
    const name = uniqueName();
    await page.click('button:has-text("New Character")');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Barbarian');
    await page.click('.modal button:has-text("Create")');
    await page.waitForTimeout(1000);

    await page.locator('.character-card').filter({ hasText: name }).click();
    const abilityValues = await page.locator('.ability-box .abil-value').allTextContents();
    expect(abilityValues.length).toBeGreaterThanOrEqual(6);
  });
});
