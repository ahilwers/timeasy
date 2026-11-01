import { test as setup, expect } from '@playwright/test';
import fs from 'fs';

const authFile = 'playwright/.auth/user.json';
const API_URL = 'http://localhost:8080/api/v1';
const KEYCLOAK_TOKEN_URL = 'http://localhost:8180/realms/timeasy/protocol/openid-connect/token';

/**
 * Performs a full Keycloak login once, persists the session state and cleans
 * up leftover e2e test data via the backend API so every run starts clean.
 */
setup('authenticate', async ({ page, request }) => {
  fs.mkdirSync('playwright/.auth', { recursive: true });

  // The auth guard kicks in on the dashboard and redirects to Keycloak.
  await page.goto('/dashboard');
  await expect(page).toHaveURL(/localhost:8180\/realms\/timeasy/);

  await page.locator('#username').fill('user1');
  await page.locator('#password').fill('password');
  await page.getByRole('button', { name: 'Sign In' }).click();

  // Back in the app, on the dashboard.
  await page.waitForURL(/localhost:4200\/dashboard/, { timeout: 30_000 });
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();

  // Save the authenticated state.
  await page.context().storageState({ path: authFile });

  // Clean up leftovers from previous runs (projects/entries named e2e-*).
  try {
    const tokenResponse = await request.post(KEYCLOAK_TOKEN_URL, {
      form: {
        grant_type: 'password',
        client_id: 'timeasy-webclient',
        username: 'user1',
        password: 'password',
      },
    });
    const { access_token } = await tokenResponse.json();
    const api = await request.newContext({
      baseURL: API_URL,
      extraHTTPHeaders: { Authorization: `Bearer ${access_token}` },
    });

    // stop/DELETE a leftover open (running) timer entry
    const openEntry: { id?: string } = await api.get('/timeentries/lastopen').json().catch(() => ({}));
    if (openEntry?.id) {
      await api.delete(`/timeentries/${openEntry.id}`);
    }

    const projects: { id: string; name?: string }[] = await api.get('/projects').json().catch(() => []);
    for (const project of projects) {
      if (project.name?.startsWith('e2e-')) await api.delete(`/projects/${project.id}`);
    }

    const entries: { id: string; description?: string }[] = await api.get('/timeentries').json().catch(() => []);
    for (const entry of entries) {
      if (entry.description?.startsWith('e2e-')) await api.delete(`/timeentries/${entry.id}`);
    }
  } catch (error) {
    console.warn('E2E data cleanup failed (continuing anyway):', error);
  }
});
