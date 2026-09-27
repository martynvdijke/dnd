import { test, expect } from './fixtures.js';
import { NAV_TIMEOUT, login, waitLoadingDone, waitModalClosed } from './helpers.js';

const uniqueName = () => `Wiz-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function createAndOpen(page: import('@playwright/test').Page, name: string, race = 'Human', cls = 'Fighter') {
  await page.getByTestId('new-character').click();
  await page.fill('#newName', name);
  await page.fill('#newRace', race);
  await page.fill('#newClass', cls);
  await page.click('.modal button:has-text("Create")');
  await waitModalClosed(page);
  await page.locator('.character-card').filter({ hasText: name }).click();
  await waitLoadingDone(page);
  await expect(page.getByTestId('sheet-view')).toBeVisible();
}

test.describe('Player experience', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('guided builder creates a character through all steps', async ({ page }) => {
    const name = uniqueName();
    await page.getByTestId('new-character').click();
    await page.getByTestId('guided-builder').click();
    await expect(page.getByTestId('character-wizard')).toBeVisible();

    await page.fill('#wizName', name);
    await page.fill('#wizRace', 'Elf');
    await page.fill('#wizClass', 'Wizard');
    await page.getByTestId('wizard-next').click();

    await expect(page.locator('#wizAbil-str')).toBeVisible();
    await page.getByTestId('wizard-next').click();

    await expect(page.locator('#wizSkill-stealth')).toBeVisible();
    await page.locator('#wizSkill-stealth').check();
    await page.getByTestId('wizard-back').click();
    await expect(page.locator('#wizAbil-str')).toBeVisible();
    await page.getByTestId('wizard-next').click();
    await page.locator('#wizSkill-stealth').check();
    await page.getByTestId('wizard-next').click();

    await page.getByTestId('wizard-create').click();
    await waitModalClosed(page);
    await expect(page.getByText(name).first()).toBeVisible({ timeout: NAV_TIMEOUT });
  });

  test('inline rules help explains a sheet stat', async ({ page }) => {
    const name = uniqueName();
    await createAndOpen(page, name);

    await page.locator('#statsSection [data-rule="exhaustion"]').first().click();
    await expect(page.locator('#genericModal')).toBeVisible();
    await expect(page.locator('#genericModalTitle')).toHaveText('Exhaustion');
    await expect(page.locator('#genericModalBody')).toContainText('long rest');
  });

  test('accessibility: skip link, live region, and tab semantics', async ({ page }) => {
    await expect(page.getByTestId('skip-link')).toBeAttached();
    await page.getByTestId('skip-link').focus();
    await expect(page.getByTestId('skip-link')).toBeFocused();
    await expect(page.locator('main#main')).toHaveCount(1);
    await expect(page.locator('#toastContainer')).toHaveAttribute('aria-live', 'polite');

    const name = uniqueName();
    await createAndOpen(page, name);

    const activeTab = page.locator('#tabBar [role="tab"][aria-selected="true"]');
    await expect(activeTab).toHaveCount(1);
    await expect(activeTab).toHaveText(/Stats/);

    await activeTab.focus();
    await page.keyboard.press('ArrowRight');
    await expect(page.locator('#tabBar [role="tab"][aria-selected="true"]')).toHaveText(/Combat/);
    await expect(page.locator('#combatSection')).toBeVisible();
  });
});

test.describe('Player experience on tablet', () => {
  test.use({ viewport: { width: 768, height: 1024 } });

  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('builder and sheet work at 768x1024', async ({ page }) => {
    const name = uniqueName();
    await page.getByTestId('new-character').click();
    await page.getByTestId('guided-builder').click();
    await expect(page.getByTestId('character-wizard')).toBeVisible();
    await page.fill('#wizName', name);
    await page.getByTestId('wizard-next').click();
    await page.getByTestId('wizard-next').click();
    await page.getByTestId('wizard-next').click();
    await page.getByTestId('wizard-create').click();
    await waitModalClosed(page);

    await waitLoadingDone(page);
    await expect(page.getByTestId('sheet-view')).toBeVisible();
    await expect(page.locator('#tabBar [role="tab"][aria-selected="true"]')).toBeVisible();
  });
});
