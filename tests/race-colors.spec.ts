import { test, expect } from './fixtures.js';
import { login, NAV_TIMEOUT, getApiToken } from './helpers.js';

async function getCSRFToken(page: any): Promise<string> {
  const resp = await page.request.get('/api/csrf-token');
  const data = await resp.json();
  return data.token;
}

test.describe('Race Colors', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('GET /api/race-colors returns array', async ({ page }) => {
    const resp = await page.request.get('/api/race-colors', {
      headers: { 'X-CSRF-Token': await getCSRFToken(page) },
    });
    expect(resp.status()).toBe(200);
    const data = await resp.json();
    expect(Array.isArray(data)).toBe(true);
    // Should have seeded defaults
    expect(data.length).toBeGreaterThanOrEqual(1);
    const elf = data.find((r: any) => r.race_name === 'Elf');
    expect(elf).toBeDefined();
    expect(elf.color).toBeTruthy();
  });

  test('PUT /api/race-colors updates colors', async ({ page }) => {
    const csrf = await getCSRFToken(page);
    const apiToken = await getApiToken(page);
    const newColors = [
      { race_name: 'Elf', color: '#00FF00' },
      { race_name: 'Dwarf', color: '#0000FF' },
    ];
    const putResp = await page.request.put('/api/race-colors', {
      data: newColors,
      headers: { 'X-CSRF-Token': csrf, Authorization: `Bearer ${apiToken}` },
    });
    expect(putResp.status()).toBe(200);
    const data = await putResp.json();
    expect(data).toEqual({ ok: true });

    // Verify
    const getResp = await page.request.get('/api/race-colors', {
      headers: { 'X-CSRF-Token': csrf },
    });
    const colors = await getResp.json();
    const elf = colors.find((r: any) => r.race_name === 'Elf');
    expect(elf.color).toBe('#00FF00');
    const dwarf = colors.find((r: any) => r.race_name === 'Dwarf');
    expect(dwarf.color).toBe('#0000FF');
  });

  test('race colors appear in character list', async ({ page }) => {
    // Create a character with a known race
    const csrf = await getCSRFToken(page);
    const apiToken = await getApiToken(page);
    await page.request.put('/api/race-colors', {
      data: [{ race_name: 'TestElf', color: '#AA00AA' }],
      headers: { 'X-CSRF-Token': csrf, Authorization: `Bearer ${apiToken}` },
    });

    await page.request.post('/api/characters', {
      data: { name: 'RaceColorChar', race: 'TestElf', class: 'Wizard', level: 1, str: 10, dex: 10, con: 10, int: 14, wis: 10, cha: 10, hp_max: 10, hp_current: 10 },
      headers: { 'X-CSRF-Token': csrf, Authorization: `Bearer ${apiToken}` },
    });

    // Navigate to character list
    await page.goto('/', { waitUntil: 'domcontentloaded' });
    // Wait for app to load
    await page.waitForFunction(() => {
      const o = document.getElementById('loadingOverlay');
      return o && o.classList.contains('d-none');
    }, { timeout: 10000 });

    // Click Characters nav
    await page.locator('#appSidebar button[data-nav="characters"]').click();
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

    // The character card should show a colored badge
    const badge = page.locator('.character-card').filter({ hasText: 'RaceColorChar' }).locator('.badge');
    await expect(badge).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(badge).toHaveText('TestElf');
    // Check the style has the color
    const style = await badge.getAttribute('style');
    expect(style).toContain('#AA00AA');
  });
});

test.describe('File Picker and Portrait Upload', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('upload API works for portrait images', async ({ page }) => {
    const csrf = await getCSRFToken(page);
    const apiToken = await getApiToken(page);
    // Create a minimal valid PNG
    const pngHeader = Buffer.from([
      0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
    ]);
    const resp = await page.request.post('/api/upload', {
      multipart: {
        image: {
          name: `portrait-test-${Date.now()}.png`,
          mimeType: 'image/png',
          buffer: pngHeader,
        },
      },
      headers: { 'X-CSRF-Token': csrf, Authorization: `Bearer ${apiToken}` },
    });
    expect(resp.status()).toBe(200);
    const data = await resp.json();
    expect(data).toHaveProperty('url');
    expect(data).toHaveProperty('id');
  });
});
