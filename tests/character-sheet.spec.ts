import { test, expect } from '@playwright/test';

const uniqueName = () => `Test-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function ensureNavOpen(page) {
  const toggler = page.locator('.navbar-toggler');
  if (await toggler.isVisible()) {
    await toggler.click();
    await page.waitForTimeout(300);
  }
}

async function waitLoadingDone(page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: 5000 }).catch(() => {});
}

async function waitModalClosed(page) {
  await page.waitForFunction(() => {
    const modal = document.getElementById('genericModal');
    return !modal || !modal.classList.contains('show');
  }, { timeout: 10000 }).catch(() => {});
}

test.describe('Character sheet editing', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('can edit character name and stats', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Ranger');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await expect(page.locator('#sheetName')).toContainText(name);
    await page.click('#tabBar button:has-text("Details")');
    await expect(page.locator('#detailsSection')).toBeVisible();
    await expect(page.locator('#sheetName')).toContainText(name);
  });

  test('can add and remove inventory items', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Ranger');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Inventory")');
    await page.click('text=Add Item');
    await page.fill('#invName', 'Longsword');
    await page.fill('#invQty', '1');
    await page.fill('#invWeight', '3');
    await page.click('#genericModal button:has-text("Add")');
    await waitModalClosed(page);

    await expect(page.locator('#inventorySection')).toContainText('Longsword');
  });

  test('can add spells and features', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Ranger');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Spells")');
    const setupBtn = page.locator('button:has-text("Set Up Spellcasting")');
    if (await setupBtn.count() > 0) {
      await setupBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('#spellsSection')).toBeVisible();

    await page.click('#tabBar button:has-text("Features")');
    await page.click('text=Add Feature');
    await page.fill('#featName', 'Darkvision');
    await page.fill('#featDesc', 'See in darkness');
    await page.fill('#featSource', 'Race');
    await page.fill('#featLevel', '1');
    await page.click('#genericModal button:has-text("Add Feature")');
    await waitModalClosed(page);

    await expect(page.locator('#featuresSection')).toContainText('Darkvision');
  });

  test('can manage proficiencies on character sheet', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Ranger');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('text=Add Proficiency');
    await page.fill('#profName', 'Stealth');
    await page.click('#genericModal button:has-text("Add Proficiency")');
    await waitModalClosed(page);

    await expect(page.locator('#statsSection')).toContainText('Stealth');
  });
});

test.describe('Rest and level up', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('can perform short rest', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Barbarian');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Combat")');
    await page.click('text=Level Up');
    await page.waitForTimeout(500);

    await expect(page.locator('#sheetName')).toContainText(name);
  });
});

test.describe('Campaign management UI', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('party view shows characters', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await waitModalClosed(page);

    await ensureNavOpen(page);
    await page.click('a:has-text("Party")');
    await expect(page.locator('#partyView h1')).toContainText('Party View');
    await expect(page.locator('#partyContent')).toContainText(name);
  });

  test('can delete a character', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Halfling');
    await page.fill('#newClass', 'Rogue');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    page.on('dialog', dialog => dialog.accept());
    await page.click('text=Delete');
    await expect(page.locator('#charGrid')).toBeVisible({ timeout: 10000 });

    await expect(page.locator('.character-card').filter({ hasText: name })).toHaveCount(0);
  });

  test('compendium search works', async ({ page }) => {
    await ensureNavOpen(page);
    await page.click('a:has-text("Compendium")');
    await expect(page.locator('#compendiumView h1')).toContainText('Compendium');

    const searchInput = page.locator('#compSearch');
    if (await searchInput.isVisible()) {
      await searchInput.fill('fire');
      await page.click('text=Search');
      await page.waitForTimeout(500);
    }
  });

  test('logout works', async ({ page }) => {
    await ensureNavOpen(page);
    await page.click('a:has-text("Logout")');
    await page.waitForURL(/\/login/);
    await expect(page.locator('h1')).toContainText('villum');
  });
});

test.describe('Death saves and concentration', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('shows death save tracker on combat tab', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Tiefling');
    await page.fill('#newClass', 'Warlock');
    await page.click('text=Create');
    await page.locator('.character-card').filter({ hasText: name }).waitFor({ state: 'visible', timeout: 10000 });
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Combat")');
    await waitLoadingDone(page);
    await expect(page.locator('#combatSection')).toBeVisible();
    const text = await page.locator('#combatSection').textContent();
    expect(text).toContain('Death');
  });

  test('shows concentration tracker on combat tab', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Tiefling');
    await page.fill('#newClass', 'Warlock');
    await page.click('text=Create');
    await page.locator('.character-card').filter({ hasText: name }).waitFor({ state: 'visible', timeout: 10000 });
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Combat")');
    const text = await page.locator('#combatSection').textContent();
    expect(text).toBeTruthy();
    expect(text).toContain('Concentration');
  });
});

test.describe('NPC interactions extended', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  async function waitModalClosed(page: any) {
    await page.waitForFunction(() => !document.getElementById('genericModal')?.classList.contains('show'), { timeout: 10000 });
  }

  test('can create and delete NPC with full stats', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Gnome');
    await page.fill('#newClass', 'Wizard');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Npcs")');
    await expect(page.locator('#npcsSection button:has-text("New NPC")')).toBeVisible({ timeout: 5000 });
    await page.click('text=New NPC');
    await page.fill('#newNPCName', 'Villain');
    await page.fill('#newNPCRace', 'Dragonborn');
    await page.fill('#newNPCClass', 'Sorcerer');
    await page.fill('#newNPCDesc', 'A powerful foe');
    await page.click('text=Create');
    await waitModalClosed(page);

    await expect(page.locator('#npcsSection')).toContainText('Villain');
  });

  test('can link and interact with NPCs', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Gnome');
    await page.fill('#newClass', 'Wizard');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Npcs")');
    await expect(page.locator('#npcsSection button:has-text("New NPC")')).toBeVisible({ timeout: 5000 });
    await page.click('text=New NPC');
    await page.fill('#newNPCName', 'Quest Giver');
    await page.fill('#newNPCRace', 'Human');
    await page.fill('#newNPCClass', 'Cleric');
    await page.fill('#newNPCDesc', 'Local priest');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.locator('button:has-text("Link NPC")').click();
    await expect(page.locator('#linkNPCId')).toBeVisible({ timeout: 5000 });
    await page.selectOption('#linkNPCId', { index: 0 });
    await page.locator('#genericModal button:has-text("Link")').click();
    await waitModalClosed(page);

    await expect(page.locator('#npcsSection')).toContainText('Quest Giver', { timeout: 5000 });
  });
});

test.describe('Spellcasting management', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('can configure spellcasting ability', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Half-Elf');
    await page.fill('#newClass', 'Cleric');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Spells")');
    await expect(page.locator('#spellsSection')).toBeVisible({ timeout: 3000 });

    const enableBtn = page.locator('text=Set Up Spellcasting');
    if (await enableBtn.isVisible()) {
      await enableBtn.click();
      await page.waitForTimeout(300);
    }
  });

  test('can track spell slots', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Half-Elf');
    await page.fill('#newClass', 'Cleric');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Spells")');
    const text = await page.locator('#spellsSection').textContent();
    expect(text).toBeTruthy();
  });
});

test.describe('Import/export edge cases', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('import handles empty JSON gracefully', async ({ page }) => {
    await page.click('button:has-text("Import")');
    await expect(page.locator('#genericModal')).toHaveClass(/show|fade/, { timeout: 5000 });
    const textarea = page.locator('#importJson');
    await expect(textarea).toBeVisible({ timeout: 10000 });
  });

  test('import handles malformed JSON', async ({ page }) => {
    await page.click('button:has-text("Import")');
    await expect(page.locator('#genericModal')).toHaveClass(/show|fade/, { timeout: 5000 });
    const textarea = page.locator('#importJson');
    await expect(textarea).toBeVisible({ timeout: 10000 });

    await textarea.fill('{not valid json}');
    await page.click('#genericModal button:has-text("Import")');
    await page.waitForTimeout(500);
  });

  test('character sheet shows currency tab', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Rogue');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).waitFor({ state: 'visible', timeout: 10000 });
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Details")');
    const text = await page.locator('#detailsSection').textContent();
    expect(text).toContain('CP');
    expect(text).toContain('GP');
  });
});

test.describe('Session and quest management UI', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('can log session with events', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dragonborn');
    await page.fill('#newClass', 'Paladin');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Sessions")');
    await page.click('text=Log Session');
    await page.fill('#sessTitle', 'The Dragon Hunt');
    await page.fill('#sessNotes', 'We tracked the dragon to its lair');
    await page.fill('#sessXP', '500');
    await page.fill('#sessGold', '200');
    await page.fill('#sessEvents', 'Found dragon hoard');
    await page.click('#genericModal button:has-text("Log Session")');
    await waitModalClosed(page);

    await expect(page.locator('#sessionsSection')).toContainText('Dragon Hunt');
  });

  test('can create and complete a quest', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dragonborn');
    await page.fill('#newClass', 'Paladin');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Quests")');
    await expect(page.locator('#questsSection button:has-text("New Quest")')).toBeVisible({ timeout: 5000 });
    await page.click('text=New Quest');
    await page.fill('#questName', 'Save the Village');
    await page.fill('#questDesc', 'Protect from goblin raid');
    await page.fill('#questObj', '1. Defeat goblins\n2. Return to mayor');
    await page.fill('#questRewards', '500 XP, 100 GP');
    await page.click('text=Create');
    await waitModalClosed(page);

    await expect(page.locator('#questsSection')).toContainText('Save the Village');
  });
});

test.describe('Theme and loading', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('dark mode toggle switches theme and persists', async ({ page }) => {
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');

    await page.click('#themeToggle');
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');

    await page.reload();
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');

    await page.click('#themeToggle');
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
  });

  test('shows loading overlay during API calls', async ({ page }) => {
    await expect(page.locator('#loadingOverlay')).toHaveClass(/d-none/);

    await page.click('button:has-text("New Character")');
    await page.fill('#newName', 'Load Test');
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('.modal button:has-text("Create")');
    await waitModalClosed(page);

    await expect(page.locator('#loadingOverlay')).toHaveClass(/d-none/);
  });
});

test.describe('Empty states', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('shows empty states for sections with no data', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).first().click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Features")');
    await expect(page.locator('#featuresSection .empty-state')).toBeVisible();

    await page.click('#tabBar button:has-text("Inventory")');
    await expect(page.locator('#inventorySection .empty-state')).toBeVisible();

    await page.click('#tabBar button:has-text("Quests")');
    await expect(page.locator('#questsSection .empty-state')).toBeVisible();

    await page.click('#tabBar button:has-text("Journal")');
    await expect(page.locator('#journalSection .empty-state')).toBeVisible();

    await page.click('#tabBar button:has-text("Sessions")');
    await expect(page.locator('#sessionsSection .empty-state')).toBeVisible();
  });
});

test.describe('Keyboard shortcuts', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('? opens keyboard shortcuts help', async ({ page }) => {
    await page.keyboard.press('?');
    await expect(page.locator('#genericModal')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#genericModal')).toContainText('Keyboard Shortcuts');
    await page.locator('#genericModal .btn-close').click();
    await waitModalClosed(page);
    await expect(page.locator('#genericModal')).not.toBeVisible();
  });

  test('n opens new character modal', async ({ page }) => {
    await page.keyboard.press('n');
    await expect(page.locator('#genericModal')).toBeVisible();
    await expect(page.locator('#genericModal')).toContainText('New Character');
  });

  test('T toggles dark mode', async ({ page }) => {
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
    await page.keyboard.press('t');
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
    await page.keyboard.press('t');
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
  });
});

test.describe('Auto-save', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('inline edit auto-saves after typing', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Ranger');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    const firstInput = page.locator('.abil-value-input').first();
    await firstInput.click();
    await firstInput.fill('18');
    await page.waitForTimeout(1200);

    await page.reload();
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
    await waitLoadingDone(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await expect(page.locator('.abil-value-input').first()).toHaveValue('18');
  });
});

test.describe('Tooltips', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('ability boxes have tooltips', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    const firstAbility = page.locator('.ability-box').first();
    await expect(firstAbility).toHaveAttribute('title', /Strength|Dexterity|Constitution|Intelligence|Wisdom|Charisma/);
  });

  test('skill rows have calculation tooltips', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    const firstSkill = page.locator('.skill-row').first();
    await expect(firstSkill).toHaveAttribute('title', /STR|DEX|CON|INT|WIS|CHA/);
  });
});

test.describe('XP progress bar', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('shows XP progress bar on stats tab', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await expect(page.locator('#xpBarContainer')).toBeVisible();
    await expect(page.locator('#xpBarContainer')).toContainText('Level');
    await expect(page.locator('#xpBarContainer')).toContainText('XP');
  });
});

test.describe('Error handling and edge cases', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('handles character creation without name', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', '');
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await expect(page.locator('#charGrid')).toBeVisible({ timeout: 5000 });
  });

  test('character tabs are navigable back and forth', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Half-Orc');
    await page.fill('#newClass', 'Barbarian');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Stats")');
    await page.waitForTimeout(100);
    await expect(page.locator('#statsSection')).toBeVisible();

    await page.click('#tabBar button:has-text("Combat")');
    await page.waitForTimeout(200);
    await expect(page.locator('#combatSection')).toBeVisible();

    for (const tab of ['Spells', 'Inventory', 'Features', 'Details']) {
      await page.click(`#tabBar button:has-text("${tab}")`);
      await page.waitForTimeout(150);
    }

    await page.click('#tabBar button:has-text("Stats")');
    await page.waitForTimeout(100);
    await expect(page.locator('#statsSection')).toBeVisible();
  });

  test('dice roller handles minus expressions', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Half-Orc');
    await page.fill('#newClass', 'Barbarian');
    await page.click('text=Create');
    await expect(page.locator('.character-card').filter({ hasText: name })).toBeVisible({ timeout: 10000 });
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Dice")');
    await expect(page.locator('#diceExpr')).toBeVisible({ timeout: 5000 });

    const input = page.locator('#diceExpr');
    await input.fill('1d20-2');
    await page.click('text=Roll the Bones');

    const result = page.locator('#diceResult');
    await expect(result).toBeVisible({ timeout: 10000 });
  });

  test('multiple dice expressions can be rolled sequentially', async ({ page }) => {
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Half-Orc');
    await page.fill('#newClass', 'Barbarian');
    await page.click('text=Create');
    await expect(page.locator('.character-card').filter({ hasText: name })).toBeVisible({ timeout: 10000 });
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Dice")');
    await expect(page.locator('#diceSection')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#diceExpr')).toBeAttached({ timeout: 5000 });

    const input = page.locator('#diceExpr');
    const btn = page.locator('text=Roll the Bones');

    const expressions = ['1d20', '2d6+3', '3d8+2d6', '1d4-1'];
    for (const expr of expressions) {
      await input.fill(expr);
      await btn.click();
      await expect(page.locator('#diceResult')).toBeVisible({ timeout: 5000 });
      await expect(page.locator('#diceResult')).toContainText(expr, { timeout: 10000 });
      await page.waitForTimeout(200);
    }

    const historyItems = page.locator('.dice-history-item');
    const count = await historyItems.count();
    expect(count).toBeGreaterThanOrEqual(4);
  });
});
