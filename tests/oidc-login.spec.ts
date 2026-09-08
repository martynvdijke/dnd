import { test, expect } from './fixtures.js';

// OIDC is not configured in test servers (no OIDC_* env): password login is
// the only path and the SSO button must stay hidden.
test.describe('OIDC login (disabled)', () => {
  test('login page hides the SSO button when OIDC is disabled', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByTestId('login-submit')).toBeVisible();
    await expect(page.getByTestId('oidc-login')).toBeHidden();
  });

  test('OIDC status reports disabled and login redirects are absent', async ({ page }) => {
    const status = await page.request.get('/api/auth/oidc/status');
    expect(status.ok()).toBe(true);
    expect(await status.json()).toEqual({ enabled: false });

    const login = await page.request.get('/api/auth/oidc/login');
    expect(login.status()).toBe(404);
  });
});
