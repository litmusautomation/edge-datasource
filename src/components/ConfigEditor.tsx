import React, { useState } from 'react';
import { Field, Input, SecretInput, Switch } from '@grafana/ui';
import {
  DataSourcePluginOptionsEditorProps,
  onUpdateDatasourceJsonDataOption,
  onUpdateDatasourceSecureJsonDataOption,
  updateDatasourcePluginResetOption,
} from '@grafana/data';
import { ConfigSection, DataSourceDescription } from '@grafana/plugin-ui';
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

      <ConfigSection title="Connection">
        <Field description="Enable to connect to a remote Litmus Edge instance.">
          <Switch label="External Litmus Edge" value={externalEdge} onChange={onToggleExternalEdge} />
        </Field>

        {externalEdge && (
          <>
            <Field
              label="Hostname"
              required
              description={
                <>
                  Hostname or IP, e.g. <code>172.17.0.1</code>.
                  <br />
                  If your Edge URL includes a port, use the same form — e.g. <code>172.17.0.1:8443</code>.
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
                  Token for accessing the NATS Proxy.{' '}
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
      </ConfigSection>

      <hr />

      <ConfigSection
        title="Topic Autocomplete"
        description={
          <>
            Suggests matching topics as you type in the query editor.
            <br />
            Makes it easier to discover and select the right data stream.
          </>
        }
      >
        <Field>
          <Switch label="Enable topic autocomplete" value={autocompleteEnabled} onChange={onToggleAutocomplete} />
        </Field>

        {autocompleteEnabled && (
          <Field
            label="API Token"
            description={
              <>
                Create an API Token under Access Control &gt; Tokens.{' '}
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
      </ConfigSection>
    </>
  );
}
