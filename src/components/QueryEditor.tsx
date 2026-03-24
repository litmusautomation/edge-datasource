import React, { useCallback } from 'react';
import { InlineFieldRow, InlineField, Input } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { EdgeDataSourceOptions, EdgeQuery } from '../types';

type Props = QueryEditorProps<DataSource, EdgeQuery, EdgeDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const topicError = getTopicError(query.topic || '');

  const onTopicChange = useCallback(
    (event: React.FormEvent<HTMLInputElement>) => {
      onChange({ ...query, topic: event.currentTarget.value });
    },
    [onChange, query]
  );

  return (
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
        <Input
          name="topic"
          required
          placeholder='e.g. "enterprise.site.area.line.machine.sensor"'
          value={query.topic}
          onBlur={onRunQuery}
          onChange={onTopicChange}
        />
      </InlineField>
    </InlineFieldRow>
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
