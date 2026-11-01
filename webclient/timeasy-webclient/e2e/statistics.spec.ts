import { test, expect } from './fixtures';

test.describe('Statistics', () => {
  test('weekly overview loads', async ({ page }) => {
    await page.goto('/statistics/weeklyoverview');
    // the heading contains the calendar week, e.g. "Weekly Overview 44/2026"
    await expect(page.getByRole('heading', { name: /Weekly Overview/ })).toBeVisible();
    // the overview table (one row per weekday) must be rendered
    await expect(page.getByText('Monday')).toBeVisible();
    await expect(page.getByText('Sunday')).toBeVisible();
  });
});
