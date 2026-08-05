import { test, expect } from './fixtures.js';
import { clickNavItem, login, NAV_TIMEOUT } from './helpers.js';

const uniqueName = () => `Party-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('Party Management', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.waitForTimeout(300);
  });

  test('Create a party via API', async ({ page }) => {
    const name = uniqueName();
    const result = await page.evaluate(async (n) => {
      return (window as any).api('POST', '/api/parties', {
        name: n,
        description: 'A new adventuring party',
      });
    }, name);
    expect(result.id).toBeGreaterThan(0);
  });

  test('Get party by ID returns correct data', async ({ page }) => {
    const name = uniqueName();
    const created = await page.evaluate(async (n) => {
      return (window as any).api('POST', '/api/parties', {
        name: n, description: 'Test description',
      });
    }, name);
    expect(created.id).toBeGreaterThan(0);

    const party = await page.evaluate(async (id) => {
      return (window as any).api('GET', `/api/parties/${id}`);
    }, created.id);
    expect(party.name).toBe(name);
    expect(party.description).toBe('Test description');
    expect(party.id).toBe(created.id);
    expect(party.user_id).toBeGreaterThan(0);
  });

  test('Get non-existent party returns 404', async ({ page }) => {
    const result = await page.evaluate(async () => {
      try {
        await (window as any).api('GET', '/api/parties/99999999');
        return { found: true };
      } catch (e: any) {
        return { found: false, error: e.message };
      }
    });
    expect(result.found).toBe(false);
  });

  test('List parties returns array', async ({ page }) => {
    const name = uniqueName();
    await page.evaluate(async (n) => {
      return (window as any).api('POST', '/api/parties', {
        name: n, description: 'List test',
      });
    }, name);

    const parties = await page.evaluate(async () => {
      return (window as any).api('GET', '/api/parties');
    });
    expect(Array.isArray(parties)).toBe(true);
    expect(parties.length).toBeGreaterThan(0);
    const found = parties.some((p: any) => p.name === name);
    expect(found).toBe(true);
  });

  test('Update party name and description', async ({ page }) => {
    const name = uniqueName();
    const created = await page.evaluate(async (n) => {
      return (window as any).api('POST', '/api/parties', {
        name: n, description: 'Original',
      });
    }, name);

    const newName = name + '-renamed';
    await page.evaluate(async ({ id, newName }) => {
      return (window as any).api('PUT', `/api/parties/${id}`, {
        name: newName, description: 'Updated description',
      });
    }, { id: created.id, newName });

    const updated = await page.evaluate(async (id) => {
      return (window as any).api('GET', `/api/parties/${id}`);
    }, created.id);
    expect(updated.name).toBe(newName);
    expect(updated.description).toBe('Updated description');
  });

  test('Delete party removes it from listing', async ({ page }) => {
    const name = uniqueName();
    const created = await page.evaluate(async (n) => {
      return (window as any).api('POST', '/api/parties', {
        name: n, description: 'Delete me',
      });
    }, name);

    await page.evaluate(async (id) => {
      return (window as any).api('DELETE', `/api/parties/${id}`);
    }, created.id);

    const parties = await page.evaluate(async () => {
      return (window as any).api('GET', '/api/parties');
    });
    const found = parties.some((p: any) => p.id === created.id);
    expect(found).toBe(false);
  });

  test('Create and list party factions', async ({ page }) => {
    const name = uniqueName();
    const created = await page.evaluate(async (n) => {
      return (window as any).api('POST', '/api/parties', {
        name: n, description: 'Faction test',
      });
    }, name);

    const faction = await page.evaluate(async (partyId) => {
      return (window as any).api('POST', `/api/parties/${partyId}/factions`, {
        name: 'Arcane Brotherhood',
        description: 'Wizards and sorcerers',
        type: 'guild',
        headquarters: 'The Spire',
      });
    }, created.id);
    expect(faction.id).toBeGreaterThan(0);

    const factions = await page.evaluate(async (partyId) => {
      return (window as any).api('GET', `/api/parties/${partyId}/factions`);
    }, created.id);
    expect(factions.length).toBeGreaterThan(0);
    expect(factions.some((f: any) => f.name === 'Arcane Brotherhood')).toBe(true);
  });

  test('List party uploads returns empty array initially', async ({ page }) => {
    const name = uniqueName();
    const created = await page.evaluate(async (n) => {
      return (window as any).api('POST', '/api/parties', {
        name: n, description: 'Upload test',
      });
    }, name);

    const uploads = await page.evaluate(async (partyId) => {
      return (window as any).api('GET', `/api/parties/${partyId}/uploads`);
    }, created.id);
    expect(Array.isArray(uploads)).toBe(true);
    expect(uploads.length).toBe(0);
  });

  test('Party view UI loads and shows party data', async ({ page }) => {
    const name = uniqueName();
    await page.evaluate(async (n) => {
      await (window as any).api('POST', '/api/parties', {
        name: n, description: 'UI test party',
      });
    }, name);

    // Navigate to Party View
    await clickNavItem(page, 'party', 'party');
    await page.waitForTimeout(1500);

    // Check the party content area contains the party name
    await expect(page.locator('#partyContent')).toContainText(name, { timeout: NAV_TIMEOUT });
    await expect(page.locator('#partyContent')).toContainText('Party View', { timeout: NAV_TIMEOUT });
  });
});
