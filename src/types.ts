import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface EdgeQuery extends DataQuery {
  topic?: string;
}

export const DEFAULT_QUERY: Partial<EdgeQuery> = {
  topic: '',
};

/**
 * These are options configured for each DataSource instance
 */
export interface EdgeDataSourceOptions extends DataSourceJsonData {
  hostname: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface EdgeSecureJsonData {
  token?: string;
}
