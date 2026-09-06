import { test, expect } from './fixtures.js';
import { waitLoadingDone, clickSecondaryNavItem, login, NAV_TIMEOUT } from './helpers.js';

const uniqueName = () => `OSC-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

/** Load HTMX content into oneshotSection */
async function loadHtmx(page, url, target?: string) {
  await page.evaluate(async ({ u, t }) => {
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

/** Create a one-shot adventure via API and return its ID */
async function createOneShot(page, title: string) {
  return page.evaluate(async (t) => {
    return (window as any).api('POST', '/api/oneshot-adventures', {
      title: t, template: 'custom', difficulty: 'medium', estimated_minutes: 120,
    });
  }, title);
}

/** Create a one-shot adventure with acts/scenes via template */
async function createGeneratedOneShot(page, title: string) {
  return page.evaluate(async (t) => {
    return (window as any).api('POST', '/api/oneshot-adventures/generate', {
      title: t, template: 'five_room_dungeon', difficulty: 'easy', estimated_minutes: 60,
    });
  }, title);
}

test.describe('One-Shot Content Features', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

// ─── Acts & Scenes API CRUD ───

test.describe('Acts & Scenes API CRUD', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
  });

  test('Create act for a one-shot', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const result = await page.evaluate(async ({ id, t }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${id}/acts`, {
        title: t, number: 1,
      });
    }, { id: adv.id, t: title + ' Act' });
    expect(result.id).toBeGreaterThan(0);
  });

  test('List acts via adventure detail', async ({ page }) => {
    const title = uniqueName();
    const adv = await createGeneratedOneShot(page, title);
    expect(adv.id).toBeGreaterThan(0);
    const detail = await page.evaluate(async (id) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
    }, adv.id);
    expect(detail.acts.length).toBeGreaterThan(0);
  });

  test('Update act title', async ({ page }) => {
    const title = uniqueName();
    const adv = await createGeneratedOneShot(page, title);
    const detail = await page.evaluate(async (id) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
    }, adv.id);
    const actId = detail.acts[0].id;
    const newTitle = 'Updated ' + uniqueName();

    await page.evaluate(async ({ id, t }) => {
      return (window as any).api('PUT', `/api/oneshot-acts/${id}`, {
        title: t, number: 1, sort_order: 1,
      });
    }, { id: actId, t: newTitle });

    const updated = await page.evaluate(async (id) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
    }, adv.id);
    const act = updated.acts.find((a: any) => a.id === actId);
    expect(act).toBeDefined();
    expect(act.title).toBe(newTitle);
  });

  test('Delete an act', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const act = await page.evaluate(async ({ id, t }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${id}/acts`, {
        title: t, number: 1,
      });
    }, { id: adv.id, t: title + ' Act' });

    await page.evaluate(async (id) => {
      return (window as any).api('DELETE', `/api/oneshot-acts/${id}`);
    }, act.id);

    const detail = await page.evaluate(async (id) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
    }, adv.id);
    expect(!detail.acts || !detail.acts.some((a: any) => a.id === act.id)).toBe(true);
  });

  test('Create scene within act', async ({ page }) => {
    const title = uniqueName();
    const adv = await createGeneratedOneShot(page, title);
    const detail = await page.evaluate(async (id) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
    }, adv.id);
    const actId = detail.acts[0].id;

    const result = await page.evaluate(async ({ actId, t }) => {
      return (window as any).api('POST', `/api/oneshot-acts/${actId}/scenes`, {
        title: t, scene_type: 'combat', estimated_minutes: 15,
      });
    }, { actId, t: title + ' Scene' });
    expect(result.id).toBeGreaterThan(0);
  });

  test('Update scene details', async ({ page }) => {
    const title = uniqueName();
    const adv = await createGeneratedOneShot(page, title);
    const detail = await page.evaluate(async (id) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
    }, adv.id);
    const sceneId = detail.acts[0].scenes[0].id;
    const newTitle = 'Updated Scene ' + uniqueName();

    await page.evaluate(async ({ id, t }) => {
      return (window as any).api('PUT', `/api/oneshot-scenes/${id}`, {
        title: t, scene_type: 'exploration', estimated_minutes: 20,
      });
    }, { id: sceneId, t: newTitle });

    const updated = await page.evaluate(async (id) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
    }, adv.id);
    const scene = updated.acts[0].scenes.find((s: any) => s.id === sceneId);
    expect(scene).toBeDefined();
    expect(scene.title).toBe(newTitle);
  });

  test('Delete a scene', async ({ page }) => {
    const title = uniqueName();
    const adv = await createGeneratedOneShot(page, title);
    const detail = await page.evaluate(async (id) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
    }, adv.id);
    const sceneId = detail.acts[0].scenes[0].id;

    await page.evaluate(async (id) => {
      return (window as any).api('DELETE', `/api/oneshot-scenes/${id}`);
    }, sceneId);

    const updated = await page.evaluate(async (id) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
    }, adv.id);
    const scenes = updated.acts[0].scenes || [];
    expect(scenes.some((s: any) => s.id === sceneId)).toBe(false);
  });
});

// ─── Session Pacing ───

test.describe('Session Pacing', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
  });

  test('Start and get pacing session', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const act = await page.evaluate(async ({ id, t }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${id}/acts`, {
        title: t, number: 1,
      });
    }, { id: adv.id, t: title + ' Act' });

    const result = await page.evaluate(async ({ advId, actId }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/pacing/start`, {
        current_act_id: actId,
      });
    }, { advId: adv.id, actId: act.id });
    expect(result.id).toBeGreaterThan(0);

    // Get pacing session (via session-pacing/:id)
    const pacing = await page.evaluate(async (sid) => {
      return (window as any).api('GET', `/api/session-pacing/${sid}`);
    }, result.id);
    expect(pacing).toBeTruthy();
    expect(pacing.status).toBe('running');
  });

  test('Pause and resume pacing session', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const act = await page.evaluate(async ({ id, t }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${id}/acts`, {
        title: t, number: 1,
      });
    }, { id: adv.id, t: title + ' Act' });

    const started = await page.evaluate(async ({ advId, actId }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/pacing/start`, {
        current_act_id: actId,
      });
    }, { advId: adv.id, actId: act.id });

    // Pause (via session-pacing/:id/pause)
    const paused = await page.evaluate(async (sid) => {
      return (window as any).api('POST', `/api/session-pacing/${sid}/pause`);
    }, started.id);
    expect(paused.status).toBe('paused');

    // Resume (via session-pacing/:id/resume)
    const resumed = await page.evaluate(async (sid) => {
      return (window as any).api('POST', `/api/session-pacing/${sid}/resume`);
    }, started.id);
    expect(resumed.status).toBe('running');
  });

  test('Advance and complete pacing session', async ({ page }) => {
    const title = uniqueName();
    const adv = await createGeneratedOneShot(page, title);
    const detail = await page.evaluate(async (id) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
    }, adv.id);
    const actId = detail.acts[0].id;

    const started = await page.evaluate(async ({ advId, actId }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/pacing/start`, {
        current_act_id: actId,
      });
    }, { advId: adv.id, actId });

    // Advance (via session-pacing/:id/next-scene)
    const advanced = await page.evaluate(async (sid) => {
      return (window as any).api('POST', `/api/session-pacing/${sid}/next-scene`);
    }, started.id);
    expect(advanced).toBeTruthy();

    // Complete (via session-pacing/:id/complete)
    const completed = await page.evaluate(async (sid) => {
      return (window as any).api('POST', `/api/session-pacing/${sid}/complete`);
    }, started.id);
    expect(completed.status).toBe('completed');
  });
});

// ─── Clues ───

test.describe('Clues', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
  });

  test('Create a clue', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const result = await page.evaluate(async ({ advId, t }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/clues`, {
        adventure_id: advId, title: t, description: 'A mysterious clue',
        clue_type: 'object', is_red_herring: false,
      });
    }, { advId: adv.id, t: 'Clue ' + uniqueName() });
    expect(result.id).toBeGreaterThan(0);
  });

  test('List clues for a one-shot', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const clueTitle = 'List Clue ' + uniqueName();
    await page.evaluate(async ({ advId, t }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/clues`, {
        adventure_id: advId, title: t, description: 'List test',
        clue_type: 'witness', is_red_herring: false,
      });
    }, { advId: adv.id, t: clueTitle });

    const clues = await page.evaluate(async (advId) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${advId}/clues`);
    }, adv.id);
    expect(Array.isArray(clues)).toBe(true);
    expect(clues.some((c: any) => c.title === clueTitle)).toBe(true);
  });

  test('Update a clue', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const clueTitle = 'Update Clue ' + uniqueName();
    const created = await page.evaluate(async ({ advId, t }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/clues`, {
        adventure_id: advId, title: t, description: 'Original',
        clue_type: 'location', is_red_herring: false,
      });
    }, { advId: adv.id, t: clueTitle });

    const newTitle = clueTitle + '-updated';
    await page.evaluate(async ({ id, t }) => {
      return (window as any).api('PUT', `/api/clues/${id}`, {
        title: t, description: 'Updated clue', clue_type: 'object', is_red_herring: true,
      });
    }, { id: created.id, t: newTitle });

    const clues = await page.evaluate(async (advId) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${advId}/clues`);
    }, adv.id);
    const updated = clues.find((c: any) => c.id === created.id);
    expect(updated).toBeDefined();
    expect(updated.title).toBe(newTitle);
  });

  test('Delete a clue', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const created = await page.evaluate(async ({ advId, t }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/clues`, {
        adventure_id: advId, title: t, description: 'Delete me',
        clue_type: 'object', is_red_herring: false,
      });
    }, { advId: adv.id, t: 'Delete Clue ' + uniqueName() });

    await page.evaluate(async (id) => {
      return (window as any).api('DELETE', `/api/clues/${id}`);
    }, created.id);

    const clues = await page.evaluate(async (advId) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${advId}/clues`);
    }, adv.id);
    expect(clues.some((c: any) => c.id === created.id)).toBe(false);
  });

  test('Reveal and hide a clue', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const created = await page.evaluate(async ({ advId, t }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/clues`, {
        adventure_id: advId, title: t, description: 'Toggle me',
        clue_type: 'object', is_red_herring: false,
      });
    }, { advId: adv.id, t: 'Reveal Clue ' + uniqueName() });

    // Reveal
    const revealed = await page.evaluate(async (id) => {
      return (window as any).api('POST', `/api/clues/${id}/reveal`);
    }, created.id);
    expect(revealed.status).toBe('revealed');

    // Hide
    const hidden = await page.evaluate(async (id) => {
      return (window as any).api('POST', `/api/clues/${id}/hide`);
    }, created.id);
    expect(hidden.status).toBe('hidden');
  });
});

// ─── Prep Checklist ───

test.describe('Prep Checklist', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
  });

  test('Create checklist item', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const result = await page.evaluate(async ({ advId }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/checklist`, {
        adventure_id: advId, item: 'Print maps', category: 'props', is_checked: false,
      });
    }, { advId: adv.id });
    expect(result.id).toBeGreaterThan(0);
  });

  test('List checklist for a one-shot', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    await page.evaluate(async ({ advId }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/checklist`, {
        adventure_id: advId, item: 'Prepare minis', category: 'props', is_checked: false,
      });
    }, { advId: adv.id });

    const items = await page.evaluate(async (advId) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${advId}/checklist`);
    }, adv.id);
    expect(Array.isArray(items)).toBe(true);
    expect(items.some((i: any) => i.item === 'Prepare minis')).toBe(true);
  });

  test('Update checklist item', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const created = await page.evaluate(async ({ advId }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/checklist`, {
        adventure_id: advId, item: 'Old item', category: 'notes', is_checked: false,
      });
    }, { advId: adv.id });

    await page.evaluate(async (id) => {
      return (window as any).api('PUT', `/api/prep-checklist/${id}`, {
        item: 'Updated item', is_checked: true,
      });
    }, created.id);

    const items = await page.evaluate(async (advId) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${advId}/checklist`);
    }, adv.id);
    const updated = items.find((i: any) => i.id === created.id);
    expect(updated).toBeDefined();
    expect(updated.is_checked).toBe(true);
  });

  test('Delete checklist item', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const created = await page.evaluate(async ({ advId }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/checklist`, {
        adventure_id: advId, item: 'Delete me', category: 'other', is_checked: false,
      });
    }, { advId: adv.id });

    await page.evaluate(async (id) => {
      return (window as any).api('DELETE', `/api/prep-checklist/${id}`);
    }, created.id);

    const items = await page.evaluate(async (advId) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${advId}/checklist`);
    }, adv.id);
    expect(!items || !items.some((i: any) => i.id === created.id)).toBe(true);
  });
});

// ─── DM Notes ───

test.describe('DM Notes', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
  });

  test('Create DM note', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const noteTitle = 'Note ' + uniqueName();
    const result = await page.evaluate(async ({ advId, t }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/notes`, {
        adventure_id: advId, title: t, content: 'DM secret content',
      });
    }, { advId: adv.id, t: noteTitle });
    expect(result.id).toBeGreaterThan(0);
  });

  test('List DM notes for a one-shot', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const noteTitle = 'List Note ' + uniqueName();
    await page.evaluate(async ({ advId, t }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/notes`, {
        adventure_id: advId, title: t, content: 'List test content',
      });
    }, { advId: adv.id, t: noteTitle });

    const notes = await page.evaluate(async (advId) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${advId}/notes`);
    }, adv.id);
    expect(Array.isArray(notes)).toBe(true);
    expect(notes.some((n: any) => n.title === noteTitle)).toBe(true);
  });

  test('Update DM note', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const noteTitle = 'Update Note ' + uniqueName();
    const created = await page.evaluate(async ({ advId, t }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/notes`, {
        adventure_id: advId, title: t, content: 'Original content',
      });
    }, { advId: adv.id, t: noteTitle });

    const newTitle = noteTitle + '-updated';
    await page.evaluate(async ({ id, t }) => {
      return (window as any).api('PUT', `/api/dm-notes/${id}`, {
        title: t, content: 'Updated content',
      });
    }, { id: created.id, t: newTitle });

    const notes = await page.evaluate(async (advId) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${advId}/notes`);
    }, adv.id);
    const updated = notes.find((n: any) => n.id === created.id);
    expect(updated).toBeDefined();
    expect(updated.title).toBe(newTitle);
  });

  test('Delete DM note', async ({ page }) => {
    const title = uniqueName();
    const adv = await createOneShot(page, title);
    const noteTitle = 'Delete Note ' + uniqueName();
    const created = await page.evaluate(async ({ advId, t }) => {
      return (window as any).api('POST', `/api/oneshot-adventures/${advId}/notes`, {
        adventure_id: advId, title: t, content: 'Delete me',
      });
    }, { advId: adv.id, t: noteTitle });

    await page.evaluate(async ({ advId, noteId }) => {
      return (window as any).api('DELETE', `/api/oneshot-adventures/${advId}/notes/${noteId}`);
    }, { advId: adv.id, noteId: created.id });

    const notes = await page.evaluate(async (advId) => {
      return (window as any).api('GET', `/api/oneshot-adventures/${advId}/notes`);
    }, adv.id);
    expect(notes.some((n: any) => n.id === created.id)).toBe(false);
  });
  });
});
