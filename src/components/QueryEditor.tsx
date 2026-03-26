import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { InlineFieldRow, InlineField, Input, Text, Icon, AsyncSelect } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { EdgeDataSourceOptions, EdgeQuery } from '../types';
import { getTopicError } from '../topicValidation';

type AutocompleteStatus = 'loading' | 'ready' | 'no_token' | 'unauthorized' | 'unreachable';
type SearchHint = 'none' | 'short' | 'nomatch';

type Props = QueryEditorProps<DataSource, EdgeQuery, EdgeDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const topicError = useMemo(() => getTopicError(query.topic || ''), [query.topic]);
  const [autocompleteStatus, setAutocompleteStatus] = useState<AutocompleteStatus>('loading');
  const [searchHint, setSearchHint] = useState<SearchHint>('none');
  const [probeKey, setProbeKey] = useState(0);

  const queryRef = useRef(query);
  useEffect(() => {
    queryRef.current = query;
  });

  const requestIdRef = useRef(0);
  const cachedTopicsRef = useRef<string[]>([]);

  useEffect(() => {
    cachedTopicsRef.current = [];
  }, [datasource, probeKey]);

  // Probe to detect API token status. Runs on mount and on retry.
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
  }, [datasource, probeKey]);

  const retryProbe = useCallback(() => {
    setAutocompleteStatus('loading');
    setProbeKey((k) => k + 1);
  }, []);

  const onInputChange = useCallback(
    (e: React.FormEvent<HTMLInputElement>) => {
      onChange({ ...queryRef.current, topic: e.currentTarget.value });
    },
    [onChange]
  );

  const onSelectInputChange = useCallback(
    (value: string, actionMeta: { action: string }) => {
      if (actionMeta.action === 'input-change') {
        onChange({ ...queryRef.current, topic: value });
      }
    },
    [onChange]
  );

  const onTopicSelect = useCallback(
    (item: SelectableValue<string>) => {
      const nextTopic = item.value ?? '';
      onChange({ ...queryRef.current, topic: nextTopic });
      if (!getTopicError(nextTopic)) {
        onRunQuery();
      }
    },
    [onChange, onRunQuery]
  );

  const onInputBlur = useCallback(() => {
    if (!getTopicError(queryRef.current.topic || '')) {
      onRunQuery();
    }
  }, [onRunQuery]);

  const loadTopicOptions = useCallback(
    async (inputValue: string): Promise<Array<SelectableValue<string>>> => {
      if (inputValue.length < 2) {
        setSearchHint(inputValue.length === 0 ? 'none' : 'short');
        requestIdRef.current++;
        return [];
      }

      const currentRequestId = ++requestIdRef.current;
      try {
        const response = await datasource.searchTopics(inputValue);

        if (currentRequestId !== requestIdRef.current) {
          return [];
        }

        if (response.error) {
          if (response.error === 'unauthorized') {
            setAutocompleteStatus('unauthorized');
          } else if (response.error === 'unreachable') {
            setAutocompleteStatus('unreachable');
          }
          setSearchHint('none');
          return [];
        }

        if (response.topics.length > 0) {
          const seen = new Set(cachedTopicsRef.current);
          for (const topic of response.topics) {
            if (!seen.has(topic)) {
              seen.add(topic);
              cachedTopicsRef.current.push(topic);
            }
          }
          if (cachedTopicsRef.current.length > 200) {
            cachedTopicsRef.current = cachedTopicsRef.current.slice(-200);
          }
          setSearchHint('none');
          return response.topics.map((topic) => ({ value: topic, label: topic }));
        }

        const needle = inputValue.toLowerCase();
        const localCandidates = new Set<string>(cachedTopicsRef.current);
        if (queryRef.current.topic) {
          localCandidates.add(queryRef.current.topic);
        }
        const localMatches = Array.from(localCandidates).filter((topic) => topic.toLowerCase().includes(needle));
        if (localMatches.length > 0) {
          setSearchHint('none');
          return localMatches.slice(0, 15).map((topic) => ({ value: topic, label: topic }));
        }

        setSearchHint('nomatch');
        return [];
      } catch {
        if (currentRequestId !== requestIdRef.current) {
          return [];
        }
        setAutocompleteStatus('unreachable');
        setSearchHint('none');
        return [];
      }
    },
    [datasource]
  );

  const isAutocompleteReady = autocompleteStatus === 'ready';
  const noOptionsMessage =
    searchHint === 'short'
      ? 'Start typing to see topics from Edge'
      : searchHint === 'nomatch'
        ? 'No topics match your search'
        : 'Start typing to see topics from Edge';

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
          {isAutocompleteReady ? (
            <AsyncSelect<string>
              placeholder="Search for a topic..."
              value={query.topic ? { value: query.topic, label: query.topic } : null}
              loadOptions={loadTopicOptions}
              onInputChange={onSelectInputChange}
              onChange={onTopicSelect}
              onBlur={onInputBlur}
              menuPlacement="top"
              noOptionsMessage={noOptionsMessage}
              defaultOptions={false}
              allowCustomValue={false}
              isClearable={false}
              openMenuOnFocus={false}
              closeMenuOnSelect
              blurInputOnSelect={false}
              maxMenuHeight={200}
            />
          ) : (
            <Input
              name="topic"
              required
              value={query.topic ?? ''}
              onChange={onInputChange}
              onBlur={onInputBlur}
              placeholder="e.g. devicehub.alias.demo.sensor_name"
              loading={autocompleteStatus === 'loading'}
            />
          )}
        </InlineField>
      </InlineFieldRow>

      <AutocompleteHint status={autocompleteStatus} onRetry={retryProbe} />
    </>
  );
}

function AutocompleteHint({ status, onRetry }: { status: AutocompleteStatus; onRetry: () => void }) {
  if (status === 'no_token') {
    return (
      <Text element="p" variant="bodySmall" color="secondary">
        <Icon name="info-circle" size="sm" /> Add an API token in datasource settings to enable autocomplete.
      </Text>
    );
  }
  if (status === 'unauthorized') {
    return (
      <Text element="p" variant="bodySmall" color="warning">
        <Icon name="exclamation-triangle" size="sm" /> Autocomplete unavailable: API token is invalid or expired.
      </Text>
    );
  }
  if (status === 'unreachable') {
    return (
      <Text element="p" variant="bodySmall" color="warning">
        <Icon name="exclamation-triangle" size="sm" /> Autocomplete unavailable: could not reach Edge API.{' '}
        <button type="button" onClick={onRetry}>
          Retry
        </button>
      </Text>
    );
  }
  return null;
}

const TopicTooltip = () => (
  <div>
    <p>
      <b>NATS topic to subscribe to</b>
    </p>
    <p>
      The dot-separated subject published by Litmus Edge, e.g. <code>devicehub.alias.demo.sensor_name</code>
    </p>
    <p>
      <a target="_blank" rel="noreferrer" href="https://docs.litmus.io/litmusedge/product-features/devicehub/tags">
        View documentation
      </a>
    </p>
  </div>
);
