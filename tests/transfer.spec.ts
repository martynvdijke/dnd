import { test, expect } from './fixtures.js';
import { NAV_TIMEOUT, login, waitLoadingDone } from './helpers.js';

test.describe('Data Transfer UI', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await waitLoadingDone(page);
  });

  test('transfer dialogs are wired into the app bundle', async ({ page }) => {
    const hasExport = await page.evaluate(() => typeof (window as any).showTransferExport);
    const hasImport = await page.evaluate(() => typeof (window as any).showTransferImport);
    expect(hasExport).toBe('function');
    expect(hasImport).toBe('function');

    await page.evaluate(() => (window as any).showTransferExport());
    await expect(page.locator('#genericModalTitle')).toHaveText('Export Data');
    await expect(page.locator('.transfer-type-chip')).toHaveCount(13);
  });

  test('campaign dashboard exposes campaign export', async ({ page }) => {
    const cid = await page.evaluate(async () => {
      const created = await (window as any).api('POST', '/api/campaigns', {
        name: 'Export Camp', party_name: 'Testers', description: '', dm_notes: '',
      });
      return created.id as number;
    });

    await page.evaluate((id) => (window as any).showCampaignDashboard(id, 'Export Camp'), cid);
    await expect(page.locator('#campaignDashContent')).toContainText('Export Campaign', { timeout: NAV_TIMEOUT });
  });
});
