import { test, expect } from './fixtures.js';
import { clickNavItem, login, NAV_TIMEOUT, waitLoadingDone, waitModalClosed } from './helpers.js';

test.describe('edit affordances', () => {
  test('NPC edit from the party view', async ({ page }) => {
    await login(page);

    const ids = await page.evaluate(async () => {
      const name = 'NPC-' + Date.now();
      const campaign = await (window as any).api('POST', '/api/campaigns', { name: 'NPCCamp-' + Date.now(), party_name: 'P' });
      const char = await (window as any).api('POST', '/api/characters', { name: 'EditNPCChar-' + Date.now(), race: 'Human', class: 'Fighter' });
      await (window as any).api('POST', `/api/campaigns/${campaign.id}/characters`, { character_id: char.id });
      const npc = await (window as any).api('POST', '/api/npcs', { name, race: 'Elf', class: 'Wizard' });
      await (window as any).api('POST', `/api/characters/${char.id}/npcs`, { npc_id: npc.id, relationship: 'ally' });
      return { npcName: name };
    });

    await clickNavItem(page, 'party', 'party');
    await page.locator('#partySubTabBar').waitFor({ state: 'visible', timeout: NAV_TIMEOUT });
    await page.locator('#partySubTabBar button:has-text("NPCs")').click();
    await waitLoadingDone(page);

    const row = page.locator('#partyNpcList li').filter({ hasText: ids.npcName });
    await expect(row).toBeVisible({ timeout: NAV_TIMEOUT });
    await row.locator('button[title="Edit NPC"]').click();

    await expect(page.locator('#editNPCName')).toBeVisible({ timeout: 5000 });
    const newName = ids.npcName + ' Updated';
    await page.locator('#editNPCName').fill(newName);
    await page.getByRole('button', { name: 'Save Changes' }).click();
    await waitModalClosed(page);

    await expect(page.locator('#partyNpcList li').filter({ hasText: newName })).toBeVisible({ timeout: NAV_TIMEOUT });
  });

  test('crafting recipe edit', async ({ page }) => {
    await login(page);

    const ctx = await page.evaluate(async () => {
      const char = await (window as any).api('POST', '/api/characters', { name: 'CraftChar-' + Date.now(), race: 'Human', class: 'Artificer' });
      const charId = char.id;
      const uniq = 'Recipe-' + Date.now() + '-' + Math.random().toString(36).slice(2, 6);
      await (window as any).api('POST', '/api/crafting/recipes', {
        name: uniq,
        description: 'test recipe',
        category: 'potion',
        difficulty_dc: 10,
        crafting_time_hours: 2,
        required_tools: '[]',
        required_materials: '[]',
        result_item_name: 'Test Potion',
        result_item_category: 'potion',
        result_quantity: 1,
        result_description: '',
        notes: '',
      });
      return { charId, recipeName: uniq };
    });

    await page.evaluate((id) => (window as any).openChar(id), ctx.charId);
    await waitLoadingDone(page);

    await page.click('text=Crafting');
    await expect(page.locator('#craftingSection')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#craftingSection')).toContainText(ctx.recipeName, { timeout: 10000 });

    await page.locator('#craftingSection button[title="Edit recipe"]').first().click();
    await expect(page.locator('#recipeName')).toBeVisible({ timeout: 5000 });
    const newName = ctx.recipeName + '-Edited';
    await page.locator('#recipeName').fill(newName);
    await page.getByRole('button', { name: 'Save Recipe' }).click();
    await waitModalClosed(page);

    await expect(page.locator('#craftingSection')).toContainText(newName, { timeout: 10000 });
  });

  test('act-NPC edit via the act details editor', async ({ page }) => {
    test.slow(); // slow: template generation + HTMX + modal
    await login(page);

    const data = await page.evaluate(async () => {
      const adv = await (window as any).api('POST', '/api/oneshot-adventures/generate', { title: 'OS-' + Date.now(), template: 'five_room_dungeon', difficulty: 'easy', estimated_minutes: 60 });
      const detail = await (window as any).api('GET', `/api/oneshot-adventures/${adv.id}`);
      const actId = detail.acts[0].id;
      const created = await (window as any).api('POST', `/api/oneshot-acts/${actId}/npcs`, { name: 'Bandit', role: 'enemy', notes: '' });
      return { actId, npcId: created.id };
    });

    // Drive the real editor entry point (same function the act-details pen calls).
    await page.evaluate(({ npcId, actId }) => (window as any).editActNpc(npcId, actId), data);
    await expect(page.locator('#actNpcName')).toBeVisible({ timeout: 8000 });
    await page.locator('#actNpcName').fill('Bandit Chief');
    await page.locator('#genericModalBody button').filter({ hasText: /^Save$/ }).last().click({ timeout: 10000 });

    await expect.poll(async () => {
      const res = await page.request.get(`/api/oneshot-acts/${data.actId}/npcs`);
      const list = (await res.json()) as any[];
      return list.some((x) => x.name === 'Bandit Chief');
    }, { timeout: 15000 }).toBe(true);
  });
});
