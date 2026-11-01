import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright E2E tests for the timeasy web client.
 *
 * Modes (set via E2E_MODE):
 *
 *   mock  (default) — API and Keycloak are intercepted in the browser context
 *                     (see e2e/support/). No backend, database, Keycloak or
 *                     Docker required. This is the mode used by CI.
 *
 *   live            — real Go API on http://localhost:8080 and real Keycloak
 *                     on http://localhost:8180 (realm "timeasy"). Useful for
 *                     local integration checks. Login/logout tests only run
 *                     in this mode.
 *
 * The Angular dev server is started automatically via webServer.
 *
 * Run:
 *   npm run e2e          # mock mode, desktop Chromium (fast local loop)
 *   npm run e2e:all      # mock mode, desktop + mobile (what CI runs)
 *   npm run e2e:mobile   # mock mode, mobile viewport only
 *   npm run e2e:live     # live mode against local API + Keycloak
 *   npm run e2e:headed   # mock mode with a visible browser
 *
 * Pass E2E_REUSE=1 to reuse an already running `ng serve`. Videos are off by
 * default; set PW_VIDEO=1 to record them for debugging.
 */
const PORT = Number(process.env.E2E_PORT ?? 4200);
const baseURL = `http://localhost:${PORT}`;
const isLiveMode = process.env.E2E_MODE === 'live';

export default defineConfig({
  testDir: './e2e',
  outputDir: './test-results',
  fullyParallel: false,
  forbidOnly: !process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  timeout: 60_000,
  expect: { timeout: 10_000 },
  reporter: process.env.CI ? [['github'], ['html', { open: 'never' }]] : [['html', { open: 'never' }]],

  use: {
    baseURL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    // Video recording is the most expensive part of a run; opt in with
    // PW_VIDEO=1 when debugging a flaky test.
    video: process.env.PW_VIDEO === '1' ? 'retain-on-failure' : 'off',
    locale: 'en-US',
  },

  projects: isLiveMode
    ? [
        // Setup project: logs in via Keycloak and stores the session state.
        { name: 'setup', testMatch: /auth\.setup\.ts/, use: { ...devices['Desktop Chrome'] } },
        {
          name: 'chromium',
          use: { ...devices['Desktop Chrome'], storageState: 'playwright/.auth/user.json' },
          dependencies: ['setup'],
        },
        {
          name: 'mobile-chrome',
          use: { ...devices['Pixel 7'], storageState: 'playwright/.auth/user.json' },
          dependencies: ['setup'],
        },
      ]
    : [
        // Mock mode: every test context installs its own auth/API mocks.
        {
          name: 'chromium',
          testIgnore: /auth\.setup\.ts/,
          use: { ...devices['Desktop Chrome'] },
        },
        {
          name: 'mobile-chrome',
          testIgnore: /auth\.setup\.ts/,
          use: { ...devices['Pixel 7'] },
        },
      ],

  webServer: [
    {
      command: 'npm run start -- --port ' + PORT + (process.env.E2E_REUSE ? '' : ' --no-live-reload'),
      url: baseURL,
      reuseExistingServer: !!process.env.E2E_REUSE,
      timeout: 180_000,
    },
  ],
});
