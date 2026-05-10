import { test, expect } from '@playwright/test';

const uniqueName = () => `Camp-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('Campaign features', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
  });

  test('Locations tab exists and can link a location', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Wizard');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await page.waitForTimeout(300);

    await page.click('text=Locations');
    await expect(page.locator('#locationsSection h5').first()).toContainText('Linked Locations');

    await page.click('text=New Location');
    await page.fill('#newLocName', 'Waterdeep');
    await page.fill('#newLocDesc', 'The City of Splendors');
    await page.click('text=Create');
    await page.waitForTimeout(300);

    await page.click('text=Link Location');
    await page.selectOption('#linkLocId', { index: 0 });
    await page.click('#genericModal button:has-text("Link")');
    await page.waitForTimeout(300);

    await expect(page.locator('#locationsSection')).toContainText('Waterdeep');
  });

  test('NPCs tab works', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Wizard');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await page.waitForTimeout(300);

    await page.click('text=Npcs');
    await expect(page.locator('#npcsSection h5').first()).toContainText('Related NPCs');

    await page.click('text=New NPC');
    await page.fill('#newNPCName', 'Elminster');
    await page.fill('#newNPCRace', 'Human');
    await page.fill('#newNPCClass', 'Wizard');
    await page.fill('#newNPCDesc', 'The Sage of Shadowdale');
    await page.click('text=Create');
    await page.waitForTimeout(300);

    await page.click('text=Link NPC');
    await page.selectOption('#linkNPCId', { index: 0 });
    await page.click('#genericModal button:has-text("Link")');
    await page.waitForTimeout(300);

    await expect(page.locator('#npcsSection')).toContainText('Elminster');
  });

  test('Sessions tab allows logging sessions', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Wizard');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await page.waitForTimeout(300);

    await page.click('text=Sessions');
    await expect(page.locator('#sessionsSection h5').first()).toContainText('Session Log');

    await page.click('text=Log Session');
    await page.fill('#sessTitle', 'Session 1: The Beginning');
    await page.fill('#sessNotes', 'Our heroes met in a tavern...');
    await page.fill('#sessXP', '300');
    await page.fill('#sessGold', '50');
    await page.fill('#sessEvents', 'Met the mysterious stranger');
    await page.click('#genericModal button:has-text("Log Session")');
    await page.waitForTimeout(300);

    await expect(page.locator('#sessionsSection')).toContainText('Session 1');
    await expect(page.locator('#sessionsSection')).toContainText('300 XP');
  });

  test('Quests tab works', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Wizard');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await page.waitForTimeout(300);

    await page.click('text=Quests');
    await expect(page.locator('#questsSection h5').first()).toContainText('Quests');

    await page.click('text=New Quest');
    await page.fill('#questName', 'Find the Lost Crown');
    await page.fill('#questDesc', 'The king has lost his crown');
    await page.fill('#questObj', '1. Enter the dungeon\n2. Defeat the boss');
    await page.fill('#questRewards', '1000 XP, Royal Favor');
    await page.click('text=Create');
    await page.waitForTimeout(300);

    await expect(page.locator('#questsSection')).toContainText('Find the Lost Crown');
  });

  test('Journal tab works', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Wizard');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await page.waitForTimeout(300);

    await page.click('text=Journal');
    await expect(page.locator('#journalSection h5').first()).toContainText('Character Journal');

    await page.click('text=Write Entry');
    await page.fill('#journalTitle', 'Day 1');
    await page.fill('#journalEntry', 'Today was the first day of my adventure...');
    await page.click('#genericModal button:has-text("Save")');
    await page.waitForTimeout(300);

    await expect(page.locator('#journalSection')).toContainText('Day 1');
    await expect(page.locator('#journalSection')).toContainText('first day of my adventure');
  });

  test('Graph tab loads visualization', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Wizard');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await page.waitForTimeout(300);

    await page.click('text=Locations');
    await expect(page.locator('#locationsSection h5').first()).toBeVisible();
    await page.click('text=New Location');
    await page.fill('#newLocName', 'Neverwinter');
    await page.fill('#newLocDesc', 'A city');
    await page.click('text=Create');
    await page.waitForTimeout(300);
    await page.click('text=Link Location');
    await page.selectOption('#linkLocId', { index: 0 });
    await page.click('#genericModal button:has-text("Link")');
    await page.waitForTimeout(300);

    await page.click('text=Sessions');
    await expect(page.locator('#sessionsSection h5').first()).toBeVisible();
    await page.click('text=Log Session');
    await page.fill('#sessTitle', 'Session Test');
    await page.fill('#sessXP', '0');
    await page.fill('#sessGold', '0');
    await page.click('#genericModal button:has-text("Log Session")');
    await page.waitForTimeout(300);

    await page.click('text=Graph');
    await expect(page.locator('#graphContainer')).toBeVisible({ timeout: 5000 });
  });
});
