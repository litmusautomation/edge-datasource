import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { InlineFieldRow, InlineField, Combobox, Text, Icon, type ComboboxOption } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { EdgeDataSourceOptions, EdgeQuery } from '../types';

type AutocompleteStatus = 'loading' | 'ready' | 'no_token' | 'unauthorized' | 'unreachable';

type Props = QueryEditorProps<DataSource, EdgeQuery, EdgeDataSourceOptions>;

const EMPTY_OPTIONS: Array<ComboboxOption<string>> = [];

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const [touched, setTouched] = useState(!!query.topic);
  const topicError = useMemo(() => (touched ? getTopicError(query.topic || '') : ''), [touched, query.topic]);
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
          setAutocompleteStatus(
            res.error === 'unauthorized' ? 'unauthorized' : res.error === 'unreachable' ? 'unreachable' : 'no_token'
          );
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

  // Use a ref so the callback always sees the latest query without needing it as a dep.
  const queryRef = useRef(query);
  useEffect(() => {
    queryRef.current = query;
  });

  const onTopicSelect = useCallback(
    (option: ComboboxOption<string>) => {
      setTouched(true);
      onChange({ ...queryRef.current, topic: option.value });
      onRunQuery();
    },
    [onChange, onRunQuery]
  );

  const requestIdRef = useRef(0);
  const timeoutRef = useRef<ReturnType<typeof setTimeout>>();
  const pendingResolveRef = useRef<((v: Array<ComboboxOption<string>>) => void) | null>(null);

  // Clean up on unmount: cancel debounce and invalidate in-flight requests.
  useEffect(() => {
    const refs = { timeout: timeoutRef, requestId: requestIdRef, pendingResolve: pendingResolveRef };
    return () => {
      clearTimeout(refs.timeout.current);
      refs.requestId.current++;
      if (refs.pendingResolve.current) {
        refs.pendingResolve.current([]);
        refs.pendingResolve.current = null;
      }
    };
  }, []);

  const loadOptions = useMemo(() => {
    if (autocompleteStatus !== 'ready') {
      return [];
    }

    return (inputValue: string): Promise<Array<ComboboxOption<string>>> => {
      // Resolve the previous Promise so Combobox never waits on an orphan.
      if (pendingResolveRef.current) {
        pendingResolveRef.current([]);
        pendingResolveRef.current = null;
      }
      clearTimeout(timeoutRef.current);

      if (inputValue.length < 2) {
        return Promise.resolve(EMPTY_OPTIONS);
      }

      const currentRequestId = ++requestIdRef.current;

      return new Promise((resolve) => {
        pendingResolveRef.current = resolve;

        timeoutRef.current = setTimeout(async () => {
          // Timer fired — this resolve is no longer "pending a debounce".
          pendingResolveRef.current = null;

          try {
            const response = await datasource.searchTopics(inputValue);

            // Discard stale response if a newer request was fired.
            if (currentRequestId !== requestIdRef.current) {
              resolve([]);
              return;
            }

            if (response.error) {
              // Only 'unauthorized' is permanent — disable autocomplete.
              // 'unreachable' is transient — keep alive for next keystroke.
              if (response.error === 'unauthorized') {
                setAutocompleteStatus('unauthorized');
              }
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
            placeholder="e.g. devicehub.alias.demo.sensor_name"
          />
        </InlineField>
      </InlineFieldRow>

      <AutocompleteHint status={autocompleteStatus} />
    </>
  );
}

function AutocompleteHint({ status }: { status: AutocompleteStatus }) {
  if (status === 'no_token') {
    return (
      <Text element="p" variant="bodySmall" color="secondary" italic>
        <Icon name="info-circle" size="sm" /> Tip: add an API Token in datasource settings to enable topic
        autocomplete.
      </Text>
    );
  }
  if (status === 'unauthorized') {
    return (
      <Text element="p" variant="bodySmall" color="warning">
        <Icon name="exclamation-triangle" size="sm" /> Topic autocomplete unavailable — API token may be invalid or
        expired.
      </Text>
    );
  }
  if (status === 'unreachable') {
    return (
      <Text element="p" variant="bodySmall" color="warning">
        <Icon name="exclamation-triangle" size="sm" /> Topic autocomplete unavailable — could not reach the Edge API.
      </Text>
    );
  }
  return null;
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
      <b>NATS topic to subscribe to</b>
    </p>
    <p>
      The dot-separated subject published by Litmus Edge, e.g.{' '}
      <code>devicehub.alias.demo.sensor_name</code>
    </p>
    <p>
      <a target="_blank" rel="noreferrer" href="https://docs.litmus.io/litmusedge/product-features/devicehub/tags">
        View documentation
      </a>
    </p>
  </div>
);
