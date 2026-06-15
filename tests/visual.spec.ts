import { test, expect } from '@playwright/test';
import { login, waitLoadingDone, clickNavItem, isMobile } from './helpers.js';

test.describe('Visual regression', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await waitLoadingDone(page);
  });

  test('@visual nav sidebar on desktop', async ({ page }) => {
    test.info().annotations.push({ type: 'visual', description: 'Desktop navigation sidebar' });
    if (await isMobile(page)) test.skip();
    await page.waitForTimeout(500);
    await expect(page.locator('#appSidebar')).toHaveScreenshot('nav-sidebar.png');
  });

  test('@visual compendium legacy tabs', async ({ page }) => {
    test.info().annotations.push({ type: 'visual', description: 'Compendium page with legacy tabs' });
    await clickNavItem(page, 'Compendium', 'compendium');
    await expect(page.locator('#compendiumView')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#compendiumTabs')).toBeVisible({ timeout: 5000 });
    // Select a stable anchor — use compendiumTabs row (content below changes with scroll)
    await page.locator('#compTabRaces').click();
    await expect(page.locator('#compRaces .card').first()).toBeVisible({ timeout: 10000 });
    await page.waitForTimeout(300);
    await expect(page.locator('#compendiumTabs')).toHaveScreenshot('compendium-legacy-tabs.png');
  });

  test('@visual compendium with dynamic schema tab', async ({ page }) => {
    test.info().annotations.push({ type: 'visual', description: 'Compendium with dynamic schema tab visible' });
    const timestamp = Date.now();
    const typeName = `vis-${timestamp}`;

    // Create schema + entry via admin API
    const schema: any = await page.evaluate((tn) => {
      return (window as any).api('POST', '/api/admin/compendium-schemas', {
        type_name: tn,
        display_name: 'Visual Test',
        fields: [{ name: 'name', type: 'text', required: true }, { name: 'description', type: 'text' }],
      });
    }, typeName);
    await page.evaluate((sid: number) => {
      return (window as any).api('POST', `/api/admin/compendium-schemas/${sid}/entries`, {
        data: { name: 'Visual Entry', description: 'Screenshot reference' },
      });
    }, schema.id);

    await clickNavItem(page, 'Compendium', 'compendium');
    await expect(page.locator('#compSchemaTabs')).toBeVisible({ timeout: 8000 });
    await expect(page.locator('#compSchemaTabs .nav-link').first()).toBeVisible({ timeout: 3000 });
    const tab = page.locator('#compSchemaTabs button.nav-link').filter({ hasText: 'Visual Test' }).first();
    await expect(tab).toBeVisible({ timeout: 8000 });
    await tab.click();
    await expect(page.locator('#compSchemaContent .card').first()).toBeVisible({ timeout: 5000 });
    await page.waitForTimeout(300);
    await expect(page.locator('#compSchemaTabs')).toHaveScreenshot('compendium-schema-tab.png');

    // Cleanup
    const entries: any = await page.evaluate((sid) => {
      return (window as any).api('GET', `/api/admin/compendium-schemas/${sid}/entries`);
    }, schema.id);
    for (const e of entries.entries || []) {
      await page.evaluate((eid: number) => {
        return (window as any).api('DELETE', `/api/admin/compendium-entries/${eid}`);
      }, e.id);
    }
    await page.evaluate((sid: number) => {
      return (window as any).api('DELETE', `/api/admin/compendium-schemas/${sid}`);
    }, schema.id);
  });

  test('@visual character list view', async ({ page }) => {
    test.info().annotations.push({ type: 'visual', description: 'Characters view heading' });
    await clickNavItem(page, 'Characters', 'characters');
    await expect(page.locator('#charactersView')).toBeVisible({ timeout: 5000 });
    await page.waitForTimeout(300);
    // Characters view content is dynamic — snapshot the heading row for stability
    await expect(page.locator('#charactersView h1')).toHaveScreenshot('characters-view.png');
  });
});
