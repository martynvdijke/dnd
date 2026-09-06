import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT } from './helpers.js';

test.describe('Session Mode', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('session mode topbar element exists in DOM', async ({ page }) => {
    const topbar = page.locator('#sessionModeTopbar');
    await expect(topbar).toBeAttached({ timeout: NAV_TIMEOUT });
    await expect(topbar).not.toBeVisible();
    await expect(topbar.locator('.sm-campaign')).toContainText('Session Mode');
    await expect(page.locator('#sessionModeExitBtn')).toContainText('Exit');
  });

  test('session mode CSS classes are defined in stylesheet', async ({ page }) => {
    const hasStyles = await page.evaluate(() => {
      const sheets = document.styleSheets;
      for (let i = 0; i < sheets.length; i++) {
        try {
          const rules = sheets[i].cssRules;
          if (!rules) continue;
          for (let j = 0; j < rules.length; j++) {
            const cssText = rules[j].cssText || '';
            if (cssText.includes('.session-mode')) return true;
          }
        } catch (e) { continue; }
      }
      return false;
    });
    expect(hasStyles).toBeTruthy();
  });

  test('session mode topbar is present in HTML', async ({ page }) => {
    const html = await page.locator('#sessionModeTopbar').innerHTML();
    expect(html).toContain('Session Mode');
    expect(html).toContain('Exit');
  });
});
