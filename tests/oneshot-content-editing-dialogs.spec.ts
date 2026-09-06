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
  test.describe('Inline Editing', () => {
    test('Update act duration', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThan(0);
      const actId = detail.acts[0].id;

      const result = await page.evaluate(async ({ id, minutes }) => {
        return (window as any).api('PATCH', `/api/oneshot-acts/${id}/duration`, {
          estimated_minutes: minutes,
        });
      }, { id: actId, minutes: 45 });
      expect(result.ok).toBe(true);
      expect(result.estimated_minutes).toBe(45);
    });

    test('Update scene duration', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThan(0);
      expect(detail.acts[0].scenes.length).toBeGreaterThan(0);
      const sceneId = detail.acts[0].scenes[0].id;

      const result = await page.evaluate(async ({ id, minutes }) => {
        return (window as any).api('PATCH', `/api/oneshot-scenes/${id}/duration`, {
          estimated_minutes: minutes,
        });
      }, { id: sceneId, minutes: 20 });
      expect(result.ok).toBe(true);
      expect(result.estimated_minutes).toBe(20);
    });

    test('Reorder acts', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThanOrEqual(2);

      // Reverse the order
      const reversed = [...detail.acts].reverse().map((a: any) => a.id);
      const result = await page.evaluate(async ({ id, order }) => {
        return (window as any).api('PUT', `/api/oneshot-adventures/${id}/acts/reorder`, { order });
      }, { id: adv.id, order: reversed });
      expect(result.ok).toBe(true);

      // Verify the order changed
      const updated = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(updated.acts[0].id).toBe(reversed[0]);
    });

    test('Reorder scenes within an act', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThan(0);

      const actId = detail.acts[0].id;
      // Ensure at least 2 scenes for reorder
      if (detail.acts[0].scenes.length < 2) {
        await page.evaluate(async ({ actId, title }) => {
          return (window as any).api('POST', `/api/oneshot-acts/${actId}/scenes`, {
            title, scene_type: 'exploration', estimated_minutes: 15,
          });
        }, { actId, title: 'Extra Scene' });
        const updated = await page.evaluate(async (id) => {
          return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
        }, adv.id);
        detail.acts = updated.acts;
      }
      expect(detail.acts[0].scenes.length).toBeGreaterThanOrEqual(2);

      const reversed = [...detail.acts[0].scenes].reverse().map((s: any) => s.id);
      const result = await page.evaluate(async ({ id, order }) => {
        return (window as any).api('PUT', `/api/oneshot-acts/${id}/scenes/reorder`, { order });
      }, { id: actId, order: reversed });
      expect(result.ok).toBe(true);

      // Verify the order changed
      const updated = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      const updatedAct = updated.acts.find((a: any) => a.id === actId);
      expect(updatedAct).toBeDefined();
      expect(updatedAct.scenes[0].id).toBe(reversed[0]);
    });
  });

  // ─── Act & Scene Editing ───

  test.describe('Act & Scene Editing', () => {
    test('Edit act title via HTMX', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThan(0);
      const actId = detail.acts[0].id;
      const newTitle = 'Edited ' + uniqueName();

      await navigateToOneShots(page);
      await loadHtmx(page, `/htmx/oneshot-adventures/${adv.id}`);
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

      // Click edit button on first act
      const editBtn = page.locator(`.sortable-act[data-id="${actId}"] .btn-outline-secondary`).first();
      await editBtn.click();
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

      // Fill modal and submit
      const titleInput = page.locator('#genericModalBody input[name="title"]');
      await titleInput.fill(newTitle);
      const submitBtn = page.locator('#genericModalBody button[class*="btn-primary"]');
      await submitBtn.click();
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

      // Verify updated title in detail
      await expect(page.locator('#oneshotSection')).toContainText(newTitle, { timeout: NAV_TIMEOUT });
    });

    test('Edit scene details via HTMX', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThan(0);
      expect(detail.acts[0].scenes.length).toBeGreaterThan(0);
      const sceneId = detail.acts[0].scenes[0].id;
      const newTitle = 'Edited Scene ' + uniqueName();

      await navigateToOneShots(page);
      await loadHtmx(page, `/htmx/oneshot-adventures/${adv.id}`);
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

      // Click edit button on first scene
      const editBtn = page.locator(`.sortable-scene[data-id="${sceneId}"] .btn-outline-secondary`).first();
      await editBtn.click();
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

      // Fill modal
      const titleInput = page.locator('#genericModalBody input[name="title"]');
      await titleInput.fill(newTitle);
      const submitBtn = page.locator('#genericModalBody button[class*="btn-primary"]');
      await submitBtn.click();
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

      // Verify updated title
      await expect(page.locator('#oneshotSection')).toContainText(newTitle, { timeout: NAV_TIMEOUT });
    });

    test('Edit act parent via API', async ({ page }) => {
      const title = uniqueName();
      const adv = await createOneShot(page, title);

      // Create 2 acts
      const act1 = await page.evaluate(async ({ id }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/acts`, {
          title: 'Root Act', number: 1, sort_order: 1,
        });
      }, { id: adv.id });

      const act2 = await page.evaluate(async ({ id }) => {
        return (window as any).api('POST', `/api/oneshot-adventures/${id}/acts`, {
          title: 'Child Act', number: 1, sort_order: 1,
        });
      }, { id: adv.id });

      // Move act2 under act1
      await page.evaluate(async ({ id, parentId }) => {
        return (window as any).api('PUT', `/api/oneshot-acts/${id}`, {
          title: 'Child Act', number: 1, sort_order: 1, parent_act_id: parentId,
        });
      }, { id: act2.id, parentId: act1.id });

      // Verify tree
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBe(1);
      expect(detail.acts[0].children.length).toBe(1);
      expect(detail.acts[0].children[0].title).toBe('Child Act');
    });
  });

  // ─── Scene Dialogs ───

  test.describe('Scene Dialogs', () => {
    async function openDialogModal(page, sceneId) {
      const dialogBtn = page.locator(`.sortable-scene[data-id="${sceneId}"] .btn-outline-warning`).first();
      await dialogBtn.click();
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
    }

    async function createDialog(page, sceneId, speaker, text) {
      return page.evaluate(async ({ sceneId, speaker, text }) => {
        const csrf = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || '';
        // Mutating routes require a bearer API token (stored by the SPA on
        // login under a user-scoped key).
        let apiToken = localStorage.getItem('villum-api-token');
        if (!apiToken) {
          for (let i = 0; i < localStorage.length; i++) {
            const k = localStorage.key(i);
            if (k && k.startsWith('villum-api-token-')) { apiToken = localStorage.getItem(k); break; }
          }
        }
        const formData = new URLSearchParams();
        formData.append('speaker', speaker);
        formData.append('dialog_text', text);
        const resp = await fetch(`/htmx/oneshot-scenes/${sceneId}/dialogs`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
            'X-CSRF-Token': csrf,
            ...(apiToken ? { 'Authorization': `Bearer ${apiToken}` } : {}),
          },
          body: formData.toString(),
        });
        return resp.ok;
      }, { sceneId, speaker, text });
    }

    test('Create and view scene dialogs', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      expect(detail.acts.length).toBeGreaterThan(0);
      const sceneId = detail.acts[0].scenes[0].id;

      await navigateToOneShots(page);
      await loadHtmx(page, `/htmx/oneshot-adventures/${adv.id}`);
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

      // Open dialog modal for first scene
      await openDialogModal(page, sceneId);

      // Create a dialog entry
      const speaker = 'Guard ' + uniqueName();
      const dtext = 'Halt! Who goes there?';
      const created = await createDialog(page, sceneId, speaker, dtext);
      expect(created).toBe(true);

      // Reload dialogs in modal
      await loadHtmx(page, `/htmx/oneshot-scenes/${sceneId}/dialogs`, 'genericModalBody');
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
      await expect(page.locator('#genericModalBody')).toContainText(speaker, { timeout: NAV_TIMEOUT });
      await expect(page.locator('#genericModalBody')).toContainText(dtext, { timeout: NAV_TIMEOUT });
    });

    test('Edit scene dialog via HTMX', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      const sceneId = detail.acts[0].scenes[0].id;
      const newText = 'Password? ' + uniqueName();

      await navigateToOneShots(page);
      await loadHtmx(page, `/htmx/oneshot-adventures/${adv.id}`);
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

      // Open dialog modal and create a dialog
      await openDialogModal(page, sceneId);
      const created = await createDialog(page, sceneId, 'Captain', 'Old text');
      expect(created).toBe(true);

      // Reload dialogs in modal to get dialog in DOM
      await loadHtmx(page, `/htmx/oneshot-scenes/${sceneId}/dialogs`, 'genericModalBody');
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

      // Get dialog ID from the rendered dialog list
      const dialogId = await page.evaluate(() => {
        const card = document.querySelector('.dialog-card');
        return card ? parseInt(card.getAttribute('data-id') || '0') : 0;
      });
      expect(dialogId).toBeGreaterThan(0);

      // Click edit button on dialog (HTMX)
      const editBtn = page.locator(`.dialog-card[data-id="${dialogId}"] .btn-outline-secondary`).first();
      await editBtn.click();
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

      // Change text and submit
      const textarea = page.locator('#genericModalBody textarea[name="dialog_text"]');
      await textarea.fill(newText);
      const submitBtn = page.locator('#genericModalBody button[class*="btn-primary"]');
      await submitBtn.click();
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

      // Verify updated text
      await expect(page.locator('#genericModalBody')).toContainText(newText, { timeout: NAV_TIMEOUT });
    });

    test('Delete scene dialog', async ({ page }) => {
      const title = uniqueName();
      const adv = await createGeneratedOneShot(page, title);
      const detail = await page.evaluate(async (id) => {
        return (window as any).api('GET', `/api/oneshot-adventures/${id}`);
      }, adv.id);
      const sceneId = detail.acts[0].scenes[0].id;
      const speaker = 'DeleteMe ' + uniqueName();

      await navigateToOneShots(page);
      await loadHtmx(page, `/htmx/oneshot-adventures/${adv.id}`);
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

      // Open dialog modal and create dialog
      await openDialogModal(page, sceneId);
      await createDialog(page, sceneId, speaker, 'Delete this');

      // Reload dialogs in modal
      await loadHtmx(page, `/htmx/oneshot-scenes/${sceneId}/dialogs`, 'genericModalBody');
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

      // Get dialog ID from DOM
      const dialogId = await page.evaluate(() => {
        const card = document.querySelector('.dialog-card');
        return card ? parseInt(card.getAttribute('data-id') || '0') : 0;
      });
      expect(dialogId).toBeGreaterThan(0);

      await expect(page.locator('#genericModalBody')).toContainText(speaker, { timeout: NAV_TIMEOUT });

      // Click delete button (handle hx-confirm dialog)
      page.once('dialog', dialog => dialog.accept());
      const delBtn = page.locator(`.dialog-card[data-id="${dialogId}"] .btn-outline-danger`).first();
      await delBtn.click();
      await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

      // Verify deleted - speaker should no longer be visible
      await expect(page.locator('#genericModalBody')).not.toContainText(speaker, { timeout: NAV_TIMEOUT });
    });
  });
});

// ─── Acts & Scenes API CRUD ───


});
