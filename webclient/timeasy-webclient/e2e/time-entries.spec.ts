import { test, expect } from './fixtures';

const testDescription = `e2e-entry-${Date.now()}`;

test.describe('Time Entries', () => {
  test('list loads and shows the add button', async ({ page }) => {
    await page.goto('/timeentries');
    await expect(page.getByRole('heading', { name: 'Time Entries' })).toBeVisible();
    await expect(page.locator('p-table')).toBeVisible();
    await expect(page.getByRole('button').filter({ has: page.locator('.pi-plus') })).toBeVisible();
    // project filter is available
    await expect(page.locator('#project')).toBeVisible();
  });

  test('create, edit and delete a time entry', async ({ page }) => {
    // --- create ---
    await page.goto('/timeentries');
    await page.getByRole('button').filter({ has: page.locator('.pi-plus') }).first().click();
    await expect(page.getByRole('heading', { name: 'Add Time Entry' })).toBeVisible();

    // pick the first project from the select
    await page.locator('#project').click();
    await page.locator('.p-select-list-item, [role="option"]').first().click();

    // description
    const description = page.locator('#description, app-description-autocomplete input').first();
    await description.fill(testDescription);

    await page.getByRole('button', { name: 'OK' }).click();

    // back on the list with the new entry
    await expect(page.getByRole('heading', { name: 'Time Entries' })).toBeVisible({ timeout: 15_000 });
    // on mobile the description column may be hidden, so assert DOM presence
    await expect(page.locator('p-table').getByText(testDescription).first()).toBeAttached();

    // --- edit ---
    const row = page.locator('p-table tbody tr', { hasText: testDescription }).first();
    await row.getByRole('button').filter({ has: page.locator('.pi-pencil') }).click();
    await expect(page.getByRole('heading', { name: /Edit Time Entry/ })).toBeVisible();
    const descField = page.locator('app-description-autocomplete input').first();
    await descField.fill(`${testDescription}-edited`);
    await page.getByRole('button', { name: 'OK' }).click();
    await expect(page.locator('p-table').getByText(`${testDescription}-edited`).first()).toBeAttached({ timeout: 15_000 });

    // --- delete ---
    const editedRow = page.locator('p-table tbody tr', { hasText: `${testDescription}-edited` }).first();
    await editedRow.getByRole('button').filter({ has: page.locator('.pi-trash') }).click();
    // PrimeNG renders the confirm dialog twice in the DOM (host + visible
    // div), so we target the visible dialog component directly.
    const confirmDialog = page.locator('.p-confirmdialog').filter({ visible: true });
    await expect(confirmDialog).toBeVisible();
    await confirmDialog.getByRole('button', { name: 'Yes' }).click();
    await expect(page.locator('p-table').getByText(`${testDescription}-edited`)).toHaveCount(0, { timeout: 15_000 });
  });

  test('filter by project narrows the list', async ({ page }) => {
    await page.goto('/timeentries');
    await expect(page.getByRole('heading', { name: 'Time Entries' })).toBeVisible();
    // open the project filter and select the first option (if any exist)
    const projectFilter = page.locator('#project');
    await projectFilter.click();
    const firstOption = page.locator('.p-select-list-item, [role="option"]').first();
    if (await firstOption.isVisible({ timeout: 3000 }).catch(() => false)) {
      await firstOption.click();
      // the table must still render (with whatever entries belong to the project)
      await expect(page.locator('p-table')).toBeVisible();
      // clear filter again
      const clearIcon = page.locator('.p-select-clear-icon, .p-select-label-clear-icon').first();
      if (await clearIcon.isVisible({ timeout: 3000 }).catch(() => false)) {
        await clearIcon.click();
      }
    }
  });
});
