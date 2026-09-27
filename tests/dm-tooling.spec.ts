import { test, expect } from './fixtures.js';
import { login } from './helpers.js';

const uniqueName = () => `DM-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function createCampaign(page, name: string) {
  return page.evaluate(async (n: string) => window.api('POST', '/api/campaigns', { name: n, description: 'dm tooling e2e', dm_notes: '' }), name);
}

test.describe('DM tooling', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('DM tools generate treasure, weather, names and budget', async ({ page }) => {
    const name = uniqueName();
    const camp = await createCampaign(page, name);
    await page.evaluate((c: any) => (window as any).showCampaignDashboard(c.id, c.name), camp);

    await expect(page.getByTestId('dash-combat-analytics')).toBeVisible();
    await expect(page.getByTestId('open-dm-tools')).toBeVisible();
    await page.getByTestId('open-dm-tools').click();

    await expect(page.getByTestId('dm-treasure-tier')).toBeVisible();
    await expect(page.getByTestId('dm-weather-biome')).toBeVisible();
    await expect(page.getByTestId('dm-weather-season')).toBeVisible();
    await expect(page.getByTestId('dm-name-race')).toBeVisible();
    await expect(page.getByTestId('dm-budget-levels')).toBeVisible();

    await page.getByTestId('dm-tool-treasure').click();
    await expect(page.locator('#dmToolsResult')).toContainText('Treasure Hoard');

    await page.getByTestId('dm-tool-weather').click();
    await expect(page.locator('#dmToolsResult')).toContainText('Weather');

    await page.getByTestId('dm-tool-name').click();
    await expect(page.locator('#dmToolsResult')).toContainText('Name');

    await page.getByTestId('dm-tool-budget').click();
    await expect(page.locator('#dmToolsResult')).toContainText('XP');
  });

  test('campaign DM screen shows the live aggregate', async ({ page }) => {
    const name = uniqueName();
    const camp = await createCampaign(page, name);
    await page.evaluate((c: any) => (window as any).showCampaignDashboard(c.id, c.name), camp);

    await expect(page.getByTestId('open-dm-screen')).toBeVisible();
    await page.getByTestId('open-dm-screen').click();

    await expect(page.getByTestId('dm-screen-content')).toBeVisible();
    await expect(page.getByTestId('dm-screen-content')).toContainText(name);
    await expect(page.getByTestId('dm-screen-refresh')).toBeVisible();
  });
});
