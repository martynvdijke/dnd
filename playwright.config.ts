import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: 2,
  workers: 1,
  timeout: 90000,
  expect: {
    timeout: 15000,
    toHaveScreenshot: { maxDiffPixelRatio: 0.05 },
  },
  snapshotPathTemplate: 'tests/visual-refs/{arg}{ext}',
  reporter: 'html',
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:6270',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      testIgnore: /setup\.spec\.ts/,
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      testIgnore: /setup\.spec\.ts/,
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'mobile-chrome',
      testIgnore: /setup\.spec\.ts/,
      use: { ...devices['Pixel 5'] },
    },
  ],
});
