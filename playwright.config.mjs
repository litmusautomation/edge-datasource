import path from 'node:path';
import { createRequire } from 'node:module';
import { defineConfig, devices } from '@playwright/test';

const require = createRequire(import.meta.url);
const pluginE2eAuth = path.join(path.dirname(require.resolve('@grafana/plugin-e2e')), 'auth');

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [['html', { open: 'never' }], ['list']],
  use: {
    baseURL: process.env.GRAFANA_URL || 'http://localhost:3000',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'auth',
      testDir: pluginE2eAuth,
      testMatch: [/.*\.js/],
    },
    {
      name: 'run-tests',
      use: {
        ...devices['Desktop Chrome'],
        storageState: 'playwright/.auth/admin.json',
      },
      dependencies: ['auth'],
    },
  ],
});
