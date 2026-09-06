import { test, expect } from './fixtures.js';
import { waitLoadingDone, waitModalClosed, login } from './helpers.js';

// Acceptance: user-imported compendium entries (generic schema layer) must be
// linkable from character sheet inventory / spells / features — not just the
// ~17 legacy SRD rows — and unlinking must keep the snapshotted data.

const unique = () => `E2E-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function ensureSchema(page: any, typeName: string, displayName: string, fields: any[]) {
  return page.evaluate(async ({ type_name, display_name, fields }) => {
    const schemas: any[] = await (window as any).api('GET', '/api/admin/compendium-schemas');
    let schema = schemas.find((s: any) => s.type_name === type_name);
    if (!schema) {
      schema = await (window as any).api('POST', '/api/admin/compendium-schemas', {
        type_name, display_name, fields,
      });
    }
    return schema;
  }, { type_name: typeName, display_name: displayName, fields });
}

async function createEntry(page: any, schema: any, data: any) {
  return page.evaluate(async ({ schemaId, data }) => {
    return (window as any).api('POST', `/api/admin/compendium-schemas/${schemaId}/entries`, { data });
  }, { schemaId: schema.id, data });
}

async function newCharacter(page: any, name: string) {
  await page.click('text=New Character');
  await page.fill('#newName', name);
  await page.fill('#newRace', 'Elf');
  await page.fill('#newClass', 'Ranger');
  await page.click('text=Create');
  await waitModalClosed(page);
  await page.locator('.character-card').filter({ hasText: name }).click();
  await waitLoadingDone(page);
}

test.describe('Compendium entry linking', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('links imported compendium entries to inventory, spells and features', async ({ page }) => {
    const charName = `Char-${unique()}`;
    const itemName = `Staff-${unique()}`;
    const spellName = `Fireball-${unique()}`;
    const featName = `Keen-${unique()}`;

    // Create an item-like schema + entry (what a user import produces).
    const schema = await ensureSchema(page, 'homebrew-items', 'Homebrew Items', [
      { name: 'name', label: 'Name', type: 'text', required: true, sortable: true, searchable: true },
      { name: 'category', label: 'Category', type: 'text' },
      { name: 'cost', label: 'Cost', type: 'text' },
      { name: 'weight', label: 'Weight', type: 'number' },
      { name: 'description', label: 'Description', type: 'text' },
      { name: 'level', label: 'Level', type: 'number' },
      { name: 'school', label: 'School', type: 'text' },
      { name: 'casting_time', label: 'Casting Time', type: 'text' },
      { name: 'range', label: 'Range', type: 'text' },
      { name: 'components', label: 'Components', type: 'text' },
      { name: 'duration', label: 'Duration', type: 'text' },
    ]);
    await createEntry(page, schema, {
      name: itemName, category: 'Wondrous', cost: '50 gp', weight: 3,
      description: 'An imported staff of power.',
    });
    await createEntry(page, schema, {
      name: spellName, level: 3, school: 'Evocation', casting_time: '1 action',
      range: '150 feet', components: 'V,S', duration: 'Instantaneous',
      description: 'An imported fireball.',
    });
    await createEntry(page, schema, {
      name: featName, level: 1, description: 'An imported keen sight feature.',
    });

    await newCharacter(page, charName);

    // ── Inventory (client-rendered picker) ──
    await page.click('#tabBar button:has-text("Inventory")');
    await page.locator('#inventorySection button:has-text("Link from Compendium")').click();
    await page.fill('#cpSearch', itemName);
    await expect(page.locator('.cp-item').filter({ hasText: itemName }).first()).toBeVisible();
    await page.locator('.cp-item').filter({ hasText: itemName }).first().click();
    await waitModalClosed(page);
    await expect(page.locator('#inventorySection')).toContainText(itemName);
    await expect(page.locator('#inventorySection .badge-compendium')).toBeVisible();

    // Unlink keeps the item data, drops the link badge.
    await page.locator('#inventorySection button[title="Unlink from compendium"]').click();
    await expect(page.locator('#inventorySection')).toContainText(itemName);
    await expect(page.locator('#inventorySection .badge-compendium')).toHaveCount(0);

    // ── Spells (HTMX picker) ──
    await page.click('#tabBar button:has-text("Spells")');
    const setupBtn = page.locator('button:has-text("Set Up Spellcasting")');
    if (await setupBtn.count() > 0) {
      await setupBtn.click();
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
    }
    await expect(page.locator('#spellsSection')).toBeVisible();
    await page.locator('#spellsSection button:has-text("Link from Compendium")').click();
    await expect(page.locator('#compendiumSpellSearch')).toBeVisible();
    await page.locator('#compendiumSpellSearch').pressSequentially(spellName);
    const spellRow = page.locator('#compendiumSpellResults div:has-text("' + spellName + '")');
    await expect(spellRow.first()).toBeVisible();
    await spellRow.locator('button.btn-success').click();
    await waitModalClosed(page);
    await expect(page.locator('#spellsSection')).toContainText(spellName);
    await expect(page.locator('#spellsSection .badge').filter({ hasText: 'Compendium' })).toBeVisible();

    page.once('dialog', (d) => d.accept());
    await page.locator('#spellsSection button[hx-delete*="compendium-unlink"]').click();
    await expect(page.locator('#spellsSection')).toContainText(spellName);
    await expect(page.locator('#spellsSection .badge').filter({ hasText: 'Compendium' })).toHaveCount(0);

    // ── Features (HTMX picker) ──
    await page.click('#tabBar button:has-text("Features")');
    await expect(page.locator('#featuresSection')).toBeVisible();
    await page.locator('#featuresSection button:has-text("Link from Compendium")').click();
    await expect(page.locator('#compendiumFeatureSearch')).toBeVisible();
    await page.locator('#compendiumFeatureSearch').pressSequentially(featName);
    const featRow = page.locator('#compendiumFeatureResults div:has-text("' + featName + '")');
    await expect(featRow.first()).toBeVisible();
    await featRow.locator('button.btn-success').click();
    await waitModalClosed(page);
    await expect(page.locator('#featuresSection')).toContainText(featName);
    await expect(page.locator('#featuresSection .badge').filter({ hasText: 'Compendium' })).toBeVisible();

    page.once('dialog', (d) => d.accept());
    await page.locator('#featuresSection button[hx-delete*="compendium-unlink"]').click();
    await expect(page.locator('#featuresSection')).toContainText(featName);
    await expect(page.locator('#featuresSection .badge').filter({ hasText: 'Compendium' })).toHaveCount(0);
  });
});
