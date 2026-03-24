import { test, expect } from '@grafana/plugin-e2e';

test.describe('Litmus Edge datasource', () => {
  test('config page renders with expected fields', async ({ createDataSourceConfigPage, page }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await expect(page.getByText('Type: Litmus Edge', { exact: true })).toBeVisible();
    await expect(page.getByText('Hostname *', { exact: true })).toBeVisible();
    await expect(page.getByText('Token *', { exact: true })).toBeVisible();
  });
});
