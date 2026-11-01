import { test, expect, type Page } from './fixtures';

const playButton = (page: Page) =>
  page.getByRole('button').filter({ has: page.locator('.pi-play') }).first();
const pauseButton = (page: Page) =>
  page.getByRole('button').filter({ has: page.locator('.pi-pause') }).first();

/**
 * The time tracking widget keeps its state on the backend (last open time
 * entry). If a timer is still running, the button shows "pause". Stop it
 * first to get a clean state. The widget may briefly re-enter the running
 * state after stopping, so we retry a few times.
 */
async function ensureTimerStopped(page: Page): Promise<void> {
  for (let i = 0; i < 5; i++) {
    if (await playButton(page).isVisible({ timeout: 1_000 }).catch(() => false)) {
      return;
    }
    if (await pauseButton(page).isVisible({ timeout: 1_000 }).catch(() => false)) {
      await pauseButton(page).click();
      await page.waitForTimeout(500);
    }
  }
  await expect(playButton(page)).toBeVisible({ timeout: 5_000 });
}

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
    await ensureTimerStopped(page);
  });

  test('shows the headline and the time tracking widget', async ({ page }) => {
    // PrimeNG select for project choice must be present
    await expect(page.locator('#project')).toBeVisible();
    // play button is visible; it may be enabled or disabled depending on
    // whether the widget auto-selected the last used project
    await expect(playButton(page)).toBeVisible();
  });

  test('starts and stops the timer', async ({ page }) => {
    // select a project explicitly
    await page.locator('#project').click();
    await page.locator('.p-select-list-item, [role="option"]').first().click();

    await expect(playButton(page)).toBeEnabled();
    await playButton(page).click();
    await expect(pauseButton(page)).toBeVisible({ timeout: 10_000 });

    // stop the timer again so the backend is left in a clean state
    await ensureTimerStopped(page);
    await expect(playButton(page)).toBeVisible();
  });

  test('shows weekly statistics and pie chart cards', async ({ page }) => {
    await expect(page.getByText(/Your Week/i)).toBeVisible();
    await expect(page.getByText(/Projects This Week/i)).toBeVisible();
  });
});
