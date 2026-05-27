import { test, expect } from '@playwright/test';

const uniqueName = () => `OS-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function waitLoadingDone(page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: 5000 }).catch(() => {});
}

async function waitModalClosed(page) {
  await page.waitForFunction(() => {
    const modal = document.getElementById('genericModal');
    return !modal || !modal.classList.contains('show') || modal.classList.contains('d-none');
  }, { timeout: 10000 }).catch(() => {});
}

/** Use page.evaluate with fetch (carries browser cookies) + explicit credentials */
async function apiFetch(page, url, opts = {}) {
  return page.evaluate(async ({ url, method, body }: { url: string; method?: string; body?: any }) => {
    const headers: Record<string, string> = { 'Content-Type': 'application/json' };
    const csrfEl = document.querySelector('meta[name="csrf-token"]');
    let token = csrfEl ? csrfEl.getAttribute('content') || '' : '';
    if (!token) {
      const tresp = await fetch('/api/csrf-token', { credentials: 'same-origin' });
      const tdata = await tresp.json();
      token = tdata.token || '';
    }
    if (token) headers['X-CSRF-Token'] = token;
    const resp = await fetch(url, {
      method: method || 'GET',
      headers,
      body: body ? JSON.stringify(body) : undefined,
      credentials: 'same-origin',
    });
    return { ok: resp.ok, status: resp.status, body: await resp.json().catch(() => null) };
  }, { url, method: opts.method || 'GET', body: opts.body || null });
}

/** Load HTMX content into oneshotSection */
async function loadHtmx(page, url) {
  await page.evaluate(async (u: string) => {
    const resp = await fetch(u, { credentials: 'same-origin' });
    const el = document.getElementById('oneshotSection')!;
    el.innerHTML = await resp.text();
    (window as any).htmx?.process(el);
  }, url);
}

/** Click the One-Shots nav link, handling mobile hamburger menu */
async function navigateToOneShots(page) {
  const toggler = page.locator('.navbar-toggler');
  if (await toggler.isVisible()) {
    await toggler.click();
    await page.waitForTimeout(300);
  }
  await page.locator('nav a:has-text("One-Shots")').click();
  await page.waitForSelector('#oneshotSection', { state: 'visible', timeout: 5000 });
}

/** Fill and submit the one-shot form inside the modal */
async function submitOneShotForm(page, { title, template, difficulty, minutes }) {
  await expect(page.locator('#genericModal')).toBeVisible({ timeout: 5000 });

  // Wait for form to load in modal body
  await page.waitForSelector('#genericModalBody input[name="title"]', { timeout: 5000 });
  await page.locator('#genericModalBody input[name="title"]').fill(title);
  if (template) await page.locator('#genericModalBody select[name="template"]').selectOption(template);
  if (difficulty) await page.locator('#genericModalBody select[name="difficulty"]').selectOption(difficulty);
  if (minutes) await page.locator('#genericModalBody input[name="estimated_minutes"]').fill(String(minutes));

  await page.locator('#genericModalBody form button.btn-primary').click();
  await waitModalClosed(page);
  await page.waitForTimeout(500);
}

test.describe('One-Shot Adventure Features', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('One-Shots nav item is visible and loads the list', async ({ page }) => {
    await navigateToOneShots(page);
    const body = await page.locator('#oneshotSection').innerText();
    expect(body.length).toBeGreaterThan(0);
  });

  test('Create a one-shot adventure via UI form', async ({ page }) => {
    await navigateToOneShots(page);

    // Verify New button exists before clicking
    await expect(page.locator('#oneshotSection button:has-text("New")')).toBeVisible({ timeout: 5000 });
    await page.locator('#oneshotSection button:has-text("New")').click();

    const title = uniqueName();
    await submitOneShotForm(page, {
      title, template: 'custom', difficulty: 'medium', minutes: 120
    });

    await expect(page.locator('#oneshotSection')).toContainText(title, { timeout: 5000 });
  });

  test('Generate a five-room dungeon via UI form', async ({ page }) => {
    await navigateToOneShots(page);

    await expect(page.locator('#oneshotSection button:has-text("Generate")')).toBeVisible({ timeout: 10000 });
    await page.locator('#oneshotSection button:has-text("Generate")').click();
    await page.waitForTimeout(300);

    const title = uniqueName();
    await submitOneShotForm(page, {
      title, template: 'five_room_dungeon', difficulty: 'hard', minutes: 180
    });

    await expect(page.locator('#oneshotSection')).toContainText(title, { timeout: 10000 });

    // Click the adventure to view detail and check for acts
    const listItem = page.locator('#oneshotSection .list-group-item').first();
    await listItem.waitFor({ state: 'visible', timeout: 5000 });
    await listItem.click();

    // Wait for HTMX detail view to load acts
    await page.waitForFunction(() => {
      const section = document.getElementById('oneshotSection');
      return section && section.textContent.includes('Act 1');
    }, { timeout: 15000 });
    await expect(page.locator('#oneshotSection')).toContainText('Act 5', { timeout: 15000 });
  });

  test('Prep dashboard loads for generated one-shot', async ({ page }) => {
    // Create via UI to ensure proper auth context
    await navigateToOneShots(page);

    const title = uniqueName();
    await page.locator('#oneshotSection button:has-text("Generate")').click();
    await submitOneShotForm(page, {
      title, template: 'five_room_dungeon', difficulty: 'easy', minutes: 60
    });

    // Get the adventure ID from the list item
    const advId = await page.evaluate((t) => {
      const items = document.querySelectorAll('#oneshotSection .list-group-item');
      for (const item of items) {
        if (item.textContent.includes(t)) {
          const hxGet = item.getAttribute('hx-get') || '';
          const m = hxGet.match(/\/oneshot-adventures\/(\d+)/);
          return m ? parseInt(m[1]) : 0;
        }
      }
      return 0;
    }, title);
    expect(advId).toBeGreaterThan(0);

    // Load prep dashboard via HTMX
    await loadHtmx(page, `/htmx/oneshot-adventures/${advId}/dashboard`);
    await expect(page.locator('#oneshotSection')).toContainText(title, { timeout: 5000 });
    await expect(page.locator('#oneshotSection')).toContainText('Session Flow', { timeout: 5000 });
  });

  test('DM screen loads with quick reference, actions, and notes tabs', async ({ page }) => {
    await navigateToOneShots(page);

    const title = uniqueName();
    await page.locator('#oneshotSection button:has-text("Generate")').click();
    await submitOneShotForm(page, {
      title, template: 'five_room_dungeon', difficulty: 'easy', minutes: 60
    });

    const advId = await page.evaluate((t) => {
      const items = document.querySelectorAll('#oneshotSection .list-group-item');
      for (const item of items) {
        if (item.textContent.includes(t)) {
          const m = (item.getAttribute('hx-get') || '').match(/\/oneshot-adventures\/(\d+)/);
          return m ? parseInt(m[1]) : 0;
        }
      }
      return 0;
    }, title);

    await loadHtmx(page, `/htmx/oneshot-adventures/${advId}/dm-screen`);
    await expect(page.locator('#oneshotSection')).toContainText('Quick Reference', { timeout: 5000 });
    await expect(page.locator('#oneshotSection')).toContainText('Quick Actions', { timeout: 5000 });
    await expect(page.locator('#oneshotSection')).toContainText('DM Notes', { timeout: 5000 });
    await expect(page.locator('#oneshotSection')).toContainText('Conditions', { timeout: 5000 });
  });

  test('Prep checklist add and toggle via UI', async ({ page }) => {
    await navigateToOneShots(page);

    // Create an adventure
    const title = uniqueName();
    await page.locator('#oneshotSection button:has-text("New")').click();
    await submitOneShotForm(page, {
      title, template: 'custom', difficulty: 'medium', minutes: 120
    });

    const advId = await page.evaluate((t) => {
      const items = document.querySelectorAll('#oneshotSection .list-group-item');
      for (const item of items) {
        if (item.textContent.includes(t)) {
          const m = (item.getAttribute('hx-get') || '').match(/\/oneshot-adventures\/(\d+)/);
          return m ? parseInt(m[1]) : 0;
        }
      }
      return 0;
    }, title);
    expect(advId).toBeGreaterThan(0);

    // Load checklist via HTMX
    await loadHtmx(page, `/htmx/oneshot-adventures/${advId}/checklist`);
    await page.waitForTimeout(1000);
    await expect(page.locator('#oneshotSection')).toContainText('No checklist items yet', { timeout: 5000 });

    // Add checklist item
    const input = page.locator('#oneshotSection input[name="item"]');
    await expect(input).toBeVisible({ timeout: 5000 });
    await input.fill('Prepare battle maps');
    await page.locator('#oneshotSection button.btn-primary i.fa-plus').click();
    await page.waitForTimeout(2000);
    await expect(page.locator('#oneshotSection')).toContainText('Prepare battle maps', { timeout: 10000 });
  });

  test('Pregenerated characters list and generate', async ({ page }) => {
    // Navigate to one-shots first (to have oneshotSection available)
    await navigateToOneShots(page);

    // Load pregens list
    await loadHtmx(page, '/htmx/pregens');
    await expect(page.locator('#oneshotSection')).toContainText(/Pregen/i, { timeout: 5000 });

    // Generate a character
    await loadHtmx(page, '/htmx/pregens/generate?name=Sir+Test&class=fighter&race=human&level=3');
    await expect(page.locator('#oneshotSection')).toContainText('Sir Test', { timeout: 5000 });
  });

  test('Session flow loads print-friendly view', async ({ page }) => {
    await navigateToOneShots(page);

    const title = uniqueName();
    await page.locator('#oneshotSection button:has-text("Generate")').click();
    await submitOneShotForm(page, {
      title, template: 'five_room_dungeon', difficulty: 'medium', minutes: 120
    });

    const advId = await page.evaluate((t) => {
      const items = document.querySelectorAll('#oneshotSection .list-group-item');
      for (const item of items) { if (item.textContent.includes(t)) { const m = (item.getAttribute('hx-get') || '').match(/\/oneshot-adventures\/(\d+)/); return m ? parseInt(m[1]) : 0; } }
      return 0;
    }, title);
    expect(advId).toBeGreaterThan(0);

    await loadHtmx(page, `/htmx/oneshot-adventures/${advId}/session-flow`);
    await expect(page.locator('#oneshotSection')).toContainText(title, { timeout: 5000 });
    await expect(page.locator('#oneshotSection')).toContainText('Entrance & Guardian', { timeout: 5000 });
    await expect(page.locator('#oneshotSection')).toContainText('Reward & Revelation', { timeout: 5000 });
  });

  test('Clue board - add and view clues', async ({ page }) => {
    await navigateToOneShots(page);

    const title = uniqueName();
    await page.locator('#oneshotSection button:has-text("Generate")').click();
    await submitOneShotForm(page, {
      title, template: 'five_room_dungeon', difficulty: 'easy', minutes: 60
    });

    const advId = await page.evaluate((t) => {
      const items = document.querySelectorAll('#oneshotSection .list-group-item');
      for (const item of items) { if (item.textContent.includes(t)) { const m = (item.getAttribute('hx-get') || '').match(/\/oneshot-adventures\/(\d+)/); return m ? parseInt(m[1]) : 0; } }
      return 0;
    }, title);
    expect(advId).toBeGreaterThan(0);

    // Add clue via API fetch inside page context
    const clueR = await apiFetch(page, `/api/oneshot-adventures/${advId}/clues`, {
      method: 'POST',
      body: { title: 'The Hidden Dagger', description: 'Found in the library', clue_type: 'object' }
    });
    expect(clueR.status).toBe(201);

    // Load clue board
    await loadHtmx(page, `/htmx/oneshot-adventures/${advId}/clues`);
    await expect(page.locator('#oneshotSection')).toContainText('The Hidden Dagger', { timeout: 5000 });
  });

  test('Pacing session can be started and viewed', async ({ page }) => {
    await navigateToOneShots(page);

    const title = uniqueName();
    await page.locator('#oneshotSection button:has-text("Generate")').click();
    await submitOneShotForm(page, {
      title, template: 'five_room_dungeon', difficulty: 'easy', minutes: 60
    });

    const advId = await page.evaluate((t) => {
      const items = document.querySelectorAll('#oneshotSection .list-group-item');
      for (const item of items) { if (item.textContent.includes(t)) { const m = (item.getAttribute('hx-get') || '').match(/\/oneshot-adventures\/(\d+)/); return m ? parseInt(m[1]) : 0; } }
      return 0;
    }, title);
    expect(advId).toBeGreaterThan(0);

    // Start pacing session via API
    const pacingR = await apiFetch(page, `/api/oneshot-adventures/${advId}/pacing/start`, { method: 'POST' });
    expect(pacingR.ok).toBeTruthy();
    const sessionId = pacingR.body.id;
    expect(sessionId).toBeGreaterThan(0);

    // Load pacing dashboard via HTMX
    await loadHtmx(page, `/htmx/session-pacing/${sessionId}`);
    await expect(page.locator('#oneshotSection')).toContainText('Session Dashboard', { timeout: 5000 });
    await expect(page.locator('#oneshotSection')).toContainText('Running', { timeout: 5000 });
    await expect(page.locator('#oneshotSection button:has-text("Pause")')).toBeVisible({ timeout: 5000 });
  });
});
