import { test, expect } from './fixtures';

const testProjectName = `e2e-test-project-${Date.now()}`;

test.describe('Projects', () => {
  test('list shows existing projects', async ({ page }) => {
    await page.goto('/projects');
    await expect(page.getByRole('heading', { name: 'Projects' })).toBeVisible();
    await expect(page.locator('p-table')).toBeVisible();
    // add button in the header
    await expect(page.getByRole('button').filter({ has: page.locator('.pi-plus') })).toBeVisible();
  });

  test('create, edit and delete a project', async ({ page }) => {
    // --- create ---
    await page.goto('/projects');
    await page.getByRole('button').filter({ has: page.locator('.pi-plus') }).first().click();
    await expect(page.getByRole('heading', { name: 'Add Project' })).toBeVisible();

    await page.locator('#name').fill(testProjectName);
    // a color must be chosen for the form to be valid
    await page.locator('#color').click();
    await page.locator('.p-select-list-item, [role="option"]').first().click();
    await page.getByRole('button', { name: 'OK' }).click();

    // back on the list with the new project
    await expect(page.getByRole('heading', { name: 'Projects' })).toBeVisible({ timeout: 15_000 });
    await expect(page.locator('p-table').getByText(testProjectName)).toBeAttached();

    // --- edit ---
    const row = page.locator('p-table tbody tr', { hasText: `${testProjectName}-edited` }).first()
      .or(page.locator('p-table tbody tr', { hasText: testProjectName }).first());
    await row.getByRole('button').filter({ has: page.locator('.pi-pencil') }).click();
    await expect(page.getByRole('heading', { name: /Edit Project/ })).toBeVisible();
    await page.locator('#name').fill(`${testProjectName}-edited`);
    await page.getByRole('button', { name: 'OK' }).click();
    await expect(page.locator('p-table').getByText(`${testProjectName}-edited`)).toBeAttached({ timeout: 15_000 });

    // --- delete ---
    const editedRow = page.locator('p-table tbody tr', { hasText: `${testProjectName}-edited` }).first();
    await editedRow.getByRole('button').filter({ has: page.locator('.pi-trash') }).click();
    // PrimeNG renders the confirm dialog twice in the DOM (host + visible
    // div), so we target the visible dialog component directly.
    const confirmDialog = page.locator('.p-confirmdialog').filter({ visible: true });
    await expect(confirmDialog).toBeVisible();
    await confirmDialog.getByRole('button', { name: 'Yes' }).click();
    await expect(page.locator('p-table').getByText(`${testProjectName}-edited`)).toHaveCount(0, { timeout: 15_000 });
  });
});
