import { test, expect } from './fixtures.js';
import { clickSecondaryNavItem, login, waitLoadingDone, NAV_TIMEOUT } from './helpers.js';

const uniqueName = () => `OSP-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

/** Load HTMX content into oneshotSection (or another target) */
async function loadHtmx(page, url, target?: string) {
  await page.evaluate(async ({ u, t }: { u: string; t: string | null }) => {
    const resp = await fetch(u, { credentials: 'same-origin' });
    const el = document.getElementById(t || 'oneshotSection')!;
    el.innerHTML = await resp.text();
    (window as any).htmx?.process(el);
  }, { u: url, t: target || null });
}

/** Click the One-Shots nav link, handling mobile hamburger menu */
async function navigateToOneShots(page) {
  await clickSecondaryNavItem(page, 'oneshots', 'moreNavOneshot', 'One-Shots');
  await page.waitForSelector('#oneshotSection', { state: 'visible', timeout: NAV_TIMEOUT });
}

/** Create a one-shot adventure with acts/scenes via the template endpoint */
async function createGeneratedOneShot(page, title: string) {
  return page.evaluate(async (t: string) => {
    return (window as any).api('POST', '/api/oneshot-adventures/generate', {
      title: t, template: 'five_room_dungeon', difficulty: 'easy', estimated_minutes: 60,
    });
  }, title);
}

/** Open an adventure detail fragment and wait for the prep toolbar */
async function openDetail(page, advId: number) {
  await loadHtmx(page, `/htmx/oneshot-adventures/${advId}`);
  await expect(page.getByTestId('oneshot-run')).toBeVisible({ timeout: NAV_TIMEOUT });
}

/** Read the session timer as seconds (mm:ss) */
async function timerSeconds(page) {
  const text = (await page.locator('#sessionTimer').innerText()).trim();
  const m = text.match(/(\d+):(\d+)/);
  return m ? parseInt(m[1], 10) * 60 + parseInt(m[2], 10) : -1;
}

test.describe('One-shot prep wiring', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('toolbar opens every prep/run surface and each has a back button', async ({ page }) => {
    await navigateToOneShots(page);
    const adv = await createGeneratedOneShot(page, uniqueName());
    await openDetail(page, adv.id);

    await page.getByTestId('oneshot-prep-dashboard').click();
    await expect(page.getByTestId('prep-back')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('prep-start-session')).toBeVisible();
    await page.getByTestId('prep-back').click();
    await expect(page.getByTestId('oneshot-run')).toBeVisible({ timeout: NAV_TIMEOUT });

    await page.getByTestId('oneshot-dm-screen').click();
    await expect(page.getByTestId('dmscreen-back')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#oneshotSection')).toContainText('Quick Reference');
    await page.locator('#actions-tab').click();
    await expect(page.getByTestId('dm-screen-generators')).toBeVisible();
    await page.getByTestId('dm-screen-hook').click();
    await expect.poll(async () => (await page.locator('#hookResult').innerText()).length, { timeout: NAV_TIMEOUT }).toBeGreaterThan(0);
    await page.getByTestId('dmscreen-back').click();
    await expect(page.getByTestId('oneshot-run')).toBeVisible({ timeout: NAV_TIMEOUT });

    await page.getByTestId('oneshot-clues').click();
    await expect(page.getByTestId('clues-back')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.getByTestId('clues-back').click();
    await expect(page.getByTestId('oneshot-run')).toBeVisible({ timeout: NAV_TIMEOUT });

    await page.getByTestId('oneshot-session-flow').click();
    await expect(page.getByTestId('flow-back')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.getByTestId('flow-back').click();
    await expect(page.getByTestId('oneshot-run')).toBeVisible({ timeout: NAV_TIMEOUT });

    await page.getByTestId('oneshot-pregens').click();
    await expect(page.getByTestId('pregen-new')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.getByTestId('pregens-back')).toBeVisible();
  });

  test('pacing clock advances and survives pause/resume and a reload', async ({ page }) => {
    await navigateToOneShots(page);
    const adv = await createGeneratedOneShot(page, uniqueName());
    await openDetail(page, adv.id);

    await page.getByTestId('oneshot-run').click();
    await expect(page.getByTestId('pacing-pause')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#oneshotSection')).toContainText('Session Dashboard');

    const t0 = await timerSeconds(page);
    await expect.poll(() => timerSeconds(page), { timeout: 20000 }).toBeGreaterThan(t0);

    await page.getByTestId('pacing-pause').click();
    await expect(page.getByTestId('pacing-resume')).toBeVisible({ timeout: NAV_TIMEOUT });
    const paused = await timerSeconds(page);
    expect(paused).toBeGreaterThanOrEqual(t0);

    const serverPaused = await page.evaluate(async (id: number) => {
      const s = await (window as any).api('GET', `/api/oneshot-adventures/${id}/pacing`);
      return s.elapsed_seconds as number;
    }, adv.id);
    expect(serverPaused).toBeGreaterThanOrEqual(paused);

    // Reload: the persisted elapsed must not reset; Run resumes the paused session
    await page.reload();
    await waitLoadingDone(page);
    await loadHtmx(page, `/htmx/oneshot-adventures/${adv.id}`);
    await expect(page.getByTestId('oneshot-run')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.getByTestId('oneshot-run').click();
    await expect(page.getByTestId('pacing-pause')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect.poll(() => timerSeconds(page), { timeout: 20000 }).toBeGreaterThanOrEqual(paused);
    await expect.poll(() => timerSeconds(page), { timeout: 20000 }).toBeGreaterThan(paused);
  });

  test('pregens can be created, edited and deleted from the surface', async ({ page }) => {
    await navigateToOneShots(page);
    const adv = await createGeneratedOneShot(page, uniqueName());
    await openDetail(page, adv.id);
    await page.getByTestId('oneshot-pregens').click();
    await expect(page.getByTestId('pregen-new')).toBeVisible({ timeout: NAV_TIMEOUT });

    const name = uniqueName();
    await page.getByTestId('pregen-new').click();
    await expect(page.locator('#pregenForm input[name="name"]')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.locator('#pregenForm input[name="name"]').fill(name);
    await page.locator('#pregenForm input[name="race"]').fill('Elf');
    await page.locator('#pregenForm input[name="class"]').fill('Wizard');
    await page.locator('#pregenForm button[type="submit"]').click();
    await expect(page.locator('#pregenList')).toContainText(name, { timeout: NAV_TIMEOUT });

    const card = page.locator('#pregenList [data-pregen-id]').filter({ hasText: name }).first();
    await card.locator('[data-testid="pregen-edit"]').click();
    await expect(page.locator('#pregenForm input[name="name"]')).toHaveValue(name, { timeout: NAV_TIMEOUT });
    const edited = `${name}-II`;
    await page.locator('#pregenForm input[name="name"]').fill(edited);
    await page.locator('#pregenForm button[type="submit"]').click();
    await expect(page.locator('#pregenList')).toContainText(edited, { timeout: NAV_TIMEOUT });

    page.once('dialog', (d) => d.accept());
    await page.locator('#pregenList [data-pregen-id]').filter({ hasText: edited }).first()
      .locator('[data-testid="pregen-delete"]').click();
    await expect(page.locator('#pregenList')).not.toContainText(edited, { timeout: NAV_TIMEOUT });
  });

  test('scene dialogs can be drag reordered and the order persists', async ({ page }) => {
    await navigateToOneShots(page);
    const adv = await createGeneratedOneShot(page, uniqueName());
    const detail = await page.evaluate(async (id: number) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
    }, adv.id);
    expect(detail.acts.length).toBeGreaterThan(0);
    expect(detail.acts[0].scenes.length).toBeGreaterThan(0);
    const sceneId = detail.acts[0].scenes[0].id;
    await openDetail(page, adv.id);

    await page.locator(`.sortable-scene[data-id="${sceneId}"] .btn-outline-warning`).first().click();
    await expect(page.locator('#genericModalBody .dialogs-container')).toBeVisible({ timeout: NAV_TIMEOUT });

    const speakerA = `A-${uniqueName()}`;
    const speakerB = `B-${uniqueName()}`;
    for (const speaker of [speakerA, speakerB]) {
      await page.locator('#genericModalBody button:has-text("Add Dialog")').click();
      await expect(page.locator('#dialogFormContainer form')).toBeVisible({ timeout: NAV_TIMEOUT });
      await page.locator('#genericModalBody input[name="speaker"]').fill(speaker);
      await page.locator('#genericModalBody textarea[name="dialog_text"]').fill(`Line for ${speaker}`);
      await page.locator('#dialogFormContainer button:has-text("Add Dialog")').click();
      await expect(page.locator('#genericModalBody .dialog-card').filter({ hasText: speaker })).toBeVisible({ timeout: NAV_TIMEOUT });
    }

    // Drag B above A using the sortable handle
    const src = page.locator('#genericModalBody .dialog-card').filter({ hasText: speakerB }).first();
    const dst = page.locator('#genericModalBody .dialog-card').filter({ hasText: speakerA }).first();
    const hb = await src.locator('.sortable-handle').boundingBox();
    const db = await dst.boundingBox();
    if (!hb || !db) throw new Error('dialog bounding boxes not found');
    await page.mouse.move(hb.x + hb.width / 2, hb.y + hb.height / 2);
    await page.mouse.down();
    await page.mouse.move(hb.x + hb.width / 2, hb.y + hb.height / 2 - 20, { steps: 5 });
    await page.mouse.move(db.x + db.width / 2, db.y + 5, { steps: 10 });
    await page.mouse.up();

    // Persisted order comes from the server-rendered fragment
    await expect.poll(async () => {
      return page.evaluate(async (sid: number) => {
        const resp = await fetch(`/htmx/oneshot-scenes/${sid}/dialogs`, { credentials: 'same-origin' });
        const html = await resp.text();
        const doc = new DOMParser().parseFromString(html, 'text/html');
        const texts = Array.from(doc.querySelectorAll('.dialog-card')).map((el) => el.textContent || '');
        return { a: texts.findIndex((t) => t.includes('A-')), b: texts.findIndex((t) => t.includes('B-')) };
      }, sceneId);
    }, { timeout: NAV_TIMEOUT }).toMatchObject({ a: 1, b: 0 });
  });

  test('red herring flag is settable from the clue form', async ({ page }) => {
    await navigateToOneShots(page);
    const adv = await createGeneratedOneShot(page, uniqueName());
    await openDetail(page, adv.id);
    await page.getByTestId('oneshot-clues').click();
    await expect(page.getByTestId('clues-back')).toBeVisible({ timeout: NAV_TIMEOUT });

    await page.locator('#oneshotSection button:has-text("Add Clue")').click();
    await expect(page.locator('#clueForm input[name="title"]')).toBeVisible({ timeout: NAV_TIMEOUT });
    const clueTitle = uniqueName();
    await page.locator('#clueForm input[name="title"]').fill(clueTitle);
    await page.locator('#clueForm textarea[name="description"]').fill('A suspicious lead');
    await page.locator('#clueForm #clueRedHerringNew').check();
    await page.locator('#clueForm button[type="submit"]').click();

    const card = page.locator('#oneshotSection .clue-card').filter({ hasText: clueTitle }).first();
    await expect(card).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(card.locator('.fa-skull')).toBeVisible();
  });

  test('one-shot generators are reachable from the DM tools modal', async ({ page }) => {
    await navigateToOneShots(page);
    const adv = await createGeneratedOneShot(page, uniqueName());
    await openDetail(page, adv.id);
    await page.getByTestId('oneshot-prep-dashboard').click();
    await expect(page.getByTestId('prep-generators')).toBeVisible({ timeout: NAV_TIMEOUT });

    await page.getByTestId('prep-generators').click();
    for (const id of ['dm-tool-hook', 'dm-tool-dressing', 'dm-tool-tavern', 'dm-tool-urban', 'dm-tool-road']) {
      await expect(page.getByTestId(id)).toBeVisible({ timeout: NAV_TIMEOUT });
    }

    await page.getByTestId('dm-tool-hook').click();
    await expect.poll(async () => (await page.locator('#dmToolsResult').innerText()).length, { timeout: NAV_TIMEOUT }).toBeGreaterThan(0);
    const first = await page.locator('#dmToolsResult').innerHTML();
    await page.getByTestId('dm-tool-tavern').click();
    await expect.poll(async () => (await page.locator('#dmToolsResult').innerHTML()) !== first, { timeout: NAV_TIMEOUT }).toBe(true);
  });

  test('adventure search results open the one-shot detail and survive a reload', async ({ page }) => {
    await navigateToOneShots(page);
    const title = uniqueName();
    const adv = await createGeneratedOneShot(page, title);

    await page.evaluate(({ id, t }: { id: number; t: string }) => {
      (window as any).navigateSearchResult('adventure', id, t);
    }, { id: adv.id, t: title });

    await expect.poll(() => page.evaluate(() => location.hash), { timeout: NAV_TIMEOUT }).toBe(`#/adventures/${adv.id}`);
    await expect(page.getByTestId('oneshot-run')).toBeVisible({ timeout: NAV_TIMEOUT });

    await page.reload();
    await waitLoadingDone(page);
    await expect(page.getByTestId('oneshot-run')).toBeVisible({ timeout: NAV_TIMEOUT });
  });
});
