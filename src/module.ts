import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { EdgeQuery, EdgeDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, EdgeQuery, EdgeDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
