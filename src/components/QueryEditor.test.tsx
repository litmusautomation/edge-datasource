import React from 'react';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryEditor } from './QueryEditor';
import { EdgeQuery } from '../types';

type MockAsyncOption = { value: string; label: string };
type MockAsyncSelectProps = {
  value?: MockAsyncOption | null;
  onChange: (value: MockAsyncOption) => void;
  onInputChange?: (value: string, meta: { action: string }) => void;
  onBlur?: () => void;
  loadOptions?: (query: string) => Promise<MockAsyncOption[]>;
  noOptionsMessage?: string;
  allowCustomValue?: boolean;
  isClearable?: boolean;
  menuPlacement?: 'top' | 'bottom' | 'auto';
  menuShouldPortal?: boolean;
};

type MockInputProps = {
  value: string;
  onChange: (e: { currentTarget: { value: string } }) => void;
  onBlur?: () => void;
};

let lastAsyncSelectProps: MockAsyncSelectProps | null = null;

jest.mock('@grafana/ui', () => ({
  InlineFieldRow: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  InlineField: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  AsyncSelect: (props: MockAsyncSelectProps) => {
    lastAsyncSelectProps = props;
    return (
      <input
        data-testid="topic-async-select"
        value={props.value?.value ?? ''}
        onChange={(e) => props.onInputChange?.(e.currentTarget.value, { action: 'input-change' })}
        onBlur={() => props.onBlur?.()}
      />
    );
  },
  Input: (props: MockInputProps) => (
    <input
      data-testid="topic-input"
      value={props.value}
      onChange={(e) => props.onChange({ currentTarget: { value: e.currentTarget.value } })}
      onBlur={() => props.onBlur?.()}
    />
  ),
  Text: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Icon: () => <span />,
}));

function renderEditor(searchTopics: jest.Mock, topic = '') {
  const onRunQuery = jest.fn();
  const onChangeSpy = jest.fn();

  function Harness() {
    const [query, setQuery] = React.useState<EdgeQuery>({ refId: 'A', topic });
    return (
      <QueryEditor
        query={query}
        onChange={(next) => {
          onChangeSpy(next);
          setQuery(next);
        }}
        onRunQuery={onRunQuery}
        datasource={{ searchTopics } as never}
      />
    );
  }

  render(<Harness />);
  return { onRunQuery, onChangeSpy };
}

describe('QueryEditor', () => {
  beforeEach(() => {
    lastAsyncSelectProps = null;
  });

  it('renders AsyncSelect when probe is ready', async () => {
    const searchTopics = jest.fn().mockResolvedValue({ topics: [] });
    renderEditor(searchTopics);

    await waitFor(() => expect(screen.getByTestId('topic-async-select')).toBeInTheDocument());
  });

  it('renders Input fallback when token is not configured', async () => {
    const searchTopics = jest.fn().mockResolvedValue({ topics: [], error: 'api_token_not_configured' });
    renderEditor(searchTopics);

    await waitFor(() => expect(screen.getByTestId('topic-input')).toBeInTheDocument());
  });

  it('does not allow custom values and is not clearable', async () => {
    const searchTopics = jest.fn().mockResolvedValue({ topics: [] });
    renderEditor(searchTopics);

    await waitFor(() => expect(lastAsyncSelectProps).not.toBeNull());
    expect(lastAsyncSelectProps?.allowCustomValue).toBe(false);
    expect(lastAsyncSelectProps?.isClearable).toBe(false);
    expect(lastAsyncSelectProps?.menuPlacement).toBe('top');
  });

  it('updates query while typing edited topic text', async () => {
    const searchTopics = jest.fn().mockResolvedValue({ topics: [] });
    const { onChangeSpy } = renderEditor(searchTopics, 'devicehub.alias.P1_L1_Machine3_1_S7.Temperature');

    const input = await screen.findByTestId('topic-async-select');
    await act(async () => {
      fireEvent.change(input, { target: { value: 'devicehub.alias.P1_L1_Machine3_1_S7.Temperat' } });
    });

    expect(onChangeSpy).toHaveBeenCalledWith({
      refId: 'A',
      topic: 'devicehub.alias.P1_L1_Machine3_1_S7.Temperat',
    });
  });

  it('loadOptions returns local fallback match when backend returns empty', async () => {
    const full = 'devicehub.alias.P1_L1_Machine3_1_S7.Temperature';
    const partial = 'devicehub.alias.P1_L1_Machine3_1_S7.Temperat';

    const searchTopics = jest.fn((query: string) => {
      if (query === '') {
        return Promise.resolve({ topics: [] });
      }
      return Promise.resolve({ topics: [] });
    });

    renderEditor(searchTopics, full);

    await waitFor(() => expect(lastAsyncSelectProps?.loadOptions).toBeDefined());
    const result = await lastAsyncSelectProps!.loadOptions!(partial);

    expect(result).toEqual([{ value: full, label: full }]);
  });

  it('loadOptions does not query backend for less than 2 chars', async () => {
    const searchTopics = jest.fn().mockResolvedValue({ topics: [] });
    renderEditor(searchTopics);

    await waitFor(() => expect(lastAsyncSelectProps?.loadOptions).toBeDefined());
    const result = await lastAsyncSelectProps!.loadOptions!('d');

    expect(result).toEqual([]);
    expect(searchTopics).toHaveBeenCalledTimes(1); // probe only
  });

  it('runs query when selecting a valid topic', async () => {
    const searchTopics = jest.fn().mockResolvedValue({ topics: [] });
    const { onRunQuery } = renderEditor(searchTopics);

    await waitFor(() => expect(lastAsyncSelectProps).not.toBeNull());
    await act(async () => {
      lastAsyncSelectProps?.onChange({ value: 'device.valid.topic', label: 'device.valid.topic' });
    });

    expect(onRunQuery).toHaveBeenCalledTimes(1);
  });

  it('does not run query when selecting invalid topic', async () => {
    const searchTopics = jest.fn().mockResolvedValue({ topics: [] });
    const { onRunQuery } = renderEditor(searchTopics);

    await waitFor(() => expect(lastAsyncSelectProps).not.toBeNull());
    await act(async () => {
      lastAsyncSelectProps?.onChange({ value: 'device.*', label: 'device.*' });
    });

    expect(onRunQuery).not.toHaveBeenCalled();
  });
});
