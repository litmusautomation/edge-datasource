import { test, expect } from '@grafana/plugin-e2e';

test.describe('Config editor', () => {
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
