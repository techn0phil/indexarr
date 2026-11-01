import { describe, expect, it, vi } from 'vitest';
import { act, renderHook, waitFor } from '@testing-library/react';
import { useInfiniteList } from './useInfiniteList';
import type { PaginatedResponse } from '../types';

type Item = { id: number };

const pageOf = (n: number): Item[] => [{ id: n * 10 + 1 }, { id: n * 10 + 2 }];

function deferred<T>() {
  let resolve!: (v: T) => void;
  const promise = new Promise<T>((r) => (resolve = r));
  return { promise, resolve };
}

const dupIds = (items: Item[]) => items.map((i) => i.id).filter((id, idx, arr) => arr.indexOf(id) !== idx);

describe('useInfiniteList concurrency', () => {
  it('does not load the same page twice when loadMore fires repeatedly', async () => {
    const fetchFn = vi.fn(async (page: number): Promise<PaginatedResponse<Item>> => ({
      success: true, data: pageOf(page), total: 6, page, pageSize: 2,
    }));
    const { result } = renderHook(() => useInfiniteList<Item>({ fetchFn, pageSize: 2 }));
    await waitFor(() => expect(result.current.isInitialLoading).toBe(false));
    act(() => {
      const lm = result.current.loadMore;
      lm();
      lm();
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(dupIds(result.current.items)).toEqual([]);
  });

  it('ignores responses from a previous filter set', async () => {
    const pending: Record<string, ReturnType<typeof deferred<PaginatedResponse<Item>>>> = {};
    const fetchFn = vi.fn((page: number, _ps: number, filters: Record<string, string>) => {
      const k = `${filters.sort ?? ''}:${page}`;
      pending[k] = deferred<PaginatedResponse<Item>>();
      return pending[k].promise;
    });
    const f1 = {};
    const f2 = { sort: '-year' };
    const { result, rerender } = renderHook(({ filters }) => useInfiniteList<Item>({ fetchFn, pageSize: 2, filters }), {
      initialProps: { filters: f1 as Record<string, string> },
    });
    await act(async () => pending[':1'].resolve({ success: true, data: pageOf(1), total: 6, page: 1, pageSize: 2 }));
    act(() => result.current.loadMore());
    rerender({ filters: f2 });
    await act(async () => pending['-year:1'].resolve({ success: true, data: pageOf(1), total: 6, page: 1, pageSize: 2 }));
    await act(async () => pending[':2'].resolve({ success: true, data: pageOf(1), total: 6, page: 2, pageSize: 2 }));
    expect(dupIds(result.current.items)).toEqual([]);
  });
});
