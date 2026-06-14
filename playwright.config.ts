import { defineConfig, devices } from '@playwright/test';
import os from 'os';

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: 1,
  workers: process.env.CI ? Math.max(1, Math.floor(os.cpus().length / 2)) : Math.max(1, Math.floor(os.cpus().length / 4)),
  timeout: 45000,
  expect: {
    timeout: 10000,
  },
  reporter: 'html',
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:6270',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'setup',
      testMatch: /setup\.spec\.ts/,
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'chromium',
      testIgnore: /setup\.spec\.ts/,
      use: { ...devices['Desktop Chrome'] },
      dependencies: ['setup'],
    },
    {
      name: 'firefox',
      testIgnore: /setup\.spec\.ts/,
      use: { ...devices['Desktop Firefox'] },
      dependencies: ['setup'],
    },
    {
      name: 'mobile-chrome',
      testIgnore: /setup\.spec\.ts/,
      use: { ...devices['Pixel 5'] },
      dependencies: ['setup'],
    },
  ],
  webServer: {
    command: 'rm -f /tmp/villum-test.db /tmp/villum-test.db-shm /tmp/villum-test.db-wal && if [ -f ./villum-server ]; then DB_PATH=/tmp/villum-test.db ./villum-server; else DB_PATH=/tmp/villum-test.db go run -tags sqlite_fts5 .; fi',
    url: 'http://localhost:6270/api/check-setup',
    reuseExistingServer: false,
    timeout: 30000,
  },
});
