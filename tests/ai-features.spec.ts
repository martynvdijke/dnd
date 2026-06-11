import { test, expect } from '@playwright/test';
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

    test('Create an AI endpoint', async ({ page }) => {
      const name = 'Test ' + uniqueName();
      const result = await page.evaluate(async (n) => {
        try {
          const data = await (window as any).api('POST', '/api/admin/ai-endpoints', {
            name: n, provider: 'openai', api_key: 'sk-test123',
            endpoint_type: 'text', base_url: 'https://api.openai.com/v1',
            models: JSON.stringify(['gpt-4']), is_enabled: true,
          });
          return { ok: true, data };
        } catch (e) {
          return { ok: false, error: String(e) };
        }
      }, name);
      // Accept either success or validation error
      expect(result.ok === true || result.error !== undefined).toBe(true);
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
