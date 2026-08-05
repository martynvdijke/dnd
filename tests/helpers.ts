import type { Page } from '@playwright/test';

export const LOGIN_TIMEOUT = 30000;
export const NAV_TIMEOUT = 10000;

export async function ensureNavOpen(page: Page) {
  const toggler = page.locator('.navbar-toggler');
  if (await toggler.isVisible()) {
    await toggler.click();
    await page.waitForTimeout(300);
  }
}

export async function waitLoadingDone(page: Page, timeout: number = 15000) {
  // First ensure the SPA has initialized and the API module is available
  await page.waitForFunction(() => typeof (window as any).api !== 'undefined', { timeout });
  // Then wait for the loading overlay to disappear
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
    await page.locator(`#appSidebar button[data-nav="${nav}"]`).click();
  }
}

export async function openMoreNav(page: Page) {
  if (await isMobile(page)) {
    await page.click('#moreTabBtn');
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
    await page.locator(`#appSidebar button[data-nav="${nav}"]`).click();
  }
}

export async function clickBottomTabForce(page: Page, nav: string) {
  await page.locator(`.bottom-tab-item[data-nav="${nav}"]`).click({ force: true });
  await page.waitForTimeout(300);
}
