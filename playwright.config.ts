import { defineConfig, devices } from '@playwright/test';
import os from 'os';

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? os.cpus().length : Math.max(1, os.cpus().length - 2),
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:6270',
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
    command: 'rm -f /tmp/villum-test.db /tmp/villum-test.db-shm /tmp/villum-test.db-wal && DB_PATH=/tmp/villum-test.db go run -tags fts5 main.go',
    url: 'http://localhost:6270/api/check-setup',
    reuseExistingServer: false,
    timeout: 30000,
  },
});
