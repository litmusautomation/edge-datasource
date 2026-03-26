import React, { useState } from 'react';
import { Field, Input, SecretInput, Switch } from '@grafana/ui';
import {
  DataSourcePluginOptionsEditorProps,
  onUpdateDatasourceJsonDataOption,
  onUpdateDatasourceSecureJsonDataOption,
  updateDatasourcePluginResetOption,
} from '@grafana/data';
import { DataSourceDescription } from '@grafana/plugin-ui';
import { EdgeDataSourceOptions, EdgeSecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<EdgeDataSourceOptions, EdgeSecureJsonData> {}

const ACCESS_ACCOUNT_DOCS =
  'https://docs.litmus.io/litmusedge/product-features/system/access-control/tokens/create-api-account';
const API_TOKEN_DOCS =
  'https://docs.litmus.io/litmusedge/product-features/system/access-control/tokens/create-api-token';

const WIDTH = 40;

export function ConfigEditor(props: Props) {
  const { options, onOptionsChange } = props;
  const { jsonData, secureJsonFields } = options;

  const externalEdge = !!jsonData.externalEdge;
  const [autocompleteEnabled, setAutocompleteEnabled] = useState(!!secureJsonFields?.apiToken);

  const onToggleExternalEdge = (e: React.ChangeEvent<HTMLInputElement>) => {
    const enabled = e.currentTarget.checked;
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        externalEdge: enabled,
        hostname: enabled ? jsonData.hostname : '',
      },
    });
  };

  const onToggleAutocomplete = (e: React.ChangeEvent<HTMLInputElement>) => {
    const enabled = e.currentTarget.checked;
    setAutocompleteEnabled(enabled);
    if (!enabled) {
      updateDatasourcePluginResetOption(props, 'apiToken');
    }
  };

  return (
    <>
      <DataSourceDescription
        dataSourceName="Litmus Edge"
        docsLink="https://github.com/litmusautomation/edge-datasource"
        hasRequiredFields={externalEdge}
      />

      <hr />

      <section>
        <div style={{ display: 'inline-flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
          <h4 style={{ margin: 0, fontSize: '16px', lineHeight: 1.2, color: 'var(--text-secondary)', fontWeight: 500 }}>
            Connection
          </h4>
          <Switch value={externalEdge} onChange={onToggleExternalEdge} aria-label="External Litmus Edge" />
        </div>
        <p
          style={{
            margin: '0 0 12px 0',
            color: 'var(--text-secondary)',
            opacity: 0.85,
            fontSize: '13px',
            lineHeight: 1.3,
          }}
        >
          Connect to a remote Litmus Edge instance.
        </p>

        {externalEdge && (
          <>
            <Field
              label="Hostname"
              required
              description={
                <>
                  Enter the Litmus Edge host or IP (for example: 172.17.0.1).
                  <br />
                  Include a port when needed (for example: 172.17.0.1:8443).
                </>
              }
            >
              <Input
                width={WIDTH}
                name="hostname"
                placeholder="172.17.0.1"
                value={jsonData.hostname || ''}
                onChange={onUpdateDatasourceJsonDataOption(props, 'hostname')}
              />
            </Field>
            <Field
              label="Access Account Token"
              required
              description={
                <>
                  Token used to access the NATS Proxy.{' '}
                  <a href={ACCESS_ACCOUNT_DOCS} target="_blank" rel="noreferrer">
                    Learn more
                  </a>
                </>
              }
            >
              <SecretInput
                width={WIDTH}
                placeholder="Access Account token"
                isConfigured={!!secureJsonFields?.token}
                onReset={() => updateDatasourcePluginResetOption(props, 'token')}
                onBlur={onUpdateDatasourceSecureJsonDataOption(props, 'token')}
              />
            </Field>
          </>
        )}
      </section>

      <hr />

      <section>
        <div style={{ display: 'inline-flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
          <h4 style={{ margin: 0, fontSize: '16px', lineHeight: 1.2, color: 'var(--text-secondary)', fontWeight: 500 }}>
            Autocomplete
          </h4>
          <Switch value={autocompleteEnabled} onChange={onToggleAutocomplete} aria-label="Enable topic autocomplete" />
        </div>
        <p
          style={{
            margin: '0 0 12px 0',
            color: 'var(--text-secondary)',
            opacity: 0.85,
            fontSize: '13px',
            lineHeight: 1.3,
          }}
        >
          Suggest topics as you type in the query editor to help you find the right data stream faster.
        </p>

        {autocompleteEnabled && (
          <Field
            label="API Token"
            description={
              <>
                Create an API token in Access Control &gt; Tokens.{' '}
                <a href={API_TOKEN_DOCS} target="_blank" rel="noreferrer">
                  Learn more
                </a>
              </>
            }
          >
            <SecretInput
              width={WIDTH}
              placeholder="API token"
              isConfigured={!!secureJsonFields?.apiToken}
              onReset={() => updateDatasourcePluginResetOption(props, 'apiToken')}
              onBlur={onUpdateDatasourceSecureJsonDataOption(props, 'apiToken')}
            />
          </Field>
        )}
      </section>
    </>
  );
}
