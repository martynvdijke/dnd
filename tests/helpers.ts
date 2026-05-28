import { Page } from '@playwright/test';

export async function ensureNavOpen(page: Page) {
  const navbar = page.locator('.navbar');
  if (await navbar.isVisible()) {
    const toggler = page.locator('.navbar-toggler');
    if (await toggler.isVisible()) {
      await toggler.click();
      await page.waitForTimeout(300);
    }
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
  const navbar = page.locator('.navbar');
  return !(await navbar.isVisible());
}

export async function clickNavItem(page: Page, text: string, bottomNav?: string) {
  if (await isMobile(page)) {
    if (bottomNav) {
      await page.click(`#bottomTabBar button[data-nav="${bottomNav}"]`);
    }
  } else {
    await page.click(`a:has-text("${text}")`);
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
    await ensureNavOpen(page);
    await page.click(`nav a:has-text("${desktopText}")`);
  }
}

export async function clickBottomTabForce(page: Page, nav: string) {
  await page.locator(`.bottom-tab-item[data-nav="${nav}"]`).click({ force: true });
  await page.waitForTimeout(300);
}
