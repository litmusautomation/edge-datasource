import React, { ChangeEvent } from 'react';
import { InlineField, Input } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { EdgeDataSourceOptions, EdgeQuery } from '../types';

type Props = QueryEditorProps<DataSource, EdgeQuery, EdgeDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onConstantChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, topic: event.target.value });
    // executes the query
    onRunQuery();
  };

  const { topic } = query;

  return (
    <div className="gf-form">
      <InlineField
        label="Topic/Tag"
        grow
        interactive
        tooltip={() => {
          return (
            <div>
              <p>
                <b>Topic/Tag to subscribe to</b>
              </p>
              <p>
                See{' '}
                <b>
                  <a
                    target="_blank"
                    rel="noreferrer"
                    href="https://docs.litmus.io/litmusedge/product-features/devicehub/tags"
                  >
                    Edge Topic/Tag
                  </a>
                </b>{' '}
                for more information
              </p>
              <small>*Wildcards are not allowed</small>
            </div>
          );
        }}
      >
        <Input
          required
          placeholder='e.g. "enterprise.site.area.line.machine.sensor"'
          value={topic}
          onBlur={onRunQuery}
          onChange={onConstantChange}
        />
      </InlineField>
    </div>
  );
}
