import React, { useSyncExternalStore } from 'react';
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
const DEFAULT_EDGE_DOCKER_GATEWAY_IP = '10.30.50.1';

const WIDTH = 40;
/** SecretInput is input (~width×8px) + Reset; use a smaller width on narrow viewports. */
const SECRET_INPUT_WIDTH_NARROW = 28;

const MQ_STACK_CONNECTION = '(max-width: 768px)';
const MQ_COMPACT_SECRET = '(max-width: 480px)';

const CONNECTION_FIELD_CELL_STYLE: React.CSSProperties = {
  flex: '0 0 auto',
  maxWidth: '100%',
  minWidth: 0,
};

/** Single subscription for both breakpoints (primitive snapshot for useSyncExternalStore). */
function useViewportLayout() {
  const bits = useSyncExternalStore(
    (onStoreChange) => {
      const mqStack = window.matchMedia(MQ_STACK_CONNECTION);
      const mqCompact = window.matchMedia(MQ_COMPACT_SECRET);
      const onChange = () => onStoreChange();
      mqStack.addEventListener('change', onChange);
      mqCompact.addEventListener('change', onChange);
      return () => {
        mqStack.removeEventListener('change', onChange);
        mqCompact.removeEventListener('change', onChange);
      };
    },
    () => {
      const mqStack = window.matchMedia(MQ_STACK_CONNECTION);
      const mqCompact = window.matchMedia(MQ_COMPACT_SECRET);
      return (mqStack.matches ? 1 : 0) | (mqCompact.matches ? 2 : 0);
    },
    () => 0
  );
  return {
    stackConnectionFields: (bits & 1) !== 0,
    compactSecretInput: (bits & 2) !== 0,
  };
}

export function ConfigEditor(props: Props) {
  const { options, onOptionsChange } = props;
  const { jsonData, secureJsonFields } = options;

  const externalEdge = !!jsonData.externalEdge;
  const { stackConnectionFields, compactSecretInput } = useViewportLayout();
  const secretInputWidth = compactSecretInput ? SECRET_INPUT_WIDTH_NARROW : WIDTH;

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
            Remote Connection
          </h4>
          <Switch value={externalEdge} onChange={onToggleExternalEdge} aria-label="Connect to remote Litmus Edge" />
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

        {!externalEdge && (
          <Field
            label="Edge Docker Gateway IP"
            description={
              <>
                Used when Grafana runs inside Litmus Edge. Update it if this instance uses a different gateway IP.
                Verify it with Save & test.
              </>
            }
          >
            <Input
              width={WIDTH}
              name="gatewayIp"
              placeholder={DEFAULT_EDGE_DOCKER_GATEWAY_IP}
              value={jsonData.gatewayIp || DEFAULT_EDGE_DOCKER_GATEWAY_IP}
              onChange={onUpdateDatasourceJsonDataOption(props, 'gatewayIp')}
            />
          </Field>
        )}

        {externalEdge && (
          <>
            <Field
              label="Litmus Edge Address"
              required
              description={
                <>
                  Address of your Litmus Edge instance. Use <code>host</code> or <code>host:port</code>.
                </>
              }
            >
              <Input
                width={WIDTH}
                name="hostname"
                placeholder="172.17.0.1 or 172.17.0.1:8443"
                value={jsonData.hostname || ''}
                onChange={onUpdateDatasourceJsonDataOption(props, 'hostname')}
              />
            </Field>
            <div
              style={{
                display: 'flex',
                flexDirection: stackConnectionFields ? 'column' : 'row',
                gap: '12px',
                alignItems: 'flex-start',
              }}
            >
              <div style={CONNECTION_FIELD_CELL_STYLE}>
                <Field
                  label="Access Account API Key"
                  required
                  description={
                    <>
                      API key used to access the NATS Proxy.{' '}
                      <a href={ACCESS_ACCOUNT_DOCS} target="_blank" rel="noreferrer">
                        Learn more
                      </a>
                    </>
                  }
                >
                  <SecretInput
                    width={secretInputWidth}
                    placeholder="Access Account API key"
                    isConfigured={!!secureJsonFields?.token}
                    onReset={() => updateDatasourcePluginResetOption(props, 'token')}
                    onBlur={onUpdateDatasourceSecureJsonDataOption(props, 'token')}
                  />
                </Field>
              </div>
              <div style={CONNECTION_FIELD_CELL_STYLE}>
                <Field label="NATS Proxy Port" description="Port for live data streaming. Default: 4222.">
                  <Input
                    width={WIDTH}
                    name="natsProxyPort"
                    placeholder="4222"
                    value={jsonData.natsProxyPort || ''}
                    onChange={onUpdateDatasourceJsonDataOption(props, 'natsProxyPort')}
                  />
                </Field>
              </div>
            </div>
          </>
        )}
      </section>

      <hr />

      <section>
        <h4
          style={{
            margin: '0 0 4px 0',
            fontSize: '16px',
            lineHeight: 1.2,
            color: 'var(--text-secondary)',
            fontWeight: 500,
          }}
        >
          Topic Discovery
        </h4>
        <p
          style={{
            margin: '0 0 12px 0',
            color: 'var(--text-secondary)',
            opacity: 0.85,
            fontSize: '13px',
            lineHeight: 1.3,
          }}
        >
          Add an API Token to search available topics as you type in the query editor.
        </p>

        <Field
          label="API Token"
          description={
            <>
              Optional, but recommended for topic discovery.{' '}
              <a href={API_TOKEN_DOCS} target="_blank" rel="noreferrer">
                Learn more
              </a>
            </>
          }
        >
          <SecretInput
            width={secretInputWidth}
            placeholder="API token"
            isConfigured={!!secureJsonFields?.apiToken}
            onReset={() => updateDatasourcePluginResetOption(props, 'apiToken')}
            onBlur={onUpdateDatasourceSecureJsonDataOption(props, 'apiToken')}
          />
        </Field>
      </section>
    </>
  );
}
