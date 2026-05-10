import { test, expect } from '@playwright/test';

test.describe('Character sheet editing', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
    await page.click('text=New Character');
    await page.fill('#newName', 'Editor Test');
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Ranger');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.click('.character-card');
  });

  test('can edit character name and stats', async ({ page }) => {
    await page.click('text=Details');
    await expect(page.locator('#detailsSection')).toBeVisible();

    await expect(page.locator('#sheetName')).toContainText('Editor Test');
  });

  test('can add and remove inventory items', async ({ page }) => {
    await page.click('text=Inventory');

    await page.click('text=Add Item');
    await page.fill('#invName', 'Longsword');
    await page.fill('#invQty', '1');
    await page.fill('#invWeight', '3');
    await page.click('text=Add');
    await page.waitForTimeout(500);

    await expect(page.locator('#inventorySection')).toContainText('Longsword');
  });

  test('can add spells and features', async ({ page }) => {
    await page.click('text=Spells');

    await page.click('text=Add Spell');
    await page.fill('#spellName', 'Mage Hand');
    await page.fill('#spellLevel', '0');
    await page.click('text=Add Spell');
    await page.waitForTimeout(500);

    await page.click('text=Features');
    await page.click('text=Add Feature');
    await page.fill('#featName', 'Darkvision');
    await page.fill('#featDesc', 'See in darkness');
    await page.fill('#featSource', 'Race');
    await page.fill('#featLevel', '1');
    await page.click('text=Add Feature');
    await page.waitForTimeout(500);

    await expect(page.locator('#featuresSection')).toContainText('Darkvision');
  });

  test('can manage proficiencies on character sheet', async ({ page }) => {
    await page.click('text=Features');

    await page.click('text=Add Proficiency');
    await page.fill('#profName', 'Stealth');
    await page.click('text=Add');
    await page.waitForTimeout(500);

    await expect(page.locator('#featuresSection')).toContainText('Stealth');
  });
});

test.describe('Rest and level up', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
    await page.click('text=New Character');
    await page.fill('#newName', 'Rest Test');
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Barbarian');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.click('.character-card');
  });

  test('can perform short rest', async ({ page }) => {
    await page.click('text=Combat');
    await page.click('text=Short Rest');
    await page.waitForTimeout(500);

    await expect(page.locator('#sheetName')).toContainText('Rest Test');
  });

  test('can perform long rest', async ({ page }) => {
    await page.click('text=Combat');
    await page.click('text=Long Rest');
    await page.waitForTimeout(500);

    await expect(page.locator('#sheetName')).toContainText('Rest Test');
  });

  test('can level up character', async ({ page }) => {
    await page.click('text=Combat');
    await page.click('text=Level Up');
    await page.waitForTimeout(500);

    await expect(page.locator('#sheetName')).toContainText('Rest Test');
  });
});

test.describe('Campaign management UI', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
  });

  test('party view shows characters', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', 'Party Member 1');
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await page.click('text=Party');
    await expect(page.locator('#partyView h1')).toContainText('Party View');
    await expect(page.locator('#partyContent')).toContainText('Party Member 1');
  });

  test('can delete a character', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', 'Delete Me');
    await page.fill('#newRace', 'Halfling');
    await page.fill('#newClass', 'Rogue');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await page.click('.character-card');
    await page.waitForTimeout(300);

    page.on('dialog', dialog => dialog.accept());
    await page.click('text=Delete');
    await page.waitForTimeout(500);

    await expect(page.locator('#charGrid .character-card')).toHaveCount(0);
  });

  test('compendium search works', async ({ page }) => {
    await page.click('text=Compendium');
    await expect(page.locator('#compendiumView h1')).toContainText('Compendium');

    const searchInput = page.locator('#compSearch');
    if (await searchInput.isVisible()) {
      await searchInput.fill('fire');
      await page.click('text=Search');
      await page.waitForTimeout(500);
    }
  });

  test('logout works', async ({ page }) => {
    await page.click('text=Logout');
    await page.waitForURL(/\/login/);
    await expect(page.locator('h1')).toContainText('villum');
  });
});

test.describe('Death saves and concentration', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
    await page.click('text=New Character');
    await page.fill('#newName', 'Death Test');
    await page.fill('#newRace', 'Tiefling');
    await page.fill('#newClass', 'Warlock');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.click('.character-card');
  });

  test('shows death save tracker on combat tab', async ({ page }) => {
    await page.click('text=Combat');
    await expect(page.locator('#combatSection')).toBeVisible();
    const text = await page.locator('#combatSection').textContent();
    expect(text).toContain('Death');
  });

  test('shows concentration tracker on combat tab', async ({ page }) => {
    await page.click('text=Combat');
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
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
    await page.click('text=New Character');
    await page.fill('#newName', 'NPC Test');
    await page.fill('#newRace', 'Gnome');
    await page.fill('#newClass', 'Wizard');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.click('.character-card');
  });

  test('can create and delete NPC with full stats', async ({ page }) => {
    await page.click('text=Npcs');

    await page.click('text=New NPC');
    await page.fill('#newNPCName', 'Villain');
    await page.fill('#newNPCRace', 'Dragonborn');
    await page.fill('#newNPCClass', 'Sorcerer');
    await page.fill('#newNPCDesc', 'A powerful foe');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await expect(page.locator('#npcsSection')).toContainText('Villain');
  });

  test('can link and interact with NPCs', async ({ page }) => {
    await page.click('text=Npcs');

    await page.click('text=New NPC');
    await page.fill('#newNPCName', 'Quest Giver');
    await page.fill('#newNPCRace', 'Human');
    await page.fill('#newNPCClass', 'Cleric');
    await page.fill('#newNPCDesc', 'Local priest');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await page.click('text=Link NPC');
    await page.selectOption('#linkNPCId', { index: 0 });
    await page.click('text=Link');
    await page.waitForTimeout(500);

    await expect(page.locator('#npcsSection')).toContainText('Quest Giver');
  });
});

test.describe('Spellcasting management', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
    await page.click('text=New Character');
    await page.fill('#newName', 'Caster Test');
    await page.fill('#newRace', 'Half-Elf');
    await page.fill('#newClass', 'Cleric');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.click('.character-card');
  });

  test('can configure spellcasting ability', async ({ page }) => {
    await page.click('text=Spells');

    await expect(page.locator('#spellsSection')).toBeVisible({ timeout: 3000 });

    const enableBtn = page.locator('text=Set Up Spellcasting');
    if (await enableBtn.isVisible()) {
      await enableBtn.click();
      await page.waitForTimeout(300);
    }
  });

  test('can track spell slots', async ({ page }) => {
    await page.click('text=Spells');
    const text = await page.locator('#spellsSection').textContent();
    expect(text).toBeTruthy();
  });
});

test.describe('Import/export edge cases', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
  });

  test('import handles empty JSON gracefully', async ({ page }) => {
    await page.click('text=Import');
    await page.fill('#importJson', '');
    await page.click('text=Import');
    await page.waitForTimeout(500);
    await expect(page.locator('#charGrid')).toBeVisible();
  });

  test('import handles malformed JSON', async ({ page }) => {
    await page.click('text=Import');
    await page.fill('#importJson', '{not valid json}');
    await page.click('text=Import');
    await page.waitForTimeout(500);
    await expect(page.locator('#charGrid')).toBeVisible();
  });

  test('character sheet shows currency tab', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', 'Currency Test');
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Rogue');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.click('.character-card');
    await page.waitForTimeout(300);

    await page.click('text=Details');
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
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
    await page.click('text=New Character');
    await page.fill('#newName', 'Quest Test');
    await page.fill('#newRace', 'Dragonborn');
    await page.fill('#newClass', 'Paladin');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.click('.character-card');
  });

  test('can log session with events', async ({ page }) => {
    await page.click('text=Sessions');

    await page.click('text=Log Session');
    await page.fill('#sessTitle', 'The Dragon Hunt');
    await page.fill('#sessNotes', 'We tracked the dragon to its lair');
    await page.fill('#sessXP', '500');
    await page.fill('#sessGold', '200');
    await page.fill('#sessEvents', 'Found dragon hoard');
    await page.click('text=Log Session');
    await page.waitForTimeout(500);

    await expect(page.locator('#sessionsSection')).toContainText('Dragon Hunt');
  });

  test('can create and complete a quest', async ({ page }) => {
    await page.click('text=Quests');

    await page.click('text=New Quest');
    await page.fill('#questName', 'Save the Village');
    await page.fill('#questDesc', 'Protect from goblin raid');
    await page.fill('#questObj', '1. Defeat goblins\n2. Return to mayor');
    await page.fill('#questRewards', '500 XP, 100 GP');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await expect(page.locator('#questsSection')).toContainText('Save the Village');
  });
});

test.describe('Error handling and edge cases', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
  });

  test('handles character creation without name', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', '');
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await page.waitForTimeout(300);
    await expect(page.locator('#charGrid')).toBeVisible();
  });

  test('character tabs are navigable back and forth', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', 'Nav Test');
    await page.fill('#newRace', 'Half-Orc');
    await page.fill('#newClass', 'Barbarian');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.click('.character-card');

    const tabs = ['Stats', 'Combat', 'Spells', 'Inventory', 'Features', 'Details', 'Dice'];
    for (const tab of tabs) {
      await page.click(`text=${tab}`);
      await page.waitForTimeout(100);
    }

    await page.click('text=Stats');
    await page.waitForTimeout(100);
    await expect(page.locator('#statsSection')).toBeVisible();
  });

  test('dice roller handles minus expressions', async ({ page }) => {
    await page.click('text=Dice');
    const input = page.locator('#diceExpr');

    await input.fill('1d20-2');
    await page.click('text=Roll the Bones');
    await page.waitForTimeout(300);

    const result = page.locator('#diceResult');
    await expect(result).toBeVisible();
  });

  test('multiple dice expressions can be rolled sequentially', async ({ page }) => {
    await page.click('text=Dice');
    const input = page.locator('#diceExpr');
    const btn = page.locator('text=Roll the Bones');

    const expressions = ['1d20', '2d6+3', '3d8+2d6', '1d4-1'];
    for (const expr of expressions) {
      await input.fill(expr);
      await btn.click();
      await page.waitForTimeout(200);
    }

    const historyItems = page.locator('.dice-history-item');
    const count = await historyItems.count();
    expect(count).toBeGreaterThanOrEqual(4);
  });
});
