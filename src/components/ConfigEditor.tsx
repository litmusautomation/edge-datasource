import React, { ChangeEvent } from 'react';
import { FieldSet, InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { EdgeDataSourceOptions, EdgeSecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<EdgeDataSourceOptions> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  // Secure field (only sent to the backend)
  const onTokenChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        token: event.target.value,
      },
    });
  };

  const onResetToken = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        token: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        token: '',
      },
    });
  };

  const { jsonData, secureJsonFields } = options;
  const secureJsonData = (options.secureJsonData || {}) as EdgeSecureJsonData;

  return (
    <div className="gf-form-group">
      <FieldSet label="Authentication">
        <InlineField
          label="Hostname"
          labelWidth={13}
          interactive
          tooltip={() => {
            return <p>Edge hostname to connect to</p>;
          }}
        >
          <Input
            value={jsonData.hostname || ''}
            placeholder='e.g. "127.0.0.1"'
            onReset={onResetToken}
            onChange={onTokenChange}
          />
        </InlineField>
        <InlineField
          label="Token"
          labelWidth={13}
          interactive
          tooltip={() => {
            return (
              <>
                <p>
                  See{' '}
                  <b>
                    <a
                      target="_blank"
                      rel="noreferrer"
                      href="https://docs.litmus.io/litmusedge/product-features/system/tokens/create-api-account"
                    >
                      Tokens
                    </a>
                  </b>{' '}
                  for more information
                </p>
              </>
            );
          }}
        >
          <SecretInput
            isConfigured={(secureJsonFields && secureJsonFields.token) as boolean}
            value={secureJsonData.token || ''}
            placeholder="auth token"
            onReset={onResetToken}
            onChange={onTokenChange}
          />
        </InlineField>
      </FieldSet>
    </div>
  );
}
