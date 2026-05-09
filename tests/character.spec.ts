import { test, expect } from '@playwright/test';

test.describe('Character management', () => {
  test.beforeEach(async ({ page }) => {
    // Login
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/app/);
  });

  test('shows character list', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Character Folio');
  });

  test('creates a new character', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', 'Test Character');
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Wizard');
    await page.click('text=Create');

    // Should show the character in the list
    await expect(page.locator('.character-card')).toHaveCount(1);
    await expect(page.locator('.char-name')).toContainText('Test Character');
  });

  test('opens character sheet', async ({ page }) => {
    // Create one first
    await page.click('text=New Character');
    await page.fill('#newName', 'Sheet Test');
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    // Open it
    await page.click('.character-card');
    await expect(page.locator('#sheetName')).toContainText('Sheet Test');
    await expect(page.locator('#sheetSubtitle')).toContainText('Human Fighter');
  });

  test('shows ability scores', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', 'Stats Test');
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Barbarian');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await page.click('.character-card');
    const values = await page.locator('.value').allTextContents();
    const vals = values.map(v => parseInt(v));
    // Default stats should all be 10
    expect(vals.every(v => v === 10)).toBe(true);
  });

  test('switches tabs in character sheet', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', 'Tab Test');
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Rogue');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.click('.character-card');

    const tabs = ['Combat', 'Spells', 'Inventory', 'Features', 'Details', 'Dice'];
    for (const tab of tabs) {
      await page.click(`text=${tab}`);
      await expect(page.locator('.tab.active')).toContainText(tab);
    }
  });
});
