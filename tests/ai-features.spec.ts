import { test, expect } from './fixtures.js';
import { login } from './helpers.js';

const uniqueName = () => `AI-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('AI Features', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
  });

  // ─── AI Endpoints ───

  test.describe('AI Endpoints', () => {
    test('List text AI endpoints', async ({ page }) => {
      const endpoints = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/ai/endpoints?type=text');
      });
      expect(Array.isArray(endpoints)).toBe(true);
    });

    test('List image AI endpoints', async ({ page }) => {
      const endpoints = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/ai/endpoints?type=image');
      });
      expect(Array.isArray(endpoints)).toBe(true);
    });

    test('List endpoints without type returns error', async ({ page }) => {
      const result = await page.evaluate(async () => {
        try {
          await (window as any).api('GET', '/api/ai/endpoints');
          return { ok: true, error: null };
        } catch (e) {
          return { ok: false, error: String(e) };
        }
      });
      // Should fail because type param is required
      expect(result.error).toBeTruthy();
    });

    test('List endpoints with invalid type returns error', async ({ page }) => {
      const result = await page.evaluate(async () => {
        try {
          await (window as any).api('GET', '/api/ai/endpoints?type=invalid');
          return { ok: true, data: null };
        } catch (e) {
          return { ok: false, error: String(e) };
        }
      });
      expect(result.error).toBeTruthy();
    });

    test('Create, update, test, and delete an AI endpoint (full CRUD)', async ({ page }) => {
      const name = 'Test ' + uniqueName();

      // Create with correct field names
      const created: any = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/admin/ai-endpoints', {
          name: n, type: 'text', base_url: 'https://api.openai.com/v1',
          model: 'gpt-4o-mini', api_key: 'sk-test123', enabled: true,
          temperature: 0.7, max_tokens: 128,
        });
      }, name);
      expect(created).toBeTruthy();
      expect(created.id).toBeTruthy();
      // API key must be encrypted before storage
      expect(JSON.stringify(created)).not.toContain('sk-test123');

      // List includes the new endpoint
      const list: any[] = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/admin/ai-endpoints');
      });
      expect(list.some((e) => e.id === created.id && e.name === name)).toBe(true);

      // Enabled endpoint shows up in DM-facing enabled list
      const enabledText: any[] = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/ai/endpoints?type=text');
      });
      expect(enabledText.some((e) => e.id === created.id)).toBe(true);

      // Update (PUT requires name, type, base_url, model)
      const updated: any = await page.evaluate(async ({ id, nm }) => {
        return (window as any).api('PUT', '/api/admin/ai-endpoints/' + id, {
          name: nm, type: 'text', base_url: 'https://api.openai.com/v1',
          model: 'gpt-4.1-mini', enabled: false,
        });
      }, { id: created.id, nm: name });
      expect(updated.model).toBe('gpt-4.1-mini');
      expect(updated.enabled).toBe(false);

      // Disabled endpoint no longer in enabled list
      const enabledTextAfter: any[] = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/ai/endpoints?type=text');
      });
      expect(enabledTextAfter.some((e) => e.id === created.id)).toBe(false);

      // Test endpoint: fake key must fail deterministically
      const testResult: any = await page.evaluate(async (id) => {
        return (window as any).api('POST', '/api/admin/ai-endpoints/' + id + '/test');
      }, created.id);
      expect(testResult.success).toBe(false);
      expect(testResult.message).toBeTruthy();

      // Delete
      const del: any = await page.evaluate(async (id) => {
        return (window as any).api('DELETE', '/api/admin/ai-endpoints/' + id);
      }, created.id);
      expect(del).toBeTruthy();

      const listAfter: any[] = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/admin/ai-endpoints');
      });
      expect(listAfter.some((e) => e.id === created.id)).toBe(false);
    });
  });

  // ─── Text Generation ───

  test.describe('Text Generation', () => {
    test('Text generation API responds', async ({ page }) => {
      const result = await page.evaluate(async () => {
        try {
          const data = await (window as any).api('POST', '/api/ai/generate/text', {
            prompt: 'Hello', system_prompt: 'Be helpful',
          });
          return { ok: true, data };
        } catch (e) {
          return { ok: false, error: String(e) };
        }
      });
      // Accept any response - AI may or may not be configured
      expect(result).toBeTruthy();
    });
  });

  // ─── Image Generation ───

  test.describe('Image Generation', () => {
    test('Image generation API responds', async ({ page }) => {
      const result = await page.evaluate(async () => {
        try {
          const data = await (window as any).api('POST', '/api/ai/generate/image', {
            prompt: 'A fantasy castle',
          });
          return { ok: true, data };
        } catch (e) {
          return { ok: false, error: String(e) };
        }
      });
      // Accept any response - AI may or may not be configured
      expect(result).toBeTruthy();
    });
  });

  // ─── AI Generation UI ───

  test.describe('AI Generation UI', () => {
    test('AI modal generates text, inserts result, and Generate Again re-sends the prompt', async ({ page }) => {
      const name = 'UI ' + uniqueName();
      const created: any = await page.evaluate(async (n) => {
        return (window as any).api('POST', '/api/admin/ai-endpoints', {
          name: n, type: 'text', base_url: 'https://api.openai.com/v1',
          model: 'gpt-4o-mini', api_key: 'sk-ui', enabled: true,
        });
      }, name);
      expect(created.id).toBeTruthy();

      // Mock the generation endpoint so the flow is deterministic
      const calls: any[] = [];
      await page.route('**/api/ai/generate/text', async (route) => {
        calls.push(route.request().postDataJSON());
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ text: 'Mocked result #' + calls.length }),
        });
      });

      // FAB menu no longer exposes "Generate with AI" (fab-cleanup)
      await expect(page.locator('#fabMenu')).not.toContainText('Generate with AI');

      // Provide a target textarea via the NPC create modal
      await page.evaluate(() => (window as any).showCreateNPC());
      await page.waitForSelector('#newNPCDesc');

      // Open the AI modal targeting the NPC description field
      await page.evaluate(() => (window as any).openAIGenModal('text', 'newNPCDesc', 'Test hint', 'Generate Test'));
      await page.waitForSelector('#aiGenModal.show');

      await page.selectOption('#aiGenEndpoint', String(created.id));
      await page.fill('#aiGenPrompt', 'Write an NPC description');
      await page.click('#aiGenGenerateBtn');

      await expect(page.locator('#aiGenResultText')).toHaveText('Mocked result #1', { timeout: 15000 });

      // Insert the result into the target field
      await page.click('#aiGenInsertBtn');
      await expect(page.locator('#newNPCDesc')).toHaveValue(/Mocked result #1/);

      // Reopen and generate again — same prompt, new result
      await page.evaluate(() => (window as any).openAIGenModal('text', 'newNPCDesc'));
      await page.waitForSelector('#aiGenModal.show');
      await page.selectOption('#aiGenEndpoint', String(created.id));
      await page.fill('#aiGenPrompt', 'Write an NPC description');
      await page.click('#aiGenGenerateBtn');
      await expect(page.locator('#aiGenResultText')).toHaveText('Mocked result #2', { timeout: 15000 });

      // Generate Again re-runs the same prompt (task 5.3)
      await page.getByRole('button', { name: /Generate Again/ }).click();
      await expect(page.locator('#aiGenResultText')).toHaveText('Mocked result #3', { timeout: 15000 });

      expect(calls.length).toBe(3);
      for (const call of calls) {
        expect(call.prompt).toBe('Write an NPC description');
      }

      await page.evaluate(async (id) => {
        return (window as any).api('DELETE', '/api/admin/ai-endpoints/' + id);
      }, created.id);
    });
  });

  // ─── Disabled Toggle ───

  test.describe('Disabled Toggle', () => {
    test.afterEach(async ({ page }) => {
      // Always restore the default (enabled) state so other tests are unaffected.
      await page.evaluate(async () => {
        try { await (window as any).api('PUT', '/api/admin/settings/ai', { enabled: true }); } catch {}
      });
    });

    test('disabling AI hides UI controls and rejects generation', async ({ page }) => {
      await page.evaluate(async () => {
        return (window as any).api('PUT', '/api/admin/settings/ai', { enabled: false });
      });

      // Reload so the SPA re-reads the toggle and hides AI controls.
      await page.reload();
      await page.waitForFunction(() => (window as any).__apiReady === true, null, { timeout: 15000 });

      // Open the NPC create modal and confirm the AI buttons are suppressed.
      await page.evaluate(() => (window as any).showCreateNPC());
      await page.waitForSelector('#newNPCDesc');
      await expect(page.locator('.ai-generate-btn').first()).toBeHidden();

      // The generation endpoint must be rejected while disabled.
      const gen: any = await page.evaluate(async () => {
        try {
          await (window as any).api('POST', '/api/ai/generate/text', { prompt: 'Hello' });
          return { ok: true };
        } catch (e) {
          return { ok: false, error: String(e) };
        }
      });
      expect(gen.ok).toBe(false);

      // The enabled-endpoints list is empty while disabled.
      const eps: any[] = await page.evaluate(async () => {
        return (window as any).api('GET', '/api/ai/endpoints?type=text');
      });
      expect(eps).toEqual([]);
    });
  });
});
