import { test, expect } from './fixtures.js';
import { login } from './helpers.js';

test.describe('PWA Support', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('manifest.json is linked and loadable', async ({ page }) => {
    const manifestLink = page.locator('link[rel="manifest"]');
    await expect(manifestLink).toHaveAttribute('href', '/static/manifest.json');

    const response = await page.request.get('/static/manifest.json');
    expect(response.status()).toBe(200);
    const manifest = await response.json();
    expect(manifest.name).toBeTruthy();
    expect(manifest.start_url).toBeTruthy();
    expect(manifest.display).toBe('standalone');
  });

  test('service worker js file is served correctly', async ({ page }) => {
    const response = await page.request.get('/static/sw.js');
    expect(response.status()).toBe(200);
    const body = await response.text();
    expect(body).toContain('self.addEventListener');
    expect(body).toContain('install');
    expect(body).toContain('fetch');
    expect(body).toContain('activate');
  });

  test('manifest JSON has required PWA fields', async ({ page }) => {
    const response = await page.request.get('/static/manifest.json');
    const manifest = await response.json();

    expect(manifest).toHaveProperty('name');
    expect(manifest).toHaveProperty('short_name');
    expect(manifest).toHaveProperty('start_url');
    expect(manifest).toHaveProperty('display', 'standalone');
    expect(manifest).toHaveProperty('background_color');
    expect(manifest).toHaveProperty('theme_color');
    expect(manifest).toHaveProperty('icons');
    expect(Array.isArray(manifest.icons)).toBe(true);
    expect(manifest.icons.length).toBeGreaterThanOrEqual(2);

    const sizes = manifest.icons.map((i: any) => i.sizes);
    expect(sizes).toContain('192x192');
    expect(sizes).toContain('512x512');
  });

  test('pwa.js script is loaded on the page', async ({ page }) => {
    const scripts = page.locator('script[src="/static/js/pwa.js"]');
    await expect(scripts).toHaveCount(1);
  });

  test('sw.js supports cache-first strategy', async ({ page }) => {
    const response = await page.request.get('/static/sw.js');
    const body = await response.text();
    expect(body).toContain('CACHE');
    expect(body).toContain('install');
    expect(body).toContain('activate');
  });
});
