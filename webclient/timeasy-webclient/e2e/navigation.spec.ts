import { test, expect, type Page } from './fixtures';

const menuRoutes: Array<{ label: string; url: RegExp; heading: RegExp }> = [
  { label: 'Projects', url: /\/projects$/, heading: /Projects/ },
  { label: 'Time Entries', url: /\/timeentries$/, heading: /Time Entries/ },
  { label: 'Weekly Overview', url: /\/statistics\/weeklyoverview$/, heading: /Weekly Overview/ },
  { label: 'Dashboard', url: /\/dashboard$/, heading: /Dashboard/ },
];

/**
 * On desktop the sidebar menu is wrapped in `nav.menu`; on mobile the same
 * menu lives in the `.mobile-nav` overlay which is revealed by the burger
 * button. The off-screen mobile overlay is still "visible" to Playwright,
 * so we must pick the right container for the current viewport.
 */
async function navigateViaMenu(page: Page, label: string, expectedUrl: RegExp, expectedHeading: RegExp): Promise<void> {
  const viewportWidth = page.viewportSize()?.width ?? 1280;
  const isMobile = viewportWidth < 768;

  let item;
  if (isMobile) {
    const mobileNav = page.locator('.mobile-nav');
    // The closed overlay is translated off-screen and still counts as
    // "visible" for Playwright, so we check the `open` class instead.
    const isOpen = (await mobileNav.getAttribute('class'))?.includes('open') ?? false;
    if (!isOpen) {
      await page.locator('.burger-menu').click();
    }
    item = mobileNav.getByRole('menuitem', { name: label });
  } else {
    item = page.locator('nav.menu').getByRole('menuitem', { name: label });
  }

  await item.click();
  await expect(page).toHaveURL(expectedUrl, { timeout: 15_000 });
  await expect(page.getByRole('heading', { name: expectedHeading })).toBeVisible();
}

test.describe('Navigation via menu', () => {
  for (const { label, url, heading } of menuRoutes) {
    test(`menu navigates to ${label}`, async ({ page }) => {
      await page.goto('/dashboard');
      await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();

      await navigateViaMenu(page, label, url, heading);
    });
  }
});
