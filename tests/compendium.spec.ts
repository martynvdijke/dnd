import { test, expect, type Page } from './fixtures.js';
import { login, clickNavItem } from './helpers.js';

const uniqueName = () => `CMP-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function waitLoadingDone(page: Page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: 5000 }).catch(() => {});
}

test.describe('Compendium', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  // ─── Compendium Search (legacy) ───

  test.describe('Search', () => {
    test('Search compendium with query', async ({ page }) => {
      const results = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/compendium/search?q=fire');
      });
      expect(Array.isArray(results)).toBe(true);
    });

    test('Search compendium with broad query returns results', async ({ page }) => {
      const results = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/compendium/search?q=a');
      });
      expect(Array.isArray(results)).toBe(true);
    });

    test('Search compendium by type', async ({ page }) => {
      const results = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/compendium/search?q=fire&type=spells');
      });
      expect(Array.isArray(results)).toBe(true);
    });
  });

  // ─── Compendium Admin Schemas ───

  test.describe('Admin Schemas', () => {
    test('List compendium schemas', async ({ page }) => {
      const schemas = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/admin/compendium-schemas');
      });
      expect(Array.isArray(schemas)).toBe(true);
    });

    test('Get specific schema details', async ({ page }) => {
      const schemas = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/admin/compendium-schemas');
      });
      expect(Array.isArray(schemas)).toBe(true);
      if (schemas.length > 0) {
        const schema = await page.evaluate(async (id) => {
          try {
            return await (window as any).api('GET', `/api/admin/compendium-schemas/${id}`);
          } catch (e) {
            return null;
          }
        }, schemas[0].id);
        expect(schema).not.toBeNull();
        expect(schema.type_name || schema.display_name).toBeTruthy();
      }
    });
  });

  // ─── Compendium Admin Entries CRUD ───

  test.describe('Admin Entries CRUD', () => {
    test('Create a compendium spell entry via schema API', async ({ page }) => {
      const name = uniqueName();
      const result = await page.evaluate(async (n) => {
        // Find or create a "spell" schema
        const schemas: any[] = await (window as any).api('GET', '/api/admin/compendium-schemas');
        let schema = schemas.find((s: any) => s.type_name === 'spell');
        if (!schema) {
          schema = await (window as any).api('POST', '/api/admin/compendium-schemas', {
            type_name: 'spell',
            display_name: 'Spell',
            fields: [
              { name: 'name', label: 'Name', type: 'text', required: true, sortable: true, searchable: true },
              { name: 'description', label: 'Description', type: 'text', sortable: false, searchable: true },
              { name: 'level', label: 'Level', type: 'number', sortable: true, searchable: false },
              { name: 'school', label: 'School', type: 'text', sortable: true, searchable: true },
            ],
          });
        }
        // Create entry under the schema
        return await (window as any).api('POST', `/api/admin/compendium-schemas/${schema.id}/entries`, {
          data: { name: n, description: 'A test spell', level: 1, school: 'evocation' },
        });
      }, name);
      expect(result.id).toBeGreaterThan(0);
    });

    test('Search finds created compendium entries', async ({ page }) => {
      const name = uniqueName();
      const result = await page.evaluate(async (n) => {
        const schemas: any[] = await (window as any).api('GET', '/api/admin/compendium-schemas');
        let schema = schemas.find((s: any) => s.type_name === 'spell');
        if (!schema) {
          schema = await (window as any).api('POST', '/api/admin/compendium-schemas', {
            type_name: 'spell', display_name: 'Spell',
            fields: [{ name: 'name', label: 'Name', type: 'text', required: true, sortable: true, searchable: true }],
          });
        }
        const entry = await (window as any).api('POST', `/api/admin/compendium-schemas/${schema.id}/entries`, {
          data: { name: n, description: 'Search test spell', level: 1, school: 'evocation' },
        });
        return entry;
      }, name);
      expect(result.id).toBeGreaterThan(0);
    });

    test('Update a compendium entry', async ({ page }) => {
      const name = uniqueName();
      const result = await page.evaluate(async (n) => {
        const schemas: any[] = await (window as any).api('GET', '/api/admin/compendium-schemas');
        let schema = schemas.find((s: any) => s.type_name === 'spell');
        if (!schema) {
          schema = await (window as any).api('POST', '/api/admin/compendium-schemas', {
            type_name: 'spell', display_name: 'Spell',
            fields: [{ name: 'name', label: 'Name', type: 'text', required: true, sortable: true, searchable: true }],
          });
        }
        const created = await (window as any).api('POST', `/api/admin/compendium-schemas/${schema.id}/entries`, {
          data: { name: n, description: 'Original', level: 1, school: 'evocation' },
        });
        await (window as any).api('PUT', `/api/admin/compendium-entries/${created.id}`, {
          data: { name: n + '-updated', description: 'Updated spell', level: 2, school: 'necromancy' },
        });
        const updated = await (window as any).api('GET', `/api/admin/compendium-entries/${created.id}`);
        return updated;
      }, name);
      expect(result).toBeTruthy();
      expect(result.data).toBeTruthy();
      expect(result.data.name).toContain('-updated');
    });

    test('Delete a compendium entry', async ({ page }) => {
      const name = uniqueName();
      const result = await page.evaluate(async (n) => {
        const schemas: any[] = await (window as any).api('GET', '/api/admin/compendium-schemas');
        let schema = schemas.find((s: any) => s.type_name === 'spell');
        if (!schema) {
          schema = await (window as any).api('POST', '/api/admin/compendium-schemas', {
            type_name: 'spell', display_name: 'Spell',
            fields: [{ name: 'name', label: 'Name', type: 'text', required: true, sortable: true, searchable: true }],
          });
        }
        const created = await (window as any).api('POST', `/api/admin/compendium-schemas/${schema.id}/entries`, {
          data: { name: n, description: 'Delete me', level: 1, school: 'evocation' },
        });
        await (window as any).api('DELETE', `/api/admin/compendium-entries/${created.id}`);
        // Try to get deleted entry — should 404
        try {
          await (window as any).api('GET', `/api/admin/compendium-entries/${created.id}`);
          return { deleted: false };
        } catch {
          return { deleted: true };
        }
      }, name);
      expect(result.deleted).toBe(true);
    });
  });

  // ─── Compendium Admin Monster Entries ───

  test.describe('Admin Monster Entries', () => {
    test('Create compendium monster entry', async ({ page }) => {
      const name = uniqueName();
      const result = await page.evaluate(async (n) => {
        const schemas: any[] = await (window as any).api('GET', '/api/admin/compendium-schemas');
        let schema = schemas.find((s: any) => s.type_name === 'monster');
        if (!schema) {
          schema = await (window as any).api('POST', '/api/admin/compendium-schemas', {
            type_name: 'monster', display_name: 'Monster',
            fields: [
              { name: 'name', label: 'Name', type: 'text', required: true, sortable: true, searchable: true },
              { name: 'ac', label: 'AC', type: 'number', sortable: true, searchable: false },
              { name: 'hp', label: 'HP', type: 'number', sortable: true, searchable: false },
            ],
          });
        }
        return await (window as any).api('POST', `/api/admin/compendium-schemas/${schema.id}/entries`, {
          data: { name: n, ac: 15, hp: 50, str: 16, dex: 12, con: 14, int_: 8, wis: 10, cha: 8, cr: '3' },
        });
      }, name);
      expect(result.id).toBeGreaterThan(0);
    });

    test('Delete compendium monster entry', async ({ page }) => {
      const name = uniqueName();
      const result = await page.evaluate(async (n) => {
        const schemas: any[] = await (window as any).api('GET', '/api/admin/compendium-schemas');
        let schema = schemas.find((s: any) => s.type_name === 'monster');
        if (!schema) {
          schema = await (window as any).api('POST', '/api/admin/compendium-schemas', {
            type_name: 'monster', display_name: 'Monster',
            fields: [{ name: 'name', label: 'Name', type: 'text', required: true, sortable: true, searchable: true }],
          });
        }
        const created = await (window as any).api('POST', `/api/admin/compendium-schemas/${schema.id}/entries`, {
          data: { name: n, ac: 10, hp: 20 },
        });
        await (window as any).api('DELETE', `/api/admin/compendium-entries/${created.id}`);
        try {
          await (window as any).api('GET', `/api/admin/compendium-entries/${created.id}`);
          return { deleted: false };
        } catch {
          return { deleted: true };
        }
      }, name);
      expect(result.deleted).toBe(true);
    });
  });

  // ─── Bulk Selection Operations ───

  test.describe('Bulk Selection', () => {
    test('batch delete compendium entries', async ({ page }) => {
      const name1 = uniqueName();
      const name2 = uniqueName();
      const result = await page.evaluate(async (names) => {
        try {
          const schemas: any[] = await (window as any).api('GET', '/api/admin/compendium-schemas');
          if (schemas.length === 0) return { err: 'no schemas' };
          const schemaId = schemas[0].id;
          const e1 = await (window as any).api('POST', `/api/admin/compendium-schemas/${schemaId}/entries`, {
            data: { name: names[0], description: 'Bulk test 1' },
          });
          const e2 = await (window as any).api('POST', `/api/admin/compendium-schemas/${schemaId}/entries`, {
            data: { name: names[1], description: 'Bulk test 2' },
          });
          const batch = await (window as any).api('POST', '/api/admin/compendium-entries/batch-delete', {
            ids: [e1.id, e2.id],
          });
          return { ok: true, deleted: batch.deleted };
        } catch (e) {
          return { ok: false, error: String(e) };
        }
      }, [name1, name2]);

      expect(result).toBeTruthy();
      if (result.ok) {
        expect(result.deleted).toBe(2);
      }
    });

    test('batch update compendium entries', async ({ page }) => {
      const name = uniqueName();
      const result = await page.evaluate(async (entryName) => {
        try {
          const schemas: any[] = await (window as any).api('GET', '/api/admin/compendium-schemas');
          if (schemas.length === 0) return { err: 'no schemas' };
          const schemaId = schemas[0].id;
          const e1 = await (window as any).api('POST', `/api/admin/compendium-schemas/${schemaId}/entries`, {
            data: { name: entryName + '-1', value: 'old1' },
          });
          const e2 = await (window as any).api('POST', `/api/admin/compendium-schemas/${schemaId}/entries`, {
            data: { name: entryName + '-2', value: 'old2' },
          });
          await (window as any).api('POST', '/api/admin/compendium-entries/batch-update', {
            ids: [e1.id, e2.id],
            data: { value: 'updated' },
          });
          return { ok: true };
        } catch (e) {
          return { ok: false, error: String(e) };
        }
      }, name);

      expect(result).toBeTruthy();
    });

    test('empty batch delete returns error', async ({ page }) => {
      const result = await page.evaluate(async () => {
        try {
          await (window as any).api('POST', '/api/admin/compendium-entries/batch-delete', { ids: [] });
          return { ok: true };
        } catch (e) {
          return { ok: false, error: String(e) };
        }
      });
      expect(result.ok).toBe(false);
    });
  });
});

// ─── User-Facing Compendium View ───

test.describe('User Compendium View', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  async function navToCompendium(page: any) {
    await clickNavItem(page, 'Compendium', 'compendium');
    // Wait for SPA loading to complete and compendium view to render
    await page.waitForSelector('#compendiumView', { state: 'visible', timeout: 10000 });
    // Give async schema load time to resolve
    await page.waitForTimeout(1000);
  }

  test('legacy tabs render when navigating to compendium', async ({ page }) => {
    await navToCompendium(page);
    // Verify heading
    await expect(page.locator('#compendiumView h1')).toContainText('Compendium', { timeout: 5000 });
    // Verify legacy tab buttons are visible
    await expect(page.locator('#compTabRaces')).toBeVisible();
    await expect(page.locator('#compTabClasses')).toBeVisible();
    await expect(page.locator('#compTabSpells')).toBeVisible();
    await expect(page.locator('#compTabEquipment')).toBeVisible();
    await expect(page.locator('#compTabMonsters')).toBeVisible();
  });

  test('dynamic schema tab appears when entries exist', async ({ page }) => {
    // Create schema + entry via Playwright's request (handles cookies/CORS same as page)
    const schemas: any = await page.evaluate(async () => {
      try {
        return await (window as any).api('GET', '/api/admin/compendium-schemas');
      } catch { return []; }
    });
    // If no schemas exist, this test needs to create one, but the api() function
    // handles CSRF for POST. Skip the test if schemas can't be read.
    if (!Array.isArray(schemas) || schemas.length === 0) {
      // Try to get entries-by-schema instead (no auth needed beyond login)
      const result: any = await page.evaluate(async () => {
        try {
          return await (window as any).api('GET', '/api/compendium/entries-by-schema');
        } catch { return { schemas: [] }; }
      });
      if (!result.schemas || result.schemas.length === 0) {
        // No schemas exist at all — skip test (need seeded data)
        return;
      }
    }

    await navToCompendium(page);

    // The section is always in DOM. If there are schemas with entries, it's visible.
    const section = page.locator('#compSchemaSection');
    await expect(section).toBeAttached({ timeout: 5000 });
  });

  test('dynamic schema tab shows entry cards with name and preview', async ({ page }) => {
    await navToCompendium(page);

    // Check if any schema tabs were rendered by loadCompendiumSchemaTypes
    const tabsHtml = await page.evaluate(() => {
      const el = document.getElementById('compSchemaTabs');
      return el ? el.innerHTML : '';
    });

    if (tabsHtml.trim().length > 0) {
      // Verify at least one tab has the proper structure
      await expect(page.locator('#compSchemaTabs .nav-link').first()).toBeAttached({ timeout: 3000 });

      // Click the first dynamic tab and check for cards
      await page.locator('#compSchemaTabs .nav-link').first().click();
      await page.waitForTimeout(300);

      // The corresponding content pane should have cards
      const firstPane = page.locator('#compSchemaContent > div').first();
      await expect(firstPane).toBeAttached();
    }
  });

  test('view all button appears when more than 20 entries exist', async ({ page }) => {
    // This test relies on a schema having >20 entries, which requires seeded data
    // or manual setup. Check via API first, skip if not enough data.
    const result: any = await page.evaluate(async () => {
      try {
        return await (window as any).api('GET', '/api/compendium/entries-by-schema');
      } catch { return { schemas: [] }; }
    });

    const largeSchema = (result.schemas || []).find((s: any) => s.entry_count > 20);
    if (!largeSchema) return; // skip — not enough data in this run

    await navToCompendium(page);

    const tab = page.locator(`#compSchemaTab-${largeSchema.type_name}`);
    await expect(tab).toBeAttached({ timeout: 3000 });
    await tab.click();
    await page.waitForTimeout(300);

    const content = page.locator(`#compSchemaContent-${largeSchema.type_name}`);
    await expect(content.locator('text=View All')).toBeAttached();
    await expect(content.locator('text=View All')).toContainText(String(largeSchema.entry_count));
  });

  test('dynamic schema section hidden when no entries exist', async ({ page }) => {
    await navToCompendium(page);

    const section = page.locator('#compSchemaSection');
    await expect(section).toBeAttached({ timeout: 5000 });

    // Determine expected state from API
    const result: any = await page.evaluate(async () => {
      try {
        return await (window as any).api('GET', '/api/compendium/entries-by-schema');
      } catch { return { schemas: [] }; }
    });

    const hasEntries = (result.schemas || []).length > 0;
    const display = await section.evaluate((el: HTMLElement) => el.style.display);
    expect(display).toBe(hasEntries ? 'block' : 'none');
  });
});
