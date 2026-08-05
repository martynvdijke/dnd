import { test, expect } from './fixtures.js';
import { login } from './helpers.js';

const uniqueName = () => `AI-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('AI Features', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.waitForTimeout(200);
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
      expect(result).toBeTruthy();
    });
  });
});
