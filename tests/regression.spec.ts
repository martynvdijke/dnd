import { test, expect } from './fixtures.js';
import { login, clickNavItem, clickSecondaryNavItem, isMobile, waitLoadingDone } from './helpers.js';

// Regression tests — run with: npx playwright test --grep @regression
test.describe('Regression suite', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await waitLoadingDone(page);
  });

  async function navTo(page: any, desktopNav: string, bottomNav: string, moreText?: string) {
    if (await isMobile(page)) {
      if (moreText) {
        await clickSecondaryNavItem(page, moreText, 'moreNav');
      } else {
        await clickNavItem(page, desktopNav, bottomNav);
      }
    } else {
      const navMap: Record<string, string> = {
        'Characters': 'characters', 'Compendium': 'compendium', 'Dice': 'dice',
        'Encounters': 'encounters', 'Combat': 'combatTracker',
      };
      await page.locator(`#appSidebar button[data-nav="${navMap[desktopNav]}"]`).click();
    }
  }

  test('@regression Characters view renders', async ({ page }) => {
    test.info().annotations.push({ type: 'regression', description: 'Characters view loads' });
    await navTo(page, 'Characters', 'characters');
    // On fresh DB without characters, the view still shows #charactersView
    await expect(page.locator('#charactersView')).toBeVisible({ timeout: 5000 });
  });

  test('@regression Compendium view renders', async ({ page }) => {
    test.info().annotations.push({ type: 'regression', description: 'Compendium view loads with tabs' });
    await navTo(page, 'Compendium', 'compendium');
    await expect(page.locator('#compendiumView')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#compendiumTabs')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('h1:has-text("Compendium")')).toBeVisible({ timeout: 5000 });
  });

  test('@regression Dice view renders', async ({ page }) => {
    test.info().annotations.push({ type: 'regression', description: 'Dice view loads' });
    await navTo(page, 'Dice', 'dice');
    await expect(page.locator('#diceView')).toBeVisible({ timeout: 5000 });
  });

  test('@regression Encounters view renders', async ({ page }) => {
    test.info().annotations.push({ type: 'regression', description: 'Encounters view loads' });
    await navTo(page, 'Encounters', 'encounters');
    await expect(page.locator('#encounterView')).toBeVisible({ timeout: 5000 });
  });

  test('@regression compendium legacy tabs load content', async ({ page }) => {
    test.info().annotations.push({ type: 'regression', description: 'Each legacy tab shows content' });
    await navTo(page, 'Compendium', 'compendium');
    await expect(page.locator('#compendiumView')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#compendiumTabs')).toBeVisible({ timeout: 5000 });

    const tabCount = await page.evaluate(async () => {
      const tabs = ['races', 'classes', 'spells', 'feats', 'backgrounds', 'equipment', 'monsters'];
      let loaded = 0;
      for (const tab of tabs) {
        try {
          const data = await (window as any).api('GET', `/api/compendium/${tab}`);
          if (Array.isArray(data) && data.length > 0) loaded++;
        } catch { /* tab may not exist */ }
      }
      return loaded;
    });
    expect(tabCount).toBeGreaterThanOrEqual(6);
  });

  test('@regression schema entry lifecycle', async ({ page }) => {
    test.info().annotations.push({ type: 'regression', description: 'Schema entry lifecycle end-to-end' });
    const timestamp = Date.now();
    const typeName = `lifecycle-${timestamp}`;

    // Create schema via admin API
    const createResp: any = await page.evaluate((tn) => {
      return (window as any).api('POST', '/api/admin/compendium-schemas', {
        type_name: tn,
        display_name: 'Lifecycle Test',
        fields: [{ name: 'name', type: 'text', required: true }, { name: 'description', type: 'text' }],
      });
    }, typeName);
    expect(createResp).toBeTruthy();
    const schemaId = createResp.id;
    expect(schemaId).toBeGreaterThan(0);

    // Create entry in schema
    const entryData = { name: `Test-${timestamp}`, description: 'Regression test entry' };
    const entryResp: any = await page.evaluate(({ sid, data }) => {
      return (window as any).api('POST', `/api/admin/compendium-schemas/${sid}/entries`, { data });
    }, { sid: schemaId, data: entryData });
    expect(entryResp).toBeTruthy();
    const entryId = entryResp.id;
    expect(entryId).toBeGreaterThan(0);

    // Navigate to compendium
    await navTo(page, 'Compendium', 'compendium');
    await expect(page.locator('#compendiumView')).toBeVisible({ timeout: 5000 });

    // Wait for the async schema API to load and tabs to render
    await expect(page.locator('#compSchemaTabs')).toBeVisible({ timeout: 8000 });
    await expect(page.locator('#compSchemaTabs .nav-link').first()).toBeVisible({ timeout: 3000 });

    // Use type_name to target the correct tab (avoids ambiguity from retry duplicates)
    const schemaTab = page.locator(`#compSchemaTab-${typeName}`);
    await expect(schemaTab).toBeVisible({ timeout: 8000 });

    // Click tab and verify content in the schema's content pane
    await schemaTab.click();
    const contentPane = page.locator(`#compSchemaContent-${typeName}`);
    await expect(contentPane).toBeVisible({ timeout: 3000 });
    const contentEntry = contentPane.locator('.card-body strong').filter({ hasText: `Test-${timestamp}` });
    await expect(contentEntry).toBeVisible({ timeout: 5000 });

    // Delete entry and schema
    await page.evaluate((eid: number) => {
      return (window as any).api('DELETE', `/api/admin/compendium-entries/${eid}`);
    }, entryId);
    await page.evaluate((sid: number) => {
      return (window as any).api('DELETE', `/api/admin/compendium-schemas/${sid}`);
    }, schemaId);

    // Refresh compendium view — the deleted schema's tab should be gone
    await navTo(page, 'Compendium', 'compendium');
    // Seeds (backgrounds, monsters, etc.) keep the section visible,
    // so verify the lifecycle tab is absent rather than the whole section
    await expect(page.locator(`#compSchemaTab-${typeName}`)).not.toBeVisible({ timeout: 8000 });
  });

  test('@regression API response times under 2s', async ({ page }) => {
    test.info().annotations.push({ type: 'regression', description: 'Key API endpoints respond within 2s' });
    const endpoints = [
      '/api/compendium/races',
      '/api/compendium/entries-by-schema',
    ];
    for (const endpoint of endpoints) {
      const time = await page.evaluate(async (ep) => {
        const start = performance.now();
        await (window as any).api('GET', ep);
        return performance.now() - start;
      }, endpoint);
      expect(time, `${endpoint} should respond within 2s`).toBeLessThan(2000);
    }
  });
});
