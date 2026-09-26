import { test, expect } from './fixtures.js';
import { login } from './helpers.js';

test('WLED admin: add device, save, persist, test flash does not crash', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', (e) => errors.push(e.message));

  await login(page);
  await page.goto('/admin', { waitUntil: 'domcontentloaded', timeout: 10000 });

  await page.click('#tabWledBtn');
  await expect(page.locator('#adminWled')).toBeVisible();

  // ensure at least one row exists (loadWledSettings creates one empty row)
  // click Add Device to get a fresh row
  await page.locator('#adminWled button:has-text("Add Device")').click();

  const lastRow = page.locator('.wled-device').last();
  await expect(lastRow).toBeVisible();
  await lastRow.locator('.wled-name').fill('Table');
  await lastRow.locator('.wled-url').fill('http://wled.local');
  await lastRow.locator('.wled-bri').fill('128');

  // Save
  await page.locator('#adminWled button:has-text("Save WLED Settings")').click();
  // toast appears; wait for response
  await page.waitForResponse((r) => r.url().includes('/api/admin/wled-settings') && r.request().method() === 'POST').catch(() => {});

  // Reload and assert persistence
  await page.goto('/admin', { waitUntil: 'domcontentloaded', timeout: 10000 });
  await page.click('#tabWledBtn');
  await expect(page.locator('#adminWled')).toBeVisible();

  // wait for loadWledSettings to populate
  await expect(page.locator('.wled-device').first()).toBeVisible({ timeout: 10000 });

  const nameInputs = page.locator('.wled-device .wled-name');
  // at least one row has value Table
  let found = false;
  const count = await nameInputs.count();
  for (let i = 0; i < count; i++) {
    const v = await nameInputs.nth(i).inputValue();
    if (v === 'Table') found = true;
  }
  expect(found).toBeTruthy();

  const urlInputs = page.locator('.wled-device .wled-url');
  let urlFound = false;
  const uc = await urlInputs.count();
  for (let i = 0; i < uc; i++) {
    const v = await urlInputs.nth(i).inputValue();
    if (v.includes('wled.local')) urlFound = true;
  }
  expect(urlFound).toBeTruthy();

  // Test Flash - should not crash
  await page.locator('#adminWled button:has-text("Test Flash")').click();
  await expect(page.locator('#adminWled')).toBeVisible();
  // no uncaught page errors
  expect(errors).toEqual([]);
});
