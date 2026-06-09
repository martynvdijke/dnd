import { test, expect } from '@playwright/test';

function makeTestPNG(): Buffer {
  // Valid PNG with unique content to avoid dedup hash collision
  const zlib = require('zlib');
  const seed = `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;
  const r = seed.split('').reduce((acc, c) => ((acc << 5) - acc) + c.charCodeAt(0), 0) & 0xFF;
  // 1x1 RGB pixel (filter byte 0 + R, G, B)
  const raw = Buffer.from([0x00, r, 0x00, 0x00]);
  const compressed = zlib.deflateSync(raw);
  // PNG chunk helpers
  const U32BE = (n: number) => { const b = Buffer.alloc(4); b.writeUInt32BE(n); return b; };
  const crc32 = (data: Buffer): number => {
    let c = 0xFFFFFFFF;
    const table = new Uint32Array(256);
    for (let n = 0; n < 256; n++) {
      let cv = n;
      for (let k = 0; k < 8; k++) cv = (cv & 1) ? (0xEDB88320 ^ (cv >>> 1)) : (cv >>> 1);
      table[n] = cv;
    }
    for (let i = 0; i < data.length; i++) c = table[(c ^ data[i]) & 0xFF] ^ (c >>> 8);
    return (c ^ 0xFFFFFFFF) >>> 0;
  };
  const chunk = (type: string, data: Buffer) =>
    Buffer.concat([U32BE(data.length), Buffer.from(type), data, U32BE(crc32(Buffer.concat([Buffer.from(type), data])))]);

  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(1, 0); ihdr.writeUInt32BE(1, 4);
  ihdr[8] = 8; ihdr[9] = 2; // 8-bit RGB

  return Buffer.concat([
    Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]), // PNG signature
    chunk('IHDR', ihdr),
    chunk('IDAT', compressed),
    chunk('IEND', Buffer.alloc(0)),
  ]);
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
    const filename = `test-${Date.now()}-${Math.random().toString(36).slice(2, 7)}.png`;

    const resp = await page.request.post('/api/upload', {
      multipart: {
        image: {
          name: filename,
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
    const filename2 = `test-${Date.now()}-${Math.random().toString(36).slice(2, 7)}.png`;

    // Upload an image
    const uploadResp = await page.request.post('/api/upload', {
      multipart: {
        image: {
          name: filename2,
          mimeType: 'image/png',
          buffer: pngBytes,
        },
      },
      headers: {
        'X-CSRF-Token': csrf,
      },
    });
    expect(uploadResp.status()).toBe(200);
    const uploadData = await uploadResp.json();
    expect(uploadData).toHaveProperty('id');

    // Create upload link
    const linkResp = await page.request.post('/api/upload-links', {
      data: {
        upload_id: uploadData.id,
        entity_type: 'campaign',
        entity_id: 42,
        field_name: 'map',
      },
      headers: {
        'X-CSRF-Token': csrf,
      },
    });
    expect(linkResp.status()).toBe(201);
    const linkData = await linkResp.json();
    expect(linkData).toHaveProperty('id');
    expect(linkData.entity_type).toBe('campaign');

    // Delete upload link
    const deleteResp = await page.request.delete(`/api/upload-links/${linkData.id}`, {
      headers: {
        'X-CSRF-Token': csrf,
      },
    });
    expect(deleteResp.status()).toBe(204);
  });

  test('media gallery HTMX loads and renders upload button', async ({ page }) => {
    await page.goto('/htmx/media-gallery?owner_type=campaign&owner_id=99999', { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(500);

    const uploadBtn = page.locator('button:has-text("Upload")');
    await expect(uploadBtn).toBeVisible({ timeout: 5000 });
  });
});
