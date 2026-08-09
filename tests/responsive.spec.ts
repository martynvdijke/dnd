import { test, expect } from './fixtures.js';
import { login, waitLoadingDone, NAV_TIMEOUT } from './helpers.js';

const uniqueName = () => `Resp-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function ensureNavOpen(page) {
  const toggler = page.locator('.navbar-toggler');
  if (await toggler.isVisible()) {
    await toggler.click();
    await page.waitForTimeout(300);
  }
}

async function waitModalClosed(page) {
  await page.waitForFunction(() => {
    const modal = document.getElementById('genericModal');
    return !modal || !modal.classList.contains('show');
  }, { timeout: 10000 }).catch(() => {});
}

test.describe('Responsive design', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('desktop layout works at 1280x720', async ({ page }) => {
    test.slow();
    await page.setViewportSize({ width: 1280, height: 720 });
    await expect(page.locator('.navbar')).toBeVisible();
    await expect(page.locator('.container').first()).toBeVisible();
  });

  test('mobile layout works at 767x1024', async ({ page }) => {
    test.slow();
    await page.setViewportSize({ width: 767, height: 1024 });
    // Fresh test DBs have no characters, which leaves #charGrid empty (zero
    // height) and "hidden" to Playwright. Create one so the grid renders.
    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await waitModalClosed(page);
    await waitLoadingDone(page);
    await expect(page.locator('.bottom-tab-bar')).toBeVisible();
    await expect(page.locator('#charGrid')).toContainText(name);
  });

  test('mobile layout works at 390x844 (iPhone 14)', async ({ page }) => {
    test.slow();
    await page.setViewportSize({ width: 390, height: 844 });

    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
    await expect(page.locator('#sheetName')).toBeVisible();

    await expect(page.locator('#statsSection .ability-box').first()).toBeVisible();
  });

  test('small mobile layout works at 320x568 (iPhone SE)', async ({ page }) => {
    test.slow();
    await page.setViewportSize({ width: 320, height: 568 });

    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Cleric');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await expect(page.locator('#sheetName')).toBeVisible();
    await page.click('#tabBar button:has-text("Combat")');
    await expect(page.locator('#combatSection')).toBeVisible();
  });

  test('dice roller is usable on mobile', async ({ page }) => {
    test.slow();
    await page.setViewportSize({ width: 390, height: 844 });
    await page.click('#bottomTabBar button[data-nav="dice"]');
    await expect(page.locator('#diceExpr')).toBeVisible({ timeout: NAV_TIMEOUT });

    await page.fill('#diceExpr', '1d20+5');
    await page.click('text=Roll the Bones');
    const result = page.locator('#diceResult');
    await expect(result).toBeVisible({ timeout: 10000 });
  });

  test('admin panel is responsive', async ({ page }) => {
    test.slow();
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(300);
    await page.goto('/admin', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('#adminUsers .card-header')).toContainText('Users');

    await page.click('#adminTabs button:has-text("Compendium")');
    await expect(page.locator('#adminUnifiedCompendium .card-header')).toContainText('Compendium');

    await page.click('#adminTabs button:has-text("Backup")');
    await expect(page.locator('#adminBackup .card-header').first()).toContainText('Backup Settings');
  });

  test('character grid adapts to viewport', async ({ page }) => {
    test.slow();
    const prefix = uniqueName();
    for (let i = 0; i < 3; i++) {
      await page.click('button:has-text("New Character")');
      await page.locator('#newName').waitFor({ state: 'visible', timeout: NAV_TIMEOUT });
      await page.fill('#newName', `${prefix}-${i}`);
      await page.fill('#newRace', 'Human');
      await page.fill('#newClass', 'Fighter');
      await page.click('.modal button:has-text("Create")');
      await waitModalClosed(page);
    }

    await page.setViewportSize({ width: 1280, height: 720 });
    await expect(page.locator('#charGrid')).toContainText(`${prefix}-0`, { timeout: NAV_TIMEOUT });

    await page.setViewportSize({ width: 390, height: 844 });
    await expect(page.locator('#charGrid')).toContainText(`${prefix}-0`, { timeout: NAV_TIMEOUT });
  });

  test('no horizontal page overflow on the dice view at 390px', async ({ page }) => {
    test.slow();
    await page.setViewportSize({ width: 390, height: 844 });
    await page.click('#bottomTabBar button[data-nav="dice"]');
    await expect(page.locator('#diceExpr')).toBeVisible({ timeout: NAV_TIMEOUT });

    const overflow = await page.evaluate(
      () => document.documentElement.scrollWidth - document.documentElement.clientWidth
    );
    expect(overflow).toBeLessThanOrEqual(0);
  });

  test('sheet tab bar sticks flush to the top on mobile', async ({ page }) => {
    test.slow();
    await page.setViewportSize({ width: 390, height: 844 });

    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
    await expect(page.locator('#sheetName')).toBeVisible();

    // Wait for the sheet's async sections to render enough content to scroll,
    // otherwise the sticky bar has no scroll range to pin within.
    await page.waitForFunction(
      () => document.documentElement.scrollHeight > window.innerHeight + 200,
      { timeout: NAV_TIMEOUT }
    );

    // The tab bar must be sticky at the very top (top:0) on mobile — the old
    // rule left a 56px gap reserved for the navbar, which is hidden on mobile.
    const tab = await page.evaluate(() => {
      const bar = document.getElementById('tabBar');
      if (!bar) return null;
      const cs = getComputedStyle(bar);
      return { position: cs.position, topCss: cs.top, natural: Math.round(bar.getBoundingClientRect().top) };
    });
    expect(tab).not.toBeNull();
    expect(tab.position).toBe('sticky');
    expect(tab.topCss).toBe('0px');

    // Scroll like a user and verify the bar actually pins flush at the top.
    await page.mouse.move(195, 400);
    await page.mouse.wheel(0, 4000);
    await page.waitForTimeout(400);
    const pinnedTop = await page.evaluate(() => {
      const bar = document.getElementById('tabBar');
      return bar ? Math.round(bar.getBoundingClientRect().top) : null;
    });
    expect(pinnedTop).not.toBeNull();
    expect(pinnedTop).toBeLessThanOrEqual(1);
  });

  test('sheet header actions stay on one row at 320px without page overflow', async ({ page }) => {
    test.slow();
    await page.setViewportSize({ width: 320, height: 568 });

    const name = uniqueName();
    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Cleric');
    await page.click('text=Create');
    await waitModalClosed(page);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
    await expect(page.locator('#sheetName')).toBeVisible();

    const result = await page.evaluate(() => {
      const strip = document.querySelector('#sheetView .d-flex.gap-2.no-print');
      if (!strip) return null;
      const rows = new Set(
        [...strip.querySelectorAll('.btn')].map((b) => Math.round(b.getBoundingClientRect().top))
      );
      return {
        distinctRows: rows.size,
        pageOverflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
      };
    });
    expect(result).not.toBeNull();
    expect(result.distinctRows).toBe(1); // buttons wrapped to two rows before the fix
    expect(result.pageOverflow).toBeLessThanOrEqual(0);
  });
});
