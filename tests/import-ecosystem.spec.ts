import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT } from './helpers.js';

test.beforeEach(async ({ page }) => {
  await login(page);
});

const charName = () => `ExtChar-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`;

test('character dry-run then commit and logs + rollback', async ({ page }) => {
  const name = charName();
  const fixture = JSON.stringify({
    name,
    stats: [15, 14, 13, 12, 10, 8],
    classes: [{ definition: { name: 'Fighter' }, level: 3 }],
    baseHitPoints: 30,
    inventory: [],
  });

  // open modal
  await page.evaluate(() => { (window as any).showExternalImport(); });
  await expect(page.getByTestId('ext-import-kind')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-source')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-payload')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-url')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-dedup')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-preview')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-commit')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-open-logs')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-rows')).toBeVisible({ timeout: NAV_TIMEOUT });
  // kind default character, source should be dndbeyond
  await expect(page.getByTestId('ext-import-source')).toContainText('D&D Beyond');

  // switch to compendium to exercise schema select
  await page.getByTestId('ext-import-kind').selectOption('compendium');
  await expect(page.getByTestId('ext-import-schema')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-source')).toContainText('5e.tools');
  // switch back to character
  await page.getByTestId('ext-import-kind').selectOption('character');
  await expect(page.getByTestId('ext-import-source')).toContainText('D&D Beyond');

  // also via global function
  await page.evaluate(() => (window as any).extImportKindChanged());

  // fill payload
  await page.getByTestId('ext-import-payload').fill(fixture);
  await page.getByTestId('ext-import-dedup').selectOption('skip');
  // use url field empty

  // ensure no character with that name yet
  const countBefore = await page.evaluate(async (n) => {
    const chars: any[] = await (window as any).api('GET', '/api/characters');
    return chars.filter((c: any) => c.name === n).length;
  }, name);
  expect(countBefore).toBe(0);

  // preview (dry_run true) should not create
  await page.getByTestId('ext-import-preview').click();
  await expect(page.getByTestId('ext-import-rows')).toContainText(name, { timeout: NAV_TIMEOUT });

  const countAfterPreview = await page.evaluate(async (n) => {
    const chars: any[] = await (window as any).api('GET', '/api/characters');
    return chars.filter((c: any) => c.name === n).length;
  }, name);
  expect(countAfterPreview).toBe(0);

  // commit
  await page.getByTestId('ext-import-commit').click();
  await page.waitForFunction(async (n) => {
    const chars: any[] = await (window as any).api('GET', '/api/characters');
    return chars.some((c: any) => c.name === n);
  }, name, { timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-rows')).toContainText(name, { timeout: NAV_TIMEOUT });

  // open via button data-testid open-external-import from showImport
  await page.evaluate(() => { (window as any).showImport(); });
  await expect(page.getByTestId('open-external-import')).toBeVisible({ timeout: NAV_TIMEOUT });
  await page.getByTestId('open-external-import').click();
  await expect(page.getByTestId('ext-import-kind')).toBeVisible({ timeout: NAV_TIMEOUT });

  // logs + rollback
  await page.evaluate(() => (window as any).showExternalImportLogs());
  await expect(page.getByTestId('ext-import-log').first()).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-rollback').first()).toBeVisible({ timeout: NAV_TIMEOUT });

  page.once('dialog', (d) => d.accept());
  await page.getByTestId('ext-import-rollback').first().click();

  await page.waitForFunction(async (n) => {
    const chars: any[] = await (window as any).api('GET', '/api/characters');
    return !chars.some((c: any) => c.name === n);
  }, name, { timeout: NAV_TIMEOUT });

  const countAfterRollback = await page.evaluate(async (n) => {
    const chars: any[] = await (window as any).api('GET', '/api/characters');
    return chars.filter((c: any) => c.name === n).length;
  }, name);
  expect(countAfterRollback).toBe(0);
});

test('references all required testids', async ({ page }) => {
  await page.evaluate(() => (window as any).showExternalImport());
  await expect(page.getByTestId('ext-import-kind')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-source')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-payload')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-url')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-schema')).toBeHidden({ timeout: NAV_TIMEOUT });
  // show schema by switching kind
  await page.getByTestId('ext-import-kind').selectOption('compendium');
  await expect(page.getByTestId('ext-import-schema')).toBeVisible({ timeout: NAV_TIMEOUT });
  await page.getByTestId('ext-import-kind').selectOption('character');
  await expect(page.getByTestId('ext-import-dedup')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-preview')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-commit')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-open-logs')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('ext-import-rows')).toBeVisible({ timeout: NAV_TIMEOUT });

  await page.evaluate(() => (window as any).showImport());
  await expect(page.getByTestId('open-external-import')).toBeVisible({ timeout: NAV_TIMEOUT });
  await expect(page.getByTestId('open-external-logs')).toBeVisible({ timeout: NAV_TIMEOUT });
});
