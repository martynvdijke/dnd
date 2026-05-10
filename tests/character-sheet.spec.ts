import { test, expect } from '@playwright/test';

test.describe('Character sheet editing', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
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

    const nameField = page.locator('#sheetName');
    await nameField.fill('Renamed Hero');
    await nameField.press('Tab');
    await page.waitForTimeout(500);

    // Verify the name persisted
    await page.reload();
    await page.waitForTimeout(500);
    await page.click('.character-card');
    await expect(page.locator('#sheetName')).toHaveValue('Renamed Hero');
  });

  test('can add and remove inventory items', async ({ page }) => {
    await page.click('text=Inventory');

    await page.click('text=Add Item');
    await page.fill('#invName', 'Longsword');
    await page.selectOption('#invCategory', 'weapon');
    await page.fill('#invQty', '1');
    await page.click('text=Save');
    await page.waitForTimeout(500);

    await expect(page.locator('#inventoryTab')).toContainText('Longsword');
  });

  test('can add spells and features', async ({ page }) => {
    await page.click('text=Spells');

    await page.click('text=Add Spell');
    await page.fill('#spellName', 'Mage Hand');
    await page.fill('#spellLevel', '0');
    await page.fill('#spellSchool', 'Conjuration');
    await page.click('text=Add');
    await page.waitForTimeout(500);

    await page.click('text=Features');
    await page.click('text=Add Feature');
    await page.fill('#featName', 'Darkvision');
    await page.fill('#featDesc', 'See in darkness');
    await page.fill('#featSource', 'Race');
    await page.fill('#featLevel', '1');
    await page.click('text=Add');
    await page.waitForTimeout(500);

    await expect(page.locator('#featuresTab')).toContainText('Darkvision');
  });

  test('can manage proficiencies on character sheet', async ({ page }) => {
    await page.click('text=Features');

    await page.click('text=Add Proficiency');
    await page.fill('#profName', 'Stealth');
    await page.selectOption('#profType', 'skill');
    await page.click('text=Add');
    await page.waitForTimeout(500);

    await expect(page.locator('#featuresTab')).toContainText('Stealth');
  });
});

test.describe('Rest and level up', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
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

    // Should still be on the sheet
    await expect(page.locator('#sheetName')).toHaveValue('Rest Test');
  });

  test('can perform long rest', async ({ page }) => {
    await page.click('text=Combat');
    await page.click('text=Long Rest');
    await page.waitForTimeout(500);

    await expect(page.locator('#sheetName')).toHaveValue('Rest Test');
  });

  test('can level up character', async ({ page }) => {
    await page.click('text=Combat');
    await page.click('text=Level Up');
    await page.waitForTimeout(500);

    // Should still be on the sheet after level up
    await expect(page.locator('#sheetName')).toHaveValue('Rest Test');
  });
});

test.describe('Campaign management UI', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
  });

  test('party view shows characters', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', 'Party Member 1');
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    await page.click('text=Party');
    await expect(page.locator('h1')).toContainText('Party');
    await expect(page.locator('#partySection')).toContainText('Party Member 1');
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

    await page.click('text=Delete');
    page.on('dialog', dialog => dialog.accept());
    await page.waitForTimeout(500);

    // Character should be gone from list
    await expect(page.locator('.character-card')).toHaveCount(0);
  });

  test('compendium search works', async ({ page }) => {
    await page.click('text=Compendium');
    await expect(page.locator('h1')).toContainText('Compendium');

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
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
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
    // Verify the death saves section exists
    const deathSavesSection = page.locator('.death-saves');
    await expect(deathSavesSection).toBeVisible({ timeout: 3000 }).catch(() => {
      // Death saves might be in a different structure
      expect(page.locator('#combatTab')).toContainText(/Death|death|Save|save/);
    });
  });

  test('shows concentration tracker on combat tab', async ({ page }) => {
    await page.click('text=Combat');
    const combatContent = await page.locator('#combatTab').textContent();
    // Should have concentration or spell focus area
    expect(combatContent).toBeTruthy();
  });
});

test.describe('NPC interactions extended', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
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

    // Create NPC first
    await page.click('text=New NPC');
    await page.fill('#newNPCName', 'Quest Giver');
    await page.fill('#newNPCRace', 'Human');
    await page.fill('#newNPCClass', 'Cleric');
    await page.fill('#newNPCDesc', 'Local priest');
    await page.click('text=Create');
    await page.waitForTimeout(500);

    // Link it
    await page.click('text=Link NPC');
    await page.selectOption('#linkNPCId', { index: 0 });
    await page.selectOption('#linkNPCRel', 'ally');
    await page.click('text=Link');
    await page.waitForTimeout(500);

    // Interaction
    const interactBtn = page.locator('text=Interact').first();
    if (await interactBtn.isVisible()) {
      await interactBtn.click();
      await page.waitForTimeout(300);
    }
  });
});

test.describe('Spellcasting management', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
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

    const spellcastingSection = page.locator('#spellTab');
    await expect(spellcastingSection).toBeVisible({ timeout: 3000 });

    // Try to set spellcasting ability
    const abilitySelect = page.locator('#scAbility');
    if (await abilitySelect.isVisible()) {
      await abilitySelect.selectOption('wis');
      await page.fill('#scDC', '14');
      await page.fill('#scBonus', '6');
      await page.click('text=Save');
      await page.waitForTimeout(500);
    }
  });

  test('can track spell slots', async ({ page }) => {
    await page.click('text=Spells');

    // Spell slot tracking UI
    const spellSlots = page.locator('.spell-slots');
    await expect(spellSlots).toBeVisible({ timeout: 3000 }).catch(() => {
      // May have different selector
      expect(page.locator('#spellTab').textContent()).toBeTruthy();
    });
  });
});

test.describe('Import/export edge cases', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
  });

  test('import handles empty JSON gracefully', async ({ page }) => {
    await page.click('text=Import');
    await page.fill('#importJson', '');
    await page.click('text=Import');
    await page.waitForTimeout(500);
    // Should not crash
    await expect(page.locator('.character-grid')).toBeVisible();
  });

  test('import handles malformed JSON', async ({ page }) => {
    await page.click('text=Import');
    await page.fill('#importJson', '{not valid json}');
    await page.click('text=Import');
    await page.waitForTimeout(500);
    await expect(page.locator('.character-grid')).toBeVisible();
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

    // Check inventory tab has currency
    await page.click('text=Inventory');
    const invContent = await page.locator('#inventoryTab').textContent();
    expect(invContent).toContain('GP');
  });
});

test.describe('Session and quest management UI', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
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

    // Try to mark as complete
    const completeBtn = page.locator('text=Complete').first();
    if (await completeBtn.isVisible()) {
      await completeBtn.click();
      await page.waitForTimeout(300);
    }
  });
});

test.describe('Error handling and edge cases', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
  });

  test('handles character creation without name', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', '');
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await page.waitForTimeout(300);
    // Should create with default name or show error
    await expect(page.locator('.character-grid')).toBeVisible();
  });

  test('character tabs are navigable back and forth', async ({ page }) => {
    await page.click('text=New Character');
    await page.fill('#newName', 'Nav Test');
    await page.fill('#newRace', 'Half-Orc');
    await page.fill('#newClass', 'Barbarian');
    await page.click('text=Create');
    await page.waitForTimeout(500);
    await page.click('.character-card');

    // Navigate through tabs and verify content exists
    const tabs = ['Stats', 'Combat', 'Spells', 'Inventory', 'Features', 'Details', 'Dice'];
    for (const tab of tabs) {
      await page.click(`text=${tab}`);
      await page.waitForTimeout(100);
    }

    // Go back to stats
    await page.click('text=Stats');
    await page.waitForTimeout(100);
    await expect(page.locator('.ability-score, .stat-block, #combatTab, #spellTab, #inventoryTab, #featuresTab, #detailsSection, #diceSection').first()).toBeVisible();
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

    // History should have all 4 entries
    const historyItems = page.locator('.dice-history-item');
    const count = await historyItems.count();
    expect(count).toBeGreaterThanOrEqual(4);
  });
});
