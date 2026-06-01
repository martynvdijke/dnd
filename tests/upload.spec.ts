import { test, expect } from '@playwright/test';

function makeTestPNG(): Buffer {
  // Minimal valid 1x1 red PNG
  return Buffer.from(
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
    'base64'
  );
}

async function getCSRFToken(page: { request: { get: (url: string) => Promise<{ json: () => Promise<Record<string, string>> }> } }): Promise<string> {
  const resp = await page.request.get('/api/csrf-token');
  const data = await resp.json();
  return data.token;
}

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
    await page.goto('/htmx/media-gallery?owner_type=campaign&owner_id=99999', { waitUntil: 'domcontentloaded' });

    const emptyState = page.locator('text=No Media Yet');
    await expect(emptyState).toBeVisible({ timeout: 5000 });
  });

  test('upload API works and returns valid response', async ({ page }) => {
    const pngBytes = makeTestPNG();
    const csrf = await getCSRFToken(page);

    const resp = await page.request.post('/api/upload', {
      multipart: {
        image: {
          name: 'test.png',
          mimeType: 'image/png',
          buffer: pngBytes,
        },
      },
      headers: {
        'X-CSRF-Token': csrf,
      },
    });
    expect(resp.status()).toBe(200);
    const data = await resp.json();
    expect(data).toHaveProperty('url');
    expect(data).toHaveProperty('id');
    expect(data).not.toHaveProperty('duplicate');
  });

  test('upload-links API works', async ({ page }) => {
    const pngBytes = makeTestPNG();
    const csrf = await getCSRFToken(page);

    // Upload an image via page context
    const uploadResult = await page.evaluate(async ({csrf, pngB64}) => {
      const blob = Uint8Array.from(atob(pngB64), c => c.charCodeAt(0));
      const formData = new FormData();
      formData.append('image', new Blob([blob], {type: 'image/png'}), 'test.png');
      const resp = await fetch('/api/upload', {
        method: 'POST',
        headers: {'X-CSRF-Token': csrf},
        body: formData,
      });
      return {status: resp.status, body: await resp.json()};
    }, {csrf, pngB64: pngBytes.toString('base64')});
    expect(uploadResult.status).toBe(200);
    const uploadData = uploadResult.body;

    // Create upload link via page context
    const linkResult = await page.evaluate(async ({csrf, uploadId}) => {
      const resp = await fetch('/api/upload-links', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': csrf,
        },
        body: JSON.stringify({
          upload_id: uploadId,
          entity_type: 'campaign',
          entity_id: 42,
          field_name: 'map',
        }),
      });
      return {status: resp.status, body: await resp.json()};
    }, {csrf, uploadId: uploadData.id});
    expect(linkResult.status).toBe(201);
    const linkData = linkResult.body;
    expect(linkData).toHaveProperty('id');
    expect(linkData.entity_type).toBe('campaign');

    // Delete upload link via page context
    const deleteResult = await page.evaluate(async ({csrf, linkId}) => {
      const resp = await fetch(`/api/upload-links/${linkId}`, {
        method: 'DELETE',
        headers: {'X-CSRF-Token': csrf},
      });
      return resp.status;
    }, {csrf, linkId: linkData.id});
    expect(deleteResult).toBe(204);
  });

  test('media gallery HTMX loads and renders upload button', async ({ page }) => {
    await page.goto('/htmx/media-gallery?owner_type=campaign&owner_id=99999', { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(500);

    const uploadBtn = page.locator('button:has-text("Upload")');
    await expect(uploadBtn).toBeVisible({ timeout: 5000 });
  });
});
