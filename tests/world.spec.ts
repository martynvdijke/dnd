import { test, expect } from './fixtures.js';
import { login, waitLoadingDone, waitModalClosed, NAV_TIMEOUT } from './helpers.js';

const unique = () => `W-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`;

async function createCampaign(page: any, name: string) {
  return page.evaluate(async (n: string) => (window as any).api('POST', '/api/campaigns', { name: n, description: 'world e2e', dm_notes: '' }), name);
}

async function getCampaignId(page: any, cid?: number): Promise<number> {
  if (cid) return cid;
  const id = await page.evaluate(() => (window as any).currentCampaign?.id ?? (window as any).getCurrentCampaign?.()?.id ?? null);
  if (id) return id;
  const camps = await page.evaluate(async () => (window as any).api('GET', '/api/campaigns'));
  if (camps.length) return camps[0].id;
  throw new Error('no campaign found');
}

async function openWorld(page: any) {
  // Prefer nav button, fall back to direct call
  const nav = page.locator('[data-testid="nav-world"]');
  if (await nav.first().isVisible().catch(() => false)) {
    await nav.first().click();
  } else {
    await page.evaluate(() => (window as any).showWorld());
  }
  // worldContent renders after API calls; wait for it to contain something not "Loading"
  await expect(page.locator('#worldView')).toBeVisible({ timeout: NAV_TIMEOUT });
  await page.waitForFunction(() => {
    const el = document.getElementById('worldContent');
    return el && !el.textContent?.includes('Loading world');
  }, { timeout: NAV_TIMEOUT });
}

async function openTimeline(page: any) {
  const nav = page.locator('[data-testid="nav-timeline"]');
  if (await nav.first().isVisible().catch(() => false)) {
    await nav.first().click();
  } else {
    await page.evaluate(() => (window as any).showTimeline());
  }
  await expect(page.locator('#timelineView')).toBeVisible({ timeout: NAV_TIMEOUT });
  // timelineContent is HTMX loaded
  await page.waitForFunction(() => {
    const el = document.getElementById('timelineContent');
    return el && !el.textContent?.includes('Loading timeline');
  }, { timeout: NAV_TIMEOUT });
}

test.describe('World Overview and place↔timeline linking', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('links a timeline event to a place and shows it in the world view', async ({ page }) => {
    const suffix = unique();
    const placeName = `World E2E ${suffix}`;
    const eventTitle = `Evt ${suffix}`;

    const camp = await createCampaign(page, `WorldCamp ${suffix}`);
    const cid = camp.id as number;
    // select campaign so world/timeline use correct cid
    await page.evaluate((id: number) => (window as any).selectCampaign(id), cid);
    await waitLoadingDone(page);

    const loc = await page.evaluate(async (name: string) => (window as any).api('POST', '/api/locations', { name, type: 'city', description: 'e2e place', latitude: 40, longitude: -3 }), placeName);
    expect(loc.id).toBeGreaterThan(0);

    const evt = await page.evaluate(async (opts: any) => (window as any).api('POST', '/api/timeline', {
      campaign_id: opts.cid, title: opts.title, description: 'd', event_date: '2026-06-01', event_type: 'major', linked_entity_type: 'location', linked_entity_id: opts.locId,
    }), { cid, title: eventTitle, locId: loc.id });
    expect(evt.id).toBeGreaterThan(0);

    await openWorld(page);

    // World directory should list the place
    await expect(page.locator('#worldContent')).toContainText(placeName, { timeout: NAV_TIMEOUT });
    // click the place button to open detail
    const placeBtn = page.locator('#worldContent').getByRole('button', { name: placeName });
    // fallback: button contains name substring
    if (await placeBtn.count() > 0) {
      await placeBtn.first().click();
    } else {
      await page.evaluate((id: number) => (window as any).showPlaceDetail(id), loc.id);
    }
    await expect(page.locator('#worldPlaceDetail')).toContainText(eventTitle, { timeout: NAV_TIMEOUT });
    await expect(page.locator('#worldPlaceDetail')).toContainText('Timeline events', { timeout: NAV_TIMEOUT });
  });

  test('timeline form persists a linked place', async ({ page }) => {
    const suffix = unique();
    const placeName = `World HTMX ${suffix}`;
    const eventTitle = `HTMX Evt ${suffix}`;

    const camp = await createCampaign(page, `HTMXCamp ${suffix}`);
    const cid = camp.id as number;
    await page.evaluate((id: number) => (window as any).selectCampaign(id), cid);
    await waitLoadingDone(page);

    const loc = await page.evaluate(async (name: string) => (window as any).api('POST', '/api/locations', { name, type: 'village', description: 'htmx place', latitude: 41, longitude: -2 }), placeName);
    expect(loc.id).toBeGreaterThan(0);

    await openTimeline(page);

    // Open New Event modal - button from timeline_list.html
    const newBtn = page.getByRole('button', { name: /New Event/ });
    await expect(newBtn).toBeVisible({ timeout: NAV_TIMEOUT });
    await newBtn.click();
    await expect(page.locator('#genericModal')).toContainText('New Event', { timeout: NAV_TIMEOUT });
    // form fields inside modal
    const modal = page.locator('#genericModal');
    await expect(modal.locator('input[name="title"]')).toBeVisible({ timeout: NAV_TIMEOUT });
    await modal.locator('input[name="title"]').fill(eventTitle);
    const dateInput = modal.locator('input[name="event_date"]');
    await dateInput.fill('2026-06-15');
    const campIdInput = modal.locator('input[name="campaign_id"]');
    if (await campIdInput.isVisible().catch(() => false)) await campIdInput.fill(String(cid));
    // Wait for place select to be populated from /api/locations
    const tlPlace = modal.locator('#tlPlace');
    await expect(tlPlace).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.waitForFunction(() => {
      const s = document.getElementById('tlPlace') as HTMLSelectElement | null;
      return s && s.options.length > 1;
    }, { timeout: NAV_TIMEOUT });
    await tlPlace.selectOption(String(loc.id));
    // linked_entity_type hidden field should become 'location'
    await expect(modal.locator('#tlPlaceType')).toHaveValue('location', { timeout: 5000 });

    // Submit HTMX create
    const submit = modal.getByRole('button', { name: /Create Event/ });
    await expect(submit).toBeVisible({ timeout: NAV_TIMEOUT });
    // Wait for HTMX swap to complete: timelineContent should contain new title or modal closes
    await Promise.all([
      page.waitForResponse((r) => r.url().includes('/htmx/timeline') && r.request().method() === 'POST', { timeout: NAV_TIMEOUT }).catch(() => null),
      submit.click(),
    ]);
    // Modal auto-hides via script in timeline_list.html after swap; wait a bit
    await waitModalClosed(page);
    // Timeline list should now contain the new event
    await expect(page.locator('#timelineContent')).toContainText(eventTitle, { timeout: NAV_TIMEOUT });

    // Verify via JSON API that linked fields persisted
    const events = await page.evaluate(async (c: number) => (window as any).api('GET', `/api/timeline?campaign_id=${c}`), cid);
    const arr: any[] = Array.isArray(events) ? events : events.events ?? [];
    const found = arr.find((e: any) => e.title === eventTitle);
    expect(found, `event ${eventTitle} not found in GET /api/timeline`).toBeTruthy();
    expect(found.linked_entity_type).toBe('location');
    expect(String(found.linked_entity_id)).toBe(String(loc.id));
  });

  test('uploads a fantasy world map and can reset to parchment', async ({ page }) => {
    const camp = await createCampaign(page, `FantasyMap ${unique()}`);
    const cid = camp.id as number;
    await page.evaluate((id: number) => (window as any).selectCampaign(id), cid);
    await waitLoadingDone(page);

    await openWorld(page);
    // No uploaded map yet → parchment basemap and the upload affordance.
    const uploadBtn = page.getByRole('button', { name: 'Upload world map' });
    await expect(uploadBtn).toBeVisible({ timeout: NAV_TIMEOUT });
    await uploadBtn.click();

    // The shared FilePicker opens; feed it a 1x1 PNG without a native dialog.
    const png = Buffer.from(
      'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAC0lEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==',
      'base64',
    );
    await page.setInputFiles('#fpFileInput', { name: 'world.png', mimeType: 'image/png', buffer: png });
    // After the picker resolves, showWorld re-renders with the stored image.
    await expect(page.getByRole('button', { name: 'Change map' })).toBeVisible({ timeout: NAV_TIMEOUT });

    const maps = await page.evaluate(async (c: number) => (window as any).api('GET', `/api/campaigns/${c}/maps`), cid);
    const worldMap = maps.find((m: any) => m.name === 'World Map');
    expect(worldMap, 'world map row was not persisted').toBeTruthy();
    expect(worldMap.image_url).toBeTruthy();

    // Remove → parchment again.
    await page.getByRole('button', { name: 'Remove' }).click();
    await expect(page.getByRole('button', { name: 'Upload world map' })).toBeVisible({ timeout: NAV_TIMEOUT });
    const after = await page.evaluate(async (c: number) => (window as any).api('GET', `/api/campaigns/${c}/maps`), cid);
    expect(after.find((m: any) => m.name === 'World Map')?.image_url).toBe('');
  });
});
