import {
  DataSourceInstanceSettings,
  CoreApp,
  DataQueryRequest,
  DataQueryResponse,
  LiveChannelScope,
} from '@grafana/data';
import { DataSourceWithBackend, getGrafanaLiveSrv } from '@grafana/runtime';

import { EdgeQuery, EdgeDataSourceOptions, DEFAULT_QUERY } from './types';
import { Observable, merge } from 'rxjs';
import { defaults } from 'lodash';

export class DataSource extends DataSourceWithBackend<EdgeQuery, EdgeDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<EdgeDataSourceOptions>) {
    super(instanceSettings);
  }

  query(request: DataQueryRequest<EdgeQuery>): Observable<DataQueryResponse> {
    if (request.targets[0].topic === undefined || request.targets[0].topic === '') {
      throw new Error('Topic is required');
    }

    const observables = request.targets.map((target) => {
      const query = defaults(target, DEFAULT_QUERY);

      return getGrafanaLiveSrv().getDataStream({
        addr: {
          scope: LiveChannelScope.DataSource,
          stream: this.uid,
          path: query?.topic || '',
          data: {
            ...query,
          },
        },
      });
    });

    return merge(...observables);
  }

  getDefaultQuery(_: CoreApp): Partial<EdgeQuery> {
    return DEFAULT_QUERY;
  }

  filterQuery(query: EdgeQuery): boolean {
    if (query.hide) {
      return false;
    }
    return true;
  }
}
