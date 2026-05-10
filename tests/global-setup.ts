import { chromium } from '@playwright/test';

async function globalSetup() {
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();

  const baseURL = process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:6270';

  await page.goto(`${baseURL}/setup`);

  const body = await page.locator('body').textContent();
  if (body?.includes('First-Time')) {
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.fill('#confirm', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { timeout: 10000 });
  }

  await browser.close();
}

export default globalSetup;
