import {
  DataSourceInstanceSettings,
  CoreApp,
  DataQueryRequest,
  DataQueryResponse,
  LoadingState,
  LiveChannelScope,
  ScopedVars,
} from '@grafana/data';
import { DataSourceWithBackend, getGrafanaLiveSrv, getTemplateSrv } from '@grafana/runtime';

import { EdgeQuery, EdgeDataSourceOptions, TopicSearchResponse, DEFAULT_QUERY } from './types';
import { Observable, merge, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { defaults } from 'lodash';

export class DataSource extends DataSourceWithBackend<EdgeQuery, EdgeDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<EdgeDataSourceOptions>) {
    super(instanceSettings);
  }

  query(request: DataQueryRequest<EdgeQuery>): Observable<DataQueryResponse> {
    const observables = request.targets.map((target) => {
      const query = defaults(target, DEFAULT_QUERY);
      const interpolatedTopic = getTemplateSrv().replace(query.topic, request.scopedVars);

      return getGrafanaLiveSrv()
        .getDataStream({
          addr: {
            scope: LiveChannelScope.DataSource,
            stream: this.uid,
            path: interpolatedTopic || '',
            data: {
              ...query,
              topic: interpolatedTopic,
            },
          },
        })
        .pipe(
          catchError((err) =>
            of<DataQueryResponse>({
              data: [],
              state: LoadingState.Error,
              error: { message: err instanceof Error ? err.message : String(err) },
            })
          )
        );
    });

    return merge(...observables);
  }

  applyTemplateVariables(query: EdgeQuery, scopedVars: ScopedVars): EdgeQuery {
    return { ...query, topic: getTemplateSrv().replace(query.topic, scopedVars) };
  }

  getDefaultQuery(_: CoreApp): Partial<EdgeQuery> {
    return DEFAULT_QUERY;
  }

  async searchTopics(query: string): Promise<TopicSearchResponse> {
    return this.getResource<TopicSearchResponse>('topics', { query });
  }

  filterQuery(query: EdgeQuery): boolean {
    if (query.hide) {
      return false;
    }
    if (!query.topic || query.topic.trim() === '') {
      return false;
    }
    return true;
  }
}
