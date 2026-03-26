import { test, expect } from '@grafana/plugin-e2e';

test.describe('Config editor — connection mode', () => {
  test('defaults to inside-LE mode with no required fields', async ({ createDataSourceConfigPage, page }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await expect(page.getByText('Type: Litmus Edge', { exact: true })).toBeVisible();
    await expect(page.getByLabel('External Litmus Edge')).toBeVisible();

    // Hostname and token fields are hidden in inside-LE mode
    await expect(page.getByPlaceholder('172.17.0.1')).not.toBeVisible();
    await expect(page.getByPlaceholder('Access Account token')).not.toBeVisible();
  });

  test('shows hostname and token when External is toggled on', async ({ createDataSourceConfigPage, page }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await page.getByLabel('External Litmus Edge').click();

    await expect(page.getByText('Hostname *', { exact: true })).toBeVisible();
    await expect(page.getByText('Access Account Token *', { exact: true })).toBeVisible();
    await expect(page.getByPlaceholder('172.17.0.1')).toBeVisible();
    await expect(page.getByPlaceholder('Access Account token')).toBeVisible();
  });

  test('hides fields and clears hostname when External is toggled off', async ({
    createDataSourceConfigPage,
    page,
  }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    // Toggle ON, fill hostname
    await page.getByLabel('External Litmus Edge').click();
    await page.getByPlaceholder('172.17.0.1').fill('10.0.0.1');
    await expect(page.getByPlaceholder('172.17.0.1')).toHaveValue('10.0.0.1');

    // Toggle OFF — fields hidden, hostname cleared
    await page.getByLabel('External Litmus Edge').click();
    await expect(page.getByPlaceholder('172.17.0.1')).not.toBeVisible();
    await expect(page.getByPlaceholder('Access Account token')).not.toBeVisible();

    // Toggle back ON — hostname should be empty (was cleared)
    await page.getByLabel('External Litmus Edge').click();
    await expect(page.getByPlaceholder('172.17.0.1')).toHaveValue('');
  });

  test('inside-LE save & test fails in dev environment', async ({ createDataSourceConfigPage, page }) => {
    const configPage = await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    // Inside-LE mode (default) — gateway detection fails outside LE
    await configPage.saveAndTest();

    await expect(configPage).toHaveAlert('error');
  });

  test('save & test fails with empty external fields', async ({ createDataSourceConfigPage, page }) => {
    const configPage = await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await page.getByLabel('External Litmus Edge').click();
    await configPage.saveAndTest();

    await expect(configPage).toHaveAlert('error');
  });

  test('save & test fails with invalid credentials', async ({ createDataSourceConfigPage, page }) => {
    const configPage = await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await page.getByLabel('External Litmus Edge').click();
    await page.getByPlaceholder('172.17.0.1').fill('192.168.0.999');
    await page.getByPlaceholder('Access Account token').fill('invalid-token');

    await configPage.saveAndTest();

    await expect(configPage).toHaveAlert('error');
  });
});

test.describe('Config editor — topic autocomplete', () => {
  test('autocomplete toggle is visible in inside-LE mode', async ({ createDataSourceConfigPage, page }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await expect(page.getByLabel('Enable topic autocomplete')).toBeVisible();
    // API token hidden by default
    await expect(page.getByPlaceholder('API token')).not.toBeVisible();
  });

  test('autocomplete toggle shows API token field', async ({ createDataSourceConfigPage, page }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await page.getByLabel('Enable topic autocomplete').click();

    await expect(page.getByText('API Token', { exact: true })).toBeVisible();
    await expect(page.getByPlaceholder('API token')).toBeVisible();
  });

  test('autocomplete toggle hides API token field when turned off', async ({
    createDataSourceConfigPage,
    page,
  }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    // Toggle ON
    await page.getByLabel('Enable topic autocomplete').click();
    await expect(page.getByPlaceholder('API token')).toBeVisible();

    // Toggle OFF
    await page.getByLabel('Enable topic autocomplete').click();
    await expect(page.getByPlaceholder('API token')).not.toBeVisible();
  });

  test('autocomplete toggle works in external mode', async ({ createDataSourceConfigPage, page }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    // Switch to external mode
    await page.getByLabel('External Litmus Edge').click();
    await expect(page.getByPlaceholder('172.17.0.1')).toBeVisible();

    // Autocomplete toggle still works
    await page.getByLabel('Enable topic autocomplete').click();
    await expect(page.getByPlaceholder('API token')).toBeVisible();

    // Both external fields and autocomplete field visible together
    await expect(page.getByPlaceholder('172.17.0.1')).toBeVisible();
    await expect(page.getByPlaceholder('Access Account token')).toBeVisible();
    await expect(page.getByPlaceholder('API token')).toBeVisible();
  });
});
