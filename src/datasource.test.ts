import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSource } from './datasource';
import { EdgeDataSourceOptions, EdgeQuery } from './types';

const mockReplace = jest.fn((value: string | undefined) => value ?? '');
const mockGetDataStream = jest.fn();

const mockGetResource = jest.fn();

jest.mock('@grafana/runtime', () => ({
  DataSourceWithBackend: class {
    uid = 'test-uid';
    getResource = mockGetResource;
  },
  getGrafanaLiveSrv: () => ({ getDataStream: mockGetDataStream }),
  getTemplateSrv: () => ({ replace: mockReplace }),
}));

function makeDs(): DataSource {
  const settings = {
    uid: 'test-uid',
    jsonData: {},
  } as unknown as DataSourceInstanceSettings<EdgeDataSourceOptions>;
  return new DataSource(settings);
}

describe('DataSource.filterQuery', () => {
  let ds: DataSource;

  beforeEach(() => {
    ds = makeDs();
  });

  it('returns false for hidden query', () => {
    expect(ds.filterQuery({ hide: true, refId: 'A', topic: 'sensor' })).toBe(false);
  });

  it('returns false for empty topic', () => {
    expect(ds.filterQuery({ refId: 'A', topic: '' })).toBe(false);
  });

  it('returns false for whitespace-only topic', () => {
    expect(ds.filterQuery({ refId: 'A', topic: '   ' })).toBe(false);
  });

  it('returns false for undefined topic', () => {
    expect(ds.filterQuery({ refId: 'A' } as EdgeQuery)).toBe(false);
  });

  it('returns true for a valid topic', () => {
    expect(ds.filterQuery({ refId: 'A', topic: 'device.sensor' })).toBe(true);
  });
});

describe('DataSource.applyTemplateVariables', () => {
  let ds: DataSource;

  beforeEach(() => {
    ds = makeDs();
    mockReplace.mockClear();
  });

  it('calls getTemplateSrv().replace with topic and scopedVars', () => {
    const scopedVars = { myVar: { text: 'val', value: 'val' } };
    mockReplace.mockReturnValueOnce('device.sensor');
    const result = ds.applyTemplateVariables({ refId: 'A', topic: '$myVar' }, scopedVars);
    expect(mockReplace).toHaveBeenCalledWith('$myVar', scopedVars);
    expect(result.topic).toBe('device.sensor');
  });

  it('handles undefined topic without throwing', () => {
    mockReplace.mockReturnValueOnce('');
    const result = ds.applyTemplateVariables({ refId: 'A' } as EdgeQuery, {});
    expect(result.topic).toBe('');
  });
});

describe('DataSource.searchTopics', () => {
  let ds: DataSource;

  beforeEach(() => {
    ds = makeDs();
    mockGetResource.mockClear();
  });

  it('calls getResource with correct path and params', async () => {
    const expected = { topics: ['topic.a'], error: undefined };
    mockGetResource.mockResolvedValueOnce(expected);

    const result = await ds.searchTopics('temp');
    expect(mockGetResource).toHaveBeenCalledWith('topics', { query: 'temp' });
    expect(result).toEqual(expected);
  });

  it('passes empty query string', async () => {
    mockGetResource.mockResolvedValueOnce({ topics: [], error: 'api_token_not_configured' });

    const result = await ds.searchTopics('');
    expect(mockGetResource).toHaveBeenCalledWith('topics', { query: '' });
    expect(result.error).toBe('api_token_not_configured');
  });
});
