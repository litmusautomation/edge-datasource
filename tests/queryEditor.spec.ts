import { test, expect } from '@grafana/plugin-e2e';

test.describe('Query editor', () => {
  test('renders topic field in explore', async ({ explorePage, page }) => {
    await explorePage.datasource.set('Litmus Edge');

    await expect(page.getByText('Topic', { exact: true })).toBeVisible();
    await expect(page.getByPlaceholder('e.g. "enterprise.site.area.line.machine.sensor"')).toBeVisible();
  });

  test('shows validation error for empty topic', async ({ explorePage, page }) => {
    await explorePage.datasource.set('Litmus Edge');

    // Topic field starts empty — validation error should be visible.
    await expect(page.getByText('Topic is required')).toBeVisible();
  });

  test('shows validation error for wildcard topic', async ({ explorePage, page }) => {
    await explorePage.datasource.set('Litmus Edge');

    const topicInput = page.getByPlaceholder('e.g. "enterprise.site.area.line.machine.sensor"');
    await topicInput.fill('device.*.temperature');

    await expect(page.getByText('Wildcards are not allowed')).toBeVisible();
  });

  test('shows validation error for invalid topic tokens', async ({ explorePage, page }) => {
    await explorePage.datasource.set('Litmus Edge');

    const topicInput = page.getByPlaceholder('e.g. "enterprise.site.area.line.machine.sensor"');
    await topicInput.fill('device..temperature');

    await expect(page.getByText(/Invalid topic:/)).toBeVisible();
  });

  test('clears validation error for valid topic', async ({ explorePage, page }) => {
    await explorePage.datasource.set('Litmus Edge');

    const topicInput = page.getByPlaceholder('e.g. "enterprise.site.area.line.machine.sensor"');
    await topicInput.fill('enterprise.site.area.line.cell.tag');

    await expect(page.getByText('Topic is required')).not.toBeVisible();
    await expect(page.getByText('Wildcards are not allowed')).not.toBeVisible();
  });
});
