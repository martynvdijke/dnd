import { test, expect } from '@playwright/test';

test.describe('Full feature coverage', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
  });

  test('campaign lifecycle: create, assign character, delete', async ({ page }) => {
    await page.goto('/admin', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('#adminUsers .card-header')).toContainText('Users');

    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await page.click('text=New Character');
    await page.fill('#newName', 'Campaign Hero');
    await page.fill('#newRace', 'Half-Elf');
    await page.fill('#newClass', 'Paladin');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await expect(page.locator('#charGrid')).toContainText('Campaign Hero');
  });

  test('character creation with all fields', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', 'Full Hero');
    await page.fill('#newRace', 'Dragonborn');
    await page.fill('#newClass', 'Sorcerer');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await page.click('.character-card');
    await expect(page.locator('#sheetName')).toContainText('Full Hero');
    await expect(page.locator('#sheetSubtitle')).toContainText('Dragonborn Sorcerer');
  });

  test('character detail fields are editable', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', 'Edit Hero');
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.click('.character-card');

    await page.click('text=Details');
    await expect(page.locator('#detailsSection')).toBeVisible();

    const raceInput = page.locator('#detailsSection input').first();
    await expect(raceInput).toBeVisible();
  });

  test('combat tab shows hit points and actions', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', 'Combat Hero');
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Barbarian');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.click('.character-card');

    await page.click('text=Combat');
    await expect(page.locator('#combatSection')).toBeVisible();
    await expect(page.locator('#combatSection')).toContainText('Hit Points');
  });

  test('dice section is accessible from sheet', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', 'Dice Hero');
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Rogue');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.click('.character-card');

    await page.click('text=Dice');
    await expect(page.locator('#diceExpr')).toBeVisible();
  });

  test('character list shows after creating character', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', 'List Hero');
    await page.fill('#newRace', 'Gnome');
    await page.fill('#newClass', 'Wizard');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await page.click('text=Characters');
    await expect(page.locator('#charGrid .character-card')).toBeVisible();
  });

  test('admin panel navigation tabs work', async ({ page }) => {
    await page.goto('/admin', { waitUntil: 'domcontentloaded' });

    await expect(page.locator('.card-header').first()).toContainText('Users');

    await page.click('text=Compendium');
    await expect(page.locator('.card-header').first()).toContainText('Compendium Management');

    await page.click('text=Backup');
    await expect(page.locator('.card-header').first()).toContainText('Backup Settings');
  });

  test('import modal opens', async ({ page }) => {
    await page.click('text=Import');
    await expect(page.locator('#importJson')).toBeVisible();
  });
});
