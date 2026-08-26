import { beforeEach, describe, expect, it, vi } from 'vitest';
import { act, renderHook, waitFor } from '@testing-library/react';
import { useInfiniteList } from './useInfiniteList';
import type { PaginatedResponse } from '../types';

type Item = { id: number; title: string };

describe('useInfiniteList', () => {
  const emptyFilters: Record<string, string> = {};

  const page1: Item[] = [
    { id: 1, title: 'Movie 1' },
    { id: 2, title: 'Movie 2' },
  ];

  const page2: Item[] = [
    { id: 3, title: 'Movie 3' },
    { id: 4, title: 'Movie 4' },
  ];

  const mockFetchFn = vi.fn<
    (page: number, pageSize: number, filters: Record<string, string>) => Promise<PaginatedResponse<Item>>
  >();

  beforeEach(() => {
    mockFetchFn.mockReset();
  });

  it('loads initial page on mount', async () => {
    mockFetchFn.mockResolvedValueOnce({
      success: true,
      data: page1,
      total: 4,
      page: 1,
      pageSize: 2,
    });

    const { result } = renderHook(() =>
      useInfiniteList<Item>({ fetchFn: mockFetchFn, pageSize: 2, filters: emptyFilters })
    );

    expect(result.current.isInitialLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isInitialLoading).toBe(false);
    });

    expect(mockFetchFn).toHaveBeenCalledWith(1, 2, {});
    expect(result.current.items).toEqual(page1);
    expect(result.current.total).toBe(4);
    expect(result.current.hasMore).toBe(true);
  });

  it('appends items when loadMore is called and hasMore is true', async () => {
    mockFetchFn
      .mockResolvedValueOnce({
        success: true,
        data: page1,
        total: 4,
        page: 1,
        pageSize: 2,
      })
      .mockResolvedValueOnce({
        success: true,
        data: page2,
        total: 4,
        page: 2,
        pageSize: 2,
      });

    const { result } = renderHook(() =>
      useInfiniteList<Item>({ fetchFn: mockFetchFn, pageSize: 2, filters: emptyFilters })
    );

    await waitFor(() => {
      expect(result.current.items).toHaveLength(2);
    });

    act(() => {
      result.current.loadMore();
    });

    await waitFor(() => {
      expect(result.current.items).toHaveLength(4);
    });

    expect(mockFetchFn).toHaveBeenNthCalledWith(1, 1, 2, {});
    expect(mockFetchFn).toHaveBeenNthCalledWith(2, 2, 2, {});
    expect(result.current.hasMore).toBe(false);
  });

  it('sets error when response.success is false', async () => {
    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    mockFetchFn.mockResolvedValueOnce({
      success: false,
      error: 'Network error',
      data: [],
      total: 0,
      page: 1,
      pageSize: 50,
    });

    const { result } = renderHook(() =>
      useInfiniteList<Item>({ fetchFn: mockFetchFn, filters: emptyFilters })
    );

    await waitFor(() => {
      expect(result.current.error).toBe('Network error');
    });

    consoleErrorSpy.mockRestore();
  });

  it('resets list state and reloads first page on reset()', async () => {
    mockFetchFn.mockResolvedValue({
      success: true,
      data: page1,
      total: 2,
      page: 1,
      pageSize: 2,
    });

    const { result } = renderHook(() =>
      useInfiniteList<Item>({ fetchFn: mockFetchFn, pageSize: 2, filters: emptyFilters })
    );

    await waitFor(() => {
      expect(result.current.items).toHaveLength(2);
    });

    act(() => {
      result.current.reset();
    });

    await waitFor(() => {
      expect(result.current.items).toHaveLength(2);
    });

    expect(result.current.page).toBe(1);
    expect(mockFetchFn).toHaveBeenCalledTimes(2);
  });

  it('re-fetches when filters change', async () => {
    mockFetchFn.mockResolvedValue({
      success: true,
      data: page1,
      total: 2,
      page: 1,
      pageSize: 2,
    });

    const { rerender } = renderHook(
      ({ filters }) =>
        useInfiniteList<Item>({
          fetchFn: mockFetchFn,
          pageSize: 2,
          filters,
        }),
      { initialProps: { filters: { status: 'available' } } }
    );

    await waitFor(() => {
      expect(mockFetchFn).toHaveBeenCalledTimes(1);
    });

    rerender({ filters: { status: 'missing' } });

    await waitFor(() => {
      expect(mockFetchFn).toHaveBeenCalledTimes(2);
    });
    expect(mockFetchFn).toHaveBeenLastCalledWith(1, 2, { status: 'missing' });
  });
});
