import { test, expect } from '@grafana/plugin-e2e';

test.describe('smoke', () => {
  test('home loads', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('Welcome to Grafana')).toBeVisible();
  });
});
