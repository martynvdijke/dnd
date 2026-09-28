import { test, expect } from './fixtures.js';
import { login } from './helpers.js';

test.describe('localization', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('switches navigation language and persists the choice', async ({ page }) => {
    const charsLabel = page.locator('.sidebar-nav-item[data-nav="characters"] .sidebar-label');
    await expect(charsLabel).toHaveText('Characters');

    await page.getByTestId('locale-select').selectOption('nl');
    await expect(charsLabel).toHaveText('Personages');

    // The choice survives a reload.
    await page.reload();
    await expect(page.getByTestId('locale-select')).toHaveValue('nl');
    await expect(page.locator('.sidebar-nav-item[data-nav="characters"] .sidebar-label')).toHaveText('Personages');
  });
});
