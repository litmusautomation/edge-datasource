import React from 'react';
import { Field, Input, SecretInput } from '@grafana/ui';
import {
  DataSourcePluginOptionsEditorProps,
  onUpdateDatasourceJsonDataOption,
  onUpdateDatasourceSecureJsonDataOption,
  updateDatasourcePluginResetOption,
} from '@grafana/data';
import { ConfigSection, DataSourceDescription } from '@grafana/plugin-ui';
import { EdgeDataSourceOptions, EdgeSecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<EdgeDataSourceOptions, EdgeSecureJsonData> {}

const WIDTH = 40;

export function ConfigEditor(props: Props) {
  const { options } = props;
  const { jsonData, secureJsonFields } = options;

  return (
    <>
      <DataSourceDescription
        dataSourceName="Litmus Edge"
        docsLink="https://docs.litmus.io/litmusedge/"
        hasRequiredFields
      />

      <hr />

      <ConfigSection title="Connection">
        <Field label="Hostname" required description="Litmus Edge hostname or IP address">
          <Input
            width={WIDTH}
            name="hostname"
            placeholder="127.0.0.1"
            value={jsonData.hostname || ''}
            onChange={onUpdateDatasourceJsonDataOption(props, 'hostname')}
          />
        </Field>
      </ConfigSection>

      <hr />

      <ConfigSection title="Authentication">
        <Field label="Token" required description="API token with NATS proxy read access">
          <SecretInput
            width={WIDTH}
            placeholder="auth token"
            isConfigured={!!secureJsonFields?.token}
            onReset={() => updateDatasourcePluginResetOption(props, 'token')}
            onBlur={onUpdateDatasourceSecureJsonDataOption(props, 'token')}
          />
        </Field>
      </ConfigSection>
    </>
  );
}
