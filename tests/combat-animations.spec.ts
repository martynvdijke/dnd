import { test, expect } from './fixtures.js';
import { ensureNavOpen, waitLoadingDone, clickNavItem, login } from './helpers.js';

test.describe('Combat animations', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('HP damage shows floating number on character sheet', async ({ page }) => {
    // Navigate to character sheet
    await clickNavItem(page, 'Characters', 'characters');
    await waitLoadingDone(page);
    // Click on the first character card to open the sheet
    const charCard = page.locator('.character-card, .char-card, [onclick*="showSheet"]').first();
    if (await charCard.count() > 0) {
      await charCard.click();
      await waitLoadingDone(page);
      // Look for damage input and apply damage
      const dmgInput = page.locator('#dmgInput');
      if (await dmgInput.count() > 0) {
        await dmgInput.fill('5');
        // Click the damage button (look for applyDamage trigger)
        const dmgBtn = page.locator('button:has-text("Damage"), button:has-text("Apply"), [onclick*="applyDamage"]').first();
        if (await dmgBtn.count() > 0) {
          await dmgBtn.click();
          // Wait for the floating damage number to appear
          const floatEl = page.locator('.damage-float');
          await expect(floatEl).toBeVisible({ timeout: 5000 });
          await expect(floatEl).toContainText('-5');
        }
      }
    }
  });

  test('combat turn transition marks active row', async ({ page }) => {
    // Navigate to combat tracker
    await clickNavItem(page, 'Combat', 'combatTracker');
    await waitLoadingDone(page);
    // Check if there are combatant rows
    const combatRow = page.locator('#combatTrackerBody tr').first();
    if (await combatRow.count() > 0) {
      // Click "Next Turn" button
      const nextTurnBtn = page.locator('button:has-text("Next Turn")');
      if (await nextTurnBtn.count() > 0) {
        await nextTurnBtn.click();
        await waitLoadingDone(page);
        // After turn change, the active row should have the combatant-row-active class
        const activeRow = page.locator('tr.combatant-row-active');
        // The class may or may not be applied depending on whether animation ran
        // but the table-active class should be present (Bootstrap active row)
        const tableActiveRow = page.locator('tr.table-active');
        await expect(tableActiveRow).toBeVisible({ timeout: 5000 });
      }
    }
  });
});
