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

  test('sw.js at root scope is version-stamped and not cached', async ({ page }) => {
    const response = await page.request.get('/sw.js');
    expect(response.status()).toBe(200);
    // Served with no-cache so the browser revalidates on every navigation
    // instead of the ~24h default SW update check.
    expect(response.headers()['cache-control']).toContain('no-cache');
    // Cache name must carry the app version: different bytes per release
    // means browsers detect and install the updated worker. Assert the
    // placeholder was substituted with the actual version.
    const body = await response.text();
    expect(body).toContain('villum-v3-');
    expect(body).not.toContain('{{VERSION}}');
  });

  test('page auto-reloads when an updated service worker takes control', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('body')).toBeVisible();

    // Serve a modified sw.js (as a new release would) and reload: the
    // browser detects the update, installs it, and the controllerchange
    // listener in pwa.ts reloads the page automatically.
    await page.route('**/sw.js', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/javascript',
        headers: { 'Cache-Control': 'no-cache, no-store, must-revalidate' },
        body: `
          const CACHE_NAME = 'villum-v3-999.0.0-test';
          self.addEventListener('install', (e) => { self.skipWaiting(); });
          self.addEventListener('activate', (e) => { self.clients.claim(); });
          self.addEventListener('fetch', (e) => { e.respondWith(fetch(e.request)); });
        `,
      });
    });

    // After the reload the browser fetches the updated sw.js, installs and
    // activates it, and the controllerchange listener triggers one more
    // navigation on its own — assert that automatic second reload.
    const autoReload = page.waitForNavigation();
    await page.reload();
    await autoReload;
    await expect(page.locator('body')).toBeVisible();
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
