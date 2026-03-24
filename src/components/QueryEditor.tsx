import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { InlineFieldRow, InlineField, Combobox, Alert, type ComboboxOption } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { EdgeDataSourceOptions, EdgeQuery } from '../types';

type AutocompleteStatus = 'loading' | 'ready' | 'no_token' | 'unauthorized' | 'unreachable';

type Props = QueryEditorProps<DataSource, EdgeQuery, EdgeDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const topicError = getTopicError(query.topic || '');
  const [autocompleteStatus, setAutocompleteStatus] = useState<AutocompleteStatus>('loading');

  // Probe on mount to detect API token status.
  useEffect(() => {
    let cancelled = false;
    datasource
      .searchTopics('')
      .then((res) => {
        if (cancelled) {
          return;
        }
        if (res.error) {
          setAutocompleteStatus(res.error === 'unauthorized' ? 'unauthorized' : res.error === 'unreachable' ? 'unreachable' : 'no_token');
        } else {
          setAutocompleteStatus('ready');
        }
      })
      .catch(() => {
        if (!cancelled) {
          setAutocompleteStatus('unreachable');
        }
      });
    return () => {
      cancelled = true;
    };
  }, [datasource]);

  const onTopicSelect = useCallback(
    (option: ComboboxOption<string>) => {
      onChange({ ...query, topic: option.value });
      onRunQuery();
    },
    [onChange, onRunQuery, query]
  );

  // Debounced async loader for Combobox.
  const timeoutRef = useRef<ReturnType<typeof setTimeout>>();
  const loadOptions = useMemo(() => {
    if (autocompleteStatus !== 'ready') {
      return [];
    }

    return (inputValue: string): Promise<Array<ComboboxOption<string>>> => {
      clearTimeout(timeoutRef.current);

      if (inputValue.length < 2) {
        return Promise.resolve([]);
      }

      return new Promise((resolve) => {
        timeoutRef.current = setTimeout(async () => {
          try {
            const response = await datasource.searchTopics(inputValue);
            if (response.error) {
              setAutocompleteStatus(response.error === 'unauthorized' ? 'unauthorized' : 'unreachable');
              resolve([]);
              return;
            }
            resolve(response.topics.map((t) => ({ value: t, label: t })));
          } catch {
            resolve([]);
          }
        }, 300);
      });
    };
  }, [datasource, autocompleteStatus]);

  return (
    <>
      <InlineFieldRow>
        <InlineField
          label="Topic"
          labelWidth={8}
          grow
          error={topicError}
          invalid={!!topicError}
          interactive
          tooltip={<TopicTooltip />}
        >
          <Combobox<string>
            options={loadOptions}
            value={query.topic || null}
            onChange={onTopicSelect}
            createCustomValue
            placeholder='Type to search topics or enter manually'
          />
        </InlineField>
      </InlineFieldRow>

      {autocompleteStatus === 'no_token' && (
        <Alert title="" severity="info">
          Tip: add an API Token in datasource settings to enable topic autocomplete.
        </Alert>
      )}
      {autocompleteStatus === 'unauthorized' && (
        <Alert title="" severity="warning">
          Topic autocomplete unavailable — API token may be invalid or expired.
        </Alert>
      )}
      {autocompleteStatus === 'unreachable' && (
        <Alert title="" severity="warning">
          Topic autocomplete unavailable — could not reach the Edge API.
        </Alert>
      )}
    </>
  );
}

function getTopicError(subject: string): string {
  if (!subject || subject === '') {
    return 'Topic is required';
  }

  const tokens = subject.split('.');
  if (tokens.some((token) => token === '>' || token === '*')) {
    return 'Wildcards are not allowed';
  }

  const isValidToken = (token: string): boolean => /^[^\s.]+$/.test(token);
  if (!tokens.every(isValidToken)) {
    return `Invalid topic: [${subject}]`;
  }

  return '';
}

const TopicTooltip = () => (
  <div>
    <p>
      <b>Topic to subscribe to</b>
    </p>
    <p>
      See{' '}
      <b>
        <a target="_blank" rel="noreferrer" href="https://docs.litmus.io/litmusedge/product-features/devicehub/tags">
          Edge Topic
        </a>
      </b>{' '}
      for more information
    </p>
    <small>
      * <u>Topic</u> and <u>Tag</u> are used interchangeably in the documentation. In the context of this plugin, they
      are the same.
    </small>
  </div>
);
