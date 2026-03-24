import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { EdgeQuery, EdgeDataSourceOptions, EdgeSecureJsonData } from './types';

export const plugin = new DataSourcePlugin<DataSource, EdgeQuery, EdgeDataSourceOptions, EdgeSecureJsonData>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
