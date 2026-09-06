import { test, expect } from './fixtures.js';
import { login, waitLoadingDone, waitModalClosed, NAV_TIMEOUT } from './helpers.js';

const uniqueName = () => `Craft-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('Crafting System', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('crafting tab shows recipes', async ({ page }) => {
    const name = uniqueName();
    await page.click('button:has-text("New Character")');
    await page.locator('#newName').waitFor({ state: 'visible', timeout: NAV_TIMEOUT });
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Artificer');
    await page.click('.modal button:has-text("Create")');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('text=Crafting');
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
    await expect(page.locator('#craftingSection')).toContainText('Known Recipes');
    await expect(page.locator('#craftingSection')).toContainText('Potion of Healing');
  });

  test('starts crafting via API', async ({ page }) => {
    const name = uniqueName();
    await page.click('button:has-text("New Character")');
    await page.locator('#newName').waitFor({ state: 'visible', timeout: NAV_TIMEOUT });
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Wizard');
    await page.click('.modal button:has-text("Create")');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).waitFor({ state: 'visible', timeout: 10000 });

    // Get character ID from the card's onclick attribute
    const cardOnclick = await page.locator('.character-card').filter({ hasText: name }).getAttribute('onclick');
    const charId = cardOnclick ? parseInt(cardOnclick.replace(/\D/g, '')) : null;
    expect(charId).toBeTruthy();

    // Start crafting via direct API call
    const result = await page.evaluate(async (cid) => {
      try {
        const recipes = await window.api('GET', '/api/crafting/recipes');
        const recipe = recipes.find((r: any) => r.name === 'Potion of Healing');
        if (!recipe) return { err: 'Recipe not found' };
        await window.api('POST', `/api/characters/${cid}/crafting`, {
          recipe_id: recipe.id,
          name: recipe.name,
          total_hours_required: recipe.crafting_time_hours,
          dc: recipe.difficulty_dc,
          materials_allocated: '[]',
          notes: '',
        });
        return { ok: true };
      } catch (e: any) {
        return { err: e.message };
      }
    }, charId);

    expect(result.err).toBeUndefined();
    if (result.err) return;

    // Switch to crafting tab to see the project
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
    await page.click('text=Crafting');
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
    await expect(page.locator('#craftingSection')).toContainText('Potion of Healing');
    await expect(page.locator('#craftingSection')).toContainText('In Progress');
  });

  test('advances crafting progress via API', async ({ page }) => {
    const name = uniqueName();
    await page.click('button:has-text("New Character")');
    await page.locator('#newName').waitFor({ state: 'visible', timeout: NAV_TIMEOUT });
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Gnome');
    await page.fill('#newClass', 'Artificer');
    await page.click('.modal button:has-text("Create")');
    await waitModalClosed(page);

    const cardOnclick = await page.locator('.character-card').filter({ hasText: name }).getAttribute('onclick');
    const charId = cardOnclick ? parseInt(cardOnclick.replace(/\D/g, '')) : null;
    expect(charId).toBeTruthy();

    // Start crafting
    const startResult = await page.evaluate(async (cid) => {
      try {
        const recipes = await window.api('GET', '/api/crafting/recipes');
        const recipe = recipes.find((r: any) => r.name === 'Potion of Healing');
        if (!recipe) return { err: 'Recipe not found' };
        await window.api('POST', `/api/characters/${cid}/crafting`, {
          recipe_id: recipe.id,
          name: recipe.name,
          total_hours_required: recipe.crafting_time_hours,
          dc: recipe.difficulty_dc,
          materials_allocated: '[]',
          notes: '',
        });
        return { ok: true };
      } catch (e: any) {
        return { err: e.message };
      }
    }, charId);
    expect(startResult.err).toBeUndefined();

    // Advance via API
    await page.evaluate(async (cid) => {
      const projects = await window.api('GET', `/api/characters/${cid}/crafting`);
      const project = projects.find((p: any) => p.status === 'in-progress');
      if (project) {
        await window.api('PUT', `/api/crafting/${project.id}`, { progress_hours: project.total_hours_required });
      }
    }, charId);

    // Switch to crafting tab
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
    await page.click('text=Crafting');
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
    await expect(page.locator('#craftingSection')).toContainText('Potion of Healing');
    await expect(page.locator('#craftingSection')).toContainText('In Progress');
  });
});
