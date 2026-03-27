import { test, expect } from '@grafana/plugin-e2e';

const remoteEdgeSwitchName = 'Connect to remote Litmus Edge';
const litmusEdgeAddressPlaceholder = '172.17.0.1 or 172.17.0.1:8443';
const accessAccountTokenPlaceholder = 'Access Account token';

async function setRemoteConnection(page: any, enabled: boolean) {
  const remoteConnectionSwitch = page.getByRole('switch', { name: remoteEdgeSwitchName });

  if (enabled) {
    await remoteConnectionSwitch.check({ force: true });
  } else {
    await remoteConnectionSwitch.uncheck({ force: true });
  }
}

test.describe('Config editor — connection mode', () => {
  test('defaults to inside-LE mode with no required fields', async ({ createDataSourceConfigPage, page }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await expect(page.getByText('Type: Litmus Edge', { exact: true })).toBeVisible();
    await expect(page.getByRole('switch', { name: remoteEdgeSwitchName })).toBeVisible();
    await expect(page.getByText('Remote Connection', { exact: true })).toBeVisible();

    // Remote-only fields are hidden by default
    await expect(page.getByPlaceholder(litmusEdgeAddressPlaceholder)).not.toBeVisible();
    await expect(page.getByPlaceholder(accessAccountTokenPlaceholder)).not.toBeVisible();

    // Topic discovery is always available
    await expect(page.getByText('Topic Discovery', { exact: true })).toBeVisible();
    await expect(page.getByRole('textbox', { name: 'API token' })).toBeVisible();
  });

  test('shows remote connection fields when enabled', async ({ createDataSourceConfigPage, page }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await setRemoteConnection(page, true);

    await expect(page.getByText('Litmus Edge Address *', { exact: true })).toBeVisible();
    await expect(page.getByText('NATS Proxy Port', { exact: true })).toBeVisible();
    await expect(page.getByText('Access Account Token *', { exact: true })).toBeVisible();
    await expect(page.getByPlaceholder(litmusEdgeAddressPlaceholder)).toBeVisible();
    await expect(page.getByPlaceholder('4222')).toBeVisible();
    await expect(page.getByPlaceholder(accessAccountTokenPlaceholder)).toBeVisible();
  });

  test('hides remote fields and clears address when disabled', async ({ createDataSourceConfigPage, page }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    // Toggle ON, fill address
    await setRemoteConnection(page, true);
    await page.getByPlaceholder(litmusEdgeAddressPlaceholder).fill('10.0.0.1');
    await expect(page.getByPlaceholder(litmusEdgeAddressPlaceholder)).toHaveValue('10.0.0.1');

    // Toggle OFF — fields hidden, address cleared
    await setRemoteConnection(page, false);
    await expect(page.getByPlaceholder(litmusEdgeAddressPlaceholder)).not.toBeVisible();
    await expect(page.getByPlaceholder(accessAccountTokenPlaceholder)).not.toBeVisible();

    // Toggle back ON — address should be empty (was cleared)
    await setRemoteConnection(page, true);
    await expect(page.getByPlaceholder(litmusEdgeAddressPlaceholder)).toHaveValue('');
  });

  test('inside-LE save & test fails in dev environment', async ({ createDataSourceConfigPage, page }) => {
    const configPage = await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    // Inside-LE mode (default) — gateway detection fails outside LE
    await configPage.saveAndTest();

    await expect(configPage).toHaveAlert('error');
  });

  test('save & test fails with empty remote connection fields', async ({ createDataSourceConfigPage, page }) => {
    const configPage = await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await setRemoteConnection(page, true);
    await configPage.saveAndTest();

    await expect(configPage).toHaveAlert('error');
  });

  test('save & test fails with invalid credentials', async ({ createDataSourceConfigPage, page }) => {
    const configPage = await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await setRemoteConnection(page, true);
    await page.getByPlaceholder(litmusEdgeAddressPlaceholder).fill('192.168.0.999');
    await page.getByPlaceholder(accessAccountTokenPlaceholder).fill('invalid-token');

    await configPage.saveAndTest();

    await expect(configPage).toHaveAlert('error');
  });
});

test.describe('Config editor — topic discovery', () => {
  test('autocomplete token field is visible in inside-LE mode', async ({ createDataSourceConfigPage, page }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await expect(page.getByText('Topic Discovery', { exact: true })).toBeVisible();
    await expect(page.getByRole('textbox', { name: 'API token' })).toBeVisible();
  });

  test('autocomplete token field stays visible in remote mode', async ({ createDataSourceConfigPage, page }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await setRemoteConnection(page, true);

    await expect(page.getByPlaceholder(litmusEdgeAddressPlaceholder)).toBeVisible();
    await expect(page.getByPlaceholder(accessAccountTokenPlaceholder)).toBeVisible();
    await expect(page.getByRole('textbox', { name: 'API token' })).toBeVisible();
  });

  test('autocomplete token field remains available while switching modes', async ({
    createDataSourceConfigPage,
    page,
  }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await expect(page.getByRole('textbox', { name: 'API token' })).toBeVisible();

    await setRemoteConnection(page, true);
    await expect(page.getByPlaceholder(litmusEdgeAddressPlaceholder)).toBeVisible();
    await expect(page.getByRole('textbox', { name: 'API token' })).toBeVisible();

    await setRemoteConnection(page, false);
    await expect(page.getByRole('textbox', { name: 'API token' })).toBeVisible();
    await expect(page.getByPlaceholder(litmusEdgeAddressPlaceholder)).not.toBeVisible();
  });

  test('remote mode shows remote fields together with autocomplete token', async ({
    createDataSourceConfigPage,
    page,
  }) => {
    await createDataSourceConfigPage({ type: 'litmus-edge-datasource' });

    await setRemoteConnection(page, true);

    await expect(page.getByPlaceholder(litmusEdgeAddressPlaceholder)).toBeVisible();
    await expect(page.getByPlaceholder('4222')).toBeVisible();
    await expect(page.getByPlaceholder(accessAccountTokenPlaceholder)).toBeVisible();
    await expect(page.getByRole('textbox', { name: 'API token' })).toBeVisible();
  });
});
