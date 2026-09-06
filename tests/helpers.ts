import type { Page } from '@playwright/test';

export const LOGIN_TIMEOUT = 60000;
export const NAV_TIMEOUT = 15000;

export async function ensureNavOpen(page: Page) {
  const toggler = page.locator('.navbar-toggler');
  if (await toggler.isVisible()) {
    await toggler.click();
    await page.waitForTimeout(300);
  }
}

export async function waitLoadingDone(page: Page, timeout: number = 30000) {
  // First ensure the SPA has initialized and the API module is available
  await page.waitForFunction(() => typeof (window as any).api !== 'undefined', { timeout });
  // Then wait for the API token to be provisioned (authoritative signal set by
  // init() after ensureApiToken completes). The loading overlay alone is
  // unreliable because of its 200ms debounce.
  await page.waitForFunction(() => (window as any).__apiReady === true, { timeout });
  // Finally wait for the loading overlay to disappear
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout });
}

/**
 * Log in as admin user and wait for the SPA to fully initialize.
 * Prefer this over inline login to avoid test flakiness on slower runtimes (mobile-chrome in CI).
 */
export async function login(page: Page, timeout: number = LOGIN_TIMEOUT) {
  await page.goto('/login', { waitUntil: 'domcontentloaded' });
  await page.fill('#username', 'admin');
  await page.fill('#password', 'testpassword123');
  // Use Promise.all to avoid race between click and navigation listener
  await Promise.all([
    page.waitForURL('/', { timeout, waitUntil: 'domcontentloaded' }),
    page.getByTestId('login-submit').click(),
  ]);
  await waitLoadingDone(page, timeout);
}

/**
 * Create (or reuse) an API token for the logged-in session and return its
 * secret. Mutating API routes require a bearer API token in addition to the
 * session cookie, so direct `page.request` mutations must attach it.
 */
export async function getApiToken(page: Page): Promise<string> {
  const csrfResp = await page.request.get('/api/csrf-token');
  const csrfData = await csrfResp.json();
  const csrf = csrfData.token;

  const createResp = await page.request.post('/api/tokens', {
    data: { name: 'e2e-test' },
    headers: { 'X-CSRF-Token': csrf },
  });
  if (createResp.status() !== 201) {
    throw new Error(`failed to create API token: ${createResp.status()}`);
  }
  const data = await createResp.json();
  return data.token;
}

export async function waitModalClosed(page: Page) {
  await page.waitForFunction(() => {
    const modal = document.getElementById('genericModal');
    return !modal || !modal.classList.contains('show');
  }, { timeout: 10000 }).catch(() => {});
}

export async function isMobile(page: Page): Promise<boolean> {
  const viewport = page.viewportSize();
  // Use viewport width as a deterministic check (<768px = Bootstrap's md breakpoint)
  // Falls back to navbar-toggler visibility for runtime DOM checks
  if (viewport) return viewport.width < 768;
  const toggler = page.locator('.navbar-toggler');
  return await toggler.isVisible().catch(() => false);
}

export async function clickNavItem(page: Page, nav: string, bottomNav?: string) {
  if (await isMobile(page)) {
    if (bottomNav) {
      await page.locator(`#bottomTabBar button[data-nav="${bottomNav}"]`).click({ force: true });
    }
  } else {
    await page.locator(`#appSidebar [data-testid="nav-${nav}"]`).click();
  }
}

export async function openMoreNav(page: Page) {
  if (await isMobile(page)) {
    await page.locator('[data-testid="nav-more"]').click();
    await page.waitForTimeout(300);
  }
}

export async function clickSecondaryNavItem(page: Page, nav: string, moreId: string, label?: string) {
  if (await isMobile(page)) {
    await openMoreNav(page);
    // Bottom sheet buttons are dynamically created without IDs, click by text
    await page.locator('#bottom-sheet-more-nav button').filter({ hasText: label ?? nav }).click();
    // Wait for bottom sheet close animation (onclick calls closeBottomSheet)
    await page.waitForTimeout(500);
  } else {
    await page.locator(`#appSidebar [data-testid="nav-${nav}"]`).click();
  }
}

export async function clickBottomTabForce(page: Page, nav: string) {
  await page.locator(`.bottom-tab-item[data-nav="${nav}"]`).click({ force: true });
  await page.waitForTimeout(300);
}
