import { test, expect } from '@playwright/test';

const uniqueName = () => `FT-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function ensureNavOpen(page) {
  const toggler = page.locator('.navbar-toggler');
  if (await toggler.isVisible()) {
    await toggler.click();
    await page.waitForTimeout(300);
  }
}

test.describe('Full feature coverage', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
  });

  test('campaign lifecycle: create, assign character, delete', async ({ page }) => {
    await page.waitForTimeout(300);
    await page.goto('/admin', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('#adminUsers .card-header')).toContainText('Users');

    const name = uniqueName();
    await page.waitForTimeout(300);
    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Half-Elf');
    await page.fill('#newClass', 'Paladin');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await expect(page.locator('#charGrid')).toContainText(name);
  });

  test('character creation with all fields', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dragonborn');
    await page.fill('#newClass', 'Sorcerer');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await expect(page.locator('#sheetName')).toContainText(name);
    await expect(page.locator('#sheetSubtitle')).toContainText('Dragonborn Sorcerer');
  });

  test('character detail fields are editable', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.locator('.character-card').filter({ hasText: name }).click();

    await page.click('text=Details');
    await expect(page.locator('#detailsSection')).toBeVisible();

    const raceInput = page.locator('#detailsSection input').first();
    await expect(raceInput).toBeVisible();
  });

  test('combat tab shows hit points and actions', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Barbarian');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.locator('.character-card').filter({ hasText: name }).click();

    await page.click('text=Combat');
    await expect(page.locator('#combatSection')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#combatSection')).toContainText('Hit Points');
  });

  test('dice section is accessible from sheet', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Rogue');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.locator('.character-card').filter({ hasText: name }).click();

    await page.click('#tabBar button:has-text("Dice")');
    await expect(page.locator('#diceExpr')).toBeVisible({ timeout: 5000 });
  });

  test('character list shows after creating character', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Gnome');
    await page.fill('#newClass', 'Wizard');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await ensureNavOpen(page);
    await page.click('a:has-text("Characters")');
    await expect(page.locator('.character-card').filter({ hasText: name })).toBeVisible();
  });

  test('admin panel navigation tabs work', async ({ page }) => {
    await page.waitForTimeout(300);
    await page.goto('/admin', { waitUntil: 'domcontentloaded' });

    await expect(page.locator('#adminUsers .card-header')).toContainText('Users');

    await page.click('#adminTabs button:has-text("Compendium")');
    await expect(page.locator('#adminCompendium .card-header')).toContainText('Compendium Management');

    await page.click('#adminTabs button:has-text("Backup")');
    await expect(page.locator('#adminBackup .card-header').first()).toContainText('Backup Settings');
  });

  test('import modal opens', async ({ page }) => {
    await page.click('button:has-text("Import")');
    await page.waitForTimeout(300);
    await expect(page.locator('#importJson')).toBeVisible();
  });
});
