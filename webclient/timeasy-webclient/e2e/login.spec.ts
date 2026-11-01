import { test, expect, isMockMode } from './fixtures';

// These tests manage their own session; they must not reuse a stored one.
test.use({ storageState: { cookies: [], origins: [] } });

test.describe('Session', () => {
  test.skip(!isMockMode, 'mocked session is only relevant in mock mode');

  test('mocked login lands on the dashboard with a user profile', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();

    // profile is served by the mocked Keycloak /account endpoint
    await page.locator('app-header p-button button').filter({ has: page.locator('.pi-user') }).click();
    await expect(page.getByRole('button', { name: 'user1' })).toBeVisible();
  });
});

test.describe('Login / Logout (live Keycloak)', () => {
  test.skip(isMockMode, 'requires a running Keycloak (E2E_MODE=live)');

  async function login(page: import('@playwright/test').Page): Promise<void> {
    await page.goto('/dashboard');
    await page.waitForURL(/localhost:8180\/realms\/timeasy/, { timeout: 30_000 });
    await page.locator('#username').fill('user1');
    await page.locator('#password').fill('password');
    await page.getByRole('button', { name: 'Sign In' }).click();
    await page.waitForURL(/localhost:4200\/dashboard/, { timeout: 30_000 });
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
  }

  test('unauthenticated user is redirected to Keycloak and can log in', async ({ page }) => {
    await login(page);
  });

  test('login with wrong credentials shows an error', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForURL(/localhost:8180\/realms\/timeasy/, { timeout: 30_000 });
    await page.locator('#username').fill('user1');
    await page.locator('#password').fill('wrong-password');
    await page.getByRole('button', { name: 'Sign In' }).click();
    await expect(page.getByText(/Invalid username or password/i)).toBeVisible({ timeout: 10_000 });
  });

  test('logged-in user can log out via the user menu', async ({ page }) => {
    // log in fresh so the logout never touches the shared storage state
    await login(page);

    await page.locator('app-header p-button button').filter({ has: page.locator('.pi-user') }).click();
    const logoutButton = page.getByRole('button', { name: 'Logout' });
    await expect(logoutButton).toBeVisible({ timeout: 5_000 });
    await logoutButton.click();

    await page.waitForURL(/localhost:8180\/realms\/timeasy/, { timeout: 30_000 });
  });
});
