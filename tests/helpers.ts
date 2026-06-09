import { Page } from '@playwright/test';

const desktopNavMap: Record<string, string> = {
  'Characters': 'characters',
  'Dice': 'dice',
  'Party': 'party',
  'Compendium': 'compendium',
  'Encounters': 'encounters',
  'Timeline': 'timeline',
  'One-Shots': 'oneshots',
  'Combat': 'combat',
  'Factions': 'factions',
  'Shops': 'shops',
  'Admin': 'admin',
};

export async function ensureNavOpen(page: Page) {
  const toggler = page.locator('.navbar-toggler');
  if (await toggler.isVisible()) {
    await toggler.click();
    await page.waitForTimeout(300);
  }
}

export async function waitLoadingDone(page: Page) {
  await page.waitForFunction(() => {
    const o = document.getElementById('loadingOverlay');
    return o && o.classList.contains('d-none');
  }, { timeout: 5000 }).catch(() => {});
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

export async function clickNavItem(page: Page, text: string, bottomNav?: string) {
  if (await isMobile(page)) {
    if (bottomNav) {
      await page.click(`#bottomTabBar button[data-nav="${bottomNav}"]`);
    }
  } else {
    const nav = desktopNavMap[text];
    if (nav) {
      await page.locator(`#appSidebar button[data-nav="${nav}"]`).click();
    }
  }
}

export async function openMoreNav(page: Page) {
  if (await isMobile(page)) {
    await page.click('#moreTabBtn');
    await page.waitForTimeout(300);
  }
}

export async function clickSecondaryNavItem(page: Page, desktopText: string, moreId: string) {
  if (await isMobile(page)) {
    await openMoreNav(page);
    await page.evaluate((id) => {
      const btn = document.getElementById(id);
      if (btn) btn.click();
    }, moreId);
    await page.waitForTimeout(300);
  } else {
    const nav = desktopNavMap[desktopText];
    if (nav) {
      await page.locator(`#appSidebar button[data-nav="${nav}"]`).click();
    }
  }
}

export async function clickBottomTabForce(page: Page, nav: string) {
  await page.locator(`.bottom-tab-item[data-nav="${nav}"]`).click({ force: true });
  await page.waitForTimeout(300);
}
