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
  const { options } = props;
  const { jsonData, secureJsonFields } = options;

  const [autocompleteEnabled, setAutocompleteEnabled] = useState(!!secureJsonFields?.apiToken);

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
        hasRequiredFields
      />

      <hr />

      <ConfigSection title="Connection">
        <Field
          label="Hostname"
          required
          description="Litmus Edge hostname or IP address. Append :port if it differs from the default (e.g., 192.168.1.100:8443)."
        >
          <Input
            width={WIDTH}
            name="hostname"
            placeholder="192.168.1.100"
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
