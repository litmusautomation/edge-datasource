import { test, expect } from '@grafana/plugin-e2e';
import type { Page } from '@playwright/test';

async function getTopicInput(page: Page) {
  const topicInput = page.locator('[role="combobox"], input[name="topic"]').first();
  await expect(topicInput).toBeVisible({ timeout: 15000 });
  return topicInput;
}

test.describe('Query editor', () => {
  test('renders topic field in explore', async ({ explorePage, page }) => {
    await explorePage.datasource.set('Litmus Edge');

    await expect(page.getByText('Topic', { exact: true })).toBeVisible();

    const topicInput = await getTopicInput(page);
    await expect(topicInput).toBeVisible();
  });

  test('shows validation error for empty topic', async ({ explorePage, page }) => {
    await explorePage.datasource.set('Litmus Edge');

    // Topic field starts empty — validation error should be visible.
    await expect(page.getByText('Topic is required')).toBeVisible();
  });

  test('shows validation error for wildcard topic', async ({ explorePage, page }) => {
    await explorePage.datasource.set('Litmus Edge');

    const topicInput = await getTopicInput(page);
    await topicInput.fill('device.*.temperature');

    await expect(page.getByText('Wildcards are not allowed')).toBeVisible();
  });

  test('shows validation error for invalid topic tokens', async ({ explorePage, page }) => {
    await explorePage.datasource.set('Litmus Edge');

    const topicInput = await getTopicInput(page);
    await topicInput.fill('device..temperature');

    await expect(page.getByText(/Invalid topic:/)).toBeVisible();
  });

  test('clears validation error for valid topic', async ({ explorePage, page }) => {
    await explorePage.datasource.set('Litmus Edge');

    const topicInput = await getTopicInput(page);
    await topicInput.fill('enterprise.site.area.line.cell.tag');

    await expect(page.getByText('Topic is required')).not.toBeVisible();
    await expect(page.getByText('Wildcards are not allowed')).not.toBeVisible();
  });

  test('keeps value while editing selected topic', async ({ explorePage, page }) => {
    await explorePage.datasource.set('Litmus Edge');

    const full = 'devicehub.alias.P1_L1_Machine3_1_S7.Temperature';
    const partial = 'devicehub.alias.P1_L1_Machine3_1_S7.Temperat';

    const topicInput = await getTopicInput(page);
    await topicInput.fill(full);
    await expect(page.getByText('Topic is required')).not.toBeVisible();

    await topicInput.click({ force: true });
    await expect(page.getByText('Topic is required')).not.toBeVisible();

    await topicInput.fill(partial);
    await expect(topicInput).toHaveValue(partial);
  });

  test('autocomplete dropdown opens above the topic input', async ({ explorePage, page }) => {
    await explorePage.datasource.set('Litmus Edge');

    const combo = page.locator('[role="combobox"]').first();
    if ((await combo.count()) === 0) {
      test.skip(true, 'Autocomplete combobox is not available (API token likely not configured)');
    }

    await combo.click();
    await combo.fill('te');

    const firstOption = page.getByRole('option').first();
    await expect(firstOption).toBeVisible({ timeout: 15000 });

    const inputBox = await combo.boundingBox();
    const optionBox = await firstOption.boundingBox();

    expect(inputBox).not.toBeNull();
    expect(optionBox).not.toBeNull();

    if (inputBox && optionBox) {
      expect(optionBox.y + optionBox.height).toBeLessThanOrEqual(inputBox.y + 1);
    }
  });
});
