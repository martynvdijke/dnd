import { test, expect } from './fixtures.js';
import { NAV_TIMEOUT, login, waitLoadingDone } from './helpers.js';

test.describe('Table Screen', () => {
  test('shared table screen shows campaign, party and initiative', async ({ page }) => {
    await login(page);
    await waitLoadingDone(page);

    const { token } = await page.evaluate(async () => {
      const api = (window as any).api;
      const camp = await api('POST', '/api/campaigns', { name: 'Table Camp', party_name: 'The Party' });
      const ch = await api('POST', '/api/characters', { name: 'Aria', race: 'Elf', class: 'Wizard' });
      await api('POST', `/api/campaigns/${camp.id}/characters`, { character_id: ch.id });
      await api('POST', '/api/combat', {
        campaign_id: camp.id, name: 'Goblin Boss', type: 'monster',
        initiative_roll: 15, initiative_mod: 2, hp_max: 30, hp_current: 20, ac: 14,
      });
      const share = await api('POST', '/api/share', { entity_type: 'table', entity_id: camp.id });
      return { token: share.token as string };
    });

    await page.goto('/share/' + token);
    await expect(page.locator('#campName')).toHaveText('Table Camp', { timeout: NAV_TIMEOUT });
    await expect(page.locator('body')).toContainText('The Party');
    await expect(page.locator('body')).toContainText('Goblin Boss');
    await expect(page.locator('body')).toContainText('Aria');
  });
});
