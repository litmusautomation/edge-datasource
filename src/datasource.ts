import { DataSourceInstanceSettings, CoreApp } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';

import { EdgeQuery, EdgeDataSourceOptions, DEFAULT_QUERY } from './types';

export class DataSource extends DataSourceWithBackend<EdgeQuery, EdgeDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<EdgeDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<EdgeQuery> {
    return DEFAULT_QUERY;
  }
}
