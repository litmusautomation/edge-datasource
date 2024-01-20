import React, { ChangeEvent, useCallback } from 'react';
import { InlineField, Input } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { EdgeDataSourceOptions, EdgeQuery } from '../types';

type Props = QueryEditorProps<DataSource, EdgeQuery, EdgeDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const { topic } = query;

  const onTopicChange = useCallback(
    (event: ChangeEvent<HTMLInputElement>) => {
      onChange({ ...query, topic: event.target.value });
    },
    [onChange, query]
  );

  const ExecuteQuery = useCallback(() => {
    const error = getTopicError(topic || '');
    if (error) {
      console.log(error);
      return;
    }
    onRunQuery();
  }, [onRunQuery, topic]);

  const topicError = getTopicError(topic || '');

  return (
    <div className="gf-form">
      <InlineField label="Topic" grow interactive tooltip={TooltipContent} error={topicError} invalid={!!topicError}>
        <Input
          required
          placeholder='e.g. "enterprise.site.area.line.machine.sensor"'
          value={topic}
          onBlur={ExecuteQuery}
          onChange={onTopicChange}
        />
      </InlineField>
    </div>
  );
}

function getTopicError(subject: string): string {
  if (!subject || subject === '') {
    return 'Topic is required';
  }

  const tokens = subject.split('.');
  const hasInvalidTokens = tokens.some((token) => token === '>' || token === '*');

  if (hasInvalidTokens) {
    return 'Wildcards are not allowed';
  }

  const isValidToken = (token: string): boolean => /^[^\s.]+$/.test(token);
  const success = tokens.every(isValidToken);

  if (!success) {
    return `Invalid topic: [${subject}]`;
  }

  return '';
}

const TooltipContent = () => (
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
