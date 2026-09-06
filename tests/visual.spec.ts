import { test, expect } from './fixtures.js';
import { login, waitLoadingDone, clickNavItem, isMobile, NAV_TIMEOUT } from './helpers.js';

test.describe('Visual regression', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await waitLoadingDone(page);
  });

  test('@visual nav sidebar on desktop', async ({ page }) => {
    test.info().annotations.push({ type: 'visual', description: 'Desktop navigation sidebar' });
    if (await isMobile(page)) test.skip();
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
    await expect(page.locator('#appSidebar')).toHaveScreenshot('nav-sidebar.png');
  });

  test('@visual compendium legacy tabs', async ({ page }) => {
    test.info().annotations.push({ type: 'visual', description: 'Compendium page with legacy tabs' });
    await clickNavItem(page, 'compendium', 'compendium');
    await expect(page.locator('#compendiumView')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#compendiumTabs')).toBeVisible({ timeout: NAV_TIMEOUT });
    // Select a stable anchor — use compendiumTabs row (content below changes with scroll)
    await page.locator('#compTabRaces').click();
    await expect(page.locator('#compRaces .card').first()).toBeVisible({ timeout: 10000 });
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
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

    await clickNavItem(page, 'compendium', 'compendium');
    await expect(page.locator('#compSchemaTabs')).toBeVisible({ timeout: 8000 });
    await expect(page.locator('#compSchemaTabs .nav-link').first()).toBeVisible({ timeout: 3000 });
    // Use type_name to target the correct tab (avoids ambiguity from retry duplicates)
    const tab = page.locator(`#compSchemaTab-${typeName}`);
    await expect(tab).toBeVisible({ timeout: 8000 });
    await tab.click();
    // Target the specific schema's content pane by type_name
    await expect(page.locator(`#compSchemaContent-${typeName} .card`).first()).toBeVisible({ timeout: NAV_TIMEOUT });

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
    await clickNavItem(page, 'characters', 'characters');
    await expect(page.locator('#charactersView')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
    // Characters view content is dynamic — snapshot the heading row for stability
    await expect(page.locator('#charactersView h1')).toHaveScreenshot('characters-view.png');
  });
});
