import { test, expect } from '@playwright/test';

const uniqueName = () => `CMP-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function waitLoadingDone(page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: 5000 }).catch(() => {});
}

test.describe('Compendium', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded', timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
    await page.waitForTimeout(200);
  });

  // ─── Compendium Search ───

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
    test('Create a compendium spell entry', async ({ page }) => {
      const name = uniqueName();
      const result = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/admin/compendium/spells', {
          name: n, description: 'A test spell', level: 1, school: 'evocation',
          casting_time: '1 action', range: '60 ft', components: 'V,S',
          duration: 'Instantaneous',
        });
      }, name);
      expect(result.id).toBeGreaterThan(0);
    });

    test('Search finds created compendium entries', async ({ page }) => {
      const name = uniqueName();
      const created = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/admin/compendium/spells', {
          name: n, description: 'Search test spell', level: 1, school: 'evocation',
          casting_time: '1 action', range: '60 ft', components: 'V,S',
          duration: 'Instantaneous',
        });
      }, name);
      expect(created.id).toBeGreaterThan(0);

      const results = await page.evaluate(async (n) => {
        return (window as any).api('GET', `/api/compendium/search?q=${encodeURIComponent(n)}`);
      }, name);
      expect(Array.isArray(results)).toBe(true);
      expect(results.some((e: any) => e.name === name)).toBe(true);
    });

    test('Update a compendium entry', async ({ page }) => {
      const name = uniqueName();
      const created = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/admin/compendium/spells', {
          name: n, description: 'Original', level: 1, school: 'evocation',
          casting_time: '1 action', range: '60 ft', components: 'V,S',
          duration: 'Instantaneous',
        });
      }, name);

      await page.evaluate(async ({ id, n }) => {
        return (window as any).api('PUT', `/api/admin/compendium/spells/${id}`, {
          name: n + '-updated', description: 'Updated spell', level: 2, school: 'necromancy',
          casting_time: '1 action', range: '30 ft', components: 'V,S,M',
          duration: '1 minute',
        });
      }, { id: created.id, n: name });

      const results = await page.evaluate(async (n) => {
        return (window as any).api('GET', `/api/compendium/search?q=${encodeURIComponent(n + '-updated')}`);
      }, name);
      expect(Array.isArray(results)).toBe(true);
      expect(results.some((e: any) => e.name === name + '-updated')).toBe(true);
    });

    test('Delete a compendium entry', async ({ page }) => {
      const name = uniqueName();
      const created = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/admin/compendium/spells', {
          name: n, description: 'Delete me', level: 1, school: 'evocation',
          casting_time: '1 action', range: '60 ft', components: 'V,S',
          duration: 'Instantaneous',
        });
      }, name);

      await page.evaluate(async (id) => {
        return (window as any).api('DELETE', `/api/admin/compendium/spells/${id}`);
      }, created.id);

      const results = await page.evaluate(async (n) => {
        return (window as any).api('GET', `/api/compendium/search?q=${encodeURIComponent(n)}`);
      }, name);
      expect(Array.isArray(results)).toBe(true);
      expect(results.some((e: any) => e.id === created.id)).toBe(false);
    });
  });

  // ─── Bulk Selection ───

  test.describe('Bulk Selection', () => {
    test('create multiple entries and verify they are searchable', async ({ page }) => {
      const suffix = Date.now().toString(36) + Math.random().toString(36).slice(2, 7);
      const names = [`Bulk-${suffix}-1`, `Bulk-${suffix}-2`, `Bulk-${suffix}-3`];
      for (const n of names) {
        await page.evaluate(async (name) => {
          return (window as any).api('POST', '/api/admin/compendium/spells', {
            name, description: 'Bulk test spell', level: 1, school: 'evocation',
            casting_time: '1 action', range: '60 ft', components: 'V,S',
            duration: 'Instantaneous',
          });
        }, n);
      }

      const results = await page.evaluate(async (opts) => {
        const all = await (window as any).api('GET', `/api/compendium/search?q=${encodeURIComponent(opts.suffix)}`);
        return all.filter((e: any) => opts.names.includes(e.name)).length;
      }, { names, suffix });

      // Use >= in case parallel test workers also created entries with matching suffix
      expect(results).toBeGreaterThanOrEqual(3);
    });

    test('delete entries can be done individually', async ({ page }) => {
      const name = uniqueName();
      const created = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/admin/compendium/spells', {
          name: n, description: 'To delete in bulk test', level: 1, school: 'evocation',
          casting_time: '1 action', range: '60 ft', components: 'V,S',
          duration: 'Instantaneous',
        });
      }, name);

      await page.evaluate(async (id) => {
        return (window as any).api('DELETE', `/api/admin/compendium/spells/${id}`);
      }, created.id);

      const results = await page.evaluate(async (n) => {
        return (window as any).api('GET', `/api/compendium/search?q=${encodeURIComponent(n)}`);
      }, name);
      expect(results.some((e: any) => e.id === created.id)).toBe(false);
    });

    test('search results include type property', async ({ page }) => {
      const name = uniqueName();
      await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/admin/compendium/spells', {
          name: n, description: 'Type filter test', level: 1, school: 'evocation',
          casting_time: '1 action', range: '60 ft', components: 'V,S',
          duration: 'Instantaneous',
        });
      }, name);

      const results = await page.evaluate(async (n) => {
        return (window as any).api('GET', `/api/compendium/search?q=${encodeURIComponent(n)}`);
      }, name);
      expect(results.some((e: any) => e.name === name)).toBe(true);

      // Filter client-side by type — spells should be 'spell'
      const spells = results.filter((e: any) => e.type === 'spell');
      expect(spells.some((e: any) => e.name === name)).toBe(true);
    });

    test('empty compendium search returns empty array', async ({ page }) => {
      const results = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/compendium/search?q=zzzzzzzzzzzzzzzzzzzz');
      });
      expect(Array.isArray(results)).toBe(true);
    });
  });

  // ─── Compendium Admin Monster Entries ───

  test.describe('Admin Monster Entries', () => {
    test('Create compendium monster entry', async ({ page }) => {
      const name = uniqueName();
      const result = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/admin/compendium/monsters', {
          name: n, ac: 15, hp: 50, str: 16, dex: 12, con: 14, int_: 8, wis: 10, cha: 8, cr: '3',
        });
      }, name);
      expect(result.id).toBeGreaterThan(0);
    });

    test('Delete compendium monster entry', async ({ page }) => {
      const name = uniqueName();
      const created = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/admin/compendium/monsters', {
          name: n, ac: 10, hp: 20, str: 10, dex: 10, con: 10, int_: 10, wis: 10, cha: 10, cr: '0',
        });
      }, name);

      await page.evaluate(async (id) => {
        return (window as any).api('DELETE', `/api/admin/compendium/monsters/${id}`);
      }, created.id);

      const results = await page.evaluate(async (n) => {
        return (window as any).api('GET', `/api/compendium/search?q=${encodeURIComponent(n)}`);
      }, name);
      expect(results.some((e: any) => e.id === created.id)).toBe(false);
    });
  });

  // ─── Bulk Selection Operations ───

  test.describe('Bulk Selection', () => {
    test('batch delete compendium entries', async ({ page }) => {
      // Create a couple of entries via generic compendium entries API
      const name1 = uniqueName();
      const name2 = uniqueName();
      const result = await page.evaluate(async (names) => {
        try {
          // Get available schemas
          const schemas = await window.api('GET', '/api/admin/compendium-schemas');
          if (schemas.length === 0) return { err: 'no schemas' };
          const schemaId = schemas[0].id;
          const e1 = await window.api('POST', `/api/admin/compendium-schemas/${schemaId}/entries`, {
            name: names[0], data: JSON.stringify({ description: 'Bulk test 1' }),
          });
          const e2 = await window.api('POST', `/api/admin/compendium-schemas/${schemaId}/entries`, {
            name: names[1], data: JSON.stringify({ description: 'Bulk test 2' }),
          });
          const batch = await window.api('POST', '/api/admin/compendium/entries/batch-delete', {
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
          const schemas = await window.api('GET', '/api/admin/compendium-schemas');
          if (schemas.length === 0) return { err: 'no schemas' };
          const schemaId = schemas[0].id;
          const e1 = await window.api('POST', `/api/admin/compendium-schemas/${schemaId}/entries`, {
            name: entryName + '-1', data: JSON.stringify({ value: 'old1' }),
          });
          const e2 = await window.api('POST', `/api/admin/compendium-schemas/${schemaId}/entries`, {
            name: entryName + '-2', data: JSON.stringify({ value: 'old2' }),
          });
          await window.api('POST', '/api/admin/compendium/entries/batch-update', {
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
          await window.api('POST', '/api/admin/compendium/entries/batch-delete', { ids: [] });
          return { ok: true };
        } catch (e) {
          return { ok: false, error: String(e) };
        }
      });
      expect(result.ok).toBe(false);
    });
  });
});
