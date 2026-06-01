import { test, expect } from '@playwright/test';

test.describe('File upload and media gallery', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);
  });

  test('empty media gallery shows empty state', async ({ page }) => {
    // Direct HTMX load for a non-existent entity — shows empty state
    await page.goto('/htmx/media-gallery?owner_type=campaign&owner_id=99999', { waitUntil: 'domcontentloaded' });

    const emptyState = page.locator('text=No Media Yet');
    await expect(emptyState).toBeVisible({ timeout: 5000 });
  });

  test('upload API works and returns valid response', async ({ page }) => {
    // Test the upload API directly via page.request
    const pngBytes = Buffer.from(
      'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
      'base64'
    );
    const boundary = '----TestBoundary' + Date.now();
    const body = [
      `--${boundary}`,
      'Content-Disposition: form-data; name="image"; filename="test.png"',
      'Content-Type: image/png',
      '',
      pngBytes.toString('binary'),
      `--${boundary}--`,
    ].join('\r\n');

    const resp = await page.request.post('/api/upload', {
      headers: { 'Content-Type': `multipart/form-data; boundary=${boundary}` },
      data: Buffer.from(body, 'binary'),
    });
    expect(resp.status()).toBe(200);
    const data = await resp.json();
    expect(data).toHaveProperty('url');
    expect(data).toHaveProperty('id');
    expect(data).not.toHaveProperty('duplicate');
  });

  test('upload-links API works', async ({ page }) => {
    // Upload an image
    const pngBytes = Buffer.from(
      'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
      'base64'
    );
    const boundary = '----TestBoundary' + Date.now();
    const body = [
      `--${boundary}`,
      'Content-Disposition: form-data; name="image"; filename="test.png"',
      'Content-Type: image/png',
      '',
      pngBytes.toString('binary'),
      `--${boundary}--`,
    ].join('\r\n');

    const uploadResp = await page.request.post('/api/upload', {
      headers: { 'Content-Type': `multipart/form-data; boundary=${boundary}` },
      data: Buffer.from(body, 'binary'),
    });
    expect(uploadResp.status()).toBe(200);
    const uploadData = await uploadResp.json();

    // Create upload link
    const linkResp = await page.request.post('/api/upload-links', {
      data: {
        upload_id: uploadData.id,
        entity_type: 'campaign',
        entity_id: 42,
        field_name: 'map',
      },
    });
    expect(linkResp.status()).toBe(201);
    const linkData = await linkResp.json();
    expect(linkData).toHaveProperty('id');
    expect(linkData.entity_type).toBe('campaign');

    // Delete upload link
    const deleteResp = await page.request.delete(`/api/upload-links/${linkData.id}`);
    expect(deleteResp.status()).toBe(204);
  });

  test('media gallery HTMX loads and renders upload button', async ({ page }) => {
    await page.goto('/htmx/media-gallery?owner_type=campaign&owner_id=99999', { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(500);

    // Gallery should contain Upload button
    const uploadBtn = page.locator('button:has-text("Upload")');
    await expect(uploadBtn).toBeVisible({ timeout: 5000 });
  });
});
