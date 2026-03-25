import { test, expect } from '@grafana/plugin-e2e';

test.describe('Config editor', () => {
  test('renders all expected fields', async ({ createDataSourceConfigPage, page }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await expect(page.getByText('Type: Litmus Edge', { exact: true })).toBeVisible();
    await expect(page.getByText('Hostname *', { exact: true })).toBeVisible();
    await expect(page.getByText('Access Account Token *', { exact: true })).toBeVisible();
  });

  test('save & test fails with empty fields', async ({ createDataSourceConfigPage, page }) => {
    const configPage = await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await configPage.saveAndTest();

    await expect(configPage).toHaveAlert('error');
  });

  test('save & test fails with invalid credentials', async ({ createDataSourceConfigPage, page }) => {
    const configPage = await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await page.getByPlaceholder('172.17.0.1').fill('192.168.0.999');
    await page.getByPlaceholder('Access Account token').fill('invalid-token');

    await configPage.saveAndTest();

    await expect(configPage).toHaveAlert('error');
  });
});
